package service

import (
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) getVariableFilter(_ context.Context, getReq *req.GetVariablesReq) bson.M {
	filter := bson.M{}

	if getReq.Type != 0 {
		filter["type"] = getReq.Type
	}

	if getReq.ProjectID != "" {
		filter["project_id"] = getReq.ProjectID
	}

	if getReq.Key != "" {
		filter["key"] = getReq.Key
	}

	return filter
}

func (s *Service) GetProjectVariables(ctx context.Context, projectID string) ([]*entity.Variable, error) {
	filter := s.getVariableFilter(ctx, &req.GetVariablesReq{
		ProjectID: projectID,
		Type:      entity.ProjectVariableType,
	})

	return s.dao.FindVariables(ctx, filter, dao.MongoFindOptionWithSortByIDAsc)
}

func (s *Service) GetVariables(ctx context.Context, getReq *req.GetVariablesReq, hiddenVal bool) ([]*resp.Variable, int, error) {
	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit

	filter := s.getVariableFilter(ctx, getReq)

	variables, err := s.dao.FindVariables(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := s.CountVariable(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*resp.Variable, 0)

	err = deepcopy.Copy(&variables).To(&res)
	if err != nil {
		return nil, 0, errors.Wrap(errcode.InternalError, err.Error())
	}

	userIDs := make([]string, 0)

	for idx, variable := range res {
		userIDs = append(userIDs, variables[idx].OwnerID, variables[idx].EditorID)

		if hiddenVal {
			variable.Value = ""
		}
	}

	usersInfo, err := s.GetUsersInfo(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	for idx, variable := range res {
		variable.Owner = usersInfo[variables[idx].OwnerID]
		variable.Editor = usersInfo[variables[idx].EditorID]
	}

	return res, count, nil
}

func (s *Service) CheckVariableKeyUnique(ctx context.Context, getReq *req.GetVariablesReq) (bool, error) {
	filter := s.getVariableFilter(ctx, getReq)

	count, err := s.CountVariable(ctx, filter)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (s *Service) CountVariable(ctx context.Context, filter bson.M) (int, error) {
	return s.dao.CountVariable(ctx, filter)
}

func (s *Service) UpdateSingleVariable(ctx context.Context, variableID string, updateReq *req.UpdateVariableReq) error {
	changeMap := make(map[string]interface{})
	changeMap["update_time"] = time.Now()
	changeMap["editor_id"] = updateReq.OperatorID

	if updateReq.Key != "" {
		changeMap["key"] = updateReq.Key
	}

	if updateReq.Value != "" {
		changeMap["value"] = updateReq.Value
	}

	return s.dao.UpdateSingleVariableByID(ctx, variableID, bson.M{
		"$set": changeMap,
	})
}

func (s *Service) CreateSingleVariable(ctx context.Context, createReq *req.CreateVariableReq) (*entity.Variable, error) {
	now := time.Now()

	variable := &entity.Variable{
		ID:         primitive.NewObjectID(),
		Key:        createReq.Key,
		Value:      createReq.Value,
		Type:       createReq.Type,
		OwnerID:    createReq.OperatorID,
		EditorID:   createReq.OperatorID,
		CreateTime: &now,
		UpdateTime: &now,
	}

	if createReq.Type == entity.ProjectVariableType {
		variable.ProjectID = createReq.ProjectID
	}

	err := s.dao.CreateSingleVariable(ctx, variable)
	if err != nil {
		return nil, err
	}

	return variable, nil
}

func (s *Service) DeleteSingleVariableByID(ctx context.Context, variableID,
	operatorID string) error {
	return s.dao.DeleteSingleVariableByID(ctx, variableID, operatorID)
}

func (s *Service) FindSingleVariableByID(ctx context.Context, id string) (*entity.Variable, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	return s.GetVariableByObjectID(ctx, objectID)
}

func (s *Service) GetVariableByObjectID(ctx context.Context, objectID primitive.ObjectID) (*entity.Variable, error) {
	return s.dao.FindSingleVariable(ctx, bson.M{"_id": objectID})
}
