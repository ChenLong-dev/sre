package service

import (
	"rulai/config"
	"rulai/models/entity"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func (s *Service) getApolloConfig(ctx context.Context, appID, cluster, namespace string,
	env entity.AppEnvName) (res map[string]string, err error) {
	host := config.Conf.Apollo.StgHost
	if env == entity.AppEnvPrd {
		host = config.Conf.Apollo.PrdHost
	}

	configURL := fmt.Sprintf("%s/configfiles/json/%s/%s/%s", host, appID, cluster, namespace)
	res = make(map[string]string)

	err = s.httpClient.Builder().
		URL(configURL).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(&res)
	if err != nil {
		return nil, errors.Wrap(_errcode.ApolloInternalError, err.Error())
	}

	return res, nil
}
