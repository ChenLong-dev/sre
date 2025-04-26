package service

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_PrintAMQPMessageBody(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		err := s.PrintAMQPMessageBody(context.Background(), []byte("this is a test"))
		assert.Nil(t, err)
	})
}

func TestService_PrintKafkaMessageBody(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		err := s.PrintKafkaMessageBody(context.Background(), &sarama.ConsumerMessage{
			Value:     nil,
			Topic:     "example",
			Partition: 0,
			Offset:    0,
		})
		assert.Nil(t, err)
	})
}
