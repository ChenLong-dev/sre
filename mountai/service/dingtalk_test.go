package service

import (
	"rulai/config"
	"rulai/models/req"

	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendRobotMsgToDingTalkByGroup(t *testing.T) {
	t.Run("TestDingDingDingMsg", func(t *testing.T) {
		res, err := s.SendRobotMsgToDingTalk(context.Background(), &req.SendRobotMessageReq{
			Token: config.Conf.DingTalk.GroupTokens[p0DeployDingToken],
			Message: req.RobotMarkdownMessageReq{
				DingTalkMarkdownMessageReq: req.DingTalkMarkdownMessageReq{
					Msgtype: req.DingTalkMsgTypeMarkdown,
					Markdown: req.DingTalkMarkdownMessageBody{
						Title: "发布test deploy",
						Text:  "发布测试BFF",
					},
				},
				At: req.RobotMarkdownMessageAt{
					IsAtAll: true,
				}},
		})
		assert.Nil(t, err)
		assert.Equal(t, "", res.ErrMsg)
	})
}
