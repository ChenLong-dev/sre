package service

import (
	"rulai/config"
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"github.com/pkg/errors"

	"context"
	"fmt"
	"time"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) GetCIJobDetail(ctx context.Context, projectID string) (*resp.CIJobDetailResp, error) {
	res := new(resp.CIJobDetailResp)
	ciJob, err := s.dao.GetCIJob(ctx, bson.M{
		"project_id": projectID,
	})
	if err != nil {
		return nil, err
	}

	err = deepcopy.Copy(ciJob).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	gitProject, err := s.GetGitlabSingleProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	res.AllowMergeSwitch = gitProject.OnlyAllowMergeIfPipelineSucceeds

	return res, nil
}

func (s *Service) getJenkinsCIJobURL(projectName string) string {
	return fmt.Sprintf("%s/view/AMS-CI/job/%s",
		config.Conf.JenkinsCI.GoJenkins.BaseURL, s.getJenkinsCIJobName(projectName))
}

func (s *Service) getJenkinsCIJobName(projectName string) string {
	return fmt.Sprintf("%s-%s", "ci", projectName)
}

func (s *Service) CreateProjectCIJob(ctx context.Context, projectID, projectName string,
	createReq *req.CreateProjectCIJobReq) error {
	// Check if ci job already exists;
	// If user create one project not once, we must guarantee only one ci job created.
	_, err := s.GetCIJobDetail(ctx, projectID)
	if err != nil && !errcode.EqualError(errcode.NoRowsFoundError, err) {
		return err
	}
	if err == nil {
		return errors.Wrap(_errcode.ProjectCIJobExistsError, "项目ci job已存在")
	}

	createReq.CIJobName = s.getJenkinsCIJobName(projectName)
	createReq.HookURL = s.getGitlabCIHookURL(createReq.CIJobName)
	createReq.ViewURL = s.getJenkinsCIJobURL(projectName)

	if !createReq.AllowMergeSwitch.IsZero() {
		editErr := s.EditGitlabProject(ctx, projectID, &req.EditGitlabProjectReq{
			OnlyAllowMergeIfPipelineSucceeds: createReq.AllowMergeSwitch.ValueOrZero(),
		})
		if editErr != nil {
			return editErr
		}
	}

	err = s.CreateJenkinsCIJob(ctx, projectID, createReq)
	if err != nil {
		return err
	}

	// 落库
	now := time.Now()
	ciJob := &entity.CIJob{
		ID:                  primitive.NewObjectID(),
		Name:                createReq.CIJobName,
		ProjectID:           projectID,
		ViewURL:             createReq.ViewURL,
		MessageNotification: createReq.MessageNotification,
		PipelineStages:      createReq.PipelineStages,
		DeployBranchName:    createReq.DeployBranchName,
		HookURL:             createReq.HookURL,
		CreateTime:          &now,
		UpdateTime:          &now,
	}

	err = s.dao.CreateCIJob(ctx, ciJob)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteProjectCIJob(ctx context.Context, projectID, projectName string) error {
	err := s.DeleteGitlabProjectCIHook(ctx, projectID, projectName)
	if err != nil && !errcode.EqualError(_errcode.GitlabHookNotExistsError, err) {
		return err
	}

	_, err = s.GetCIJobDetail(ctx, projectID)
	if err != nil && !errcode.EqualError(errcode.NoRowsFoundError, err) {
		return err
	}

	err = s.dao.DeleteCIJob(ctx, bson.M{
		"project_id": projectID,
	})
	return err
}

func (s *Service) generateUpdateProjectCIJobMap(updateReq *req.UpdateProjectCIJobReq) map[string]interface{} {
	change := make(map[string]interface{})
	change["update_time"] = time.Now()
	change["message_notification"] = updateReq.MessageNotification
	change["pipeline_stages"] = updateReq.PipelineStages
	if len(updateReq.DeployBranchName) != 0 {
		change["deploy_branch_name"] = updateReq.DeployBranchName
	}
	if updateReq.HookURL != "" {
		change["hook_url"] = updateReq.HookURL
	}
	if updateReq.ViewURL != "" {
		change["view_url"] = updateReq.ViewURL
	}

	return change
}

func (s *Service) UpdateProjectCIJob(ctx context.Context, projectID, projectName, ciJobID string,
	updateReq *req.UpdateProjectCIJobReq) error {
	if !updateReq.AllowMergeSwitch.IsZero() {
		err := s.EditGitlabProject(ctx, projectID, &req.EditGitlabProjectReq{
			OnlyAllowMergeIfPipelineSucceeds: updateReq.AllowMergeSwitch.ValueOrZero(),
		})
		if err != nil {
			return err
		}
	}

	err := s.ApplyJenkinsCIJob(ctx, projectID, &req.CreateProjectCIJobReq{
		CIJobName: s.getJenkinsCIJobName(projectName),
	})
	if err != nil {
		return err
	}

	// 落库
	changeMap := s.generateUpdateProjectCIJobMap(updateReq)
	err = s.dao.UpdateCIJob(ctx, ciJobID, bson.M{
		"$set": changeMap,
	})
	return err
}

// GetProjectCIJobs returns ci jobs by page and limit
func (s *Service) GetProjectCIJobs(ctx context.Context, getReq *req.GetProjectCIJobs) ([]*resp.CIJobDetailResp, error) {
	res := make([]*resp.CIJobDetailResp, 0)
	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit

	ciJobs, err := s.dao.GetCIJobs(ctx, bson.M{}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})
	if err != nil {
		return nil, err
	}

	if err = deepcopy.Copy(ciJobs).To(&res); err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}
