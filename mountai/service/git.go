package service

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"gitlab.shanhai.int/sre/library/log"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/httpclient"
)

var (
	frameworkVersionRegexp = regexp.MustCompile(`gitlab\.whjzjx\.cn/shanghai/zeus [\w.]+`)
	libraryVersionRegexp   = regexp.MustCompile(`gitlab\.whjzjx\.cn/shanghai/zeus [\w.]+`)
)

// 校验git token
func (s *Service) AuthGitlabToken(ctx context.Context, getReq *req.UserAuthLoginReq) (*resp.GitAuthTokenResp, error) {
	res := new(resp.GitAuthTokenResp)

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/oauth/token", config.Conf.Git.Host)).
		Headers(httpclient.NewFormURLEncodedHeader()).
		FormBody(
			httpclient.NewForm().
				Add("grant_type", "password").
				Add("username", getReq.UserName).
				Add("password", getReq.Password),
		).
		Method(http.MethodPost).
		AccessStatusCode(http.StatusOK, http.StatusBadRequest).
		Fetch(ctx).
		DecodeJSON(res)

	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	if res.Error == "invalid_grant" {
		return nil, errors.Wrap(_errcode.GitlabUserAuthError, res.ErrorDesc)
	}

	if res.Error != "" {
		return nil, errors.Wrap(_errcode.GitlabInternalError, res.ErrorDesc)
	}

	return res, nil
}

// 获取git用户信息
func (s *Service) GetGitlabUserProfile(ctx context.Context, token string) (*resp.GitUserProfileResp, error) {
	res := new(resp.GitUserProfileResp)

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/user", config.Conf.Git.Host)).
		QueryParams(httpclient.NewUrlValue().Add("access_token", token)).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}

// 获取git项目信息
func (s *Service) GetGitlabSingleProject(ctx context.Context, id string) (*resp.GitProjectResp, error) {
	res := new(resp.GitProjectResp)

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/projects/%s", config.Conf.Git.Host, id)).
		QueryParams(httpclient.NewUrlValue().Add("private_token", config.Conf.Git.Token)).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}

// 获取git项目的分支列表
func (s *Service) GetGitlabSingleProjectBranches(ctx context.Context, id string,
	getReq *req.GetRepoBranchesReq) ([]*resp.GitBranchResp, error) {
	res := make([]*resp.GitBranchResp, 0)
	query := httpclient.NewUrlValue().Add("private_token", config.Conf.Git.Token)

	if getReq.Page != 0 {
		query.Add("page", strconv.Itoa(getReq.Page))
	}

	if getReq.PageSize != 0 {
		query.Add("per_page", strconv.Itoa(getReq.PageSize))
	}

	if getReq.Keyword != "" {
		query.Add("search", getReq.Keyword)
	}

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/projects/%s/repository/branches", config.Conf.Git.Host, id)).
		QueryParams(query).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(&res)
	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}

// 获取git项目的所有分支列表
func (s *Service) GetGitlabSingleProjectAllBranches(ctx context.Context, id string) ([]*resp.GitBranchResp, error) {
	res := make([]*resp.GitBranchResp, 0)
	pageSize := 100
	page := 1

	for {
		branches, err := s.GetGitlabSingleProjectBranches(ctx, id, &req.GetRepoBranchesReq{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return nil, err
		}

		res = append(res, branches...)
		page++

		if len(branches) < pageSize {
			break
		}
	}

	return res, nil
}

// 获取git项目的分支信息
func (s *Service) GetGitlabSingleProjectBranch(ctx context.Context, id, branchName string) (*resp.GitBranchResp, error) {
	res := new(resp.GitBranchResp)

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/projects/%s/repository/branches/%s", config.Conf.Git.Host, id, url.PathEscape(branchName))).
		QueryParams(httpclient.NewUrlValue().Add("private_token", config.Conf.Git.Token)).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}

// 获取git项目文件
func (s *Service) GetGitlabProjectRawFile(ctx context.Context, id, ref, filePath string) (string, error) {
	res, err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/projects/%s/repository/files/%s/raw", config.Conf.Git.Host, id, url.PathEscape(filePath))).
		QueryParams(httpclient.NewUrlValue().Add("private_token", config.Conf.Git.Token).Add("ref", ref)).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		Fetch(ctx).
		Body()
	if err != nil {
		return "", errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) GetQTFrameworkVersion(ctx context.Context, id, branchName string) (*resp.GetQTFrameworkVersionResp, error) {
	res := new(resp.GetQTFrameworkVersionResp)

	// 获取源文件
	modFile, err := s.GetGitlabProjectRawFile(ctx, id, branchName, "go.mod")
	if err != nil {
		// 若不存在，则直接返回空
		if strings.Contains(err.Error(), strconv.Itoa(http.StatusNotFound)) {
			return res, nil
		}
		return nil, err
	}

	// 正则匹配
	frameworkLine := frameworkVersionRegexp.FindAllString(modFile, -1)
	if len(frameworkLine) > 0 {
		frameworkStrings := strings.Split(frameworkLine[0], " ")
		if len(frameworkStrings) == 2 {
			res.FrameworkVersion = frameworkStrings[1]
		}
	}

	libraryLine := libraryVersionRegexp.FindAllString(modFile, -1)
	if len(libraryLine) > 0 {
		libraryStrings := strings.Split(libraryLine[0], " ")
		if len(libraryStrings) == 2 {
			res.LibraryVersion = libraryStrings[1]
		}
	}

	return res, nil
}

// GetGitlabUser 获取git用户信息
func (s *Service) GetGitlabUser(ctx context.Context, id string) (*resp.GitUserResp, error) {
	res := new(resp.GitUserResp)
	query := httpclient.NewUrlValue().Add("private_token", config.Conf.Git.Token)

	response := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/users/%s", config.Conf.Git.Host, id)).
		QueryParams(query).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		AccessStatusCode(http.StatusOK, http.StatusNotFound).
		Fetch(ctx)

	if response.Error() != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, response.Error().Error())
	}

	if response.StatusCode == http.StatusNotFound {
		return nil, _errcode.GitLabUserNotFound
	}

	err := response.DecodeJSON(res)

	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}

// GetGitlabProjectMembers 分页获取项目成员
func (s *Service) GetGitlabProjectMembers(ctx context.Context, id string,
	getReq *req.GetProjectMembersReq) ([]*resp.GitProjectMemberResp, error) {
	res := new([]*resp.GitProjectMemberResp)
	query := httpclient.NewUrlValue().Add("private_token", config.Conf.Git.Token)

	if getReq.Page != 0 {
		query.Add("page", strconv.Itoa(getReq.Page))
	}

	if getReq.PageSize != 0 {
		query.Add("per_page", strconv.Itoa(getReq.PageSize))
	}

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/projects/%s/members/all", config.Conf.Git.Host, id)).
		QueryParams(query).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return *res, nil
}

// GetGitlabProjectAllActiveMembers 获取一个项目下所有活跃的成员列表
func (s *Service) GetGitlabProjectAllActiveMembers(ctx context.Context, id string) ([]*resp.GitProjectMemberResp, error) {
	members, e := s.dao.GetGitlabProjectActiveMembersFromCache(ctx, id)

	if e != nil {
		return nil, e
	}

	if members != nil {
		return members, nil
	}

	res := make([]*resp.GitProjectMemberResp, 0)
	pageSize := 100
	page := 1

	// membersIDIndexMap 用于去重
	membersIDIndexMap := make(map[int]int)
	var index int

	for {
		members, err := s.GetGitlabProjectMembers(ctx, id, &req.GetProjectMembersReq{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return nil, err
		}

		for _, member := range members {
			if member.State == entity.GitMemberStateActive {
				i, ok := membersIDIndexMap[member.ID]
				if !ok {
					res = append(res, member)
					membersIDIndexMap[member.ID] = index
					index++
				} else if member.AccessLevel > res[i].AccessLevel {
					res[i] = member
				}
			}
		}

		page++

		if len(members) < pageSize {
			break
		}
	}

	e = s.dao.SetGitlabProjectActiveMembersToCache(ctx, id, res)
	if e != nil {
		log.Errorc(ctx, "occurred error during set gitlab members info to cache, error: %s", e)
	}

	return res, nil
}

// 给gitlab项目添加hook，该hook用于触发CI流程
func (s *Service) AddGitlabProjectCIHook(ctx context.Context, projectID, hookURL string) error {
	hooks, err := s.ListGitlabProjectAllHooks(ctx, projectID)
	if err != nil {
		return err
	}

	// 检查hook是否已经存在
	for _, hook := range hooks {
		if hook.URL == hookURL {
			return errors.Wrapf(_errcode.GitlabHookExistsError, "hook %s already exists", hookURL)
		}
	}

	err = s.AddGitlabProjectHook(ctx, &req.AddGitlabProjectHookReq{
		ProjectID: projectID,
		HookDetail: &req.GitlabProjectHookDetailReq{
			URL:                   hookURL,
			Token:                 config.Conf.JenkinsCI.GitlabSecretToken,
			PushEvents:            true,
			MergeRequestsEvents:   true,
			EnableSSLVerification: true,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// 添加gitlab钩子
func (s *Service) AddGitlabProjectHook(ctx context.Context, addReq *req.AddGitlabProjectHookReq) error {
	err := s.httpClient.Builder().
		Method(http.MethodPost).
		URL(fmt.Sprintf("%s/api/v4/projects/%s/hooks", config.Conf.Git.Host, addReq.ProjectID)).
		Headers(httpclient.NewJsonHeader().Add("Private-Token", config.Conf.Git.CIToken)).
		JsonBody(addReq.HookDetail).
		AccessStatusCode(http.StatusCreated, http.StatusOK).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return nil
}

// 列出gitlab项目下的hook，带分页
func (s *Service) ListGitlabProjectHooks(ctx context.Context, projectID string,
	listReq *req.ListGitlabProjectHooksReq) ([]resp.ListGitlabProjectHooksResp, error) {
	res := make([]resp.ListGitlabProjectHooksResp, 0)
	query := httpclient.NewUrlValue().Add("private_token", config.Conf.Git.CIToken)

	if listReq.Page != 0 {
		query.Add("page", strconv.Itoa(listReq.Page))
	}

	if listReq.PageSize != 0 {
		query.Add("per_page", strconv.Itoa(listReq.PageSize))
	}

	err := s.httpClient.Builder().
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/api/v4/projects/%s/hooks", config.Conf.Git.Host, projectID)).
		QueryParams(query).
		Headers(httpclient.GetDefaultHeader()).
		Fetch(ctx).
		DecodeJSON(&res)
	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}

// 列出gitlab项目所有hook
func (s *Service) ListGitlabProjectAllHooks(ctx context.Context, projectID string) ([]resp.ListGitlabProjectHooksResp, error) {
	res := make([]resp.ListGitlabProjectHooksResp, 0)
	pageSize := 20
	page := 1

	for {
		hooks, err := s.ListGitlabProjectHooks(ctx, projectID, &req.ListGitlabProjectHooksReq{
			Page:     page,
			PageSize: pageSize,
		})
		if err != nil {
			return nil, err
		}

		res = append(res, hooks...)
		page++

		if len(hooks) < pageSize {
			break
		}
	}

	return res, nil
}

// gitlab hook url
func (s *Service) getGitlabCIHookURL(jenkinsProjectName string) string {
	return fmt.Sprintf("%s/project/%s", config.Conf.JenkinsCI.GoJenkins.BaseURL, jenkinsProjectName)
}

// 删除gitlab hook
func (s *Service) DeleteGitlabProjectCIHook(ctx context.Context, projectID, projectName string) error {
	hooks, err := s.ListGitlabProjectAllHooks(ctx, projectID)
	if err != nil {
		return err
	}

	hookID := -1
	hookURL := s.getGitlabCIHookURL(s.getJenkinsCIJobName(projectName))
	for _, hook := range hooks {
		if hook.URL == hookURL {
			hookID = hook.ID
			break
		}
	}

	if hookID == -1 {
		return errors.Wrapf(_errcode.GitlabHookNotExistsError, "%s is not exist", hookURL)
	}

	err = s.DeleteGitlabHook(ctx, projectID, hookID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteGitlabHook(ctx context.Context, projectID string, hookID int) error {
	err := s.httpClient.Builder().
		Method(http.MethodDelete).
		URL(fmt.Sprintf("%s/api/v4/projects/%s/hooks/%s", config.Conf.Git.Host, projectID, strconv.Itoa(hookID))).
		Headers(httpclient.GetDefaultHeader().Add("Private-Token", config.Conf.Git.CIToken)).
		AccessStatusCode(http.StatusOK, http.StatusNoContent).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return nil
}

func (s *Service) EditGitlabProject(ctx context.Context, projectID string, editReq *req.EditGitlabProjectReq) error {
	err := s.httpClient.Builder().
		Method(http.MethodPut).
		URL(fmt.Sprintf("%s/api/v4/projects/%s", config.Conf.Git.Host, projectID)).
		Headers(httpclient.GetDefaultHeader().Add("Private-Token", config.Conf.Git.CIToken)).
		JsonBody(editReq).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return nil
}

// EditGitlabProjectHookByURL edit hook by hook url
func (s *Service) EditGitlabProjectHookByURL(ctx context.Context, projectID, hookURL string,
	editDetail *req.GitlabProjectHookDetailReq) error {
	hooks, err := s.ListGitlabProjectAllHooks(ctx, projectID)
	if err != nil {
		return err
	}

	// 检查hook是否已经存在
	for _, hook := range hooks {
		if hook.URL == hookURL {
			return s.EditGitlabProjectHook(ctx, projectID, strconv.Itoa(hook.ID), editDetail)
		}
	}

	return errors.Wrapf(_errcode.GitlabHookNotExistsError, "hook does not exist, hook url: %s", hookURL)
}

// EditGitlabProjectHook edit gitlab hook
func (s *Service) EditGitlabProjectHook(ctx context.Context, projectID, hookID string,
	editDetail *req.GitlabProjectHookDetailReq) error {
	err := s.httpClient.Builder().
		Method(http.MethodPut).
		URL(fmt.Sprintf("%s/api/v4/projects/%s/hooks/%s",
			config.Conf.Git.Host, projectID, hookID)).
		Headers(httpclient.NewJsonHeader().Add("Private-Token", config.Conf.Git.CIToken)).
		JsonBody(editDetail).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return nil
}

// 获取git log
func (s *Service) GetGitlabProjectCommitMsg(ctx context.Context, projectID, id string) (*resp.GitCommitResp, error) {
	res := new(resp.GitCommitResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/projects/%s/repository/commits/%s", config.Conf.Git.Host, projectID, id)).
		QueryParams(httpclient.NewUrlValue().Add("private_token", config.Conf.Git.Token)).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) GetGitlabSingleProjectTags(ctx context.Context, id string) ([]*resp.GitTagResp, error) {
	res := make([]*resp.GitTagResp, 0)
	query := httpclient.NewUrlValue().Add("private_token", config.Conf.Git.Token)

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/api/v4/projects/%s/repository/tags", config.Conf.Git.Host, id)).
		QueryParams(query).
		Headers(httpclient.GetDefaultHeader()).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(&res)
	if err != nil {
		return nil, errors.Wrap(_errcode.GitlabInternalError, err.Error())
	}

	return res, nil
}
