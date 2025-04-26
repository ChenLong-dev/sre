package service

import (
	"rulai/models/req"
	_errcode "rulai/utils/errcode"

	"context"
	"testing"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func TestService_GetLogStoreDetail(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		res, err := s.GetLogStoreDetail(context.Background(), &req.AliGetLogStoreDetailReq{
			ProjectName: "k8s-stg",
			StoreName:   "app-framework-example-stg",
		})
		assert.Nil(t, err)
		assert.Equal(t, "app-framework-example-stg", res.LogStoreName)
	})
	t.Run("logStoreNotExist", func(t *testing.T) {
		res, err := s.GetLogStoreDetail(context.Background(), &req.AliGetLogStoreDetailReq{
			ProjectName: "k8s-stg",
			StoreName:   "not-exist",
		})
		assert.True(t, errcode.EqualError(_errcode.AliResourceNotFoundError, err))
		assert.Nil(t, res)
	})
	t.Run("logProjectNotExist", func(t *testing.T) {
		res, err := s.GetLogStoreDetail(context.Background(), &req.AliGetLogStoreDetailReq{
			ProjectName: "not-exist",
			StoreName:   "app-framework-example-stg",
		})
		assert.True(t, true, errcode.EqualError(_errcode.AliResourceNotFoundError, err))
		assert.Nil(t, res)
	})
}

func TestService_CreateAliLogStoreShipper(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		logStore, err := s.GetAliLogStore(context.Background(), &req.AliGetLogStoreDetailReq{
			ProjectName: "k8s-dev1",
			StoreName:   "app-framework",
		})
		assert.Nil(t, err)
		_, err = logStore.GetShipper(s.getAliLogStoreColdStorageShipperName("app-framework"))
		assert.Nil(t, err)

		err = s.createAliLogStoreShipperWhenNotExist(context.Background(), &req.CreateAliLogStoreShipperReq{
			TargetType:      sls.OSSShipperType,
			ProjectName:     "app-framework",
			LogStoreName:    "app-framework",
			LogStoreProject: "k8s-dev1",
			ShipperName:     s.getAliLogStoreColdStorageShipperName("app-framework"),
			CreateOSSShipperReq: &req.CreateAliLogStoreOSSShipperReq{
				ShipperName:    s.getAliLogStoreColdStorageShipperName("app-framework"),
				OSSBucket:      "k8s-dev-log-cold",
				OSSPrefix:      "app-framework",
				RoleArn:        "acs:ram::1378641383022900:role/aliyunlogdefaultrole",
				BufferInterval: 300,
				BufferSize:     256,
				CompressType:   "none",
				PathFormat:     "event_date=%Y-%m-%d/hour=%H/%M",
				Format:         "json",
			},
		})
		assert.Nil(t, err)

		err = logStore.DeleteShipper(s.getAliLogStoreColdStorageShipperName("app-framework"))
		assert.Nil(t, err)
	})
}
