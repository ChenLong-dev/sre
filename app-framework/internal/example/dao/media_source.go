package dao

import (
	"context"
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/entity"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 发送创建消息到kafka
func (d *Dao) SendMediaSourceToKafka(ctx context.Context, source *entity.MediaSourceNotice) (*sarama.ProducerMessage, error) {
	data, err := json.Marshal(source)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, errors.Wrapf(errcode.InternalError, "%s", err)
	}

	// ==========================
	// 发送kafka消息
	// ==========================
	message := &sarama.ProducerMessage{
		Topic: "media_resource_staging",
		Value: sarama.StringEncoder(data),
	}
	partition, offset, err := d.KafkaProducer.SendMessage(message)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, errors.Wrapf(errcode.InternalError,
			"push message to media_resource_staging error: partition=%d offset=%d error=%s",
			partition, offset, err)
	}

	return message, nil
}
