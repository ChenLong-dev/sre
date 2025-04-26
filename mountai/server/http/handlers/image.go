package handlers

import (
	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/service"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
	"rulai/utils/response"
	"strconv"
	"time"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func CreateImageJob(c *gin.Context) {
	projectID := c.Param("project_id")
	createReq := new(req.CreateImageJobReq)
	err := c.ShouldBindJSON(createReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	if createReq.BuildArg != "" && createReq.BuildArgsTemplateID != "" {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "镜像构建参数和镜像构建模版不能同时存在"))
		return
	}

	project, err := service.SVC.GetProjectDetail(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	// 判断创建镜像是否已经存在
	image, err := service.SVC.GetJenkinsBuildImage(c, project.Name, createReq.BranchName, createReq.CommitID)
	if err != nil && !errcode.EqualError(errcode.NoRowsFoundError, err) {
		response.JSON(c, nil, err)
		return
	}

	if image != nil {
		response.JSON(c, nil, _errcode.DuplicatedImageTagError)
		return
	}

	jobs, err := service.SVC.GetImageJobs(c, &req.GetImageJobsReq{
		BaseListRequest: models.BaseListRequest{
			Limit: 1,
			Page:  1,
		},
		BranchName:  createReq.BranchName,
		Status:      null.StringFrom(string(resp.JenkinsJobResultRunning)),
		ProjectName: project.Name,
		ProjectID:   projectID,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	if len(jobs) > 0 {
		response.JSON(c, nil, errors.Wrapf(_errcode.OtherRunningImageJobExistsError, "build id:%s", jobs[0].BuildID))
		return
	}

	userID, userName, userEmail := "", "", ""
	if createReq.UserID == "" {
		operator, ok := c.Get(utils.ContextUserClaimKey)
		if !ok {
			response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, "operator is not exist"))
			return
		}
		userClaim := operator.(entity.UserClaims)
		userName = userClaim.Name
		userEmail = userClaim.Email
		userID = userClaim.Id
	} else {
		// Generate default user
		userInfo, e := service.SVC.GetDefaultUserProfile(c, createReq.UserID)
		if e != nil {
			response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", e))
			return
		}
		userID = userInfo.ID
		userName = userInfo.Name
		userEmail = userInfo.Email
	}
	userIntID, err := strconv.Atoi(userID)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}
	syncToken, err := utils.GenerateJWTToken(c, &resp.GitUserProfileResp{
		ID:    userIntID,
		Name:  userName,
		Email: userEmail,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	createReq.SyncToken = syncToken
	createReq.UserID = userID

	// 创建打包任务
	err = service.SVC.CreateImageJob(c, createReq, project)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func GetImageJobDetail(c *gin.Context) {
	buildID := c.Param("build_id")
	projectID := c.Param("project_id")

	project, err := service.SVC.GetProjectDetail(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	image, err := service.SVC.GetImageJobDetail(c, &req.GetImageJobDetailReq{
		BuildID:     buildID,
		ProjectName: project.Name,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, image, nil)
}

func GetImageJobs(c *gin.Context) {
	getReq := new(req.GetImageJobsReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	projectID := c.Param("project_id")
	project, err := service.SVC.GetProjectDetail(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	getReq.ProjectName = project.Name
	getReq.ProjectID = projectID

	images, err := service.SVC.GetImageJobs(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	count, err := service.SVC.GetImageJobsCount(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{
		List:  images,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: count,
	}, nil)
}

func GetImageTags(c *gin.Context) {
	getReq := new(req.GetImageTagsReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	projectID := c.Param("project_id")
	project, err := service.SVC.GetProjectDetail(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}
	getReq.ProjectName = project.Name

	aliRepoTagsResp, err := service.SVC.AliGetRepoTags(c, &req.AliGetRepoTagsReq{
		ProjectName: getReq.ProjectName,
		Page:        getReq.Page,
		Size:        getReq.Limit,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	tags := aliRepoTagsResp.Data.Tags
	total := aliRepoTagsResp.Data.Total

	templateIDs := make([]string, 0)

	// 分离tag
	res := make([]*resp.ImageTagResp, 0)
	for _, tag := range tags {
		image, intErr := service.SVC.GetLastJenkinsBuildImage(c, &req.GetImageBuildReq{
			ProjectName: getReq.ProjectName,
			ImageTag:    tag.Tag,
		})

		if intErr != nil && !errcode.EqualError(errcode.NoRowsFoundError, intErr) {
			response.JSON(c, nil, intErr)
			return
		}

		ret := &resp.ImageTagResp{
			Version:    service.SVC.GetAliImageRepoURL(getReq.ProjectName, tag.Tag),
			CreateTime: time.UnixMilli(int64(tag.ImageCreate)).Local().Format(utils.DefaultTimeFormatLayout),
			UpdateTime: time.UnixMilli(int64(tag.ImageUpdate)).Local().Format(utils.DefaultTimeFormatLayout),
		}

		if image != nil {
			ret.Description = image.Description
			ret.BranchName = image.BranchName
			ret.CommitID = image.CommitID
			ret.BuildArgWithMask = image.BuildArgWithMask
			ret.BuildArgsTemplateID = image.BuildArgsTemplateID
			ret.ImageTag = image.ImageTag
			if image.BuildArgsTemplateID != "" {
				templateIDs = append(templateIDs, image.BuildArgsTemplateID)
			}
		} else {
			version, branch, intErr := service.SVC.HardDecodeImageTag(tag.Tag)
			// 兼容历史镜像，忽略不符合格式的镜像
			if errcode.EqualError(_errcode.InvalidImageVersionError, intErr) {
				log.Errorc(c, "project %s contains invalid image tag: %s", projectID, tag.Tag)
				continue
			}
			if intErr != nil {
				response.JSON(c, nil, intErr)
				return
			}

			ret.BranchName = branch
			ret.CommitID = version
		}

		res = append(res, ret)
	}

	templates, err := service.SVC.GetImageArgsTemplatesByIDs(c, templateIDs)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	gitTags, err := service.SVC.GetGitlabSingleProjectTags(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	for _, item := range res {
		if item.BuildArgsTemplateID != "" {
			item.Template = templates[item.BuildArgsTemplateID]
		}

		for _, v := range gitTags {
			if v.Commit.ShortId != item.CommitID {
				continue
			}

			tmp := new(resp.ImageTagResp)
			err := deepcopy.Copy(item).To(tmp)
			if err != nil {
				response.JSON(c, nil, err)
				return
			}

			tmp.BranchName = "tags"
			tmp.CommitID = v.Name
			res = append(res, tmp)
		}
	}

	response.JSON(c, models.BaseListResponse{
		List:  res,
		Limit: getReq.Limit,
		Page:  getReq.Page,
		Count: total,
	}, nil)
}

func DeleteImageJob(c *gin.Context) {
	buildID := c.Param("build_id")
	projectID := c.Param("project_id")

	project, err := service.SVC.GetProjectDetail(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	err = service.SVC.DeleteImageJob(c, &req.DeleteImageJobReq{
		BuildID:     buildID,
		ProjectName: project.Name,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func CacheImageJob(c *gin.Context) {
	buildID := c.Param("build_id")
	projectID := c.Param("project_id")

	project, err := service.SVC.GetProjectDetail(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	image, err := service.SVC.GetImageJobDetail(c, &req.GetImageJobDetailReq{
		BuildID:     buildID,
		ProjectName: project.Name,
	})
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	// TODO: remove after 前端镜像构建参数移除
	// 兼容不使用镜像参数模版的镜像构建
	if image.BuildArgWithMask != "" && image.BuildArgsTemplateID == "" {
		err = service.SVC.UpdateProject(c, projectID, &req.UpdateProjectReq{
			ImageArgs: map[string]string{
				image.BranchName: image.BuildArgWithMask,
			},
		})
		if err != nil {
			response.JSON(c, nil, err)
			return
		}
	}

	err = service.SVC.CreateJenkinsBuildImage(c, image, project)
	if err != nil {
		log.Errorc(c, "cache image %s fail: %s", image.BuildID, err)
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, nil, nil)
}

func GetLastImageArgs(c *gin.Context) {
	projectID := c.Param("project_id")

	args, err := service.SVC.GetProjectLastJenkinsBuildArgs(c, projectID)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, args, nil)
}
