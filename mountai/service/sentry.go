package service

import (
	"rulai/config"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/httpclient"
)

// 创建Sentry项目
func (s *Service) CreateSentryProject(ctx context.Context, projectReq *req.CreateSentryProjectReq) (*resp.CreateSentryProjectResp, error) {
	createProjectRes := new(resp.CreateSentryProjectResp)
	createResp := s.httpClient.Builder().
		Method(http.MethodPost).
		URL(
			fmt.Sprintf("%s/api/0/teams/%s/%s/projects/",
				config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, projectReq.TeamSlug),
		).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		JsonBody(map[string]string{"name": projectReq.ProjectName, "slug": projectReq.ProjectSlug}).
		AccessStatusCode(http.StatusCreated, http.StatusOK).
		Fetch(ctx)
	err := createResp.DecodeJSON(&createProjectRes)
	if err != nil {
		if createResp.StatusCode == http.StatusConflict {
			return nil, errors.Wrapf(_errcode.SentryProjectExistsError, "%s", err.Error())
		}

		return nil, errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}

	return createProjectRes, nil
}

// 启动sentry 钉钉通知
func (s *Service) EnableSentryDingDing(ctx context.Context, enableReq *req.EnableSentryDingDingReq) error {
	err := s.httpClient.Builder().
		Method(http.MethodPost).
		URL(
			fmt.Sprintf("%s/api/0/projects/%s/%s/plugins/DingDing/",
				config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, enableReq.ProjectSlug),
		).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		AccessStatusCode(http.StatusCreated, http.StatusOK).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}

	return nil
}

// GetSentryDingDing 获取sentry钉钉配置
func (s *Service) GetSentryDingDing(ctx context.Context, getReq *req.GetSentryDingDingReq) (*resp.SentryDingDingResp, error) {
	dingResp := new(resp.SentryDingDingResp)
	err := s.httpClient.Builder().
		Method(http.MethodGet).
		URL(
			fmt.Sprintf("%s/api/0/projects/%s/%s/plugins/DingDing/",
				config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, getReq.ProjectSlug),
		).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		AccessStatusCode(http.StatusCreated, http.StatusOK).
		Fetch(ctx).
		DecodeJSON(dingResp)
	if err != nil {
		return nil, errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}
	return dingResp, nil
}

// UpdateSentryDingDing 更新sentry 钉钉配置
func (s *Service) UpdateSentryDingDing(ctx context.Context, updateReq *req.UpdateSentryDingDingReq) error {
	err := s.httpClient.Builder().
		Method(http.MethodPut).
		URL(
			fmt.Sprintf("%s/api/0/projects/%s/%s/plugins/DingDing/",
				config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, updateReq.ProjectSlug),
		).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		JsonBody(map[string]string{"access_token": updateReq.AccessToken}).
		AccessStatusCode(http.StatusCreated, http.StatusOK).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}
	return nil
}

// CheckAndUpdateProjectSentryDingDing 检测并更新sentry的钉钉
func (s *Service) CheckAndUpdateProjectSentryDingDing(ctx context.Context,
	projectDetail *resp.ProjectDetailResp, appDetail *resp.AppDetailResp) error {
	dingResp, err := s.GetSentryDingDing(ctx, &req.GetSentryDingDingReq{
		ProjectSlug: appDetail.SentryProjectSlug,
	})
	if err != nil {
		return err
	}

	if len(dingResp.Config) == 0 || dingResp.Config[0].Value == projectDetail.Team.DingHook {
		return nil
	}

	err = s.UpdateSentryDingDing(ctx, &req.UpdateSentryDingDingReq{
		ProjectSlug: appDetail.SentryProjectSlug,
		AccessToken: projectDetail.Team.DingHook,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetSentryProject 获取sentry项目
func (s *Service) GetSentryProject(ctx context.Context, getReq *req.GetSentryProjectReq) (*resp.SentryProjectDetailResp, error) {
	// 获取sentry的project信息
	projectResp := new(resp.SentryProjectDetailResp)
	err := s.httpClient.Builder().
		Method(http.MethodGet).
		URL(
			fmt.Sprintf("%s/api/0/projects/%s/%s/",
				config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, getReq.ProjectSlug),
		).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		Fetch(ctx).
		DecodeJSON(projectResp)
	if err != nil {
		return nil, errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}
	return projectResp, nil
}

// AddSentryProjectTeam 增加sentry团队
func (s *Service) AddSentryProjectTeam(ctx context.Context, addReq *req.AddSentryProjectTeamReq) error {
	// 添加新team
	err := s.httpClient.Builder().
		Method(http.MethodPost).
		URL(
			fmt.Sprintf("%s/api/0/projects/%s/%s/teams/%s/",
				config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, addReq.ProjectSlug, addReq.TeamSlug),
		).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		AccessStatusCode(http.StatusCreated, http.StatusOK).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}
	return nil
}

// RemoveSentryProjectTeam 删除sentry团队
func (s *Service) RemoveSentryProjectTeam(ctx context.Context, removeReq *req.RemoveSentryProjectTeamReq) error {
	// 删除原team
	err := s.httpClient.Builder().
		Method(http.MethodDelete).
		URL(
			fmt.Sprintf("%s/api/0/projects/%s/%s/teams/%s/",
				config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, removeReq.ProjectSlug, removeReq.TeamSlug),
		).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		AccessStatusCode(http.StatusCreated, http.StatusOK).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}
	return nil
}

// CheckAndUpdateProjectSentryTeam 检测并更新sentry的团队
func (s *Service) CheckAndUpdateProjectSentryTeam(ctx context.Context,
	projectDetail *resp.ProjectDetailResp, appDetail *resp.AppDetailResp, oldTeamSlug string) error {
	if projectDetail.Team.SentrySlug == oldTeamSlug {
		return nil
	}

	projectResp, err := s.GetSentryProject(ctx, &req.GetSentryProjectReq{
		ProjectSlug: appDetail.SentryProjectSlug,
	})
	if err != nil {
		return err
	}

	flgAdd := false
	flgRmv := false
	for _, team := range projectResp.Teams {
		if team.Slug == projectDetail.Team.SentrySlug {
			flgAdd = true
		}
		if team.Slug == oldTeamSlug {
			flgRmv = true
		}
	}

	// 添加团队
	if !flgAdd {
		err = s.AddSentryProjectTeam(ctx, &req.AddSentryProjectTeamReq{
			ProjectSlug: appDetail.SentryProjectSlug,
			TeamSlug:    projectDetail.Team.SentrySlug,
		})
		if err != nil {
			return err
		}
	}

	// 删除团队
	if flgRmv {
		err = s.RemoveSentryProjectTeam(ctx, &req.RemoveSentryProjectTeamReq{
			ProjectSlug: appDetail.SentryProjectSlug,
			TeamSlug:    oldTeamSlug,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// 获取项目的Sentry dsn
func (s *Service) GetSentryProjectKeys(ctx context.Context, getReq *req.GetSentryProjectKeyReq) ([]*resp.GetSentryProjectKeyResp, error) {
	getProjectKeyResp := make([]*resp.GetSentryProjectKeyResp, 1)
	err := s.httpClient.Builder().
		Method(http.MethodGet).
		URL(
			fmt.Sprintf("%s/api/0/projects/%s/%s/keys/",
				config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, getReq.ProjectSlug),
		).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		Fetch(ctx).
		DecodeJSON(&getProjectKeyResp)
	if err != nil {
		return nil, errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}

	return getProjectKeyResp, nil
}

func (s *Service) GetSentryProjectPublicDsn(ctx context.Context, getReq *req.GetSentryProjectKeyReq) (string, error) {
	getProjectKeyResp, err := s.GetSentryProjectKeys(ctx, getReq)
	if err != nil {
		return "", err
	}

	if len(getProjectKeyResp) == 0 {
		return "", errors.New("sentry project key is empty")
	}

	if getProjectKeyResp[0].ProjectDsn.Public == "" {
		return "", errors.New("sentry project public dsn is empty")
	}

	return getProjectKeyResp[0].ProjectDsn.Public, nil
}

// 创建Sentry team
func (s *Service) CreateSentryTeam(ctx context.Context, createReq *req.CreateSentryTeamReq) (*resp.CreateSentryTeamResp, error) {
	createTeamResp := new(resp.CreateSentryTeamResp)
	err := s.httpClient.Builder().
		Method(http.MethodPost).
		URL(fmt.Sprintf("%s/api/0/organizations/sentry/teams/", config.Conf.SentrySystem.Host)).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		JsonBody(map[string]string{"name": createReq.TeamName, "slug": createReq.TeamSlug}).
		AccessStatusCode(http.StatusCreated, http.StatusOK).
		Fetch(ctx).
		DecodeJSON(&createTeamResp)
	if err != nil {
		return nil, errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}

	return createTeamResp, nil
}

// 删除Sentry项目
func (s *Service) DeleteSentryProject(ctx context.Context, deleteReq *req.DeleteSentryProjectReq) error {
	err := s.httpClient.Builder().
		Method(http.MethodDelete).
		URL(fmt.Sprintf("%s/api/0/projects/%s/%s/", config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization, deleteReq.ProjectSlug)).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		AccessStatusCode(http.StatusOK, http.StatusNoContent).
		Fetch(ctx).
		Error()
	if err != nil {
		return errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}

	return nil
}

// 获取Sentry所有team
func (s *Service) GetSentryTeams(ctx context.Context) ([]*resp.GetSentryTeamsResp, error) {
	teamsResp := make([]*resp.GetSentryTeamsResp, 0)
	err := s.httpClient.Builder().
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/api/0/organizations/%s/teams/", config.Conf.SentrySystem.Host, config.Conf.SentrySystem.Organization)).
		Headers(httpclient.NewJsonHeader().Add("Authorization", fmt.Sprintf("Bearer %s", config.Conf.SentrySystem.AuthToken))).
		Fetch(ctx).
		DecodeJSON(&teamsResp)
	if err != nil {
		return nil, errors.Wrapf(_errcode.SentryInternalError, "%s", err.Error())
	}

	return teamsResp, nil
}

// 创建app的sentry项目
func (s *Service) CreateAppSentry(ctx context.Context, team *resp.TeamDetailResp,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) (*resp.CreateAppSentryResp, error) {
	projectSlug := fmt.Sprintf("%s_%s", project.Name, app.Name)
	token, err := team.GetDingAccessToken()
	if err != nil {
		return nil, err
	}

	sentryProject, err := s.CreateSentryProject(ctx, &req.CreateSentryProjectReq{
		ProjectName: projectSlug,
		TeamSlug:    team.SentrySlug,
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	sentryProjectPublicDsn, err := s.GetSentryProjectPublicDsn(ctx, &req.GetSentryProjectKeyReq{
		ProjectSlug: sentryProject.ProjectSlug,
	})
	if err != nil {
		return nil, err
	}

	err = s.EnableSentryDingDing(ctx, &req.EnableSentryDingDingReq{
		ProjectSlug: projectSlug,
	})
	if err != nil {
		return nil, err
	}

	err = s.UpdateSentryDingDing(ctx, &req.UpdateSentryDingDingReq{
		ProjectSlug: projectSlug,
		AccessToken: token,
	})
	if err != nil {
		return nil, err
	}

	return &resp.CreateAppSentryResp{
		SentryProjectPublicDsn: sentryProjectPublicDsn,
		SentryProjectSlug:      projectSlug,
	}, nil
}
