package service

import (
	"rulai/models/entity"
	"rulai/models/req"

	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_GetAppConfig(t *testing.T) {
	t.Run("yaml", func(t *testing.T) {
		data, err := s.GetAppConfig(context.Background(), &req.GetConfigManagerFileReq{
			ProjectID:  "1449",
			EnvName:    entity.AppEnvStg,
			CommitID:   "c91da6fae5653c09b84e46d5438de3be234f1057",
			IsDecrypt:  true,
			FormatType: req.ConfigManagerFormatTypeYaml,
		})
		assert.Nil(t, err)
		fmt.Printf("%#v\n", data)
		assert.Contains(t, data.Config, "rpc")
	})
	t.Run("json", func(t *testing.T) {
		data, err := s.GetAppConfig(context.Background(), &req.GetConfigManagerFileReq{
			ProjectID:  "1449",
			EnvName:    entity.AppEnvStg,
			CommitID:   "c91da6fae5653c09b84e46d5438de3be234f1057",
			IsDecrypt:  true,
			FormatType: req.ConfigManagerFormatTypeJSON,
		})
		assert.Nil(t, err)
		fmt.Printf("%#v\n", data)
		assert.Equal(t, "abcdefg", data.Config.(map[string]interface{})["config.txt"])
	})
}

func TestService_GetProjectResourceFromConfig(t *testing.T) {
	t.Run("app-framework", func(t *testing.T) {
		project, err := s.GetProjectDetail(context.Background(), "1449")
		assert.Nil(t, err)

		res, err := s.GetProjectResourceFromConfig(context.Background(), &req.GetProjectResourceFromConfigReq{
			EnvName: "prd",
		}, project)
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("user-system", func(t *testing.T) {
		project, err := s.GetProjectDetail(context.Background(), "1319")
		assert.Nil(t, err)

		res, err := s.GetProjectResourceFromConfig(context.Background(), &req.GetProjectResourceFromConfigReq{
			EnvName: "prd",
		}, project)
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("pop-api", func(t *testing.T) {
		project, err := s.GetProjectDetail(context.Background(), "1113")
		assert.Nil(t, err)

		res, err := s.GetProjectResourceFromConfig(context.Background(), &req.GetProjectResourceFromConfigReq{
			EnvName: "prd",
		}, project)
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("seo-admin", func(t *testing.T) {
		project, err := s.GetProjectDetail(context.Background(), "1470")
		assert.Nil(t, err)

		res, err := s.GetProjectResourceFromConfig(context.Background(), &req.GetProjectResourceFromConfigReq{
			EnvName: "prd",
		}, project)
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})
}
