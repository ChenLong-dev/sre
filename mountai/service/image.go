package service

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	"gitlab.shanhai.int/sre/library/log"

	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	infraJenkins "gitlab.shanhai.int/sre/gojenkins"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var imageUserProfileLocalCache = make(map[string]*resp.UserProfileResp)

// 获取镜像版本
func (s *Service) getImageVersion(projectName, tagName string) string {
	return fmt.Sprintf("%s:%s", projectName, tagName)
}

// 获取镜像标签
func (s *Service) getImageTag(branchName, commitID, argsHash string) string {
	branchNameRegexp := regexp.MustCompile(`[^\w\-]`)
	branchName = branchNameRegexp.ReplaceAllString(branchName, "_")
	if argsHash != "" {
		return fmt.Sprintf("%s-%s-%s", commitID, branchName, argsHash[:8])
	}

	return fmt.Sprintf("%s-%s", commitID, branchName)
}

// 获取短commit id
func (s *Service) getShortCommitID(id string) (string, error) {
	if len(id) < 8 {
		return "", errors.Wrapf(_errcode.InvalidImageCommitIDError, "commit id:%s", id)
	}

	return id[0:8], nil
}

// 获取最后提交记录
func (s *Service) getLastComment(ctx context.Context, raw *infraJenkins.BuildResponse, projectID, id string) (string, error) {
	if len(raw.ChangeSets) > 0 {
		for _, v := range raw.ChangeSets {
			if v.Kind == "git" {
				return v.Items[len(v.Items)-1].Msg, nil
			}
		}
	}

	// 重复打包从 gitlab 获取
	res, err := s.GetGitlabProjectCommitMsg(ctx, projectID, id)
	if err != nil {
		return "", err
	}

	return res.Title, nil
}

// 获取镜像仓库地址
func (s *Service) GetImageRepoURL(version string) string {
	return fmt.Sprintf("crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/infra/%s", version)
}

// 初始化jenkins模版
func (s *Service) initJenkinsConfigTemplate(ctx context.Context, project *resp.ProjectDetailResp,
	createReq *req.CreateImageJobReq) (*entity.JenkinsConfigTemplate, error) {
	gitProject, err := s.GetGitlabSingleProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	timeout := createReq.Timeout
	if timeout == 0 {
		timeout = config.Conf.Other.BuildJobTimeout
	}

	return &entity.JenkinsConfigTemplate{
		ProjectSSHUrl: gitProject.HTTPURL,
		Timeout:       timeout,
	}, nil
}

// 获取镜像数量
func (s *Service) GetImageJobsCount(ctx context.Context, getReq *req.GetImageJobsReq) (int, error) {
	job, err := s.jenkinsClient.GetJob(s.getImageJobName(getReq.ProjectName))
	if err != nil {
		// 不存在，返回空
		if err.Error() == strconv.Itoa(http.StatusNotFound) {
			return 0, nil
		}
		return 0, errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	buildIDs, err := job.GetAllBuildIds()
	if err != nil {
		return 0, errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	return len(buildIDs), nil
}

// 获取镜像详情
func (s *Service) GetImageJobDetail(ctx context.Context, getReq *req.GetImageJobDetailReq) (*resp.ImageDetailResp, error) {
	job, err := s.jenkinsClient.GetJob(s.getImageJobName(getReq.ProjectName))
	if err != nil {
		return nil, errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	buildID, err := strconv.Atoi(getReq.BuildID)
	if err != nil {
		return nil, errors.Wrap(errcode.InvalidParams, err.Error())
	}

	jobBuild, err := job.GetBuild(int64(buildID))
	if err != nil {
		return nil, errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	res, paramMap, err := s.formatImageDetailFromJob(ctx, job, jobBuild)
	if err != nil {
		return nil, err
	}

	userProfile, err := s.getImageBuildUserFromCache(ctx, paramMap["UserID"])
	if err != nil {
		return nil, err
	}

	res.UserProfile = userProfile

	return res, nil
}

// 获取镜像列表
func (s *Service) GetImageJobs(ctx context.Context, getReq *req.GetImageJobsReq) ([]*resp.ImageListResp, error) {
	res := make([]*resp.ImageListResp, 0)

	job, err := s.jenkinsClient.GetJob(s.getImageJobName(getReq.ProjectName))
	if err != nil {
		// 不存在，返回空
		if err.Error() == strconv.Itoa(http.StatusNotFound) {
			return res, nil
		}
		return nil, errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	buildIDs, err := job.GetAllBuildIds()
	if err != nil {
		return nil, errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	// 由于jenkins没有分页以及相关条件过滤，只能手动分页
	eligibleCount := 0
	skipCount := (getReq.Page - 1) * getReq.Limit

	for i := 0; i < len(buildIDs); i++ {
		id := buildIDs[i].Number

		jobBuild, err := job.GetBuild(id)
		if err != nil {
			return nil, errors.Wrap(_errcode.JenkinsInternalError, err.Error())
		}

		paramMap := make(map[string]string)

		for _, param := range jobBuild.GetParameters() {
			paramMap[param.Name] = param.Value
		}

		shortCommitID, err := s.getShortCommitID(paramMap["CommitID"])
		if err != nil {
			return nil, errors.Wrap(errcode.InvalidParams, err.Error())
		}

		lastComment, e := s.getLastComment(ctx, jobBuild.Raw, getReq.ProjectID, paramMap["CommitID"])
		if e != nil {
			log.Warnc(ctx, "get commit comment err: %s", e.Error())
		}

		// 过滤
		if getReq.BranchName != "" && paramMap["BranchName"] != getReq.BranchName {
			continue
		}

		if !getReq.Status.IsZero() && jobBuild.GetResult() != getReq.Status.ValueOrZero() {
			continue
		}

		// 确认条件符合的数量和skip数量
		eligibleCount++
		if eligibleCount <= skipCount {
			continue
		} else if eligibleCount > skipCount+getReq.Limit {
			break
		}

		userProfile, err := s.getImageBuildUserFromCache(ctx, paramMap["UserID"])
		if err != nil {
			return nil, err
		}

		cur := &resp.ImageListResp{
			BuildID:             strconv.Itoa(int(id)),
			JobURL:              jobBuild.GetUrl(),
			Status:              resp.JenkinsJobResult(jobBuild.GetResult()),
			BranchName:          paramMap["BranchName"],
			BuildArgsTemplateID: paramMap["BuildArgsTemplateID"],
			BuildArgWithMask:    paramMap["BuildArgWithMask"],
			ImageTag:            paramMap["ImageTag"],
			CommitID:            shortCommitID,
			LastComment:         lastComment,
			CreateTime:          jobBuild.GetTimestamp().Format(utils.ImageTimeFormatLayout),
			ImageRepoURL:        paramMap["ImageRepoUrl"],
			Description:         paramMap["Description"],
			UserProfile:         userProfile,
			Duration:            s.formatImageBuildDuration(jobBuild),
		}
		res = append(res, cur)
	}

	return res, nil
}

func (s *Service) getImageJobName(projectName string) string {
	return fmt.Sprintf("rulai-image-%s", projectName)
}

// 创建镜像
func (s *Service) CreateImageJob(ctx context.Context, createReq *req.CreateImageJobReq, project *resp.ProjectDetailResp) error {
	shortCommitID, err := s.getShortCommitID(createReq.CommitID)
	if err != nil {
		return err
	}

	tpl, err := s.initJenkinsConfigTemplate(ctx, project, createReq)
	if err != nil {
		return err
	}

	jenkinsConfig, err := s.RenderTemplate(ctx, "./template/jenkins/ImageConfig.xml", tpl)
	if err != nil {
		return err
	}

	jobName := s.getImageJobName(project.Name)

	job, err := s.jenkinsClient.GetJob(jobName)
	if err != nil {
		if err.Error() != strconv.Itoa(http.StatusNotFound) {
			return errors.Wrap(_errcode.JenkinsInternalError, err.Error())
		}

		// 不存在则创建
		_, err = s.jenkinsClient.CreateJob(jenkinsConfig, jobName)
		if err != nil {
			return errors.Wrap(_errcode.JenkinsInternalError, err.Error())
		}
	} else {
		// 更新配置
		err = job.UpdateConfig(jenkinsConfig)
		if err != nil {
			return errors.Wrap(_errcode.JenkinsInternalError, err.Error())
		}
	}

	// 优先使用模版
	if createReq.BuildArgsTemplateID != "" {
		// 渲染镜像参数模版
		imageArg, imageArgWithMask, intErr := s.RenderImageArgsTemplate(ctx, createReq.BuildArgsTemplateID, project.ID)
		if intErr != nil {
			return errors.Wrap(_errcode.JenkinsInternalError, intErr.Error())
		}

		createReq.BuildArg = imageArg
		createReq.BuildArgWithMask = imageArgWithMask
	} else {
		// 为兼容参数构建
		createReq.BuildArgWithMask = createReq.BuildArg
	}

	hash := ""
	buildArg := ""
	buildArgWithMask := ""
	if createReq.BuildArg != "" {
		buildArg = strings.ReplaceAll(createReq.BuildArg, "--build-arg ", "--opt build-arg:")
		buildArgWithMask = strings.ReplaceAll(createReq.BuildArgWithMask, "--build-arg ", "--opt build-arg:")
		hash = strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(createReq.BuildArg))))
	}

	tag := s.getImageTag(createReq.BranchName, shortCommitID, hash)
	version := s.getImageVersion(project.Name, tag)
	repoURL := s.GetHuaWeiImageRepoURL(version)

	_, err = s.jenkinsClient.BuildJob(jobName, map[string]string{
		"ProjectName":         project.Name,
		"BranchName":          createReq.BranchName,
		"ImageRepoUrl":        repoURL,
		"BuildArg":            buildArg,
		"BuildArgsTemplateID": createReq.BuildArgsTemplateID,
		"BuildArgWithMask":    buildArgWithMask,
		"CommitID":            createReq.CommitID,
		"ImageTag":            tag,
		// 为了同步镜像构建参数
		"SyncHost":     fmt.Sprintf("%s/api/v1/projects/%s", config.Conf.Other.AMSHost, project.ID),
		"SyncJWTToken": createReq.SyncToken,
		"Description":  createReq.Description,
		// 构建完成后缓存镜像信息
		"ImageCacheHost": fmt.Sprintf("%s/api/v1/projects/%s/images/jobs", config.Conf.Other.AMSHost, project.ID),
		"UserID":         createReq.UserID,
	})
	if err != nil {
		return errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	return nil
}

// 停止镜像任务
func (s *Service) DeleteImageJob(ctx context.Context, deleteReq *req.DeleteImageJobReq) error {
	job, err := s.jenkinsClient.GetJob(s.getImageJobName(deleteReq.ProjectName))
	if err != nil {
		return errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	buildID, err := strconv.Atoi(deleteReq.BuildID)
	if err != nil {
		return errors.Wrap(errcode.InvalidParams, err.Error())
	}

	jobBuild, err := job.GetBuild(int64(buildID))
	if err != nil {
		return errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	_, err = jobBuild.Stop()
	if err != nil {
		return errors.Wrap(_errcode.JenkinsInternalError, err.Error())
	}

	return nil
}

func (s *Service) formatImageDetailFromJob(_ context.Context, job *infraJenkins.Job, jobBuild *infraJenkins.Build) (
	res *resp.ImageDetailResp, paramMap map[string]string, err error) {
	paramMap = make(map[string]string)

	for _, param := range jobBuild.GetParameters() {
		paramMap[param.Name] = param.Value
	}

	shortCommitID, err := s.getShortCommitID(paramMap["CommitID"])
	if err != nil {
		return nil, nil, errors.Wrap(errcode.InvalidParams, err.Error())
	}

	res = &resp.ImageDetailResp{
		JobName:             job.GetName(),
		BuildID:             jobBuild.Info().ID,
		JobURL:              jobBuild.GetUrl(),
		Status:              resp.JenkinsJobResult(jobBuild.GetResult()),
		ImageRepoURL:        paramMap["ImageRepoUrl"],
		BuildArgsTemplateID: paramMap["BuildArgsTemplateID"],
		BuildArgWithMask:    paramMap["BuildArgWithMask"],
		BuildArg:            paramMap["BuildArg"],
		BranchName:          paramMap["BranchName"],
		ImageTag:            paramMap["ImageTag"],
		Description:         paramMap["Description"],
		CommitID:            shortCommitID,
		ConsoleOutput:       jobBuild.GetConsoleOutput(),
		CreateTime:          jobBuild.GetTimestamp().Format(utils.ImageTimeFormatLayout),
		Timestamp:           jobBuild.GetTimestamp(),
		Duration:            s.formatImageBuildDuration(jobBuild),
	}
	return res, paramMap, nil
}

// GetBranchNameFromImageVersion 从镜像地址中获取分支名，注：有些分支名含有 `-`
// crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/infra/app-common:d9ce00ac-master
func (s *Service) ExtraInfoFromImageVersion(imageVersion string) (projectName, version, branch string, err error) {
	partArr := strings.Split(imageVersion, "/")
	if len(partArr) != 3 {
		err = _errcode.InvalidImageVersionError
		return
	}

	partArr = strings.Split(partArr[2], ":")
	if len(partArr) != 2 {
		err = _errcode.InvalidImageVersionError
		return
	}
	projectName = partArr[0]

	build, err := s.GetLastJenkinsBuildImage(context.Background(), &req.GetImageBuildReq{
		ProjectName: projectName,
		ImageTag:    partArr[1],
	})
	if err != nil && !errcode.EqualError(errcode.NoRowsFoundError, err) {
		return
	}

	if build != nil {
		version = s.getImageVersion(projectName, build.ImageTag)
		branch = build.BranchName

		return
	}

	version, branch, err = s.HardDecodeImageTag(partArr[1])
	if err != nil {
		return
	}

	return projectName, version, branch, nil
}

// 硬解码 为兼容早期数据
// tag：commitID-branchName
func (s *Service) HardDecodeImageTag(tag string) (version, branch string, err error) {
	strArr := strings.Split(tag, "-")
	if len(strArr) <= 1 {
		err = _errcode.InvalidImageVersionError
		return
	}

	version = strArr[0]
	branch = strings.Join(strArr[1:], "-")
	return
}

func (s *Service) GetLastJenkinsBuildImage(ctx context.Context, getReq *req.GetImageBuildReq) (*entity.JenkinsBuildImage, error) {
	return s.dao.GetLastJenkinsBuildImage(ctx, s.getImageBuildFilter(ctx, getReq))
}

func (s *Service) getImageBuildFilter(_ context.Context, getReq *req.GetImageBuildReq) bson.M {
	filter := bson.M{}

	if getReq.BranchName != "" {
		filter["branch_name"] = getReq.BranchName
	}

	if getReq.CommitID != "" {
		filter["commit_id"] = getReq.CommitID
	}

	if getReq.ProjectName != "" {
		filter["project_name"] = getReq.ProjectName
	}

	if getReq.ImageTag != "" {
		filter["image_tag"] = getReq.ImageTag
	}

	return filter
}

func (s *Service) CreateJenkinsBuildImage(ctx context.Context, image *resp.ImageDetailResp, project *resp.ProjectDetailResp) error {
	build := &entity.JenkinsBuildImage{
		ID:                  primitive.NewObjectID(),
		ProjectID:           project.ID,
		ProjectName:         project.Name,
		ImageTag:            image.ImageTag,
		BuildID:             image.BuildID,
		JobName:             image.JobName,
		JobURL:              image.JobURL,
		ImageRepoURL:        image.ImageRepoURL,
		BuildArgWithMask:    image.BuildArgWithMask,
		BuildArgsTemplateID: image.BuildArgsTemplateID,
		BranchName:          image.BranchName,
		CommitID:            image.CommitID,
		CreateTime:          &image.Timestamp,
		Description:         image.Description,
		UserID:              image.UserProfile.ID,
	}
	return s.dao.CreateJenkinsBuildImage(ctx, build)
}

func (s *Service) getImageBuildUserFromCache(ctx context.Context, userID string) (*resp.UserProfileResp, error) {
	if userID == "" {
		return nil, nil
	}

	user, ok := imageUserProfileLocalCache[userID]
	if ok {
		return user, nil
	}

	return s.GetUserInfo(ctx, userID)
}

func (s *Service) GetProjectLastJenkinsBuildArgs(ctx context.Context, projectID string) ([]*resp.ImageLastArgsResp, error) {
	images, err := s.dao.GetProjectLastJenkinsBuildImages(ctx, projectID)
	if err != nil {
		return nil, err
	}

	res := make([]*resp.ImageLastArgsResp, 0)
	err = deepcopy.Copy(&images).To(&res)

	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// Format image build time, eg:"04m21s"
func (s *Service) formatImageBuildDuration(build *infraJenkins.Build) string {
	duration := time.Duration(build.GetDuration()) * time.Millisecond
	// Build job is running
	if duration == 0 {
		duration = time.Since(build.GetTimestamp())
	}
	durationTime := time.Unix(int64(duration.Seconds()), 0)
	if durationTime.Minute() == 0 {
		return fmt.Sprintf("%02ds", durationTime.Second())
	}
	return fmt.Sprintf("%02dm%02ds", durationTime.Minute(), durationTime.Second())
}
