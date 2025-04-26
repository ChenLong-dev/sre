package dao

import (
	"rulai/models/entity"

	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDao_DingTalkMessageRecord(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		now := time.Now()
		_, err := d.CreateDingTalkMessageRecord(context.Background(), &entity.DingTalkMessageRecord{
			ID:      primitive.NewObjectID(),
			UserIDs: []string{"171"},
			Content: entity.DingTalkMessageContent{
				ActionType: entity.SubscribeActionDeploy,
				AppID:      primitive.NewObjectID(),
				ProjectID:  "1107",
				Env:        entity.AppEnvStg,
			},
			TaskID:     123123,
			CreateTime: &now,
			UpdateTime: &now,
		})
		assert.Nil(t, err)
	})
}
