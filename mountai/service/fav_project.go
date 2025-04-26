package service

import (
	"context"
	"time"

	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetFavProjectResp 获取收藏返回
func (s *Service) GetFavProjectResp(ctx context.Context, favProject *entity.FavProject) (*resp.FavProjectResp, error) {
	res := new(resp.FavProjectResp)

	err := deepcopy.Copy(favProject).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	res.User, err = s.GetUserInfo(ctx, favProject.UserID)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	res.Project, err = s.GetProjectDetail(ctx, favProject.ProjectID)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// CreateFavProject 创建收藏项目
func (s *Service) CreateFavProject(ctx context.Context, createReq *req.CreateFavProjectReq,
	operatorID string) (*resp.FavProjectResp, error) {
	now := time.Now()
	favProject := &entity.FavProject{
		ID:         primitive.NewObjectID(),
		UserID:     operatorID,
		CreateTime: &now,
		UpdateTime: &now,
	}
	err := deepcopy.Copy(createReq).To(favProject)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	_, err = s.dao.CreateFavProject(ctx, favProject)
	if err != nil {
		return nil, err
	}

	return s.GetFavProjectResp(ctx, favProject)
}

// getFavProjectsFilter 获取收藏列表过滤条件
func (s *Service) getFavProjectsFilter(_ context.Context, getReq *req.GetFavProjectReq) bson.M {
	filter := bson.M{}

	if getReq.UserID != "" {
		filter["user_id"] = getReq.UserID
	}

	if getReq.ProjectID != "" {
		filter["project_id"] = getReq.ProjectID
	}

	return filter
}

// GetFavProjects 获取收藏列表
func (s *Service) GetFavProjects(ctx context.Context, getReq *req.GetFavProjectReq) ([]*resp.FavProjectResp, error) {
	filter := s.getFavProjectsFilter(ctx, getReq)

	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit
	favProjects, err := s.dao.GetFavProjectList(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})
	if err != nil {
		return nil, err
	}

	res := make([]*resp.FavProjectResp, 0)

	for _, item := range favProjects {
		favProjectResp, err := s.GetFavProjectResp(ctx, item)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		res = append(res, favProjectResp)
	}

	return res, nil
}

// DeleteFavProject 删除收藏项目
func (s *Service) DeleteFavProject(ctx context.Context, userID, projectID string) error {
	return s.dao.DeleteFavProject(ctx, bson.M{"project_id": projectID, "user_id": userID})
}

// DeleteFavProjectByProjectID 删除收藏项目
func (s *Service) DeleteFavProjectByProjectID(ctx context.Context, projectID string) error {
	return s.dao.DeleteFavProject(ctx, bson.M{"project_id": projectID})
}

// CountFavProject 获取收藏数量
func (s *Service) CountFavProject(ctx context.Context, getReq *req.GetFavProjectReq) (int, error) {
	filter := s.getFavProjectsFilter(ctx, getReq)

	return s.dao.CountFavProjects(ctx, filter)
}

// IsProjectFav 项目是否被收藏
func (s *Service) IsProjectFav(ctx context.Context, projectID, userID string) (bool, error) {
	favProject, err := s.dao.GetFavProject(ctx, bson.M{
		"user_id":    userID,
		"project_id": projectID,
	})
	if favProject != nil {
		return true, nil
	}
	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		return false, nil
	}

	return false, err
}
