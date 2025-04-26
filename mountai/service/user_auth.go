package service

import (
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/pkg/errors"

	"context"
	"strconv"
	"time"

	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
)

// 保存用户信息
func (s *Service) SaveUserInfo(ctx context.Context, user *resp.GitUserProfileResp, token string) (*resp.UserProfileResp, error) {
	_, err := s.dao.FindSingleUserAuth(ctx, bson.M{
		"_id": strconv.Itoa(user.ID),
	})
	now := time.Now()
	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		err = s.dao.CreateSingleUserAuth(ctx, &entity.UserAuth{
			ID:         strconv.Itoa(user.ID),
			Name:       user.Name,
			AvatarURL:  user.AvatarURL,
			Email:      user.Email,
			Token:      token,
			CreateTime: &now,
			UpdateTime: &now,
		})
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		err = s.dao.UpdateSingleUserAuth(ctx, strconv.Itoa(user.ID), bson.M{
			"$set": bson.M{
				"token":       token,
				"name":        user.Name,
				"avatar_url":  user.AvatarURL,
				"email":       user.Email,
				"update_time": now,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return &resp.UserProfileResp{
		ID:        strconv.Itoa(user.ID),
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Email:     user.Email,
		Token:     token,
	}, nil
}

func (s *Service) GetUserInfo(ctx context.Context, userID string) (*resp.UserProfileResp, error) {
	userInfo, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &resp.UserProfileResp{
		ID:        userInfo.ID,
		Name:      userInfo.Name,
		AvatarURL: userInfo.AvatarURL,
		Email:     userInfo.Email,
	}, nil
}

func (s *Service) GetUserByID(ctx context.Context, id string) (*entity.UserAuth, error) {
	return s.dao.FindSingleUserAuth(ctx, bson.M{"_id": id})
}

// 是否有权限
func (s *Service) ValidateHasPermission(ctx context.Context, validateReq *req.ValidateHasPermissionReq) error {
	// k8s系统用户拥有权限
	if validateReq.OperatorID == entity.K8sSystemUserID {
		return nil
	}

	// 先做不需要远程调用的校验
	memberID, err := strconv.Atoi(validateReq.OperatorID)
	if err != nil {
		return errors.Wrap(errcode.InvalidParams, "operator id is invalid")
	}

	members, err := s.GetGitlabProjectAllActiveMembers(ctx, validateReq.ProjectID)
	if err != nil {
		return err
	}

	var member *resp.GitProjectMemberResp
	for _, gitMember := range members {
		if gitMember.ID == memberID {
			member = gitMember
			break
		}
	}

	if member == nil {
		return errors.Wrapf(_errcode.GitlabUserNoPermissionError, "operator(%d) is not a valid git member", memberID)
	}

	// 删除project, 删除app，更改app名权限校验
	switch validateReq.OperateType {
	case entity.OperateTypeDeleteProject, entity.OperateTypeDeleteApp, entity.OperateTypeCorrectAppName,
		entity.OperateTypeReadVariableValue, entity.OperateTypeUpdateVariableValue,
		entity.OperateTypeCreateVariableValue, entity.OperateTypeDeleteVariableValue, entity.OperateTypeDeleteJob:
		if member.AccessLevel == entity.GitMemberAccessOwner || member.AccessLevel == entity.GitMemberAccessMaintainer {
			return nil
		}

		// 设置 Kong 权重权限校验
	case entity.OperateTypeSetAppClusterKongWeights:
		if member.AccessLevel == entity.GitMemberAccessOwner || member.AccessLevel == entity.GitMemberAccessMaintainer {
			return nil
		}

		// developer在stg和fat上有调整Kong权重的权限
		if member.AccessLevel == entity.GitMemberAccessDeveloper &&
			(validateReq.CreateTaskEnvName == entity.AppEnvFat || validateReq.CreateTaskEnvName == entity.AppEnvStg) {
			return nil
		}

	case entity.OperateTypeCreateTask:
		if member.AccessLevel == entity.GitMemberAccessOwner || member.AccessLevel == entity.GitMemberAccessMaintainer {
			return nil
		}

		// developer在prd和pre上有restart权限
		if member.AccessLevel == entity.GitMemberAccessDeveloper &&
			(validateReq.CreateTaskEnvName == entity.AppEnvPrd || validateReq.CreateTaskEnvName == entity.AppEnvPre) &&
			validateReq.CreateTaskAction == entity.TaskActionRestart {
			return nil
		}
		// developer在stg和fat上有所有权限
		if member.AccessLevel == entity.GitMemberAccessDeveloper &&
			(validateReq.CreateTaskEnvName == entity.AppEnvFat || validateReq.CreateTaskEnvName == entity.AppEnvStg) {
			return nil
		}

	default:
	}

	return errors.Wrapf(_errcode.GitlabUserNoPermissionError,
		"operator(%d) with git level(%d) has no permission for operation(%s)", memberID, member.AccessLevel, validateReq.OperateType)
}

func (s *Service) GetUsersInfo(ctx context.Context, userIDs []string) (map[string]*resp.UserProfileResp, error) {
	filter := s.getUsersFilter(ctx, &req.GetUsersReq{UserIDs: userIDs})
	usersInfo, err := s.dao.FindUserAuth(ctx, filter, dao.MongoFindOptionWithSortByIDAsc)
	if err != nil {
		return nil, err
	}

	profiles := make(map[string]*resp.UserProfileResp, len(usersInfo))

	for _, userInfo := range usersInfo {
		profiles[userInfo.ID] = &resp.UserProfileResp{
			ID:        userInfo.ID,
			Name:      userInfo.Name,
			AvatarURL: userInfo.AvatarURL,
			Email:     userInfo.Email,
		}
	}

	return profiles, nil
}

func (s *Service) GetUsersCount(ctx context.Context, getReq *req.GetUsersReq) (int, error) {
	filter := s.getUsersFilter(ctx, getReq)

	res, err := s.dao.CountUser(ctx, filter)
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (s *Service) GetUsers(ctx context.Context, getReq *req.GetUsersReq) ([]*resp.UserProfileResp, error) {
	filter := s.getUsersFilter(ctx, getReq)

	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit

	usersInfo, err := s.dao.FindUserAuth(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})

	if err != nil {
		return nil, err
	}

	profiles := make([]*resp.UserProfileResp, 0)

	for _, userInfo := range usersInfo {
		userProfileResp := &resp.UserProfileResp{
			ID:        userInfo.ID,
			Name:      userInfo.Name,
			AvatarURL: userInfo.AvatarURL,
			Email:     userInfo.Email,
		}

		profiles = append(profiles, userProfileResp)
	}

	return profiles, nil
}

func (s *Service) getUsersFilter(_ context.Context, userReq *req.GetUsersReq) bson.M {
	filter := bson.M{}

	if len(userReq.UserIDs) > 0 {
		filter["_id"] = bson.M{
			"$in": userReq.UserIDs,
		}
	}

	if userReq.Keyword != "" {
		filter["$or"] = bson.A{
			bson.M{
				"email": bson.M{
					"$regex": userReq.Keyword,
				},
			},
			bson.M{
				"name": bson.M{
					"$regex": userReq.Keyword,
				},
			},
		}
	}

	return filter
}

// GetDefaultUserProfile generate default user information
func (s *Service) GetDefaultUserProfile(ctx context.Context, userID string) (*resp.UserProfileResp, error) {
	userInfo, err := s.GetUserInfo(ctx, userID)
	if err != nil {
		if !errcode.EqualError(errcode.NoRowsFoundError, err) {
			return nil, err
		}
		k8sSystemUser, getErr := s.GetUserInfo(ctx, entity.K8sSystemUserID)
		if getErr != nil {
			return nil, getErr
		}
		return k8sSystemUser, nil
	}

	return userInfo, nil
}

// CheckAndSyncUser 检查和同步用户记录
func (s *Service) CheckAndSyncUser(ctx context.Context, userID string) error {
	_, err := s.dao.FindSingleUserAuth(ctx, bson.M{
		"_id": userID,
	})

	if errcode.EqualError(errcode.NoRowsFoundError, err) {
		user, intErr := s.GetGitlabUser(ctx, userID)
		if intErr != nil {
			return intErr
		}

		now := time.Now()
		intErr = s.dao.CreateSingleUserAuth(ctx, &entity.UserAuth{
			ID:         strconv.Itoa(user.ID),
			Name:       user.Name,
			AvatarURL:  user.AvatarURL,
			Email:      user.Email,
			CreateTime: &now,
			UpdateTime: &now,
		})

		if intErr != nil {
			return intErr
		}

		return nil
	}

	return err
}
