// 当前公共库http客户端未抽象成接口类型
// 暂时是远端调用型式的单元测试
package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

func Test_Service_QDNSRecordOperations(t *testing.T) {
	ctx := context.TODO()

	t.Run("QDNS 解析记录流程测试", func(t *testing.T) {
		ts := time.Now().UnixNano()
		domainName := fmt.Sprintf("ams-unittest-QDNS-record-%d", ts)
		ip := "1.2.3.4"
		operator := fmt.Sprintf("ams-unittest-QDNS-record-operator-%d", ts)
		ttl := "77"
		createReq := &req.CreateQDNSRecordReq{
			DomainType:       entity.ARecord,
			DomainRecordName: domainName,
			DomainValue:      ip,
			DomainName:       AliK8sPrivateZone,
			TTL:              ttl,
			DomainUpdater:    operator,
		}
		createRes, err := s.CreateQDNSRecord(ctx, createReq)
		assert.Nil(t, err)
		assert.Zero(t, createRes.Status)

		// 重复创建的特殊情况
		createRes, err = s.CreateQDNSRecord(ctx, createReq)
		assert.Nil(t, err)
		assert.Equal(t, resp.QDNSStatusDuplicateRecordResp.Status, createRes.Status)
		assert.Equal(t, resp.QDNSStatusDuplicateRecordResp.Msg, createRes.Msg)

		getReq := &req.GetQDNSRecordsReq{
			DomainType:       entity.ARecord,
			DomainRecordName: domainName,
			DomainValue:      ip,
			PrivateZone:      AliK8sPrivateZone,
			PageNumber:       1,
			PageSize:         req.GetQDNSRecordsPageSizeLimit,
		}
		records, err := s.GetQDNSRecords(ctx, getReq)
		assert.Nil(t, err)
		if assert.Len(t, records, 1) {
			assert.Empty(t, records[0].DomainController)
			assert.Equal(t, AliK8sPrivateZone, records[0].DomainName)
			assert.Equal(t, domainName, records[0].Name)
			assert.Equal(t, ttl, records[0].TTL)
			assert.Equal(t, entity.ARecord, records[0].Type)
			assert.Equal(t, ip, records[0].Value)
		}

		getReq.PageNumber = 2
		records, err = s.GetQDNSRecords(ctx, getReq)
		assert.Nil(t, err)
		assert.Len(t, records, 0)

		deleteReq := &req.DeleteQDNSRecordReq{
			DomainType:       entity.ARecord,
			DomainRecordName: domainName,
			DomainValue:      ip,
			DomainName:       AliK8sPrivateZone,
			DomainUpdater:    operator,
		}
		deleteRes, err := s.DeleteQDNSRecord(ctx, deleteReq)
		assert.Nil(t, err)
		assert.Zero(t, deleteRes.Status)

		getReq.PageNumber = 1
		records, err = s.GetQDNSRecords(ctx, getReq)
		assert.Nil(t, err)
		assert.Len(t, records, 0)
	})
}

func Test_Service_QDNSBusinessOperations(t *testing.T) {
	ctx := context.TODO()
	ts := time.Now().UnixNano()
	envName := "dev"
	domainName := fmt.Sprintf("ams-unittest-qdns-business-host-%d.qtfm.cn", ts)
	healthCheckPath := "/health"
	target1 := fmt.Sprintf("ams-unittest-QDNS-business-host-1.%s:%d", AliK8sPrivateZone, req.KongServiceUnUsefulPort)
	weight1 := 30
	modifiedWeight1 := 40
	target2 := fmt.Sprintf("ams-unittest-QDNS-business-host-2.%s:%d", AliK8sPrivateZone, req.KongServiceUnUsefulPort)
	weight2 := 70
	modifiedWeight2 := 60
	operator := fmt.Sprintf("ams-unittest-QDNS-business-operator-%d", ts)
	business := fmt.Sprintf("ams-unittest-QDNS-business-business_name-%d", ts)
	modifiedBusiness := fmt.Sprintf("ams-unittest-QDNS-business-modified-business_name-%d", ts)
	unittestKongTag := "ams-unittest-QDNS-business-tag"
	unittestKongHost1Tag := "ams-unittest-QDNS-business-host-1-tag"
	unittestKongHost2Tag := "ams-unittest-QDNS-business-host-2-tag"

	t.Run("QDNS 统一接入规则流程测试", func(t *testing.T) {
		// 删除历史遗留单元测试数据
		preListReq := &req.GetQDNSBusinessListReq{
			Targets: []string{target1, target2},
		}

		preListRes, preErr := s.getQDNSBusinessList(ctx, preListReq)
		require.NoError(t, preErr)

		for i := range preListRes {
			preDeleteReq := &req.DeleteQDNSBusinessReq{
				Env:      preListRes[i].Env,
				ID:       preListRes[i].ID,
				UserName: operator,
			}
			preErr = s.deleteQDNSBusiness(ctx, preDeleteReq)
			require.NoError(t, preErr)
			t.Logf("deleted with env=%s, id=%d", preListRes[i].Env, preListRes[i].ID)
		}

		t.Logf("removed %d old test data", len(preListRes))

		// 测试创建使用 AMS 需要的值
		qdnsEnv := s.getKongEnvByAppEnv(entity.AppEnvName(envName))
		require.NotEqual(t, entity.QDNSEnvNameUnknown, qdnsEnv)

		createReq := &req.CreateQDNSBusinessReq{
			Env:      qdnsEnv,
			UserName: operator,
			Business: business,
			Upstream: &req.CreateQDNSKongUpstreamReq{
				Slots: req.KongaDefaultSlotPtr,
				HealthChecks: &req.KongUpstreamHealthCheck{
					Active: &req.KongUpstreamActiveHealthCheck{
						Healthy: &req.KongUpstreamActiveHealthCheckHealthyConfig{
							Interval:  req.AMSKongDefaultHealthCheckIntervalPtr,
							Successes: req.AMSKongDefaultHealthCheckSuccessPtr,
						},
						Unhealthy: &req.KongUpstreamActiveHealthCheckUnhealthyConfig{
							TCPFailures:  req.AMSKongDefaultHealthCheckFailuresPtr,
							Timeouts:     req.AMSKongDefaultHealthCheckTimeoutsPtr,
							HTTPFailures: req.AMSKongDefaultHealthCheckFailuresPtr,
							Interval:     req.AMSKongDefaultHealthCheckIntervalPtr,
						},
						HTTPPath: healthCheckPath,
						Timeout:  req.AMSKongDefaultHealthCheckTimeoutInSecondPtr,
					},
					Passive: &req.KongUpstreamPassiveHealthCheck{
						Healthy: &req.KongUpstreamPassiveHealthCheckHealthyConfig{
							Successes: req.AMSKongDefaultHealthCheckSuccessPtr,
						},
						Unhealthy: &req.KongUpstreamPassiveHealthCheckUnhealthyConfig{
							HTTPFailures: req.AMSKongDefaultHealthCheckFailuresPtr,
							TCPFailures:  req.AMSKongDefaultHealthCheckFailuresPtr,
							Timeouts:     req.AMSKongDefaultHealthCheckTimeoutsPtr,
						},
					},
				},
				Tags: []string{unittestKongTag},
			},
			Service: &req.CreateQDNSKongServiceReq{
				Retries: req.AMSKongServiceDefaultRetriesPtr,
				Port:    req.KongServiceUnUsefulPort,
				Tags:    []string{unittestKongTag},
			},
			Routes: []*req.CreateQDNSKongRouteReq{
				{
					Hosts:        []string{domainName},
					Paths:        req.KongDefaultRoutePaths,
					StripPath:    req.AMSKongDefaultStripPath,
					PathHandling: entity.KongPathHandleBehaviorV1,
					Tags:         []string{unittestKongTag},
				},
			},
			Targets: []*req.UpsertQDNSKongTargetReq{
				{
					Target: target1,
					Weight: &weight1,
					Tags:   []string{unittestKongTag, unittestKongHost1Tag},
				},
				{
					Target: target2,
					Weight: &weight2,
					Tags:   []string{unittestKongTag, unittestKongHost2Tag},
				},
			},
		}
		createRes, err := s.createQDNSBusiness(ctx, createReq)
		require.Nil(t, err)
		require.Equal(t, resp.QDNSSuccessStatus, createRes.GetStatus())
		require.Nil(t, createRes.Data)

		// 获取列表暂时不需要测试分页
		listReq := &req.GetQDNSBusinessListReq{
			Targets: []string{target1, target2},
		}
		listRes, err := s.getQDNSBusinessList(ctx, listReq)
		require.Nil(t, err)
		require.Len(t, listRes, 1)
		checkQDNSBusinessCreationResult(t, createReq, listRes[0])

		// 重复创建的特殊情况, QDNS 不会认为是重复, 会创建一组完全相同(名称不同)的资源
		createRes, err = s.createQDNSBusiness(ctx, createReq)
		require.Nil(t, err)
		require.Equal(t, resp.QDNSSuccessStatus, createRes.GetStatus())

		listRes, err = s.getQDNSBusinessList(ctx, listReq)
		require.Nil(t, err)
		expectedCount := 2
		require.Len(t, listRes, expectedCount)
		for i := range listRes {
			checkQDNSBusinessCreationResult(t, createReq, listRes[i])
		}

		patchReq := &req.PatchQDNSBusinessReq{
			Env: listRes[0].Env,
			Upstream: &req.PatchQDNSKongUpstreamReq{
				ID: listRes[0].BindID,
				Targets: []*req.UpsertQDNSKongTargetReq{
					{
						Target: target1,
						Weight: &modifiedWeight1,
						Tags:   []string{unittestKongTag, unittestKongHost1Tag},
					},
					{
						Target: target2,
						Weight: &modifiedWeight2,
						Tags:   []string{unittestKongTag, unittestKongHost2Tag},
					},
				},
			},
			UserName: operator,
			Business: modifiedBusiness,
		}

		err = s.patchQDNSBusiness(ctx, patchReq)
		require.Nil(t, err)

		listRes, err = s.getQDNSBusinessList(ctx, listReq)
		require.Nil(t, err)
		require.Len(t, listRes, expectedCount)

		// 测试删除
		var deleteReq *req.DeleteQDNSBusinessReq
		for len(listRes) > 0 {
			deleteReq = &req.DeleteQDNSBusinessReq{
				Env:      listRes[0].Env,
				ID:       listRes[0].ID,
				UserName: operator,
			}
			err = s.deleteQDNSBusiness(ctx, deleteReq)
			require.Nil(t, err)

			listRes, err = s.getQDNSBusinessList(ctx, listReq)
			require.Nil(t, err)
			expectedCount--
			require.Len(t, listRes, expectedCount)
		}
	})
}

type getQDNSUpdaterTestCase struct {
	tag          string
	operatorID   string
	operatorName string
	ecode        errcode.Codes
}

func Test_Service_GetQDNSUpdater(t *testing.T) {
	ctx := context.TODO()
	tcs := generateGetQDNSUpdaterTestCases()

	for _, tc := range tcs {
		operatorID := tc.operatorID
		expOperatorName := tc.operatorName
		expEcode := tc.ecode
		tag := tc.tag

		t.Run(tag, func(t *testing.T) {
			operatorName, err := s.getQDNSUpdater(ctx, operatorID)
			assert.Equal(t, expOperatorName, operatorName)
			if expEcode == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errcode.EqualError(expEcode, err))
			}
		})
	}
}

func generateGetQDNSUpdaterTestCases() []*getQDNSUpdaterTestCase {
	return []*getQDNSUpdaterTestCase{
		{
			tag:        "无效的 operator_id",
			operatorID: "abc",
			ecode:      _errcode.K8sInternalError,
		},
		{
			tag:          "系统用户",
			operatorID:   "-1",
			operatorName: "ams_系统用户",
		},
		{
			tag:          "operator: 236",
			operatorID:   "236",
			operatorName: "ams_王晋元",
		},
	}
}

func checkQDNSBusinessCreationResult(t *testing.T, createReq *req.CreateQDNSBusinessReq, businessDetail *resp.GetQDNSBusinessDetailResp) {
	now := time.Now().Unix()
	assert.Equal(t, createReq.Env, businessDetail.Env)
	assert.NotEmpty(t, businessDetail.BindID)
	assert.Greater(t, businessDetail.ID, 0)
	assert.NotEmpty(t, businessDetail.Name)
	// assert.Equal(t, createReq.Upstream.HealthChecks.Active, businessDetail.ActiveHealthCheck)
	assert.Equal(t, 1, businessDetail.Status)
	assert.EqualValues(t, createReq.Upstream.Tags, businessDetail.Tags)
	assert.InDelta(t, now, time.Time(*businessDetail.CreateTime).Unix(), 15)
	assert.InDelta(t, now, time.Time(*businessDetail.UpdateTime).Unix(), 15)
	assert.Equal(t, createReq.UserName, businessDetail.LastModify)

	if assert.Len(t, businessDetail.Service, 1) {
		assert.Equal(t, *createReq.Service.Retries, businessDetail.Service[0].Retries)
		assert.Equal(t, createReq.Service.Port, businessDetail.Service[0].Port)
		assert.EqualValues(t, createReq.Service.Tags, businessDetail.Service[0].Tags)
	}

	// 目前 routes 只会创建一个
	if assert.Len(t, businessDetail.Route, len(createReq.Routes)) &&
		assert.Len(t, businessDetail.Route, 1) {
		m := make(map[string]*req.CreateQDNSKongRouteReq)
		for _, routeReq := range createReq.Routes {
			if assert.Len(t, routeReq.Hosts, 1) {
				m[routeReq.Hosts[0]] = routeReq
			}
		}

		for _, route := range businessDetail.Route {
			assert.Equal(t, createReq.Env, route.Env)
			assert.NotEmpty(t, route.BindID)
			assert.Greater(t, route.ID, 0)
			assert.NotEmpty(t, route.Name)

			if assert.Len(t, route.Host, 1) {
				routeReq, ok := m[route.Host[0]]
				if assert.True(t, ok) {
					assert.EqualValues(t, routeReq.Methods, route.Methods)
					assert.EqualValues(t, routeReq.Hosts, route.Host)
					assert.EqualValues(t, routeReq.Paths, route.Path)
					assert.Nil(t, routeReq.Headers, route.Headers)
					assert.Equal(t, 426, route.HTTPSRedirectStatusCode)
					assert.Equal(t, 0, route.RegexPriority)
					if routeReq.StripPath == nil || !(*routeReq.StripPath) {
						assert.Equal(t, 0, route.StripPath)
					} else {
						assert.Equal(t, 1, route.StripPath)
					}
					assert.Equal(t, routeReq.PathHandling, route.PathHandling)
					assert.EqualValues(t, routeReq.Tags, route.Tags)
					if routeReq.PreserveHost == nil || !(*routeReq.PreserveHost) {
						assert.Equal(t, 0, route.PreserveHost)
					} else {
						assert.Equal(t, 1, route.PreserveHost)
					}
				}
			}
		}
	}

	if assert.Len(t, businessDetail.Targets, len(createReq.Targets)) {
		m := make(map[string]*req.UpsertQDNSKongTargetReq)
		for _, targetReq := range createReq.Targets {
			if targetReq.Weight == nil || *targetReq.Weight > 0 {
				// 权重为 0 的 target 将不存在
				m[targetReq.Target] = targetReq
			}
		}

		for _, target := range businessDetail.Targets {
			assert.NotEmpty(t, target.ID)
			assert.InDelta(t, now, int64(target.CreatedAt), 15)
			if assert.NotNil(t, target.Upstream) {
				assert.Equal(t, businessDetail.BindID, target.Upstream.ID)
			}

			targetReq, ok := m[target.Target]
			if assert.True(t, ok) {
				assert.Equal(t, targetReq.Target, target.Target)
				if targetReq.Weight == nil {
					assert.Equal(t, 100, target.Weight)
				} else {
					assert.Equal(t, *targetReq.Weight, target.Weight)
				}
				assert.EqualValues(t, targetReq.Tags, target.Tags)
			}
		}
	}
}
