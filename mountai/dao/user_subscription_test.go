package dao

import (
	"rulai/models/entity"

	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestDao_UserSubscription(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		now := time.Now()
		appID, _ := primitive.ObjectIDFromHex("6048a5712097e742e290a90")
		upsert := true
		action := entity.SubscribeActionDeploy
		env := entity.AppEnvPrd
		userID := "2"

		err := d.UpdateUserSubscription(context.Background(), bson.M{
			"user_id":     userID,
			"app_id":      appID,
			"action_type": action,
			"env":         env,
		}, bson.M{
			"$set": bson.M{
				"user_id":      userID,
				"app_id":       appID,
				"action_type":  action,
				"env":          env,
				"updated_time": now,
			},
			"$setOnInsert": bson.M{"create_time": now},
		}, &options.UpdateOptions{Upsert: &upsert})
		assert.Nil(t, err)

		var userSub *entity.UserSubscription

		userSub, err = d.FindUserSubscription(context.Background(), bson.M{
			"user_id": userID,
		})
		assert.Nil(t, err)
		assert.Equal(t, userID, userSub.UserID)

		err = d.DeleteUserSubscription(context.Background(), bson.M{
			"user_id":     userID,
			"app_id":      appID,
			"action_type": action,
			"env":         env,
		})
		assert.Nil(t, err)
	})
}
