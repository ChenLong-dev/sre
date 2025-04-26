package service

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes/scheme"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

// ApplyVirtualServiceFromTpl 使用模板创建 VirtualService 对象
func (s *Service) ApplyVirtualServiceFromTpl(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, backendServiceName string) (*v1beta1.VirtualService, error) {
	data, err := s.RenderVirtualServiceTemplate(ctx, project, app, task, backendServiceName)
	if err != nil {
		return nil, err
	}

	return s.ApplyVirtualService(ctx, task.ClusterName, task.EnvName, []byte(data))
}

// DeleteVirtualService 删除指定的 VirtualService 对象
func (s *Service) DeleteVirtualService(ctx context.Context, clusterName entity.ClusterName, task *resp.TaskDetailResp,
	backendServiceName, namespace string) error {
	info, err := s.getClusterInfo(clusterName, string(task.EnvName))
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	ic, err := versionedclient.NewForConfig(info.config)
	if err != nil {
		return errors.Wrapf(_errcode.K8sInternalError, err.Error())
	}

	return ic.NetworkingV1beta1().VirtualServices(namespace).
		Delete(ctx, backendServiceName, metav1.DeleteOptions{})
}

// RenderVirtualServiceTemplate 渲染 VirtualService 模板
func (s *Service) RenderVirtualServiceTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, backendServiceName string) (string, error) {
	// 当前kong存在bug, 需要ingress包含集群所有的域名, 做兼容处理
	clusters := config.Conf.K8sClusters[string(task.EnvName)]

	tpl := &entity.VirtualServiceTemplate{
		Name:        project.Name + "-" + app.Name,
		Namespace:   task.Namespace,
		ProjectName: project.Name,
		AppName:     app.Name,
		Annotations: map[string]string{
			"createTime": time.Now().Format("2006-01-02 15:04:05"),
		},
		IstioGateway:            entity.DefaultIstioGateway,
		ServiceHostsWithCluster: []string{},
		ServiceName:             backendServiceName,
	}

	// 内网域名处理 {{{
	for _, cluster := range clusters {
		tpl.ServiceHostsWithCluster = append(tpl.ServiceHostsWithCluster,
			s.getAliPrivateZoneK8sFullDomainNameWithCluster(app.ServiceName, task.GetEnvName(), entity.ClusterName(cluster.Name)))
	}

	// 多集群通配域名和 ingress 公用
	tpl.ServiceHostsWithCluster = append(tpl.ServiceHostsWithCluster,
		s.getAliPrivateZoneK8sFullDomainName(backendServiceName, task.EnvName))
	// }}}

	return s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
}

// ApplyVirtualService 通过声明式 API 创建 VirtualService 对象, 如果存在则更新 VirtualService 对象
func (s *Service) ApplyVirtualService(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName,
	data []byte) (*v1beta1.VirtualService, error) {
	applyVirtualService, err := s.decodeVirtualServiceYamlData(ctx, data)
	if err != nil {
		return nil, err
	}

	_, err = s.GetVirtualServiceDetail(ctx, clusterName, envName,
		&req.VirtualServiceReq{
			Namespace: applyVirtualService.GetNamespace(),
			Name:      applyVirtualService.GetName(),
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		return s.CreateVirtualService(ctx, clusterName, envName, applyVirtualService)
	}

	return s.PatchVirtualService(ctx, clusterName, envName, applyVirtualService)
}

// PatchVirtualService patches a given VirtualService in cluster
func (s *Service) PatchVirtualService(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName,
	applyVirtualService *v1beta1.VirtualService) (*v1beta1.VirtualService, error) {
	patchData, err := json.Marshal(applyVirtualService)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	ic, err := s.ClustersetFromConfig(ctx, clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := ic.NetworkingV1beta1().VirtualServices(applyVirtualService.GetNamespace()).
		Patch(ctx, applyVirtualService.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// CreateVirtualService applies a given VirtualService object into cluster
func (s *Service) CreateVirtualService(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName,
	applyVirtualService *v1beta1.VirtualService) (*v1beta1.VirtualService, error) {
	ic, err := s.ClustersetFromConfig(ctx, clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}
	vs, err := ic.NetworkingV1beta1().VirtualServices(applyVirtualService.GetNamespace()).
		Create(ctx, applyVirtualService, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return vs, nil
}

// decodeVirtualServiceYamlData returns a VirtualService decoded from yaml data, nil for error
func (s *Service) decodeVirtualServiceYamlData(_ context.Context, yamlData []byte) (*v1beta1.VirtualService, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	vs, ok := obj.(*v1beta1.VirtualService)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not v1beta1.Ingress")
	}

	return vs, nil
}

// GetVirtualServiceDetail 获取 VirtualService 详情
func (s *Service) GetVirtualServiceDetail(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName,
	getReq *req.VirtualServiceReq) (*v1beta1.VirtualService, error) {
	ic, err := s.ClustersetFromConfig(ctx, clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := ic.NetworkingV1beta1().VirtualServices(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// ClustersetFromConfig returns istio clusterset interface with config from cluster info struct
func (s *Service) ClustersetFromConfig(ctx context.Context, clusterName entity.ClusterName,
	envName entity.AppEnvName) (versionedclient.Interface, error) {
	info, err := s.getClusterInfo(clusterName, string(envName))
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return versionedclient.NewForConfig(info.config)
}

// DescribeVirtualService returns istio description info
func (s *Service) DescribeVirtualService(ctx context.Context, clusterName entity.ClusterName, env entity.AppEnvName,
	descReq *req.DescribeVirtualServiceReq) (res *resp.DescribeVirtualServiceResp, err error) {
	vs, err := s.GetVirtualServiceDetail(ctx, clusterName, env,
		&req.VirtualServiceReq{
			Name:      descReq.Name,
			Namespace: descReq.Namespace,
			Env:       descReq.Env,
		})
	if err != nil {
		return nil, err
	}

	res = new(resp.DescribeVirtualServiceResp)

	err = deepcopy.Copy(vs).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	resourceEvents, err := s.GetK8sResourceEvents(ctx, clusterName,
		&req.GetK8sResourceEventsReq{
			Namespace: descReq.Namespace,
			Resource:  vs,
			Env:       descReq.Env,
		})
	if err != nil {
		return nil, err
	}

	res.Events = resourceEvents.Events

	return res, nil
}

// getNamespaceFromDomainName 根据 domain 提取域名所在 k8s 集群,以确定要更新的 vs 所在的集群
func (s *Service) getNamespaceFromDomainName(name string) string {
	idx := strings.Index(name, ".svc")
	if idx < 0 {
		return ""
	}

	nsLength := strings.LastIndex(name[:idx-1], ".")
	if nsLength < 0 {
		return ""
	}

	return name[nsLength+1 : idx]
}

// getServiceNameFromDomainName  根据 domain 提取 AMS 服务名
func (s *Service) getServiceNameFromDomainName(name string) string {
	idx := strings.Index(name, ".svc")
	if idx < 0 {
		return ""
	}

	nsIdx := strings.Index(name, s.getNamespaceFromDomainName(name))

	if strings.Index(name[0:nsIdx-1], ".") > 0 {
		return name[strings.Index(name[0:nsIdx-1], ".")+1 : nsIdx-1]
	}

	return name[0 : nsIdx-1]
}

// DetermineUpstreamName 根据项目的 istio_enable 属性和实际的 deployment 检测来确定是否需要替换为 istio upstream
func (s *Service) DetermineUpstreamName(ctx context.Context, request *req.DetermineUpstreamNameReq) (
	resp.DetermineUpstreamNameResp, error) {
	response := make(resp.DetermineUpstreamNameResp)

	for _, host := range request.BackendHost {
		var (
			serviceName = s.getServiceNameFromDomainName(host)
			data        = new(resp.UpstreamInfo)
		)
		data.BackendHostName = host
		data.RecommendBackendHostName = host
		data.EnableIstio = false
		response[host] = data

		// find project by service name
		filter, err := s.getAppsFilter(ctx, &req.GetAppsReq{ServiceName: serviceName})
		if err != nil {
			continue
		}

		app, err := s.dao.FindSingleApp(ctx, filter)
		if err != nil {
			continue
		}

		data.EnableIstio = app.EnableIstio
		data.RecommendBackendHostName = ""

		if entity.RegularGetClusterName.MatchString(host) {
			data.Tag = host
		}

		response[host] = data
	}

	return response, nil
}

// createIngressPrivateZoneRecord 为集群所有域名添加 Istio 私有域记录
func (s *Service) createIstioPrivateZoneRecord(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	ingressgateway, err := s.GetServiceDetail(ctx, task.ClusterName, &req.GetServiceDetailReq{
		Namespace: entity.IstioNamespace,
		Name:      entity.IstioServiceIngressGateway,
		Env:       string(task.EnvName),
	})

	if err != nil {
		return err
	}

	err = s.createPrivateZoneRecord(ctx, project, app, task, ingressgateway)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) createPrivateZoneRecord(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, svc *v1.Service) error {
	// 获取
	if len(svc.Status.LoadBalancer.Ingress) > 0 && svc.Status.LoadBalancer.Ingress[0].IP != "" {
		updater, err := s.getQDNSUpdater(ctx, task.OperatorID)
		if err != nil {
			return err
		}

		domainController := s.GetDomainControllerFromProjectOwners(project.Owners)
		domainName := s.getAliPrivateZoneK8sDomainNameWithCluster(app.ServiceName, task.GetEnvName(), task.ClusterName)
		// TODO: 当前以第一个项目负责人替代域名负责人，未来与运维商讨负责人机制
		err = s.createAliPrivateZoneRecord(ctx, domainName, svc.Status.LoadBalancer.Ingress[0].IP, updater, domainController, entity.ARecord)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkVirtualServicePrivateZoneRecord 检验 VirtualService 解析是否注册至 private_zone
func (s *Service) checkVirtualServicePrivateZoneRecord(ctx context.Context, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	svc, err := s.GetServiceDetail(ctx, task.ClusterName,
		&req.GetServiceDetailReq{
			Namespace: entity.IstioNamespace,
			Name:      entity.IstioServiceIngressGateway,
			Env:       string(task.EnvName),
		})
	if err != nil {
		return err
	}

	ingressAddress := svc.Status.LoadBalancer.Ingress[0].IP

	multiClusterSupported, err := s.CheckMultiClusterSupport(ctx, task.EnvName, app.ProjectID)
	if err != nil {
		return err
	}

	// 多集群白名单内的应用只检测集群专属域名注册情况
	if multiClusterSupported {
		domainName := s.getAliPrivateZoneK8sDomainNameWithCluster(app.ServiceName, task.GetEnvName(), task.ClusterName)
		return s.checkAliPrivateZoneRegistration(ctx, domainName, ingressAddress)
	}

	return s.checkAllAliPrivateZoneRegistration(ctx, app.ServiceName, task.EnvName, task.ClusterName, ingressAddress)
}
