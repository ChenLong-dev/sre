package service

import (
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"

	"context"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MaskSecretStr = "*****"

func (s *Service) GetImageArgsTemplatesByIDs(ctx context.Context, ids []string) (map[string]*resp.ImageArgsTemplate, error) {
	objectIDs := make([]primitive.ObjectID, len(ids))
	for idx := range ids {
		objectID, err := primitive.ObjectIDFromHex(ids[idx])
		if err != nil {
			return nil, errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
		}

		objectIDs[idx] = objectID
	}

	templates, err := s.dao.FindImageArgsTemplates(
		ctx,
		bson.M{
			"_id": bson.M{
				"$in": objectIDs,
			},
		},
		dao.MongoFindOptionWithSortByIDAsc,
	)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*resp.ImageArgsTemplate)
	for _, template := range templates {
		item := new(resp.ImageArgsTemplate)
		err = deepcopy.Copy(template).To(&item)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}
		res[item.ID] = item
	}

	return res, nil
}

func (s *Service) GetImageArgsTemplates(ctx context.Context, getReq *req.GetImageArgsTemplateReq) (
	[]*resp.ImageArgsTemplateDetail, int, error) {
	filter := s.getImageArgsTemplateFilter(ctx, getReq)

	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit

	imageArgsTemplates, err := s.dao.FindImageArgsTemplates(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := s.dao.CountImageArgsTemplate(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*resp.ImageArgsTemplateDetail, 0)

	err = deepcopy.Copy(&imageArgsTemplates).To(&res)
	if err != nil {
		return nil, 0, errors.Wrap(errcode.InternalError, err.Error())
	}

	userIDs := make([]string, 0)
	teamIDs := make([]string, 0)

	for idx := range res {
		userIDs = append(userIDs, imageArgsTemplates[idx].OwnerID)
		teamIDs = append(teamIDs, imageArgsTemplates[idx].TeamID)
	}

	usersInfo, err := s.GetUsersInfo(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	teamsInfo, err := s.GetTeamsByIDs(ctx, teamIDs)
	if err != nil {
		return nil, 0, err
	}

	for idx, imageArgsTemplate := range res {
		imageArgsTemplate.Owner = usersInfo[imageArgsTemplates[idx].OwnerID]
		imageArgsTemplate.Team = teamsInfo[imageArgsTemplates[idx].TeamID]
	}

	return res, count, nil
}

func (s *Service) getImageArgsTemplateFilter(_ context.Context, getReq *req.GetImageArgsTemplateReq) bson.M {
	filter := bson.M{}

	if getReq.TeamID != "" {
		filter["team_id"] = getReq.TeamID
	}

	if getReq.Name != "" {
		filter["name"] = getReq.Name
	}

	return filter
}

func (s *Service) CheckImageArgsTemplateNameUnique(ctx context.Context, getReq *req.GetImageArgsTemplateReq) (bool, error) {
	filter := s.getImageArgsTemplateFilter(ctx, getReq)

	count, err := s.dao.CountImageArgsTemplate(ctx, filter)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (s *Service) UpdateSingleImageArgsTemplate(ctx context.Context, id string, updateReq *req.UpdateImageArgsTemplateReq) error {
	changeMap := make(map[string]interface{})
	changeMap["update_time"] = time.Now()

	if updateReq.Name != "" {
		changeMap["name"] = updateReq.Name
	}

	if updateReq.Content != "" {
		changeMap["content"] = updateReq.Content
	}

	return s.dao.UpdateSingleImageArgsTemplate(ctx, id, bson.M{
		"$set": changeMap,
	})
}

func (s *Service) CreateSingleImageArgsTemplate(ctx context.Context, createReq *req.CreateImageArgsTemplateReq) error {
	now := time.Now()

	imageArgsTemplate := &entity.ImageArgsTemplate{
		ID:         primitive.NewObjectID(),
		Name:       createReq.Name,
		Content:    createReq.Content,
		TeamID:     createReq.TeamID,
		OwnerID:    createReq.OperatorID,
		CreateTime: &now,
		UpdateTime: &now,
	}

	return s.dao.CreateSingleImageArgsTemplate(ctx, imageArgsTemplate)
}

func (s *Service) DeleteSingleImageArgsTemplateByID(ctx context.Context, id string) error {
	return s.dao.DeleteSingleImageArgsTemplateByID(ctx, id)
}

func (s *Service) FindSingleImageArgsTemplateByID(ctx context.Context, id string) (*entity.ImageArgsTemplate, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	return s.GetImageArgsTemplateByObjectID(ctx, objectID)
}

func (s *Service) GetImageArgsTemplateByObjectID(ctx context.Context, objectID primitive.ObjectID) (*entity.ImageArgsTemplate, error) {
	return s.dao.FindSingleImageArgsTemplate(ctx, bson.M{"_id": objectID})
}

func (s *Service) RenderImageArgsTemplate(ctx context.Context, id, projectID string) (imageArg, imageArgWithMask string, err error) {
	imageArgsTemplate, err := s.FindSingleImageArgsTemplateByID(ctx, id)
	if err != nil {
		return
	}

	variables, err := s.GetProjectVariables(ctx, projectID)
	if err != nil {
		return
	}

	variableMap := make(map[string]string)
	for _, variable := range variables {
		variableMap[variable.Key] = variable.Value
	}

	imageArg, imageArgWithMask = s.renderImageArgsTemplate(ctx, imageArgsTemplate.Content, variableMap)

	return
}

func (s *Service) renderImageArgsTemplate(_ context.Context, templateContent string, variables map[string]string) (
	imageArg, imageArgWithMask string) {
	imageArg = templateContent
	imageArgWithMask = templateContent

	keyRegexp := regexp.MustCompile(`\${(\w+)}`)
	for _, submatch := range keyRegexp.FindAllStringSubmatch(templateContent, -1) {
		replaceStr := submatch[0]
		key := submatch[1]

		if value, ok := variables[key]; ok {
			imageArgWithMask = strings.ReplaceAll(imageArgWithMask, replaceStr, MaskSecretStr)
			imageArg = strings.ReplaceAll(imageArg, replaceStr, value)
		}
	}

	return
}
