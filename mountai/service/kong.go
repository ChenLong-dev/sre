// AMS 与蜻蜓 Kong 网关集成
// 当前使用的 Kong 网关版本: 1.5.1
// 对应 admin API 文档：https://docs.konghq.com/enterprise/1.5.x/admin-api/#upstream-object
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/httpclient"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

// Kong tag 格式
const (
	KongK8sTagFormatter     = "k8s-%s-%s"    // k8s-${service_name}-${env_name}
	KongK8sClusterTagPrefix = "k8s-cluster-" // k8s-${cluster_name}
	KongIstioUpstreamName   = "k8s-cluster-istio-upstream"
	KongDefaultPageSize     = 50 // 默认单页记录数
)

// createKongPrivateZoneRecordForK8sDomain 为 k8s 通配域名添加 Kong 私有域解析记录
func (s *Service) createKongPrivateZoneRecordForK8sDomain(ctx context.Context,
	envName entity.AppEnvName, app *resp.AppDetailResp, task *resp.TaskDetailResp, domainController string) error {
	updater, err := s.getQDNSUpdater(ctx, task.OperatorID)
	if err != nil {
		return err
	}

	domainName := s.getAliPrivateZoneK8sDomainName(app.ServiceName, envName)
	// LB 类型的 Service 通配域名解析到 LB 的 IP, Ingress 类型解析到 Kong 集群对应 LB 的 CNAME, 其他不处理
	switch app.ServiceExposeType {
	case entity.AppServiceExposeTypeLB:
		ip, e := s.GetServiceLoadBalancerIP(ctx, task.ClusterName, app.ServiceName, string(task.EnvName))
		if e != nil {
			return e
		}

		return s.createAliPrivateZoneRecord(ctx, domainName, ip, updater, domainController, entity.ARecord)

	case entity.AppServiceExposeTypeIngress:
		return s.createAliPrivateZoneRecord(ctx, domainName, config.Conf.Kong.Envs[string(envName)].Address,
			updater, domainController, entity.CNAMERecord)

	default:
	}

	return nil
}

// deleteKongPrivateZoneRecordForK8sDomain 删除 k8s 通配域名的 Kong 私有域解析记录
func (s *Service) deleteKongPrivateZoneRecordForK8sDomain(ctx context.Context,
	serviceExposeType entity.AppServiceExposeType, appServiceName string, task *resp.TaskDetailResp) error {
	updater, err := s.getQDNSUpdater(ctx, task.OperatorID)
	if err != nil {
		return err
	}

	domainName := s.getAliPrivateZoneK8sDomainName(appServiceName, task.GetEnvName())
	// LB 类型的 Service 通配域名删除 LB 的 IP, Ingress 类型删除 Kong 集群对应 LB 的 CNAME, 其他不处理
	switch serviceExposeType {
	case entity.AppServiceExposeTypeLB:
		ip, e := s.GetServiceLoadBalancerIP(ctx, task.ClusterName, appServiceName, string(task.EnvName))
		if e != nil {
			return e
		}

		return s.deleteAliPrivateZoneRecord(ctx, domainName, ip, updater, entity.ARecord)

	case entity.AppServiceExposeTypeIngress:
		return s.deleteAliPrivateZoneRecord(ctx, domainName, config.Conf.Kong.Envs[string(task.EnvName)].Address,
			updater, entity.CNAMERecord)

	default:
	}

	return nil
}

func (s *Service) getKongTargetHostAndPort(appServiceName string, envName entity.AppEnvName, clusterName entity.ClusterName) string {
	clusterSubDomain := s.getAliPrivateZoneK8sDomainNameWithCluster(appServiceName, envName, clusterName)
	return fmt.Sprintf("%s.%s:%d", clusterSubDomain, AliK8sPrivateZone, entity.ServiceDefaultInternalPort)
}

func (s *Service) getK8sKongTag(appServiceName string, envName entity.AppEnvName) string {
	return fmt.Sprintf(KongK8sTagFormatter, appServiceName, envName)
}

func (s *Service) getK8sClusterKongTag(clusterName entity.ClusterName) string {
	return KongK8sClusterTagPrefix + string(clusterName)
}

// getK8sClusterNameFromKongTags 从 Kong tags 中获取集群名
// 返回值依次为: 是否带有指定 k8s Kong tag, 解析得到的集群名, 错误值
func (s *Service) getK8sClusterNameFromKongTags(tags []string, k8sKongTag string) (bool, entity.ClusterName, error) {
	var clusterName entity.ClusterName
	containsK8sKongTag := false
	for _, tag := range tags {
		if strings.HasPrefix(tag, KongK8sClusterTagPrefix) {
			clusterName = entity.ClusterName(tag[len(KongK8sClusterTagPrefix):])
		} else if tag == k8sKongTag {
			containsK8sKongTag = true
		}

		if containsK8sKongTag && clusterName != "" {
			break
		}
	}

	if containsK8sKongTag && clusterName == "" {
		return containsK8sKongTag, clusterName, errors.Wrap(_errcode.KongConfigNotFoundError, "no cluster tag in tags")
	}

	return containsK8sKongTag, clusterName, nil
}

// GetKongUpstreamTargets returns a bunch of targets under the given upstream resource id
func (s *Service) GetKongUpstreamTargets(ctx context.Context, getReq *req.GetKongUpstreamTargetsReq) (
	*resp.GetKongTargetsResp, error) {
	res := new(resp.GetKongTargetsResp)

	queryParams := httpclient.NewUrlValue()
	if getReq.Offset != "" {
		queryParams.Add("offset", getReq.Offset)
	}

	if getReq.Size == "" {
		getReq.Size = strconv.Itoa(KongDefaultPageSize)
	}

	queryParams.Add("size", getReq.Size)

	hosts := getKongAPIHosts(string(getReq.EnvName))
	for _, host := range hosts {
		err := s.httpClient.Builder().
			URL(fmt.Sprintf("%s/upstreams/%s/targets", host, getReq.UpstreamName)).
			QueryParams(queryParams).
			Method(http.MethodGet).
			Fetch(ctx).
			DecodeJSON(res)

		if err != nil {
			return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
		}
	}
	return res, nil
}

// GetKongUpstreams returns a bunch of upstreams
func (s *Service) GetKongUpstreams(ctx context.Context, getReq *req.GetKongUpstreamsReq) (
	*resp.GetKongUpstreamsResp, error) {
	res := new(resp.GetKongUpstreamsResp)

	queryParams := httpclient.NewUrlValue().Add("tags", getReq.Tags)
	if getReq.Offset != "" {
		queryParams.Add("offset", getReq.Offset)
	}

	if getReq.Size == "" {
		getReq.Size = strconv.Itoa(KongDefaultPageSize)
	}

	queryParams.Add("size", getReq.Size)

	for _, host := range getKongAPIHosts(string(getReq.EnvName)) {
		err := s.httpClient.Builder().
			URL(fmt.Sprintf("%s/upstreams", host)).
			QueryParams(queryParams).
			Method(http.MethodGet).
			Fetch(ctx).
			DecodeJSON(res)

		if err != nil {
			return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
		}
	}
	return res, nil
}

// GetKongUpstreams returns a bunch of upstreams
func (s *Service) GetKongServices(ctx context.Context, getReq *req.GetKongServicesReq) (
	*resp.GetKongServicesResp, error) {
	res := new(resp.GetKongServicesResp)

	queryParams := httpclient.NewUrlValue().Add("tags", getReq.Tags)
	if getReq.Offset != "" {
		queryParams.Add("offset", getReq.Offset)
	}

	if getReq.Size == "" {
		getReq.Size = strconv.Itoa(KongDefaultPageSize)
	}

	queryParams.Add("size", getReq.Size)

	for _, host := range getKongAPIHosts(getReq.EnvName) {
		err := s.httpClient.Builder().
			URL(fmt.Sprintf("%s/services", host)).
			QueryParams(queryParams).
			Method(http.MethodGet).
			Fetch(ctx).
			DecodeJSON(res)
		if err != nil {
			return nil, errors.Wrap(_errcode.QDNSInternalError, err.Error())
		}
	}
	return res, nil
}

// DeleteKongTarget deleted specific target from upstream
func (s *Service) DeleteKongTarget(ctx context.Context, deleteReq *req.DeleteKongTargetReq) error {
	for _, host := range getKongAPIHosts(deleteReq.EnvName) {
		r := s.httpClient.Builder().
			URL(fmt.Sprintf("%s/upstreams/%s/targets/%s", host, deleteReq.UpstreamName, deleteReq.Host)).
			Method(http.MethodDelete).
			AccessStatusCode(http.StatusNoContent).
			Fetch(ctx)

		if r.StatusCode == http.StatusNoContent {
			log.Infoc(ctx, "Deleted target #%s for upstream #%s.", deleteReq.Host, deleteReq.UpstreamName)
		}
	}

	return nil
}

// CreateKongTarget 在指定的 upstream 中创建 kong target
func (s *Service) CreateKongTarget(ctx context.Context, createReq *req.CreateKongTargetReq) error {
	data, err := json.Marshal(map[string]string{
		"target": createReq.Target,
	})
	if err != nil {
		return err
	}

	hosts := getKongAPIHosts(createReq.EnvName)

	// 多集群同步...
	for _, host := range hosts {
		r := s.httpClient.Builder().
			URL(fmt.Sprintf("%s/upstreams/%s/targets/", host, createReq.UpstreamName)).
			Method(http.MethodPost).
			Body(data).
			AccessStatusCode(http.StatusCreated).
			Fetch(ctx)

		if r.StatusCode == http.StatusCreated {
			log.Infoc(ctx, "Created new target for upstream #%s with target %s", createReq.UpstreamName,
				createReq.Target)
		}
	}

	return nil
}

// getKongConfig returns specific env kong config
func getKongConfig(env string) (*config.KongEnvConfig, error) {
	if v, ok := config.Conf.Kong.Envs[env]; ok {
		return v, nil
	}
	return nil, _errcode.KongConfigNotFoundError
}

// return kong cluster manage api
func getKongAPIHosts(env string) []string {
	c, err := getKongConfig(env)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return distinctStringSlice(c.AdminHost)
}

// returns unique string elements from a string slice
func distinctStringSlice(s []string) []string {
	if len(s) == 1 {
		return s
	}

	result := make([]string, 0, len(s))
	filter := make(map[string]bool)
	for _, v := range s {
		if _, ok := filter[v]; ok {
			continue
		}
		filter[v] = true
		result = append(result, v)
	}

	return result
}
