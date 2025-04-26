package service

import (
	"rulai/models"
	"rulai/models/req"

	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/log"
)

func TestService_GetImages(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		res, err := s.GetImageJobs(context.Background(), &req.GetImageJobsReq{
			BaseListRequest: models.BaseListRequest{
				Limit: 2,
				Page:  1,
			},
			ProjectName: "app-framework",
		})
		assert.Nil(t, err)
		assert.True(t, len(res) > 0)
		for _, item := range res {
			log.Info("%#v", item)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		res, err := s.GetImageJobs(context.Background(), &req.GetImageJobsReq{
			BaseListRequest: models.BaseListRequest{
				Limit: 10000,
				Page:  10,
			},
			ProjectName: "app-framework",
		})
		assert.Nil(t, err)
		assert.True(t, len(res) == 0)
	})
}

func TestService_GetImagesCount(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		res, err := s.GetImageJobsCount(context.Background(), &req.GetImageJobsReq{
			BaseListRequest: models.BaseListRequest{
				Limit: 10,
				Page:  1,
			},
			ProjectName: "app-framework",
		})
		assert.Nil(t, err)
		assert.True(t, res > 0)
	})
}

func TestService_GetImageDetail(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		res, err := s.GetImageJobDetail(context.Background(), &req.GetImageJobDetailReq{
			BuildID:     "1",
			ProjectName: "app-framework",
		})
		assert.Nil(t, err)
		assert.Equal(t, "1", res.BuildID)
		assert.Equal(t, "app-framework", res.JobName)
		assert.NotEqual(t, "", res.ConsoleOutput)
	})
}

func TestService_getImageVersion(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		version := s.getImageVersion("project", s.getImageTag("master", "abcdef", ""))
		assert.Equal(t, "project:abcdef-master", version)
	})

	t.Run("line", func(t *testing.T) {
		version := s.getImageVersion("project", s.getImageTag("feat-76", "abcdef", ""))
		assert.Equal(t, "project:abcdef-feat-76", version)
	})

	t.Run("underline", func(t *testing.T) {
		version := s.getImageVersion("project", s.getImageTag("feat_76", "abcdef", ""))
		assert.Equal(t, "project:abcdef-feat_76", version)
	})

	t.Run("slash", func(t *testing.T) {
		version := s.getImageVersion("project", s.getImageTag("feat/76", "abcdef", ""))
		assert.Equal(t, "project:abcdef-feat_76", version)
	})

	t.Run("hash", func(t *testing.T) {
		version := s.getImageVersion("project", s.getImageTag("feat#76", "abcdef", ""))
		assert.Equal(t, "project:abcdef-feat_76", version)
	})

	t.Run("with ArgHash", func(t *testing.T) {
		version := s.getImageVersion("project", s.getImageTag("feat#76", "abcdef", "ESAD"))
		assert.Equal(t, "project:abcdef-feat_76-ESAD", version)
	})
}
