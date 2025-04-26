package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/library/base/null"
	"testing"
)

func TestService_GetAvailableVipItemList(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		list, err := s.GetAvailableVipItemList(context.Background(), &req.GetVipItemListReq{})
		assert.Nil(t, err)
		assert.Equal(t, true, len(list) > 0)
	})

	t.Run("item_type", func(t *testing.T) {
		list, err := s.GetAvailableVipItemList(context.Background(), &req.GetVipItemListReq{
			ItemType: null.StringFrom("vip"),
		})
		assert.Nil(t, err)
		assert.Equal(t, true, len(list) > 0)
	})

	t.Run("include", func(t *testing.T) {
		list, err := s.GetAvailableVipItemList(context.Background(), &req.GetVipItemListReq{
			Include: null.StringFrom("autorenew"),
		})
		assert.Nil(t, err)
		assert.Equal(t, true, len(list) > 0)
	})

	t.Run("phone_type", func(t *testing.T) {
		list, err := s.GetAvailableVipItemList(context.Background(), &req.GetVipItemListReq{
			PhoneType: null.StringFrom("ios"),
		})
		assert.Nil(t, err)
		assert.Equal(t, true, len(list) > 0)
	})
}
