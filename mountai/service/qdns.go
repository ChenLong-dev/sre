package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/base/slice"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/httpclient"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

// CreateQDNSRecord 添加QDNS解析记录
func (s *Service) CreateQDNSRecord(ctx context.Context, createReq *req.CreateQDNSRecordReq) (*resp.QDNSStandardResp, error) {
	res := new(resp.QDNSStandardResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/domain", config.Conf.QDNS.Host)).
		JsonBody(createReq).
		Method(http.MethodPost).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	return res, nil
}

// DeleteQDNSRecord 删除QDNS解析记录
func (s *Service) DeleteQDNSRecord(ctx context.Context, deleteReq *req.DeleteQDNSRecordReq) (*resp.QDNSStandardResp, error) {
	res := new(resp.QDNSStandardResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/domain", config.Conf.QDNS.Host)).
		JsonBody(deleteReq).
		Method(http.MethodDelete).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	return res, nil
}

// GetQDNSRecords 从QDNS获取解析记录
func (s *Service) GetQDNSRecords(ctx context.Context, getReq *req.GetQDNSRecordsReq) ([]*entity.QDNSRecord, error) {
	var records []*entity.QDNSRecord
	res := &resp.QDNSStandardListResp{Data: &records}
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/domain", config.Conf.QDNS.Host)).
		QueryParams(httpclient.NewUrlValue().
			Add("domain_type", string(getReq.DomainType)).
			Add("domain_record_name", getReq.DomainRecordName).
			Add("domain_value", getReq.DomainValue).
			Add("private_zone", getReq.PrivateZone).
			Add("page_number", strconv.Itoa(getReq.PageNumber)).
			Add("page_size", strconv.Itoa(getReq.PageSize))).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	if res.Status != resp.QDNSRecordSuccessStatus {
		return nil, errors.Wrapf(_errcode.QDNSInternalError,
			"Get QDNS records failed with code(%d) and msg(%s)", res.GetStatus(), res.GetMessage())
	}

	if len(records) == res.PageSize {
		// 返回数量到达分页上限记录一个警告
		log.Errorc(ctx, "Count of QDNS response data exceeds the page_size limit(%d)", req.GetQDNSRecordsPageSizeLimit)
	}

	return records, nil
}

// GetQDNSFrontendInfo 根据 access hosts 获取 QDNS 统一接入 Kong Route 中的 paths 和 hosts
func (s *Service) GetQDNSFrontendInfo(ctx context.Context,
	appType entity.AppType, serviceType entity.AppServiceType, accessHosts []string) ([]*resp.QDNSFrontendInfoResp, error) {
	accessHosts = removeDuplicateItem(accessHosts)
	if appType != entity.AppTypeService || len(accessHosts) == 0 {
		return make([]*resp.QDNSFrontendInfoResp, 0), nil
	}

	m := make(map[string][]*resp.KongRouteInfoResp)
	// accessHosts 未带端口
	for i := range accessHosts {
		accessHosts[i] = fmt.Sprintf("%s:%d", accessHosts[i], entity.ServiceDefaultInternalPort)
		m[accessHosts[i]] = nil
	}

	list, err := s.getAllQDNSBusinessListByTargets(ctx, accessHosts)
	if err != nil {
		return nil, err
	}

	for i := range list {
		info := &resp.KongRouteInfoResp{
			KongName: getKongAliasName(list[i].Env),
		}

		for _, route := range list[i].Route {
			info.Hosts = append(info.Hosts, route.Host...)
			info.Paths = append(info.Paths, route.Path...)
		}

		var targetFound bool
		for _, target := range list[i].Targets {
			if _, ok := m[target.Target]; ok {
				targetFound = true
				m[target.Target] = append(m[target.Target], info)
				break
			}
		}

		if !targetFound {
			log.Errorc(ctx, "QDNS business(id=%d) does not refers to access_hosts(%v)", list[i].ID, accessHosts)
			continue
		}
	}

	end := 0
	frontendInfo := make([]*resp.QDNSFrontendInfoResp, len(m))
	for i := range accessHosts {
		if len(m[accessHosts[i]]) == 0 {
			continue
		}

		frontendInfo[end] = &resp.QDNSFrontendInfoResp{
			AccessHost: accessHosts[i],
			RouteInfo:  m[accessHosts[i]],
		}
		end++
	}

	return frontendInfo[:end], nil
}

// SetAppClusterQDNSWeights 设置应用环境集群在 QDNS 统一接入规则中的权重
func (s *Service) SetAppClusterQDNSWeights(ctx context.Context, app *resp.AppDetailResp, setReq *req.SetAppClusterQDNSWeightsReq) error {
	// 牵涉到域名操作, 谨慎一点多校验一步
	_, ok := app.Env[setReq.Env]
	if !ok {
		return errors.Wrapf(errcode.InternalError, "env(%s) not found in app(%s)", setReq.Env, app.ID)
	}

	targets := make([]string, len(setReq.ClusterWeights))
	weightGtZeroCluster := make([]*req.AppClusterKongWeight, 0, len(setReq.ClusterWeights))
	for i := range targets {
		targets[i] = s.getKongTargetHostAndPort(app.ServiceName, setReq.Env, setReq.ClusterWeights[i].ClusterName)
		if setReq.ClusterWeights[i].Weight > 0 {
			weightGtZeroCluster = append(weightGtZeroCluster, setReq.ClusterWeights[i])
		}
	}

	list, err := s.getAllQDNSBusinessListByTargets(ctx, targets)
	if err != nil {
		return errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	k8sKongTag := s.getK8sKongTag(app.ServiceName, setReq.Env)
	// 检查是否有非 AMS 配置的统一接入规则, 根据 setReq.ForceUpdateAll 参数决定是否处理
	// 同时检查 AMS 配置的统一接入规则是否存在
	var k8sBusiness *resp.GetQDNSBusinessDetailResp
	errGroup := errcode.NewGroup(_errcode.ExtraWeightConfigExistsError)
	for _, business := range list {
		for _, tag := range business.Tags {
			if tag == k8sKongTag {
				if k8sBusiness != nil {
					return errors.Wrap(errcode.InternalError, "more than 1 k8s-domain's QDNS business found")
				}

				k8sBusiness = business
				break
			}
		}

		if k8sBusiness == nil && !setReq.ForceUpdateAll {
			errGroup = errGroup.AddChildren(
				errors.Wrapf(_errcode.ExtraWeightConfigExistsError, "business_id=%d", business.ID))
		}
	}

	if len(errGroup.Children()) > 0 {
		return errGroup
	}

	setReq.ClusterWeights = weightGtZeroCluster

	// 兼容性操作: 第一次设置权重时, 对应的 Kong 路由尚未创建
	// 域名和 Kong 路由记录分开控制, 调整权重时不处理域名解析
	if k8sBusiness == nil {
		return s.ensureQDNSBusiness(ctx, app.ServiceName, setReq)
	}

	operator, err := s.getQDNSUpdater(ctx, setReq.OperatorID)
	if err != nil {
		return err
	}

	// 强制修改所有相关联的 target
	for _, business := range list {
		// 判断是否有 target 需要删除
		targetMapping := make(map[string]*resp.KongTarget, len(business.Targets))
		for _, target := range business.Targets {
			targetMapping[target.Target] = target
		}

		var targetsToBeModified []*req.UpsertQDNSKongTargetReq
		targetsWithHealthCheckToBeEnforced := make([]*entity.KongID, 0, len(setReq.ClusterWeights))
		for _, item := range setReq.ClusterWeights {
			clusterTag := s.getK8sClusterKongTag(item.ClusterName)
			target := &req.UpsertQDNSKongTargetReq{
				// TODO: 如果未来有多云需要区分 privateZone 的域名, 这里需要改
				// 当前目标端口必定是固定的 Ingress HTTP端口, 注意未来如果有变动这里要考虑如何兼容
				Target: s.getKongTargetHostAndPort(app.ServiceName, setReq.Env, item.ClusterName),
				Weight: &item.Weight,
				Tags:   []string{k8sKongTag, clusterTag},
			}

			if business.BindID != k8sBusiness.BindID {
				target.Tags = []string{}
			}

			targetsToBeModified = append(targetsToBeModified, target)

			if _, ok := targetMapping[target.Target]; ok {
				delete(targetMapping, target.Target)
				targetsWithHealthCheckToBeEnforced = append(targetsWithHealthCheckToBeEnforced, &entity.KongID{ID: target.Target})
			}
		}

		// 多余的 Target 需要删除
		for _, target := range targetMapping {
			targetsToBeModified = append(targetsToBeModified, &req.UpsertQDNSKongTargetReq{
				// TODO: 如果未来有多云需要区分 privateZone 的域名, 这里需要改
				// 当前目标端口必定是固定的 Ingress HTTP端口, 注意未来如果有变动这里要考虑如何兼容
				Target: target.Target,
				Weight: req.ZeroIntPtr, // 删除时权重修改为0, 不需要 tag
			})
		}

		patchReq := &req.PatchQDNSBusinessReq{
			Env: business.Env,
			Upstream: &req.PatchQDNSKongUpstreamReq{
				ID:      business.BindID,
				Targets: targetsToBeModified,
			},
			UserName: operator,
			// Business: "", // TODO: 如果未来 AMS 区分业务, 则填写
		}

		// 调整后如果 upstream 下 target 数量超过 1, 需要开启 upstream 的健康检查中所有项(出现问题的 target 依靠 Kong 的健康检查及时移除及恢复时自愈)
		// 反之如果 upstream 下 target 数量未超过 1, 需要关闭 upstream 的健康检查中所有项(唯一 target 出现问题必定 502/503/504, 关闭健康检查减少无效报警)
		// 并且手动将唯一 target 设置成 healthy 状态(原因是调整时有可能 target 正处于不健康状态, 关闭健康检查后无法自行恢复)
		// 注: 操作顺序上必须先关闭健康检查, 然后手动设置健康状态
		forceToHealthy := false
		// NOTE: setReq.ClusterWeights 需要保证不重复, 不然与实际的 target 数量不一致
		if len(setReq.ClusterWeights) > 1 {
			patchReq.Upstream.HealthChecks = s.generateQDNSKongUpstreamEnabledHealthChecks(ctx, business, setReq.HealthCheckPath)
		} else {
			forceToHealthy = true
			patchReq.Upstream.HealthChecks = s.generateQDNSKongUpstreamDisabledHealthChecks(ctx, business)
		}

		err = s.patchQDNSBusiness(ctx, patchReq)
		if err != nil {
			return err
		}

		if forceToHealthy {
			updateReq := &req.UpdateQDNSKongUpstreamsTargetsHealthyReq{
				Env: business.Env,
				Upstreams: []*req.UpdateQDNSKongUpstreamTargetsHealthyReq{
					{
						ID:      business.BindID,
						Targets: targetsWithHealthCheckToBeEnforced,
					},
				},
			}

			err = s.updateQDNSKongUpstreamsTargetsHealthy(ctx, updateReq)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetAppClusterQDNSWeights 获取应用环境所有集群在 QDNS 统一接入规则中的权重
// 取 k8s 通配域名对应 Kong Upstream Target 中配置的权重，其他 Kong upstream 暂时不能保证没有发生人工修改
func (s *Service) GetAppClusterQDNSWeights(ctx context.Context,
	appServiceName string, envName entity.AppEnvName) ([]*resp.AppClusterQDNSWeightResp, error) {
	// 未支持多集群的项目不应进入该方法, 进入则获取全部集群对应域名的权重
	list, err := s.getAppClusterQDNSBusiness(ctx, appServiceName, envName)
	if err != nil {
		return nil, err
	}

	var (
		isK8sTarget bool
		clusterName entity.ClusterName
		weights     []*resp.AppClusterQDNSWeightResp
	)

	k8sKongTag := s.getK8sKongTag(appServiceName, envName)

	for _, businessDetail := range list {
		for _, target := range businessDetail.Targets {
			isK8sTarget, clusterName, err = s.getK8sClusterNameFromKongTags(target.Tags, k8sKongTag)
			if err != nil {
				return nil, err
			}

			if isK8sTarget {
				weights = append(weights, &resp.AppClusterQDNSWeightResp{
					ClusterName: clusterName,
					Weight:      target.Weight,
				})
			}
		}
		if len(weights) > 0 {
			break
		}
	}

	return weights, nil
}

// getAppClusterQDNSBusiness 获取应用环境所有集群在 QDNS 统一接入规则中的配置
func (s *Service) getAppClusterQDNSBusiness(ctx context.Context,
	appServiceName string, envName entity.AppEnvName) ([]*resp.GetQDNSBusinessDetailResp, error) {
	clusters, err := s.GetClusters(ctx, &req.GetClustersReq{
		Namespace: string(envName),
	})
	if err != nil {
		return nil, err
	}

	targetHostAndPorts := make([]string, 2*len(clusters)) // nolint:mnd
	for i := range clusters {
		targetHostAndPorts[2*i] = s.getKongTargetHostAndPort(appServiceName, envName, clusters[i].Name)
		// 考虑现有 istio 部署,再回退到 ingress 的场景,这里需要检测 istio 的 target
		e := entity.IstioNamespacePrefix + envName
		targetHostAndPorts[2*i+1] = s.getKongTargetHostAndPort(appServiceName, e, clusters[i].Name)
	}

	return s.getAllQDNSBusinessListByTargets(ctx, targetHostAndPorts)
}

// ensureQDNSBusinessWithCalculation 计算并确认QDNS统一接入规则(用于兼容老服务尚未配置权重的情形)
// 增加 istio 灰度流量处理逻辑
func (s *Service) ensureQDNSBusinessWithCalculation(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp, stage entity.TaskAction) error {
	setReq := &req.SetAppClusterQDNSWeightsReq{
		Env:             task.EnvName,
		OperatorID:      task.OperatorID,
		HealthCheckPath: task.Param.HealthCheckURL,
		EnableIstio:     null.BoolFrom(s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app)),
		ClusterName:     task.ClusterName,
		Stage:           stage,
	}

	// 计算 灰度流量权重
	upstreamTargetWeights, e := s.getUpstreamTargetsWithCalculateWeights(ctx, project, app, task, stage)
	if e != nil {
		return e
	}

	// 如果没有要调整 targets, 结束 qdns 调整
	if len(upstreamTargetWeights) == 0 {
		return nil
	}

	setReq.UpstreamTargetWeights = upstreamTargetWeights
	return s.ensureQDNSBusiness(ctx, app.ServiceName, setReq)
}

// getUpstreamTargetsWithCalculateWeights 当应用从 ingress 切换为 istio 灰度部署时,会触发此逻辑
func (s *Service) getUpstreamTargetsWithCalculateWeights(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp, stage entity.TaskAction) ([]*req.KongUpstreamTargetsWeight, error) {
	// target 权重数据
	result := make([]*req.KongUpstreamTargetsWeight, 0, 2)

	// 获取已部署 pod 数量(不包含当前正在部署的 pod)
	pods, err := s.GetPods(ctx, task.ClusterName, &req.GetPodsReq{Env: string(task.EnvName), ProjectName: project.Name,
		AppName: app.Name, Namespace: ""})
	if err != nil {
		return nil, err
	}

	// 存在 ingress 服务
	hasIngressDeployment := false
	hasIstioDeployment := false
	weights := make(map[string]int)
	for k := range pods {
		// 仅清理资源时, 需要忽略当前版本
		if slice.StrSliceContains([]string{string(entity.TaskActionDelete)}, string(stage)) &&
			pods[k].ObjectMeta.Labels["version"] == task.Version {
			continue
		}

		weights[pods[k].Namespace]++

		if strings.HasPrefix(pods[k].Namespace, entity.IstioNamespacePrefix) {
			hasIstioDeployment = true
		} else {
			hasIngressDeployment = true
		}
	}

	// 判断当前活动任务是 istio 还是 ingress
	currentEnv := "ingress"
	if strings.HasPrefix(task.Namespace, entity.IstioNamespacePrefix) {
		currentEnv = "istio"
	}

	switch stage {
	// 全量发布的时候, task.namespace 决定了流量切istio 还是 ingresss
	// 处理首次部署
	case entity.TaskActionFullDeploy, entity.TaskActionFullCanaryDeploy:
		// 全量部署, 如果全量部署到 istio-xxx ,则通过设置 ingress targets 权重为0
		result = append(result, s.setupKongUpstreamTargetsWeight(app, task, s.getOppositeEnvName(task.Namespace), 0))
		// 第一次发布,可能没有 pod 需要默认为 1
		if v, ok := weights[task.Namespace]; !ok || v == 0 {
			weights[task.Namespace] = 1
		}

		result = append(result, s.setupKongUpstreamTargetsWeight(app, task, entity.AppEnvName(task.Namespace), weights[task.Namespace]))
	// 删除task 需要重新调整 upstream target 权重
	case entity.TaskActionDelete:
		result = append(result, s.setupKongUpstreamTargetsWeight(app, task, entity.AppEnvName(task.Namespace), weights[task.Namespace]))
		env := s.getOppositeEnvName(task.Namespace)
		result = append(result, s.setupKongUpstreamTargetsWeight(app, task, env, weights[string(env)]))
	// 金丝雀发布, 需要刷新老 target 权重
	case entity.TaskActionCanaryDeploy:
		weight := 1
		// 已经部署过了
		if weights[task.Namespace] > 1 {
			weight = weights[task.Namespace]
		}

		if currentEnv == entity.Istio {
			result = append(result, s.setupKongUpstreamTargetsWeight(app, task, entity.AppEnvName(task.Namespace), weight))
			if hasIngressDeployment {
				env := s.getOppositeEnvName(task.Namespace)
				result = append(result, s.setupKongUpstreamTargetsWeight(app, task, env, weights[string(env)]))
			}
		} else {
			result = append(result, s.setupKongUpstreamTargetsWeight(app, task, entity.AppEnvName(task.Namespace), weight))
			if hasIstioDeployment {
				env := s.getOppositeEnvName(task.Namespace)
				result = append(result, s.setupKongUpstreamTargetsWeight(app, task, env, weights[string(env)]))
			}
		}
	}

	return result, nil
}

// 获取 ingress 和 istio 对立的 环境名称,全量部署的时候需要将对立命名空间下的 target 权重清零(删除)
func (s *Service) getOppositeEnvName(ns string) entity.AppEnvName {
	if strings.HasPrefix(ns, entity.IstioNamespacePrefix) {
		return entity.AppEnvName(strings.ReplaceAll(ns, entity.IstioNamespacePrefix, ""))
	}
	return entity.AppEnvName(entity.IstioNamespacePrefix + ns)
}

func (s *Service) setupKongUpstreamTargetsWeight(app *resp.AppDetailResp, task *resp.TaskDetailResp,
	envName entity.AppEnvName, weight int) *req.KongUpstreamTargetsWeight {
	return &req.KongUpstreamTargetsWeight{
		ClusterName:    task.ClusterName,
		TargetHostPort: s.getKongTargetHostAndPort(app.ServiceName, envName, task.ClusterName),
		Weight:         weight,
	}
}

// ensureQDNSBusiness 确认QDNS统一接入规则
func (s *Service) ensureQDNSBusiness(ctx context.Context, appServiceName string, setReq *req.SetAppClusterQDNSWeightsReq) error {
	operator, err := s.getQDNSUpdater(ctx, setReq.OperatorID)
	if err != nil {
		return err
	}

	qdnsEnv := s.getKongEnvByAppEnv(setReq.Env)
	if qdnsEnv == entity.QDNSEnvNameUnknown {
		return errors.Wrapf(errcode.InternalError, "no QDNS env refers to env(%s)", setReq.Env)
	}

	// upstream 信息
	tags := []string{}
	// 兼容老 tag
	k8sKongTag := s.getK8sKongTag(appServiceName, setReq.Env)
	tags = append(tags, k8sKongTag)
	fullDomainNameWithCluster := s.getAliPrivateZoneK8sFullDomainNameWithCluster(appServiceName, setReq.Env, setReq.ClusterName)
	// 切换成新的统一 tag 集群域名
	k8sKongTag = fullDomainNameWithCluster
	k8sDvDomainName := s.getAliPrivateZoneK8sFullDomainName(appServiceName, entity.AppEnvName(strings.ReplaceAll(
		string(setReq.Env), entity.IstioNamespacePrefix, "")))
	tags = append(tags, fullDomainNameWithCluster, k8sDvDomainName)

	// 校验内网服务是否存在, 这里要考虑 istio 和 ingress 混合部署的问题
	hasInternalService, kongComponents, err := s.checkBusinessHasDeployed(ctx, appServiceName,
		fullDomainNameWithCluster, setReq)
	if err != nil {
		return err
	}

	// 如果 kong 组件不存在, 则创建
	if !hasInternalService {
		// 统一接入规则不存在时创建完整的统一接入规则
		createReq := s.createQDNSCreateRequest(qdnsEnv, operator, setReq, tags, k8sDvDomainName)
		// 新建 qdns 组件,指定 upstream 的 targets 信息
		var validTargetCounter = 0
		for i := range setReq.UpstreamTargetWeights {
			clusterTag := s.getK8sClusterKongTag(setReq.UpstreamTargetWeights[i].ClusterName)
			clusterTags := append(tags, clusterTag)

			if setReq.UpstreamTargetWeights[i].Weight != 0 {
				validTargetCounter++
			}

			createReq.Targets[i] = &req.UpsertQDNSKongTargetReq{
				Target: setReq.UpstreamTargetWeights[i].TargetHostPort,
				Weight: &setReq.UpstreamTargetWeights[i].Weight,
				Tags:   clusterTags,
			}
		}

		// 没有可用的 target ,不进行 kong object 创建
		if validTargetCounter == 0 {
			return nil
		}

		// 如果预期创建的 target 不止一个, 需要开启健康检查中所有项, 否则不开启健康检查
		// NOTE: setReq.ClusterWeights 需要保证不重复, 不然与实际的 target 数量不一致
		if validTargetCounter > 1 {
			fulfillHealthChecks(createReq)
		}

		createResp, e := s.createQDNSBusiness(ctx, createReq)
		if e != nil {
			return e
		}

		if createResp.Status != resp.QDNSSuccessStatus {
			return errors.Wrapf(_errcode.QDNSInternalError,
				"QDNS responses with status(%d) and message(%s)", createResp.GetStatus(), createResp.GetMessage())
		}

		return nil
	}

	// 多个 kong 组件权重同步(多个外网域名以及内网域名同步更新循环更新权重, 支持 ingress 和 istio 互相切换之后的灰度发布)
	for _, component := range kongComponents {
		// 统一接入规则存在时需要判断是否与期望一致, 不一致时需要变更
		// 除了 Target 会有变化之外, 由于 QDNS 并不是原子性操作, 中间过程可能会出现失败, 所以 Service 和 Route 都有可能需要修改
		// Target 由于是 upsert 机制, 所以无需对比确认, 直接操作 upsert 为期望的结果即可
		var targetsToBeModified []*req.UpsertQDNSKongTargetReq
		var validTargetCounter = 0

		for _, item := range setReq.UpstreamTargetWeights {
			clusterTag := s.getK8sClusterKongTag(item.ClusterName)

			if item.Weight != 0 {
				validTargetCounter++
			}

			targetsToBeModified = append(targetsToBeModified, &req.UpsertQDNSKongTargetReq{
				Target: item.TargetHostPort,
				Weight: &item.Weight,
				Tags:   []string{k8sKongTag, clusterTag},
			})
		}

		if validTargetCounter == 0 {
			log.Errorc(ctx, "upstream: %s 更新之后没有可用的 target, 阻止此次更新", component.Name)
			continue
		}

		patchReq := &req.PatchQDNSBusinessReq{
			Env: qdnsEnv,
			Upstream: &req.PatchQDNSKongUpstreamReq{
				ID:      component.BindID,
				Targets: targetsToBeModified,
			},
			UserName: operator,
			// Business: "", // TODO: 如果未来 AMS 区分业务, 则填写
		}

		// 如果预期创建的 target 不止一个, 需要开启健康检查中所有项, 否则不开启健康检查
		// NOTE: setReq.ClusterWeights 需要保证不重复, 不然与实际的 target 数量不一致, 考虑 target weight 为 0 关闭的情况
		if validTargetCounter > 1 {
			patchReq.Upstream.HealthChecks = s.generateQDNSKongUpstreamEnabledHealthChecks(ctx, component,
				setReq.HealthCheckPath)
		} else {
			patchReq.Upstream.HealthChecks = s.generateQDNSKongUpstreamDisabledHealthChecks(ctx, component)
		}

		err = s.patchQDNSBusiness(ctx, patchReq)
		if err != nil {
			log.Errorc(ctx, "%s", err.Error())
			return err
		}
	}

	return nil
}

func (s *Service) createQDNSCreateRequest(qdnsEnv entity.QDNSEnvName, operator string,
	setReq *req.SetAppClusterQDNSWeightsReq, tags []string, k8sDvDomainName string) *req.CreateQDNSBusinessReq {
	return &req.CreateQDNSBusinessReq{
		Env:      qdnsEnv,
		UserName: operator,
		Upstream: &req.CreateQDNSKongUpstreamReq{
			Slots: req.KongaDefaultSlotPtr,
			HealthChecks: &req.KongUpstreamHealthCheck{
				// 先设置 healthchecks.active.timeout 和 healthchecks.active.http_path 两个参数, 否则它们是默认值, 开启健康检查时会有问题
				Active: &req.KongUpstreamActiveHealthCheck{
					HTTPPath: setReq.HealthCheckPath,
					Timeout:  req.AMSKongDefaultHealthCheckTimeoutInSecondPtr,
				},
			},
			Tags: tags,
		},
		Service: &req.CreateQDNSKongServiceReq{
			Retries: req.AMSKongServiceDefaultRetriesPtr,
			Port:    req.KongServiceUnUsefulPort,
			Tags:    tags,
		},
		Routes: []*req.CreateQDNSKongRouteReq{
			{
				Hosts:        []string{k8sDvDomainName},
				Paths:        req.KongDefaultRoutePaths,
				StripPath:    req.AMSKongDefaultStripPath,
				PathHandling: entity.KongPathHandleBehaviorV1,
				Tags:         tags,
				PreserveHost: setReq.EnableIstio.Ptr(),
			},
		},
		Targets: make([]*req.UpsertQDNSKongTargetReq, len(setReq.UpstreamTargetWeights)),
	}
}

func fulfillHealthChecks(createReq *req.CreateQDNSBusinessReq) {
	createReq.Upstream.HealthChecks.Active.Healthy = &req.KongUpstreamActiveHealthCheckHealthyConfig{
		Interval:  req.AMSKongDefaultHealthCheckIntervalPtr,
		Successes: req.AMSKongDefaultHealthCheckSuccessPtr,
	}
	createReq.Upstream.HealthChecks.Active.Unhealthy = &req.KongUpstreamActiveHealthCheckUnhealthyConfig{
		TCPFailures:  req.AMSKongDefaultHealthCheckFailuresPtr,
		Timeouts:     req.AMSKongDefaultHealthCheckTimeoutsPtr,
		HTTPFailures: req.AMSKongDefaultHealthCheckFailuresPtr,
		Interval:     req.AMSKongDefaultHealthCheckIntervalPtr,
	}
	createReq.Upstream.HealthChecks.Passive = &req.KongUpstreamPassiveHealthCheck{
		Healthy: &req.KongUpstreamPassiveHealthCheckHealthyConfig{
			Successes: req.AMSKongDefaultHealthCheckSuccessPtr,
		},
		Unhealthy: &req.KongUpstreamPassiveHealthCheckUnhealthyConfig{
			HTTPFailures: req.AMSKongDefaultHealthCheckFailuresPtr,
			TCPFailures:  req.AMSKongDefaultHealthCheckFailuresPtr,
			Timeouts:     req.AMSKongDefaultHealthCheckTimeoutsPtr,
		},
	}
}

// getK8sQDNSBusinessFromClusterWeights 根据待设置权重信息获取 QDNS 统一接入规则中 k8s 创建的规则
func (s *Service) getK8sQDNSBusinessFromClusterWeights(ctx context.Context,
	appServiceName string, setReq *req.SetAppClusterQDNSWeightsReq, _ string) ([]*resp.GetQDNSBusinessDetailResp, error) {
	envName := setReq.Env

	var targets = []string{
		s.getKongTargetHostAndPort(appServiceName, envName, setReq.ClusterName),
	}

	// 补全 istio 和 ingress 部署的 target  集群私有域名
	if !strings.HasPrefix(string(envName), entity.IstioNamespacePrefix) {
		envName = entity.IstioNamespacePrefix + setReq.Env
		targets = append(targets, s.getKongTargetHostAndPort(appServiceName, envName, setReq.ClusterName))
	} else {
		envName = entity.AppEnvName(strings.ReplaceAll(string(envName), entity.IstioNamespacePrefix, ""))
		s.getKongTargetHostAndPort(appServiceName, envName, setReq.ClusterName)
	}

	list, err := s.getAllQDNSBusinessListByTargets(ctx, targets)
	if err != nil {
		return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	return list, nil
}

// getAllQDNSBusinessListByTargets 从QDNS获取指定targets相关的所有统一接入规则列表
func (s *Service) getAllQDNSBusinessListByTargets(ctx context.Context, targets []string) ([]*resp.GetQDNSBusinessDetailResp, error) {
	var list []*resp.GetQDNSBusinessDetailResp
	// TODO: 清理qdns代码
	return list, nil
	getReq := &req.GetQDNSBusinessListReq{
		Targets:    targets,
		PageNumber: 1,
		PageSize:   req.GetQDNSBusinessPageSizeLimit,
	}
	for {
		cur, err := s.getQDNSBusinessList(ctx, getReq)
		if err != nil {
			return nil, err
		}

		if len(cur) == 0 {
			break
		}

		list = append(list, cur...)
		getReq.PageNumber++
	}

	return list, nil
}

// cleanQDNSBusiness 清理 k8s 应用在 QDNS 统一接入中创建的对应资源
func (s *Service) cleanQDNSBusiness(ctx context.Context, appServiceName string, task *resp.TaskDetailResp) error {
	// TODO: 清理qdns代码
	return nil
	qdnsEnv := s.getKongEnvByAppEnv(task.EnvName)
	if qdnsEnv == entity.QDNSEnvNameUnknown {
		return errors.Wrapf(errcode.InternalError, "no QDNS env refers to env(%s)", task.EnvName)
	}

	operator, err := s.getQDNSUpdater(ctx, task.OperatorID)
	if err != nil {
		return err
	}

	clusters, err := s.GetClusters(ctx, &req.GetClustersReq{
		Namespace: string(task.EnvName),
	})
	if err != nil {
		return err
	}

	targetsToBeModified := make([]*req.UpsertQDNSKongTargetReq, len(clusters))
	targetHostAndPorts := make([]string, len(clusters))
	targetHostAndPortsMapping := make(map[string]struct{})
	for i := range clusters {
		targetHostAndPorts[i] = s.getKongTargetHostAndPort(appServiceName, task.EnvName, clusters[i].Name)
		targetHostAndPortsMapping[targetHostAndPorts[i]] = struct{}{}
		targetsToBeModified[i] = &req.UpsertQDNSKongTargetReq{
			Target: targetHostAndPorts[i],
			Weight: req.ZeroIntPtr, // 删除时权重修改为0, 不需要 tag
		}
	}

	list, err := s.getAllQDNSBusinessListByTargets(ctx, targetHostAndPorts)
	if err != nil {
		return err
	}

	var (
		ok       bool
		patchReq *req.PatchQDNSBusinessReq
	)
	errGroup := errcode.NewGroup(_errcode.ExtraWeightConfigExistsError)
	for _, businessDetail := range list {
		allK8sTargets := true
		for _, target := range businessDetail.Targets {
			if _, ok = targetHostAndPortsMapping[target.Target]; !ok {
				allK8sTargets = false
				break
			}
		}

		if allK8sTargets {
			err = s.deleteQDNSBusiness(ctx, &req.DeleteQDNSBusinessReq{
				Env:      qdnsEnv,
				ID:       businessDetail.ID,
				UserName: operator,
			})
			if err != nil {
				errGroup = errGroup.AddChildren(
					errors.Wrapf(_errcode.QDNSInternalError, "delete business(id=%d) failed: %s", businessDetail.ID, err))
			}
			continue
		}

		patchReq = &req.PatchQDNSBusinessReq{
			Env: businessDetail.Env,
			Upstream: &req.PatchQDNSKongUpstreamReq{
				ID:      businessDetail.BindID,
				Targets: targetsToBeModified,
			},
			UserName: operator,
			// Business: "", // TODO: 如果未来 AMS 区分业务, 则填写
		}
		err = s.patchQDNSBusiness(ctx, patchReq)
		if err != nil {
			errGroup = errGroup.AddChildren(
				errors.Wrapf(_errcode.QDNSInternalError, "remove k8s Targets from business(id=%d) failed: %s", businessDetail.ID, err))
		}
	}

	if len(errGroup.Children()) > 0 {
		return errGroup
	}

	return nil
}

// createQDNSBusiness 创建QDNS统一接入规则
// API 文档: https://git2.qingtingfm.com/devops/qdns/-/blob/master/doc/api.md#统一接入添加接口
func (s *Service) createQDNSBusiness(ctx context.Context, createReq *req.CreateQDNSBusinessReq) (*resp.QDNSStandardResp, error) {
	createReq.ClientToken = uuid.NewV4().String()
	res := new(resp.QDNSStandardResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/kong/add_business", config.Conf.QDNS.Host)).
		Method(http.MethodPost).
		JsonBody(createReq).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	return res, nil
}

// deleteQDNSBusiness 删除QDNS统一接入规则
// API 文档: https://git2.qingtingfm.com/devops/qdns/-/blob/master/doc/api.md#统一接入删除接口
func (s *Service) deleteQDNSBusiness(ctx context.Context, deleteReq *req.DeleteQDNSBusinessReq) error {
	res := new(resp.QDNSStandardResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/kong/delete_business", config.Conf.QDNS.Host)).
		Method(http.MethodDelete).
		JsonBody(deleteReq).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	if res.Status != resp.QDNSSuccessStatus {
		return errors.Wrapf(_errcode.QDNSInternalError,
			"Delete QDNS business failed with status(%d) and message(%s)", res.GetStatus(), res.GetMessage())
	}

	return nil
}

// patchQDNSBusiness 修改QDNS统一接入规则
// API 文档: https://git2.qingtingfm.com/devops/qdns/-/blob/master/doc/api.md#统一接入更新upstream接口
func (s *Service) patchQDNSBusiness(ctx context.Context, patchReq *req.PatchQDNSBusinessReq) error {
	patchReq.ClientToken = uuid.NewV4().String()
	res := new(resp.QDNSStandardResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/kong/update_business_backend", config.Conf.QDNS.Host)).
		Method(http.MethodPatch).
		JsonBody(patchReq).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	if res.Status != resp.QDNSSuccessStatus {
		return errors.Wrapf(_errcode.QDNSInternalError,
			"Patch QDNS business failed with status(%d) and message(%s)", res.GetStatus(), res.GetMessage())
	}

	return nil
}

// getQDNSBusinessList 从QDNS获取统一接入规则列表
// API 文档: https://git2.qingtingfm.com/devops/qdns/-/blob/master/doc/api.md#统一接入查询接口
func (s *Service) getQDNSBusinessList(ctx context.Context,
	listReq *req.GetQDNSBusinessListReq) ([]*resp.GetQDNSBusinessDetailResp, error) {
	// API 传参要求传入数组的 json 格式字符串
	targetsFormat, err := json.Marshal(listReq.Targets)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	var data []*resp.GetQDNSBusinessDetailResp
	res := &resp.QDNSStandardListResp{Data: &data}
	queryParams := httpclient.NewUrlValue().Add("targets", string(targetsFormat))
	if listReq.PageNumber > 0 {
		queryParams.Add("page_number", strconv.Itoa(listReq.PageNumber))
	}
	if listReq.PageSize > 0 {
		queryParams.Add("page_size", strconv.Itoa(listReq.PageSize))
	}
	err = s.httpClient.Builder().
		URL(fmt.Sprintf("%s/kong/describe_business", config.Conf.QDNS.Host)).
		QueryParams(queryParams).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	if res.Status != resp.QDNSSuccessStatus {
		return nil, errors.Wrapf(_errcode.QDNSInternalError,
			"Get QDNS business backend failed with code(%d) and msg(%s)", res.GetStatus(), res.GetMessage())
	}

	return data, nil
}

// updateQDNSKongUpstreamsTargetsHealthy 更新QDNS统一接入多个 Kong Upstream 下多个 Target 健康检查状态
func (s *Service) updateQDNSKongUpstreamsTargetsHealthy(ctx context.Context,
	updateReq *req.UpdateQDNSKongUpstreamsTargetsHealthyReq) error {
	res := new(resp.QDNSStandardResp)
	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/kong/set_target_healthy", config.Conf.QDNS.Host)).
		Method(http.MethodPost).
		JsonBody(updateReq).
		Fetch(ctx).
		DecodeJSON(res)
	if err != nil {
		return errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	if res.GetStatus() != resp.QDNSSuccessStatus {
		return errors.Wrapf(_errcode.QDNSInternalError,
			"Update QDNS Targets Healthy for Upstreams failed with code(%d) and msg(%s)", res.GetStatus(), res.GetMessage())
	}

	return nil
}

// generateQDNSKongUpstreamEnabledHealthChecks 生成 QDNS 统一接入规则开启 Kong Upstream 健康检查参数
func (s *Service) generateQDNSKongUpstreamEnabledHealthChecks(_ context.Context,
	businessDetail *resp.GetQDNSBusinessDetailResp, httpHealthCheckPath string) *req.KongUpstreamHealthCheck {
	var activeHealthy *req.KongUpstreamActiveHealthCheckHealthyConfig
	if businessDetail.ActiveHealthCheck.Healthy.Successes == 0 {
		if activeHealthy == nil {
			activeHealthy = new(req.KongUpstreamActiveHealthCheckHealthyConfig)
		}
		activeHealthy.Successes = req.AMSKongDefaultHealthCheckSuccessPtr
	}

	if businessDetail.ActiveHealthCheck.Healthy.Interval == 0 {
		if activeHealthy == nil {
			activeHealthy = new(req.KongUpstreamActiveHealthCheckHealthyConfig)
		}
		activeHealthy.Interval = req.AMSKongDefaultHealthCheckIntervalPtr
	}

	var activeUnhealthy *req.KongUpstreamActiveHealthCheckUnhealthyConfig
	if businessDetail.ActiveHealthCheck.Unhealthy.HTTPFailures == 0 {
		if activeUnhealthy == nil {
			activeUnhealthy = new(req.KongUpstreamActiveHealthCheckUnhealthyConfig)
		}
		activeUnhealthy.HTTPFailures = req.AMSKongDefaultHealthCheckFailuresPtr
	}

	if businessDetail.ActiveHealthCheck.Unhealthy.TCPFailures == 0 {
		if activeUnhealthy == nil {
			activeUnhealthy = new(req.KongUpstreamActiveHealthCheckUnhealthyConfig)
		}
		activeUnhealthy.TCPFailures = req.AMSKongDefaultHealthCheckFailuresPtr
	}

	if businessDetail.ActiveHealthCheck.Unhealthy.Timeouts == 0 {
		if activeUnhealthy == nil {
			activeUnhealthy = new(req.KongUpstreamActiveHealthCheckUnhealthyConfig)
		}
		activeUnhealthy.Timeouts = req.AMSKongDefaultHealthCheckTimeoutsPtr
	}

	if businessDetail.ActiveHealthCheck.Unhealthy.Interval == 0 {
		if activeUnhealthy == nil {
			activeUnhealthy = new(req.KongUpstreamActiveHealthCheckUnhealthyConfig)
		}
		activeUnhealthy.Interval = req.AMSKongDefaultHealthCheckIntervalPtr
	}

	var passiveHealthy *req.KongUpstreamPassiveHealthCheckHealthyConfig
	if businessDetail.PassiveHealthCheck.Healthy.Successes == 0 {
		if passiveHealthy == nil {
			passiveHealthy = new(req.KongUpstreamPassiveHealthCheckHealthyConfig)
		}
		passiveHealthy.Successes = req.AMSKongDefaultHealthCheckSuccessPtr
	}

	var passiveUnhealthy *req.KongUpstreamPassiveHealthCheckUnhealthyConfig
	if businessDetail.PassiveHealthCheck.Unhealthy.HTTPFailures == 0 {
		if passiveUnhealthy == nil {
			passiveUnhealthy = new(req.KongUpstreamPassiveHealthCheckUnhealthyConfig)
		}
		passiveUnhealthy.HTTPFailures = req.AMSKongDefaultHealthCheckFailuresPtr
	}

	if businessDetail.PassiveHealthCheck.Unhealthy.TCPFailures == 0 {
		if passiveUnhealthy == nil {
			passiveUnhealthy = new(req.KongUpstreamPassiveHealthCheckUnhealthyConfig)
		}
		passiveUnhealthy.TCPFailures = req.AMSKongDefaultHealthCheckFailuresPtr
	}

	if businessDetail.PassiveHealthCheck.Unhealthy.Timeouts == 0 {
		if passiveUnhealthy == nil {
			passiveUnhealthy = new(req.KongUpstreamPassiveHealthCheckUnhealthyConfig)
		}
		passiveUnhealthy.Timeouts = req.AMSKongDefaultHealthCheckTimeoutsPtr
	}

	// healthchecks.active.timeout 的默认值是 1, 此时可能是默认值
	var active *req.KongUpstreamActiveHealthCheck
	if activeHealthy != nil || activeUnhealthy != nil ||
		businessDetail.ActiveHealthCheck.Timeout != *req.AMSKongDefaultHealthCheckTimeoutInSecondPtr {
		active = &req.KongUpstreamActiveHealthCheck{
			Healthy:   activeHealthy,
			Unhealthy: activeUnhealthy,
			Timeout:   req.AMSKongDefaultHealthCheckTimeoutInSecondPtr,
		}

		if businessDetail.ActiveHealthCheck.HTTPPath != httpHealthCheckPath {
			active.HTTPPath = httpHealthCheckPath
		}
	}

	var passive *req.KongUpstreamPassiveHealthCheck
	if passiveHealthy != nil || passiveUnhealthy != nil {
		passive = &req.KongUpstreamPassiveHealthCheck{
			Healthy:   passiveHealthy,
			Unhealthy: passiveUnhealthy,
		}
	}

	if active == nil && passive == nil {
		return nil
	}

	return &req.KongUpstreamHealthCheck{
		Active:  active,
		Passive: passive,
	}
}

// generateQDNSKongUpstreamDisabledHealthChecks 生成 QDNS 统一接入规则关闭 Kong Upstream 健康检查参数
func (s *Service) generateQDNSKongUpstreamDisabledHealthChecks(_ context.Context,
	businessDetail *resp.GetQDNSBusinessDetailResp) *req.KongUpstreamHealthCheck {
	var activeHealthy *req.KongUpstreamActiveHealthCheckHealthyConfig
	if businessDetail.ActiveHealthCheck.Healthy.Successes != 0 {
		if activeHealthy == nil {
			activeHealthy = new(req.KongUpstreamActiveHealthCheckHealthyConfig)
		}
		activeHealthy.Successes = req.ZeroIntPtr
	}

	if businessDetail.ActiveHealthCheck.Healthy.Interval != 0 {
		if activeHealthy == nil {
			activeHealthy = new(req.KongUpstreamActiveHealthCheckHealthyConfig)
		}
		activeHealthy.Interval = req.ZeroInt64Ptr
	}

	var activeUnhealthy *req.KongUpstreamActiveHealthCheckUnhealthyConfig
	if businessDetail.ActiveHealthCheck.Unhealthy.HTTPFailures != 0 {
		if activeUnhealthy == nil {
			activeUnhealthy = new(req.KongUpstreamActiveHealthCheckUnhealthyConfig)
		}
		activeUnhealthy.HTTPFailures = req.ZeroIntPtr
	}

	if businessDetail.ActiveHealthCheck.Unhealthy.TCPFailures != 0 {
		if activeUnhealthy == nil {
			activeUnhealthy = new(req.KongUpstreamActiveHealthCheckUnhealthyConfig)
		}
		activeUnhealthy.TCPFailures = req.ZeroIntPtr
	}

	if businessDetail.ActiveHealthCheck.Unhealthy.Timeouts != 0 {
		if activeUnhealthy == nil {
			activeUnhealthy = new(req.KongUpstreamActiveHealthCheckUnhealthyConfig)
		}
		activeUnhealthy.Timeouts = req.ZeroIntPtr
	}

	if businessDetail.ActiveHealthCheck.Unhealthy.Interval != 0 {
		if activeUnhealthy == nil {
			activeUnhealthy = new(req.KongUpstreamActiveHealthCheckUnhealthyConfig)
		}
		activeUnhealthy.Interval = req.ZeroInt64Ptr
	}

	var passiveHealthy *req.KongUpstreamPassiveHealthCheckHealthyConfig
	if businessDetail.PassiveHealthCheck.Healthy.Successes != 0 {
		if passiveHealthy == nil {
			passiveHealthy = new(req.KongUpstreamPassiveHealthCheckHealthyConfig)
		}
		passiveHealthy.Successes = req.ZeroIntPtr
	}

	var passiveUnhealthy *req.KongUpstreamPassiveHealthCheckUnhealthyConfig
	if businessDetail.PassiveHealthCheck.Unhealthy.HTTPFailures != 0 {
		if passiveUnhealthy == nil {
			passiveUnhealthy = new(req.KongUpstreamPassiveHealthCheckUnhealthyConfig)
		}
		passiveUnhealthy.HTTPFailures = req.ZeroIntPtr
	}

	if businessDetail.PassiveHealthCheck.Unhealthy.TCPFailures != 0 {
		if passiveUnhealthy == nil {
			passiveUnhealthy = new(req.KongUpstreamPassiveHealthCheckUnhealthyConfig)
		}
		passiveUnhealthy.TCPFailures = req.ZeroIntPtr
	}

	if businessDetail.PassiveHealthCheck.Unhealthy.Timeouts != 0 {
		if passiveUnhealthy == nil {
			passiveUnhealthy = new(req.KongUpstreamPassiveHealthCheckUnhealthyConfig)
		}
		passiveUnhealthy.Timeouts = req.ZeroIntPtr
	}

	var active *req.KongUpstreamActiveHealthCheck
	if activeHealthy != nil || activeUnhealthy != nil {
		active = &req.KongUpstreamActiveHealthCheck{
			Healthy:   activeHealthy,
			Unhealthy: activeUnhealthy,
		}
	}

	var passive *req.KongUpstreamPassiveHealthCheck
	if passiveHealthy != nil || passiveUnhealthy != nil {
		passive = &req.KongUpstreamPassiveHealthCheck{
			Healthy:   passiveHealthy,
			Unhealthy: passiveUnhealthy,
		}
	}

	if active == nil && passive == nil {
		return nil
	}

	return &req.KongUpstreamHealthCheck{
		Active:  active,
		Passive: passive,
	}
}

// getQDNSUpdater 获取QDNS操作人
func (s *Service) getQDNSUpdater(ctx context.Context, operatorID string) (string, error) {
	if operatorID == entity.K8sSystemUserID {
		return "ams_系统用户", nil
	}

	_, err := strconv.Atoi(operatorID)
	if err != nil {
		return "", errors.Wrapf(_errcode.K8sInternalError, "Invalid operator_id(%s)", operatorID)
	}

	operator, err := s.GetGitlabUser(ctx, operatorID)
	if err != nil {
		return "", err
	}

	return "ams_" + operator.Name, nil
}

// GetDomainControllerFromProjectOwners 获取域名负责人
// TODO: 当前以第一个项目负责人替代域名负责人，未来与运维商讨负责人机制
func (s *Service) GetDomainControllerFromProjectOwners(owners []*resp.UserProfileResp) string {
	if len(owners) > 0 {
		return owners[0].Name
	}

	return ""
}

// getKongEnv 获取 AMS 环境及命名空间对应的 Kong 集群环境
// 目前状态:
//  1. dev(AMS stg) 下无论哪个子环境均从 kong-dev 接入, 所有 AMS 环境变量非 prd 情况均认为是 dev
//  2. stg(AMS prd) 从 kong-stg 接入
//  3. prd(AMS prd) 不止一个 Kong 集群, AMS prd 环境的统一接入规则只从 kong-int 接入
func (s *Service) getKongEnvByAppEnv(appEnv entity.AppEnvName) entity.QDNSEnvName {
	if config.Conf.Env != "prd" {
		return entity.QDNSEnvNameDev
	}

	switch appEnv {
	case entity.AppEnvFat, entity.AppEnvStg:
		return entity.QDNSEnvNameStg

	case entity.AppEnvPre, entity.AppEnvPrd:
		return entity.QDNSEnvNameInt

	default:
	}

	return entity.QDNSEnvNameUnknown
}

// getKongAliasName 获取 Kong 集群别名
func getKongAliasName(qdnsEnv entity.QDNSEnvName) string {
	switch qdnsEnv {
	case entity.QDNSEnvNameDev:
		return "kong-dev"

	case entity.QDNSEnvNameStg:
		return "kong-stg"

	case entity.QDNSEnvNameMain:
		return "kong-main"

	case entity.QDNSEnvNamePortal:
		return "kong-portal"

	case entity.QDNSEnvNameInt:
		return "kong-int"

	default:
	}

	return "unknown"
}

func removeDuplicateItem(s []string) []string {
	result := make([]string, 0, len(s))
	temp := map[string]struct{}{}
	for _, item := range s {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// GetBusinessByTag 根据Tag 和 KongClusterName 获取 QDNS tag 对应的资源
func (s *Service) GetBusinessByTag(ctx context.Context, tagName string, appEnvName entity.AppEnvName) (
	*resp.GetTagResponse, error) {
	var result = new(resp.GetTagResponse)
	queryParams := httpclient.NewUrlValue().Add("tag", tagName).Add("env", string(s.getKongEnvByAppEnv(
		appEnvName)))

	err := s.httpClient.Builder().
		URL(fmt.Sprintf("%s/kong/get_tags", config.Conf.QDNS.Host)).
		QueryParams(queryParams).
		Method(http.MethodGet).
		Fetch(ctx).
		DecodeJSON(result)
	if err != nil {
		return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
	}

	if result.Status != resp.QDNSSuccessStatus {
		return nil, errors.Wrapf(_errcode.QDNSInternalError,
			"Get QDNS business backend failed with code(%d) and msg(%s)", result.Status, result.Msg)
	}

	return result, nil
}

func (s *Service) checkBusinessHasDeployed(ctx context.Context, appServiceName, fullDomainNameWithCluster string,
	setReq *req.SetAppClusterQDNSWeightsReq) (bool, []*resp.GetQDNSBusinessDetailResp, error) {
	// 使用旧版本根据 target 反查 service 的方式确定是否存在 service, route, upstream
	// 这里要考虑混合部署的方式

	k8sBusiness, err := s.getK8sQDNSBusinessFromClusterWeights(ctx, appServiceName, setReq, fullDomainNameWithCluster)
	if err != nil {
		return false, nil, err
	}

	for _, business := range k8sBusiness {
		for _, s := range business.Service {
			if slice.StrSliceContains(s.Tags, fullDomainNameWithCluster) {
				return true, k8sBusiness, nil
			}
		}
	}

	return false, nil, nil
}
