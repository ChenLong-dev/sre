package dao

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/entity"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/queue"
)

// 发送退款到AMQP
func (d *Dao) SendVipRefundNoticeToAMQP(ctx context.Context, notice *entity.VipRefundNotice) error {
	data, err := json.Marshal(notice)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return errors.Wrapf(errcode.InternalError, "%s", err)
	}

	// ==========================
	// 发送AMQP消息
	// ==========================
	err = d.AMQPSession.Push(ctx, &amqp.Publishing{
		ContentType: "text/plain",
		Body:        data,
	}, queue.PushOption{})
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return errors.Wrap(errcode.InternalError, err.Error())
	}

	return nil
}
