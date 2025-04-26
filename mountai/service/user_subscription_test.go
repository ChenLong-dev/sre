package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/utils"

	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
)

func TestService_UserSubscription(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		err := s.UserSubscribe(context.Background(), &req.UserSubscribeReq{
			EnvName: entity.AppEnvFat,
			Action:  entity.SubscribeActionDeploy,
			AppID:   "6040b1e53dec54da3bf93423",
		}, "189")

		assert.Nil(t, err)

		err = s.UserUnsubscribe(context.Background(), &req.UserUnsubscribeReq{
			EnvName: entity.AppEnvFat,
			Action:  entity.SubscribeActionDeploy,
			AppID:   "6040b1e53dec54da3bf93423",
		}, "189")

		assert.Nil(t, err)
	})
}

func TestHandleAppOpMsgEvent(t *testing.T) {
	t.Run("AppOpMsgProducer", func(t *testing.T) {
		msgVal := entity.SubscribeEventMsg{
			ActionType: entity.SubscribeActionDeploy,
			AppID:      "5f3b424f01642a88c74a9f83",
			OperatorID: "265",
			TaskID:     "5f34f075a28aabd67c8acd87",
			Env:        entity.AppEnvName("prd"),
			OpTime:     utils.FormatK8sTime(time.Now()),
		}
		err := s.PublishAppOpEvent(context.Background(), &msgVal)

		assert.Nil(t, err)
	})

	t.Run("AppOpMsgConsumer", func(t *testing.T) {
		msgVal := entity.SubscribeEventMsg{
			ActionType: entity.SubscribeActionDeploy,
			AppID:      "5f3b424f01642a88c74a9f83",
			OperatorID: "265",
			TaskID:     "60613c9ec1fefb9925643b86",
			Env:        entity.AppEnvName("prd"),
			OpTime:     utils.FormatK8sTime(time.Now()),
		}
		msgByte, err := json.Marshal(msgVal)

		assert.Nil(t, err)
		msg := sarama.ConsumerMessage{
			Value: msgByte,
		}
		err = s.HandleAppOpMsgEvent(context.Background(), &msg)

		assert.Nil(t, err)
	})
}
