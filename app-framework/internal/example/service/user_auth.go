package service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/config"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/resp"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/httpclient"
	"net/http"
)

// 用户鉴权
func (s *Service) PostUserAuthVerify(ctx context.Context, verifyReq *req.UserAuthVerifyReq) (*resp.UserAuthVerifyDetailResp, error) {
	result := new(resp.UserAuthVerifyResp)

	// 调用http客户端获取数据
	err := s.httpClient.Builder().
		Method(http.MethodPost).
		URL(fmt.Sprintf("%s/v1/tokens/verify", config.Conf.Host.UserAuth)).
		Headers(httpclient.GetDefaultHeader()).
		JsonBody(verifyReq).
		// ==========================
		// 降级后的响应
		// ==========================
		DegradedJsonResponse(resp.UserAuthVerifyResp{
			Code:    0,
			Message: "success",
			Data: &resp.UserAuthVerifyDetailResp{
				QingTingID: verifyReq.QingTingID,
				Verify:     resp.UserAuthVerifyDeny,
			},
		}).
		Fetch(ctx).
		DecodeJSON(&result)
	if err != nil {
		return nil, errors.Wrapf(errcode.InternalError, "%s", err)
	}

	return result.Data, nil
}
