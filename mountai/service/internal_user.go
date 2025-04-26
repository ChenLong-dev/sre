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

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/httpclient"
)

// GetInternalUsers gets internal users.
func (s *Service) GetInternalUsers(ctx context.Context, getReq *req.GetInternalUsersReq) ([]*entity.InternalUser, error) {
	res := new(resp.GetInternalUsersResp)
	query := httpclient.NewUrlValue()
	if getReq.Email != "" {
		query.Add("email", url.PathEscape(getReq.Email))
	}
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/v3/admin/users", config.Conf.Other.InternalUserHost)).
		QueryParams(query).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.InternalUserInternalError, err.Error())
	}
	if res.ErrCode != 0 {
		return nil, errors.Wrap(_errcode.InternalUserInternalError, res.ErrMsg)
	}

	return res.Data.InternalUsersItems, nil
}

// GetInternalSingleUser gets single internal user.
func (s *Service) GetInternalSingleUser(ctx context.Context, getReq *req.GetInternalUsersReq) (*entity.InternalUser, error) {
	users, err := s.GetInternalUsers(ctx, getReq)
	if err != nil {
		return nil, err
	}

	if len(users) < 1 {
		return nil, _errcode.NoRequiredInternalUserError
	}

	return users[0], nil
}
