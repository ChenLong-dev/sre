package service

import (
	"rulai/config"
	"rulai/dao"
	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"

	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 名称正则表达式 k8s不能以-开头结尾
var projectNameRegexp = regexp.MustCompile(`^([0-9a-z]+[0-9a-z\-]*)?[0-9a-z]$`)

// CreateProject 创建项目
func (s *Service) CreateProject(ctx context.Context, createReq *req.CreateProjectReq) (*resp.ProjectDetailResp, error) {
	// 获取仓库信息
	_, err := s.GetGitlabSingleProject(ctx, createReq.GitID)
	if err != nil {
		return nil, err
	}

	// 获取团队信息
	teamDetail, err := s.GetTeamDetail(ctx, createReq.TeamID)
	if err != nil {
		return nil, errors.Wrap(errcode.InvalidParams, err.Error())
	}

	// 获取、校验项目 owner 信息
	ownersDetail := make([]*resp.UserProfileResp, 0)
	if len(createReq.OwnerIDs) > 0 {
		ownersReq := &req.GetUsersReq{UserIDs: createReq.OwnerIDs,
			BaseListRequest: models.BaseListRequest{Page: 1, Limit: len(createReq.OwnerIDs)}}
		ownersDetail, err = s.GetUsers(ctx, ownersReq)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		if len(createReq.OwnerIDs) != len(ownersDetail) {
			return nil, errors.Wrap(errcode.InvalidParams, "contains invalid owner id")
		}
	}

	// 生成实体
	now := time.Now()
	project := &entity.Project{
		ID:                 createReq.GitID,
		Labels:             make([]string, 0),
		OwnerIDs:           make([]string, 0),
		QAEngineers:        make([]*entity.DingDingUserDetail, 0),
		OperationEngineers: make([]*entity.DingDingUserDetail, 0),
		ProductManagers:    make([]*entity.DingDingUserDetail, 0),
		ImageArgs:          make(map[string]string),
		ResourceSpec: map[entity.AppEnvName]entity.ProjectResourceSpec{
			entity.AppEnvFat: {
				CPURequestList: entity.CPUStgRequestResourceList,
				CPULimitList:   entity.CPUStgLimitResourceList,
				MemRequestList: entity.MemStgRequestResourceList,
				MemLimitList:   entity.MemStgLimitResourceList,
			},
			entity.AppEnvStg: {
				CPURequestList: entity.CPUStgRequestResourceList,
				CPULimitList:   entity.CPUStgLimitResourceList,
				MemRequestList: entity.MemStgRequestResourceList,
				MemLimitList:   entity.MemStgLimitResourceList,
			},
			entity.AppEnvPre: {
				CPURequestList: entity.CPUPrdRequestResourceList,
				CPULimitList:   entity.CPUPrdLimitResourceList,
				MemRequestList: entity.MemPrdRequestResourceList,
				MemLimitList:   entity.MemPrdLimitResourceList,
			},
			entity.AppEnvPrd: {
				CPURequestList: entity.CPUPrdRequestResourceList,
				CPULimitList:   entity.CPUPrdLimitResourceList,
				MemRequestList: entity.MemPrdRequestResourceList,
				MemLimitList:   entity.MemPrdLimitResourceList,
			},
		},
		LogStoreName: createReq.Name,
		CreateTime:   &now,
		UpdateTime:   &now,
	}
	err = deepcopy.Copy(createReq).
		SetConfig(&deepcopy.Config{
			NotZeroMode: true,
		}).
		To(project)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	// 落库
	err = s.dao.CreateSingleProject(ctx, project)
	if err != nil {
		return nil, err
	}

	// 返回
	res := &resp.ProjectDetailResp{
		Team:   teamDetail,
		Owners: ownersDetail,
	}
	err = deepcopy.Copy(project).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// 生成更新项目的map
func (s *Service) generateUpdateProjectMap(updateReq *req.UpdateProjectReq) map[string]interface{} {
	change := make(map[string]interface{})
	if updateReq.APIDocURL != "" {
		change["api_doc_url"] = updateReq.APIDocURL
	}

	if updateReq.Desc != "" {
		change["desc"] = updateReq.Desc
	}

	if updateReq.DevDocURL != "" {
		change["dev_doc_url"] = updateReq.DevDocURL
	}

	if updateReq.Language != "" {
		change["language"] = updateReq.Language
	}

	if updateReq.TeamID != "" {
		change["team_id"] = updateReq.TeamID
	}

	if len(updateReq.OwnerIDs) != 0 {
		change["owner_ids"] = updateReq.OwnerIDs
	}

	if updateReq.ApolloAppID != "" {
		change["apollo_appid"] = updateReq.ApolloAppID
	}

	if len(updateReq.Labels) != 0 {
		change["labels"] = updateReq.Labels
	}

	if updateReq.LogStoreName != "" {
		change["log_store_name"] = updateReq.LogStoreName
	}

	// 更新 istio 启用状态
	change["enable_istio"] = updateReq.EnableIstio.ValueOrZero()

	if len(updateReq.ImageArgs) > 0 {
		for branch, args := range updateReq.ImageArgs {
			// mongo不支持 `.` 以及 `$` 等符号，需要特殊处理
			branch = strings.ReplaceAll(branch, "\\", "\\\\")
			branch = strings.ReplaceAll(branch, ".", "\\u002e")
			branch = strings.ReplaceAll(branch, "$", "\\u0024")
			change[fmt.Sprintf("image_args.%s", branch)] = args
		}
	}

	if len(updateReq.QAEngineers) != 0 {
		change["qa_engineers"] = updateReq.QAEngineers
	}
	if len(updateReq.OperationEngineers) != 0 {
		change["operation_engineers"] = updateReq.OperationEngineers
	}
	if len(updateReq.ProductManagers) != 0 {
		change["product_managers"] = updateReq.ProductManagers
	}

	change["update_time"] = time.Now()
	return change
}

// 更新项目
func (s *Service) UpdateProject(ctx context.Context, id string, updateReq *req.UpdateProjectReq) error {
	// 生成更新map
	changeMap := s.generateUpdateProjectMap(updateReq)

	// 更新库
	err := s.dao.UpdateSingleProject(ctx, id, bson.M{
		"$set": changeMap,
	})
	if err != nil {
		return err
	}

	return nil
}

// 删除项目
func (s *Service) DeleteSingleProject(ctx context.Context, id string) error {
	// 数据库信息删除
	err := s.dao.DeleteSingleProject(ctx, bson.M{
		"_id": id,
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteLogStoresByProject 删除项目相关logStore
// 目前仅有阿里云日志关联在项目上, 华为云的日志关联在应用上所以在 status_worker 中清理
func (s *Service) DeleteLogStoresByProject(ctx context.Context, id string) error {
	// logstore删除
	projectDetail, err := s.GetProjectDetail(ctx, id)
	if err != nil {
		return err
	}
	clusterNames := []string{config.Conf.Other.AliLogProjectStgName, config.Conf.Other.AliLogProjectPrdName}
	for _, cluster := range clusterNames {
		_, err = s.GetLogStoreDetail(ctx, &req.AliGetLogStoreDetailReq{
			ProjectName: cluster,
			StoreName:   projectDetail.LogStoreName,
		})
		if err != nil {
			if errcode.EqualError(_errcode.AliResourceNotFoundError, err) {
				continue
			}
			return err
		}

		err = s.DeleteAliLogStore(ctx, &req.AliDeleteLogStoreReq{
			ProjectName: cluster,
			StoreName:   projectDetail.LogStoreName,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// 获取项目详情
func (s *Service) GetProjectDetail(ctx context.Context, id string) (*resp.ProjectDetailResp, error) {
	project, err := s.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}

	teamDetail, err := s.GetTeamDetail(ctx, project.TeamID)
	if errcode.EqualError(_errcode.InvalidHexStringError, err) || errcode.EqualError(errcode.NoRowsFoundError, err) {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}
	if err != nil {
		return nil, err
	}

	res := &resp.ProjectDetailResp{
		Team: teamDetail,
	}

	ownersDetail := make([]*resp.UserProfileResp, 0)
	if len(project.OwnerIDs) != 0 {
		ownersReq := &req.GetUsersReq{UserIDs: project.OwnerIDs,
			BaseListRequest: models.BaseListRequest{Page: 1, Limit: len(project.OwnerIDs)}}
		ownersDetail, err = s.GetUsers(ctx, ownersReq)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}
	}

	res.Owners = ownersDetail

	err = deepcopy.Copy(project).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	// 额外返回支持的特殊配置重命名前缀和模式
	prefixes, err := s.dao.FindAllConfigRenamePrefixes(ctx)
	if err != nil {
		return nil, err
	}

	res.ConfigRenamePrefixes = make([]*resp.ConfigRenamePrefixDetail, len(prefixes))
	for i := range prefixes {
		res.ConfigRenamePrefixes[i] = &resp.ConfigRenamePrefixDetail{
			Name:   prefixes[i].Name,
			Prefix: prefixes[i].Prefix,
		}
	}
	res.ConfigRenameModes = getConfigRenameModes()

	return res, nil
}

// 获取项目过滤器
func (s *Service) getProjectsFilter(_ context.Context, getReq *req.GetProjectsReq) bson.M {
	filter := bson.M{}

	if getReq.Keyword != "" {
		if getReq.KeywordField == "" {
			keyword := getReq.Keyword
			filter["$or"] = bson.A{
				bson.M{
					"name": bson.M{
						"$regex": keyword,
					},
				},
				bson.M{
					"desc": bson.M{
						"$regex": keyword,
					},
				},
				bson.M{
					"_id": keyword,
				},
			}
		} else {
			filter[getReq.KeywordField] = bson.M{
				"$regex": getReq.Keyword,
			}
		}
	}

	if getReq.Name != "" {
		filter["name"] = getReq.Name
	}

	if getReq.Language != "" {
		filter["language"] = getReq.Language
	}

	if getReq.TeamID != "" {
		filter["team_id"] = getReq.TeamID
	}

	if getReq.OwnerID != "" {
		filter["owner_ids"] = bson.M{"$in": []string{getReq.OwnerID}}
	}

	if getReq.Labels != "" {
		labelsArray := strings.Split(getReq.Labels, ",")
		filter["labels"] = bson.M{"$all": labelsArray}
	}

	if len(getReq.IDs) != 0 {
		filter["_id"] = bson.M{"$in": getReq.IDs}
	}

	return filter
}

// 获取项目
func (s *Service) GetProjects(ctx context.Context, getReq *req.GetProjectsReq) ([]*resp.ProjectListResp, error) {
	filter := s.getProjectsFilter(ctx, getReq)

	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit
	projects, err := s.dao.FindProjects(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})
	if err != nil {
		return nil, err
	}

	// 团队本地缓存
	teamLocalCache := make(map[string]*resp.TeamDetailResp)
	res := make([]*resp.ProjectListResp, 0)

	for _, project := range projects {
		item := new(resp.ProjectListResp)
		err = deepcopy.Copy(project).To(item)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		if cacheTeam, ok := teamLocalCache[project.TeamID]; ok {
			item.Team = &resp.TeamListResp{
				ID:             cacheTeam.ID,
				Name:           cacheTeam.Name,
				DingHook:       cacheTeam.DingHook,
				Label:          cacheTeam.Label,
				AliAlarmName:   cacheTeam.AliAlarmName,
				ExtraDingHooks: cacheTeam.ExtraDingHooks,
			}
		} else {
			teamDetail, e := s.GetTeamDetail(ctx, project.TeamID)
			if e != nil {
				return nil, errors.Wrap(errcode.InternalError, e.Error())
			}
			teamLocalCache[project.TeamID] = teamDetail

			item.Team = &resp.TeamListResp{
				ID:             teamDetail.ID,
				Name:           teamDetail.Name,
				DingHook:       teamDetail.DingHook,
				Label:          teamDetail.Label,
				AliAlarmName:   teamDetail.AliAlarmName,
				ExtraDingHooks: teamDetail.ExtraDingHooks,
			}
		}

		// 获取项目负责人
		ownersDetail := make([]*resp.UserProfileResp, 0)
		if len(project.OwnerIDs) > 0 {
			ownersReq := &req.GetUsersReq{UserIDs: project.OwnerIDs,
				BaseListRequest: models.BaseListRequest{Page: 1, Limit: len(project.OwnerIDs)}}
			ownersDetail, err = s.GetUsers(ctx, ownersReq)
			if err != nil {
				return nil, errors.Wrap(errcode.InternalError, err.Error())
			}
		}

		item.Owners = ownersDetail

		res = append(res, item)
	}

	return res, nil
}

func (s *Service) GetProjectsCount(ctx context.Context, getReq *req.GetProjectsReq) (int, error) {
	filter := s.getProjectsFilter(ctx, getReq)

	res, err := s.dao.CountProject(ctx, filter)
	if err != nil {
		return 0, err
	}

	return res, nil
}

// GetProjectUserRole 获取项目下用户的角色(权限)信息
func (s *Service) GetProjectUserRole(ctx context.Context, projectID, userID string) (*resp.ProjectMemberRoleResp, error) {
	memID, err := strconv.Atoi(userID)
	if err != nil {
		return nil, errors.Wrap(errcode.InvalidParams, err.Error())
	}

	members, err := s.GetGitlabProjectAllActiveMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	res := &resp.ProjectMemberRoleResp{AccessLevel: entity.GitMemberAccessNoAccess,
		Role: entity.GitMemberAccessRoleMap[entity.GitMemberAccessNoAccess]}
	for _, member := range members {
		if memID == member.ID {
			res.AccessLevel = member.AccessLevel
			res.Role = entity.GitMemberAccessRoleMap[member.AccessLevel]
			return res, nil
		}
	}

	return res, nil
}

func (s *Service) CheckProjectNameLegal(ctx context.Context, name string) error {
	// 校验非法字符
	if !projectNameRegexp.MatchString(name) {
		return errors.Wrapf(errcode.InvalidParams, "name:%s contains invalid char", name)
	}

	// 应用名不可重复
	count, err := s.GetProjectsCount(ctx, &req.GetProjectsReq{
		Name: name,
	})
	if err != nil {
		return errors.Wrap(errcode.InvalidParams, err.Error())
	} else if count != 0 {
		return errors.Wrapf(errcode.InvalidParams, "name is not only: %s", name)
	}

	return nil
}

// GetAmsFrontendProjectURL get ams frontend project url
func (s *Service) GetAmsFrontendProjectURL(projectID string, envName entity.AppEnvName) string {
	return fmt.Sprintf("%s/project/%s/application?envname=%s",
		config.Conf.Other.AmsFrontendHost, projectID, envName)
}

// IsP0LevelProject judges whether project is P0 level
func (s *Service) IsP0LevelProject(project *resp.ProjectDetailResp) bool {
	for _, label := range project.Labels {
		if label == string(entity.ProjectLabelP0) {
			return true
		}
	}
	return false
}

func (s *Service) IsPrdP0LevelApp(envName entity.AppEnvName, project *resp.ProjectDetailResp) bool {
	return envName == entity.AppEnvPrd && s.IsP0LevelProject(project)
}

func (s *Service) GetProjectByID(ctx context.Context, id string) (*entity.Project, error) {
	return s.dao.FindSingleProject(ctx, bson.M{"_id": id})
}

func (s *Service) GetJenkinsBuildImage(ctx context.Context, projectName, branch, commitID string) (
	*entity.JenkinsBuildImage, error) {
	return s.GetLastJenkinsBuildImage(ctx, &req.GetImageBuildReq{
		ProjectName: projectName, BranchName: branch, CommitID: commitID})
}

func getConfigRenameModes() []*resp.ConfigRenameModeDetail {
	modes := make([]*resp.ConfigRenameModeDetail, len(entity.SupportedConfigRenameModes))
	for i := range modes {
		modes[i] = &resp.ConfigRenameModeDetail{
			Enum: entity.SupportedConfigRenameModes[i],
			Name: entity.GetConfigRenameModeName(entity.SupportedConfigRenameModes[i]),
		}
	}

	return modes
}
