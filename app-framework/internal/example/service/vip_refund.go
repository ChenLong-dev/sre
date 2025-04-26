package service

import (
	"context"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/entity"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 创建媒体资源
func (s *Service) CreateRefundNotice(ctx context.Context, createReq *req.CreateVipRefundReq) error {
	// ==========================
	// 将请求模型复制到消息实体中，通过Library中的deepcopy包实现
	// ==========================
	detail := new(entity.VipRefundNotice)
	err := deepcopy.Copy(createReq).To(detail)
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return errors.Wrapf(errcode.InternalError, "%s", err)
	}

	// ==========================
	// 发送amqp消息
	// ==========================
	err = s.dao.SendVipRefundNoticeToAMQP(ctx, detail)
	if err != nil {
		return err
	}

	return nil
}
