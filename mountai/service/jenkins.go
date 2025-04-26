package service

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"context"
	"net/http"
	"strconv"
)

func (s *Service) initJenkinsCIConfigTemplate(_ context.Context,
	project *resp.ProjectDetailResp) *entity.JenkinsCIConfigTemplate {
	return &entity.JenkinsCIConfigTemplate{
		ProjectID:             project.ID,
		GitlabSecretToken:     config.Conf.JenkinsCI.GitlabSecretToken,
		PipelineBranch:        config.Conf.JenkinsCI.PipelineBranch,
		ScriptPath:            config.Conf.JenkinsCI.ScriptPath,
		GitLabConnection:      config.Conf.JenkinsCI.GitLabConnection,
		PipelineURL:           config.Conf.JenkinsCI.PipelineURL,
		PipelineCredentialsID: config.Conf.JenkinsCI.PipelineCredentialsID,
	}
}

// 创建jenkins job
func (s *Service) ApplyJenkinsCIJob(ctx context.Context, projectID string,
	createReq *req.CreateProjectCIJobReq) error {
	tpl := s.initJenkinsCIConfigTemplate(ctx, &resp.ProjectDetailResp{
		ID: projectID,
	})

	jenkinsConfig, err := s.RenderTemplate(ctx, "./template/jenkins/CIConfig.xml", tpl)
	if err != nil {
		return err
	}

	// 创建Jenkins的CI流程pipeline
	job, err := s.jenkinsCIClient.GetJob(createReq.CIJobName)
	if err != nil {
		if err.Error() != strconv.Itoa(http.StatusNotFound) {
			return errors.Wrap(_errcode.JenkinsInternalError, err.Error())
		}

		// 不存在则创建
		_, err = s.jenkinsCIClient.CreateJob(jenkinsConfig, createReq.CIJobName)
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

	return nil
}

// 创建jenkins跑CI流程job，并且给gitlab添加触发hook
func (s *Service) CreateJenkinsCIJob(ctx context.Context, projectID string,
	createReq *req.CreateProjectCIJobReq) error {
	// 创建Jenkins job
	err := s.ApplyJenkinsCIJob(ctx, projectID, createReq)
	if err != nil {
		return err
	}

	// 给项目添加Jenkins hook，以触发CI流程
	err = s.AddGitlabProjectCIHook(ctx, projectID, createReq.HookURL)
	if err != nil && !errcode.EqualError(_errcode.GitlabHookExistsError, err) {
		return err
	}

	return nil
}
