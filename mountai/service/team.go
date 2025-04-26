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
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Service) CreateTeam(ctx context.Context, createReq *req.CreateTeamReq) (*resp.TeamDetailResp, error) {
	now := time.Now()
	team := &entity.Team{
		ID:         primitive.NewObjectID(),
		CreateTime: &now,
		UpdateTime: &now,
		SentrySlug: createReq.Label,
	}
	err := deepcopy.Copy(createReq).To(team)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	// _, err = s.CreateSentryTeam(ctx, &req.CreateSentryTeamReq{
	// 	TeamName: team.SentrySlug,
	// 	TeamSlug: team.SentrySlug,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	_, err = s.dao.CreateSingleTeam(ctx, team)
	if err != nil {
		return nil, err
	}

	res := new(resp.TeamDetailResp)
	err = deepcopy.Copy(team).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

func (s *Service) generateUpdateTeamMap(updateReq *req.UpdateTeamReq) map[string]interface{} {
	change := make(map[string]interface{})
	if updateReq.AliAlarmName != "" {
		change["ali_alarm_name"] = updateReq.AliAlarmName
	}
	if updateReq.DingHook != "" {
		change["ding_hook"] = updateReq.DingHook
	}
	if updateReq.Label != "" {
		change["label"] = updateReq.Label
	}
	if updateReq.ExtraDingHooks != nil {
		change["extra_ding_hooks"] = *updateReq.ExtraDingHooks
	}
	change["update_time"] = time.Now()
	return change
}

func (s *Service) UpdateTeam(ctx context.Context, id string, updateReq *req.UpdateTeamReq) error {
	// 生成更新map
	changeMap := s.generateUpdateTeamMap(updateReq)

	err := s.dao.UpdateSingleTeam(ctx, id, bson.M{
		"$set": changeMap,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteSingleTeam(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	err = s.dao.DeleteSingleTeam(ctx, bson.M{
		"_id": objectID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetTeamDetail(ctx context.Context, id string) (*resp.TeamDetailResp, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	team, err := s.GetTeamByObjectID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	res := new(resp.TeamDetailResp)
	err = deepcopy.Copy(team).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

func (s *Service) GetTeamByObjectID(ctx context.Context, objectID primitive.ObjectID) (*entity.Team, error) {
	return s.dao.FindSingleTeam(ctx, bson.M{"_id": objectID})
}

func (s *Service) getTeamsFilter(_ context.Context, getReq *req.GetTeamsReq) (bson.M, error) {
	filter := bson.M{}

	if getReq.Keyword != "" {
		if getReq.KeywordField != "" {
			keyword := getReq.Keyword
			filter["$or"] = bson.A{
				bson.M{
					"name": bson.M{
						"$regex": keyword,
					},
				},
				bson.M{
					"label": bson.M{
						"$regex": keyword,
					},
				},
			}
		} else {
			filter[getReq.KeywordField] = bson.M{
				"$regex": getReq.Keyword,
			}
		}
	}
	if getReq.Label != "" {
		filter["label"] = getReq.Label
	}

	if len(getReq.IDs) > 0 {
		objectIDs := make([]primitive.ObjectID, len(getReq.IDs))
		for i, id := range getReq.IDs {
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				return nil, errors.Wrap(_errcode.InvalidHexStringError, err.Error())
			}
			objectIDs[i] = objectID
		}
		filter["_id"] = bson.M{
			"$in": objectIDs,
		}
	}

	return filter, nil
}

func (s *Service) GetTeams(ctx context.Context, getReq *req.GetTeamsReq) ([]*resp.TeamListResp, error) {
	filter, err := s.getTeamsFilter(ctx, getReq)
	if err != nil {
		return nil, err
	}

	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit
	teams, err := s.dao.FindTeams(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})
	if err != nil {
		return nil, err
	}

	res := make([]*resp.TeamListResp, 0)
	err = deepcopy.Copy(&teams).To(&res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

func (s *Service) GetTeamsCount(ctx context.Context, getReq *req.GetTeamsReq) (int, error) {
	filter, err := s.getTeamsFilter(ctx, getReq)
	if err != nil {
		return 0, err
	}

	res, err := s.dao.CountTeam(ctx, filter)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (s *Service) GetTeamsByIDs(ctx context.Context, ids []string) (map[string]*resp.TeamDetailResp, error) {
	objectIDs := make([]primitive.ObjectID, len(ids))
	for idx := range ids {
		objectID, err := primitive.ObjectIDFromHex(ids[idx])
		if err != nil {
			return nil, errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
		}

		objectIDs[idx] = objectID
	}

	teams, err := s.dao.FindTeams(
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

	res := make(map[string]*resp.TeamDetailResp)
	for _, team := range teams {
		item := new(resp.TeamDetailResp)
		err = deepcopy.Copy(team).To(&item)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}
		res[item.ID] = item
	}

	return res, nil
}
