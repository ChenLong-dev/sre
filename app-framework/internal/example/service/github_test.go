package service

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/app-framework/internal/example/config"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"testing"
)

func TestService_GetOwnerGithubDetail(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		result, err := s.GetOwnerGithubDetail(context.Background(), "CMonoceros")
		assert.Nil(t, err)
		assert.Equal(t, "CMonoceros", result.Login)
	})

	t.Run("unknown_user", func(t *testing.T) {
		_, err := s.GetOwnerGithubDetail(context.Background(), uuid.NewV4().String())
		assert.Equal(t, true, errcode.EqualError(errcode.InternalError, err))
	})
}

func TestService_GetOwnerGithubReposList(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		result, err := s.GetOwnerGithubReposList(context.Background(), &req.GetGithubRepositoryListReq{
			Page:  null.IntFrom(1),
			Limit: null.IntFrom(2),
			Owner: "CMonoceros",
		})
		assert.Nil(t, err)
		assert.Equal(t, 2, len(result))
	})

	t.Run("unknown_user", func(t *testing.T) {
		_, err := s.GetOwnerGithubReposList(context.Background(), &req.GetGithubRepositoryListReq{
			Page:  null.IntFrom(1),
			Limit: null.IntFrom(2),
			Owner: uuid.NewV4().String(),
		})
		assert.Equal(t, true, errcode.EqualError(errcode.InternalError, err))
	})
}

func TestService_GetOwnerGithubAggregationResp(t *testing.T) {
	goroutine.Init(config.Conf.Goroutine)

	t.Run("normal", func(t *testing.T) {
		result, err := s.GetOwnerGithubAggregationResp(context.Background(), "CMonoceros")
		assert.Nil(t, err)
		assert.Equal(t, "CMonoceros", result.Owner.Login)
		assert.Equal(t, 5, len(result.Repos))
	})

	t.Run("unknown_user", func(t *testing.T) {
		_, err := s.GetOwnerGithubAggregationResp(context.Background(), uuid.NewV4().String())
		assert.Equal(t, true, errcode.EqualError(errcode.InternalError, err))
	})
}
