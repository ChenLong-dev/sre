package huawei

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gitlab.shanhai.int/sre/library/base/null"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/httpclient"

	"rulai/config"
	"rulai/models/entity"
	_errcode "rulai/utils/errcode"

	"rulai/vendors/huawei/APIGW-go-sdk-2.0.2/core"
)

func Test_Controller_HuaweiLTSStreamOperations(t *testing.T) {
	c, err := getUnittestController()
	require.NoError(t, err)

	now := time.Now().UnixNano()
	ctx := context.TODO()

	ruleName := fmt.Sprintf("unittest-rule-%d", now)
	containerName := fmt.Sprintf("unittest-container-%d", now) // 容器名并不需要实际存在
	ttlInDays := null.IntFrom(7)
	stream := &entity.AppLTSStream{
		ClusterName: entity.ClusterZeus,
		EnvName:     entity.AppEnvStg,
		StreamName:  fmt.Sprintf("unittest-stream-%d", now),
	}

	streamID, err := c.createHuaweiLTSStream(ctx, stream, ttlInDays)
	require.NoError(t, err)
	require.NotEmpty(t, streamID)

	streamID, err = c.createHuaweiLTSStream(ctx, stream, ttlInDays)
	require.NoError(t, err)
	require.NotEmpty(t, streamID)

	// FIXME: 华为云 API 目前有 bug, 填写日志流名称时可能会报重名错误, 当前只能先填写 stream_id
	ruleID, err := c.createHuaweiAOMToLTSRule(ctx, stream.EnvName, stream.ClusterName, string(stream.EnvName),
		ruleName, containerName, streamID, stream.StreamName)
	assert.NoError(t, err)
	assert.NotEmpty(t, ruleID)
	// FIXME: 华为云 API 目前有 bug, 填写日志流名称时可能会报重名错误, 当前只能先填写 stream_id
	ruleID, err = c.createHuaweiAOMToLTSRule(ctx, stream.EnvName, stream.ClusterName, string(stream.EnvName), ruleName,
		containerName, streamID, stream.StreamName)
	assert.NoError(t, err)
	assert.NotEmpty(t, ruleID)

	// FIXME: 华为云 API 目前有 bug, 填写日志流名称时可能会报重名错误, 当前只能先填写 stream_id
	tmpRuleID, err := c.createHuaweiAOMToLTSRule(ctx, stream.EnvName, stream.ClusterName, string(stream.EnvName), ruleName,
		containerName, streamID, stream.StreamName)
	assert.True(t, errcode.EqualError(_errcode.LTSResourceAlreadyExists, err))
	assert.Empty(t, tmpRuleID)

	err = c.deleteHuaweiAOMToLTSRule(ctx, ruleID)
	assert.NoError(t, err)

	err = c.deleteHuaweiAOMToLTSRule(ctx, ruleID)
	assert.True(t, errcode.EqualError(_errcode.LTSResourceNotFound, err))

	err = c.deleteHuaweiLTSStream(ctx, stream.EnvName, stream.ClusterName, streamID)
	assert.NoError(t, err)

	err = c.deleteHuaweiLTSStream(ctx, stream.EnvName, stream.ClusterName, streamID)
	assert.True(t, errcode.EqualError(_errcode.LTSResourceNotFound, err))
}

func getUnittestController() (*Controller, error) {
	cfg := config.Read("../../config/config.yaml")
	var huaweiCfg *config.VendorConfig
	for _, vendorCfg := range cfg.Vendors {
		if vendorCfg.Name == string(entity.VendorHuawei) {
			huaweiCfg = vendorCfg
			break
		}
	}

	if huaweiCfg == nil {
		return nil, errors.New("no huawei vendor config")
	}

	stgClusterConfigs, ok := cfg.K8sClusters[string(entity.AppEnvStg)]
	if !ok {
		return nil, errors.New("no stg env cluster configs")
	}

	var stgZeusCfg *config.K8sClusterConfig
	for _, clusterCfg := range stgClusterConfigs {
		if clusterCfg.Name == string(entity.ClusterZeus) {
			stgZeusCfg = clusterCfg
		}
	}
	if stgZeusCfg == nil {
		return nil, errors.New("no zeus cluster configs")
	}

	c := &Controller{
		DisableLogConfig: huaweiCfg.DisableLogConfig,
		LogEndpoint:      huaweiCfg.LogEndpoint,
		ProjectID:        huaweiCfg.RegionID,
		HTTPClient:       httpclient.NewHttpClient(cfg.HTTPClient),
		APISigner: &core.Signer{
			Key:    huaweiCfg.AccessKeyID,
			Secret: huaweiCfg.AccessKeySecret,
		},
		Clusters: map[entity.AppEnvName]map[entity.ClusterName]*ClusterConfig{
			entity.AppEnvStg: {
				entity.ClusterZeus: {
					ClusterID:           stgZeusCfg.ClusterID,
					ClusterNameInVendor: stgZeusCfg.ClusterNameInVendor,
					LogGroupID:          stgZeusCfg.LogBucketID,
					LogGroupName:        stgZeusCfg.LogBucketName,
				},
			},
		},
	}

	return c, nil
}

func TestController_GetLTSStreamIDByStreamName(t *testing.T) {
	type args struct {
		ctx         context.Context
		envName     entity.AppEnvName
		clusterName entity.ClusterName
		streamName  string
	}
	tests := []struct {
		name         string
		args         args
		wantStreamID string
		wantErr      assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{"sms-backend-stg-zeus", args{
			ctx:         context.TODO(),
			envName:     entity.AppEnvStg,
			clusterName: entity.ClusterZeus,
			streamName:  "sms-backend-stg-zeus",
		}, "15d79a93-e3e4-4670-a55a-743da7a04aec", nil},
		{"log-not-exist", args{
			ctx:         context.TODO(),
			envName:     entity.AppEnvStg,
			clusterName: entity.ClusterZeus,
			streamName:  "log-not-exist",
		}, "", nil},
	}
	c, _ := getUnittestController()
	for _, tt := range tests {
		ctx := tt.args.ctx
		envName := tt.args.envName
		clusterName := tt.args.clusterName
		streamName := tt.args.streamName
		wantStreamID := tt.wantStreamID

		t.Run(tt.name, func(t *testing.T) {
			gotStreamID, err := c.GetLTSStreamIDByStreamName(ctx, envName,
				clusterName, streamName)
			t.Log(gotStreamID, err)
			assert.Equalf(t, wantStreamID, gotStreamID, "GetLTSStreamIDByStreamName(%v, %v, %v, %v)",
				ctx, envName, clusterName, streamName)
		})
	}
}

func TestController_GetAOMToLTSRuleIDByStream(t *testing.T) {
	type args struct {
		ctx           context.Context
		envName       entity.AppEnvName
		clusterName   entity.ClusterName
		namespace     string
		ruleName      string
		containerName string
		streamID      string
		streamName    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{
			"test", args{
				ctx:           context.TODO(),
				envName:       entity.AppEnvStg,
				clusterName:   entity.ClusterZeus,
				namespace:     "stg",
				ruleName:      "",
				containerName: "",
				streamID:      "d7a654dc-a1bd-4a80-9de5-077aef798e28",
				streamName:    "sms-backend-prd-zeus",
			}, "7757e16c-cc82-4c9d-b19e-7a8e38a25e40", nil},
	}

	c, _ := getUnittestController()
	for _, tt := range tests {
		ctx := tt.args.ctx
		envName := tt.args.envName
		clusterName := tt.args.clusterName
		namespace := tt.args.namespace
		ruleName := tt.args.ruleName
		containerName := tt.args.containerName
		streamID := tt.args.streamID
		streamName := tt.args.streamName
		want := tt.want

		t.Run(tt.name, func(t *testing.T) {
			got, err := c.GetAOMToLTSRuleIDByStream(ctx, envName, clusterName,
				namespace, ruleName, containerName, streamID, streamName)
			t.Log(got, err)
			assert.Equalf(t, want, got, "GetAOMToLTSRuleIDByStream(%v, %v, %v, %v, %v, %v, %v, %v)",
				ctx, envName, clusterName, namespace, ruleName,
				containerName, streamID, streamName)
		})
	}
}
