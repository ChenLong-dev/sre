package service

import (
	"rulai/models/entity"
	"rulai/models/resp"

	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestService_SaveAuthToken(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		_, err := s.SaveUserInfo(context.Background(), &resp.GitUserProfileResp{
			ID:       1,
			Name:     "test",
			UserName: "test",
			State:    "active",
			Email:    "test",
		}, "token")
		assert.Nil(t, err)

		auth, err := s.dao.FindSingleUserAuth(context.Background(), bson.M{
			"_id": "1",
		})

		assert.Nil(t, err)
		assert.Equal(t, "token", auth.Token)

		err = s.dao.DeleteSingleUserAuth(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)
	})

	t.Run("update", func(t *testing.T) {
		err := s.dao.CreateSingleUserAuth(context.Background(), &entity.UserAuth{
			ID:    "1",
			Token: "old",
		})
		assert.Nil(t, err)

		_, err = s.SaveUserInfo(context.Background(), &resp.GitUserProfileResp{
			ID:       1,
			Name:     "test",
			UserName: "test",
			State:    "active",
			Email:    "test",
		}, "token")
		assert.Nil(t, err)

		auth, err := s.dao.FindSingleUserAuth(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)
		assert.Equal(t, "token", auth.Token)

		err = s.dao.DeleteSingleUserAuth(context.Background(), bson.M{
			"_id": "1",
		})
		assert.Nil(t, err)
	})
}
