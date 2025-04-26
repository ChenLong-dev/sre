package service

import (
	"encoding/json"
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/pvtz"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

const (
	// 阿里云容器仓库域名
	AliDockerRepoDomain = "cr.cn-shanghai.aliyuncs.com"
	// 阿里云容器仓库命名空间
	AliDockerRepoNamespace = "shanhaii"

	// 阿里云k8s私有域
	// 接入 QDNS 时改变一下命名，避免未来出现多云时命名冲突
	AliK8sPrivateZone = "cluster.local"

	// 阿里云云解析记录开启
	AliPrivateZoneRecordEnable = "ENABLE"
)

var (
	// 阿里云私有域id本地缓存
	aliPrivateZoneIDLocalCache = make(map[string]string)
)

// 获取阿里云私有域K8s域名
func (s *Service) getAliPrivateZoneK8sDomainName(serviceName string, envName entity.AppEnvName) string {
	return fmt.Sprintf("%s%s.%s.svc", config.Conf.Ali.PrivateZonePrefix, serviceName, envName)
}

// getAliPrivateZoneK8sDomainNameWithCluster 获取阿里云私有域带集群名的K8s域名
func (s *Service) getAliPrivateZoneK8sDomainNameWithCluster(serviceName string,
	envName entity.AppEnvName, clusterName entity.ClusterName) string {
	return fmt.Sprintf("%s%s.%s.svc.%s", config.Conf.Ali.PrivateZonePrefix, serviceName, envName, clusterName)
}

// 获取阿里云私有域K8s完整域名
func (s *Service) getAliPrivateZoneK8sFullDomainName(serviceName string, envName entity.AppEnvName) string {
	return fmt.Sprintf("%s.%s", s.getAliPrivateZoneK8sDomainName(serviceName, envName), AliK8sPrivateZone)
}

// getAliPrivateZoneK8sFullDomainNameWithCluster 获取阿里云私有域带集群名的K8s完整域名
func (s *Service) getAliPrivateZoneK8sFullDomainNameWithCluster(serviceName string,
	envName entity.AppEnvName, clusterName entity.ClusterName) string {
	return fmt.Sprintf("%s.%s", s.getAliPrivateZoneK8sDomainNameWithCluster(serviceName, envName, clusterName), AliK8sPrivateZone)
}

// 获取阿里云私有域列表
func (s *Service) AliDescribeZones(ctx context.Context, getReq *req.AliDescribeZonesReq) (*pvtz.DescribeZonesResponse, error) {
	aliReq := pvtz.CreateDescribeZonesRequest()
	aliReq.Keyword = getReq.Keyword

	if getReq.Size != 0 {
		aliReq.PageSize = requests.NewInteger(getReq.Size)
	}

	if getReq.Page != 0 {
		aliReq.PageNumber = requests.NewInteger(getReq.Page)
	}

	aliResp := pvtz.CreateDescribeZonesResponse()
	err := s.wrapAliDoAction(ctx, aliReq, aliResp)

	if err != nil {
		return nil, err
	}

	return aliResp, nil
}

// 获取阿里云私有域解析记录
func (s *Service) AliDescribeZoneRecords(ctx context.Context,
	getReq *req.AliDescribeZoneRecordsReq) (*pvtz.DescribeZoneRecordsResponse, error) {
	aliReq := pvtz.CreateDescribeZoneRecordsRequest()
	aliReq.Keyword = getReq.Keyword
	aliReq.ZoneId = getReq.ZoneID

	if getReq.Size != 0 {
		aliReq.PageSize = requests.NewInteger(getReq.Size)
	}

	if getReq.Page != 0 {
		aliReq.PageNumber = requests.NewInteger(getReq.Page)
	}

	aliResp := pvtz.CreateDescribeZoneRecordsResponse()
	err := s.wrapAliDoAction(ctx, aliReq, aliResp)

	if err != nil {
		return nil, err
	}

	return aliResp, nil
}

// createPrivateZoneRecordEntryForClusterDomain 为集群专用域名创建阿里云 private zone 解析记录统一入口
func (s *Service) createPrivateZoneRecordEntryForClusterDomain(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	switch app.ServiceExposeType {
	case entity.AppServiceExposeTypeLB:
		// 如果是通过 LB 方式暴露服务，需要创建 LB 解析
		return s.createLBPrivateZoneRecordForClusterDomain(ctx, project, app, task)

	case entity.AppServiceExposeTypeIngress:
		// 如果是通过 Ingress/或者 Istio IngressGateway 方式暴露服务，需要创建 DNS 解析
		if s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app) {
			return s.createIstioPrivateZoneRecord(ctx, project, app, task)
		}
		return s.createIngressPrivateZoneRecordForClusterDomain(ctx, project, app, task)

	case entity.AppServiceExposeTypeInternal:
		return nil

	default:
	}

	return nil
}

// createPrivateZoneRecordEntry 为所有域名创建阿里云 private zone 解析记录统一入口
func (s *Service) createPrivateZoneRecordEntry(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp) error {
	switch app.ServiceExposeType {
	case entity.AppServiceExposeTypeLB:
		// 如果是通过 LB 方式暴露服务，需要创建 LB 解析
		return s.createLBPrivateZoneRecord(ctx, project, app, task)
	case entity.AppServiceExposeTypeIngress:
		// 如果启用了 istio 接入,创建 Gateway 解析
		if s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app) {
			return s.createIstioPrivateZoneRecord(ctx, project, app, task)
		}
		// 如果是通过 Ingress 方式暴露服务，需要创建 Ingress 解析
		return s.createIngressPrivateZoneRecord(ctx, project, app, task)
	case entity.AppServiceExposeTypeInternal:
		return nil
	}
	return nil
}

// deletePrivateZoneRecordEntry 删除阿里云 private zone 解析记录统一入口
func (s *Service) deletePrivateZoneRecordEntry(ctx context.Context, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	switch app.ServiceExposeType {
	case entity.AppServiceExposeTypeLB:
		// 如果是通过 LB 方式暴露服务，需要删除 LB 解析
		return s.deleteLBPrivateZoneRecord(ctx, app.ServiceName, task)
	case entity.AppServiceExposeTypeIngress:
		// 如果是通过 Ingress 方式暴露服务，需要删除 Ingress 解析
		return s.deleteIngressPrivateZoneRecord(ctx, task.ClusterName, app.ServiceName, task)
	case entity.AppServiceExposeTypeInternal:
		return nil
	}
	return nil
}

// createAliPrivateZoneRecord 添加阿里云的PrivateZone解析记录
func (s *Service) createAliPrivateZoneRecord(ctx context.Context, domainName, domainValue, updater,
	domainController string, domainType entity.DomainRecordType) error {
	createReq := &req.CreateQDNSRecordReq{
		DomainType:       domainType,
		DomainRecordName: domainName,
		DomainValue:      domainValue,
		DomainName:       AliK8sPrivateZone,
		DomainUpdater:    updater,
	}

	if domainController != "" {
		createReq.DomainController = domainController
	}

	res, err := s.CreateQDNSRecord(ctx, createReq)
	if err != nil {
		return err
	}

	// 重复域名情况下操作实际是成功的，忽略该响应，域名重复可以从 httpclient 的响应中排查
	if resp.EqualQDNSErrorCase(resp.QDNSStatusDuplicateRecordResp, res) {
		return nil
	}

	if res.Status != 0 {
		return errors.Wrapf(_errcode.QDNSInternalError, "Create QDNS record failed with code(%d) and msg(%s)", res.Status, res.Msg)
	}

	return nil
}

// deleteAliPrivateZoneRecord 删除阿里云的PrivateZone解析记录
func (s *Service) deleteAliPrivateZoneRecord(ctx context.Context, domainName, domainValue,
	updater string, domainType entity.DomainRecordType) error {
	res, err := s.DeleteQDNSRecord(ctx, &req.DeleteQDNSRecordReq{
		DomainType:       domainType,
		DomainRecordName: domainName,
		DomainValue:      domainValue,
		DomainName:       AliK8sPrivateZone,
		DomainUpdater:    updater,
	})
	if err != nil {
		return err
	}

	// 兼容 QDNS 尚未同步某些 k8s 域名记录的情形
	// TODO: 去除兼容时，应当忽略 resp.QDNSStatusRecordNotFoundResp 错误(记录日志)
	if resp.EqualQDNSErrorCase(resp.QDNSStatusRecordNotFoundResp, res) {
		log.Warnc(ctx, "Domain(%s) with domain value(%s) not found in QDNS. Fallback to ali API.", domainName, domainValue)

		_, records, err := s.searchAliPrivateZoneRecord(ctx, domainName, AliK8sPrivateZone)
		if err != nil {
			return err
		}

		// 筛选记录，如存在，则删除
		for _, record := range records.Records.Record {
			if record.Status == AliPrivateZoneRecordEnable && record.Rr == domainName && record.Value == domainValue {
				_, err = s.AliDeleteZoneRecord(ctx, int(record.RecordId))
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	if res.Status != 0 {
		return errors.Wrapf(
			_errcode.QDNSInternalError,
			"Delete QDNS record failed with code(%d) and msg(%s)",
			res.Status, res.Msg)
	}

	return nil
}

// createLBPrivateZoneRecordForClusterDomain 为集群专用域名创建 LB 解析
func (s *Service) createLBPrivateZoneRecordForClusterDomain(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	address, err := s.GetServiceLoadBalancerIP(ctx, task.ClusterName, app.ServiceName, string(task.EnvName))

	if err != nil {
		return err
	}

	if address == "" {
		log.Errorc(ctx, "LB expose type service(%s) load balancer IP is empty", app.ServiceName)
		return nil
	}

	updater, err := s.getQDNSUpdater(ctx, task.OperatorID)
	if err != nil {
		return err
	}

	domainName := s.getAliPrivateZoneK8sDomainNameWithCluster(app.ServiceName, task.EnvName, task.ClusterName)
	domainController := s.GetDomainControllerFromProjectOwners(project.Owners)
	err = s.createAliPrivateZoneRecord(ctx, domainName, address, updater, domainController, entity.ARecord)
	if err != nil {
		return err
	}

	return nil
}

// createLBPrivateZoneRecord 为所有域名创建 LB 解析
func (s *Service) createLBPrivateZoneRecord(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp) error {
	address, err := s.GetServiceLoadBalancerIP(ctx, task.ClusterName, app.ServiceName, string(task.EnvName))

	if err != nil {
		return err
	}

	if address == "" {
		log.Errorc(ctx, "LB expose type service(%s) load balancer IP is empty", app.ServiceName)
		return nil
	}

	updater, err := s.getQDNSUpdater(ctx, task.OperatorID)
	if err != nil {
		return err
	}

	domainName := s.getAliPrivateZoneK8sDomainName(app.ServiceName, task.GetEnvName())
	domainController := s.GetDomainControllerFromProjectOwners(project.Owners)
	err = s.createAliPrivateZoneRecord(ctx, domainName, address, updater, domainController, entity.ARecord)
	if err != nil {
		return err
	}

	domainName = s.getAliPrivateZoneK8sDomainNameWithCluster(app.ServiceName, task.GetEnvName(), task.ClusterName)
	err = s.createAliPrivateZoneRecord(ctx, domainName, address, updater, domainController, entity.ARecord)
	if err != nil {
		return err
	}

	return nil
}

// deleteLBPrivateZoneRecord 删除通配域名和 k8s 集群专用域名 LB 解析
func (s *Service) deleteLBPrivateZoneRecord(ctx context.Context, serviceName string,
	task *resp.TaskDetailResp) error {
	address, err := s.GetServiceLoadBalancerIP(ctx, task.ClusterName, serviceName, string(task.EnvName))

	if err != nil {
		if errcode.EqualError(_errcode.NotFindK8sServiceIngressError, err) {
			// Service 的 IP 没有就绪, 理论上没有解析, 但留个日志以便检查
			log.Warnc(ctx, "skipped delete phase of service(%s) loadbalancer private zone record by err: %s", serviceName, err)
			return nil
		}

		return err
	}

	if address == "" {
		log.Errorc(ctx, "LB expose type service(%s) load balancer IP is empty", serviceName)
		return nil
	}

	var updater string
	updater, err = s.getQDNSUpdater(ctx, task.OperatorID)
	if err != nil {
		return err
	}

	domainName := s.getAliPrivateZoneK8sDomainName(serviceName, task.GetEnvName())
	err = s.deleteAliPrivateZoneRecord(ctx, domainName, address, updater, entity.ARecord)
	if err != nil {
		return err
	}

	domainName = s.getAliPrivateZoneK8sDomainNameWithCluster(serviceName, task.EnvName, task.ClusterName)
	err = s.deleteAliPrivateZoneRecord(ctx, domainName, address, updater, entity.ARecord)
	if err != nil {
		return err
	}

	return nil
}

// checkPrivateZoneRecordExistanceForK8sDomain 检验 k8s 通配域名在 privateZone 是否存在解析记录
func (s *Service) checkPrivateZoneRecordExistanceForK8sDomain(ctx context.Context,
	envName entity.AppEnvName, appServiceName string) (bool, error) {
	domainName := s.getAliPrivateZoneK8sDomainName(appServiceName, envName)
	records, err := s.getAliPrivateZoneRecord(ctx, &req.GetQDNSRecordsReq{
		DomainRecordName: domainName,
		PrivateZone:      AliK8sPrivateZone,
		PageNumber:       1,
		PageSize:         req.GetQDNSRecordsPageSizeLimit,
	})

	if err != nil {
		return false, err
	}

	return len(records) > 0, nil
}

// getAliPrivateZoneRecord 获取阿里云PrivateZone解析记录
func (s *Service) getAliPrivateZoneRecord(ctx context.Context, getReq *req.GetQDNSRecordsReq) ([]*entity.QDNSRecord, error) {
	// 当前机制下单个IP地址的记录数(<=集群数)离分页上限很遥远，不需要多次调用
	return s.GetQDNSRecords(ctx, getReq)
}

// searchAliPrivateZoneRecord:查找阿里云私有域记录
// keyword: address or domain name
func (s *Service) searchAliPrivateZoneRecord(ctx context.Context, keyword,
	zoneName string) (string, *pvtz.DescribeZoneRecordsResponse, error) {
	zoneID, err := s.GetAliPrivateZoneID(ctx, zoneName)
	if err != nil {
		return "", nil, err
	}

	// 获取具体分区记录
	records, err := s.AliDescribeZoneRecords(ctx, &req.AliDescribeZoneRecordsReq{
		ZoneID:  zoneID,
		Keyword: keyword,
	})
	if err != nil {
		return "", nil, err
	}

	return zoneID, records, nil
}

// 删除阿里云私有域解析日志
func (s *Service) AliDeleteZoneRecord(ctx context.Context, recordID int) (*pvtz.DeleteZoneRecordResponse, error) {
	aliReq := pvtz.CreateDeleteZoneRecordRequest()
	aliReq.RecordId = requests.NewInteger(recordID)

	aliResp := pvtz.CreateDeleteZoneRecordResponse()
	err := s.wrapAliDoAction(ctx, aliReq, aliResp)

	if err != nil {
		return nil, err
	}

	return aliResp, nil
}

// 包装阿里云操作
func (s *Service) wrapAliDoAction(ctx context.Context, aliReq requests.AcsRequest, aliResp responses.AcsResponse) error {
	startTime := time.Now()
	msgMap := map[string]interface{}{
		"title":      "AlibabaCloud",
		"action":     aliReq.GetActionName(),
		"req":        aliReq,
		"resp":       aliResp,
		"start_time": startTime,
		"duration":   time.Since(startTime).Truncate(time.Millisecond).String(),
	}

	err := s.aliClient.DoAction(aliReq, aliResp)
	if err != nil {
		msgMap["error"] = err
		log.Errorv(ctx, msgMap)

		return errors.Wrap(_errcode.AliSDKInternalError, err.Error())
	}

	log.Infov(ctx, msgMap)

	if !aliResp.IsSuccess() {
		return errors.Wrapf(_errcode.AliSDKResponseError, "%s", aliResp)
	}

	return nil
}

// GetAliPrivateZoneID:获取阿里云私有域id
func (s *Service) GetAliPrivateZoneID(ctx context.Context, zoneName string) (string, error) {
	// 获取缓存
	zoneID, ok := aliPrivateZoneIDLocalCache[zoneName]
	if ok {
		return zoneID, nil
	}

	// 获取分区
	zones, err := s.AliDescribeZones(ctx, &req.AliDescribeZonesReq{
		Keyword: zoneName,
	})
	if err != nil {
		return "", err
	}

	// 定位具体的分区id
	for i := range zones.Zones.Zone {
		zone := zones.Zones.Zone[i]
		if zone.ZoneName == zoneName {
			zoneID = zone.ZoneId
			break
		}
	}

	if zoneID == "" {
		return "", errors.Wrapf(_errcode.NotFindAliPrivateZoneError, "zone name:%s", zoneName)
	}

	aliPrivateZoneIDLocalCache[zoneName] = zoneID

	return zoneID, nil
}

// 获取阿里云镜像标签
func (s *Service) AliGetRepoTags(ctx context.Context, getReq *req.AliGetRepoTagsReq) (*resp.AliGetDockerTagsResp, error) {
	aliReq := cr.CreateGetRepoTagsRequest()
	aliReq.Domain = AliDockerRepoDomain
	aliReq.RepoName = getReq.ProjectName
	aliReq.RepoNamespace = AliDockerRepoNamespace

	if getReq.Size != 0 {
		aliReq.PageSize = requests.NewInteger(getReq.Size)
	}

	if getReq.Page != 0 {
		aliReq.Page = requests.NewInteger(getReq.Page)
	}

	aliResp := cr.CreateGetRepoTagsResponse()
	err := s.wrapAliDoAction(ctx, aliReq, aliResp)

	if err != nil {
		return nil, err
	}

	res := new(resp.AliGetDockerTagsResp)
	err = json.Unmarshal(aliResp.GetHttpContentBytes(), res)

	if err != nil {
		return nil, errors.Wrap(_errcode.AliSDKResponseFormatError, err.Error())
	}

	return res, nil
}

// 获取镜像仓库地址
func (s *Service) GetAliImageRepoURL(projectName, tag string) string {
	return fmt.Sprintf("crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/shanhaii/%s:%s", projectName, tag)
}
