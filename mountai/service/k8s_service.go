package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
)

func (s *Service) initServiceTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp, serviceName string) ([]*entity.ServiceTemplate, error) {
	currentEnv, ok := app.Env[task.EnvName]
	if !ok {
		return nil, errors.Wrapf(errcode.InternalError, "couldn't find env:%s", task.EnvName)
	}

	templates := make([]*entity.ServiceTemplate, 0)
	// 默认端口
	ports := make([]*entity.ServicePortTemplate, 0)
	if app.ServiceType == entity.AppServiceTypeRestful {
		ports = append(ports, &entity.ServicePortTemplate{
			Name:         entity.ServiceDefaultHTTPName,
			Protocol:     entity.ServiceTCPProtocol,
			ExternalPort: entity.ServiceDefaultInternalPort,
			TargetPort:   int32(task.Param.TargetPort),
		})

		// 证书是挂在LB上的，需要https443端口
		if app.ServiceExposeType == entity.AppServiceExposeTypeLB {
			ports = append(ports, &entity.ServicePortTemplate{
				Name:         entity.ServiceDefaultHTTPSName,
				Protocol:     entity.ServiceTCPProtocol,
				ExternalPort: entity.ServiceDefaultInternalPort443,
				TargetPort:   int32(task.Param.TargetPort),
			})
		}
	}

	if app.ServiceType == entity.AppServiceTypeGRPC {
		ports = append(ports, &entity.ServicePortTemplate{
			Name:         entity.ServiceDefaultHTTPSName,
			Protocol:     entity.ServiceTCPProtocol,
			ExternalPort: entity.ServiceGRPCDefaultInternalPort,
			TargetPort:   int32(task.Param.TargetPort),
		})
	}

	// 额外端口
	for name, port := range task.Param.ExposedPorts {
		ports = append(ports, &entity.ServicePortTemplate{
			Name:         name,
			Protocol:     entity.ServiceTCPProtocol,
			ExternalPort: int32(port),
			TargetPort:   int32(port),
		})
	}

	// 设置默认协议
	lbProtocol := currentEnv.ServiceProtocol
	if lbProtocol == "" {
		lbProtocol = entity.LoadBalancerProtocolHTTP
	}

	if app.ServiceExposeType == entity.AppServiceExposeTypeLB {
		clusterInfo, e := s.getClusterInfo(task.ClusterName, string(task.EnvName))
		if e != nil {
			return nil, e
		}

		formatName := func(i int) (string, string) {
			if i > 0 {
				return fmt.Sprintf("%s-ext-%d", serviceName, i), fmt.Sprintf("%s-ext-%d-%s", serviceName, i, string(task.EnvName))
			}

			return serviceName, fmt.Sprintf("%s-%s", serviceName, string(task.EnvName))
		}
		// 获取 lb 实例 id
		lbs := s.GetAppLoadBalancersByEnvAndCluster(ctx, app, task.ClusterName, task.EnvName)
		for i, lb := range lbs {
			if lb.LoadBalancerID == "" || lb.LoadbalancerCertId == "" {
				return nil, errors.Wrap(errcode.InternalError, "service load balancer instance id or cert id is empty")
			}

			sn, loadBalancerName := formatName(i)
			annotations, e := s.generateServiceAnnotationsForLB(clusterInfo.vendor, loadBalancerName, lb.LoadBalancerID, lb.LoadbalancerCertId)
			if e != nil {
				return nil, e
			}

			templates = append(templates, &entity.ServiceTemplate{
				ProjectName:    project.Name,
				AppName:        app.Name,
				Name:           sn,
				AppServiceType: app.ServiceType,
				Namespace:      task.Namespace,
				Ports:          ports,
				Protocol:       lbProtocol,
				Type:           v1.ServiceTypeLoadBalancer,
				WithLB:         true,
				Annotations:    annotations,
			})
		}

	} else {
		templates = append(templates, &entity.ServiceTemplate{
			ProjectName:    project.Name,
			AppName:        app.Name,
			Name:           serviceName,
			AppServiceType: app.ServiceType,
			Namespace:      task.Namespace,
			Ports:          ports,
			Protocol:       lbProtocol,
			Type:           v1.ServiceTypeClusterIP,
			WithLB:         false,
		})
	}

	return templates, nil
}

func (s *Service) decodeServiceYamlData(_ context.Context, yamlData []byte) (*v1.Service, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}
	service, ok := obj.(*v1.Service)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not v1.Service")
	}
	return service, nil
}

// RenderServiceTemplate 渲染服务模板
func (s *Service) RenderServiceTemplates(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp, serviceName string) ([][]byte, error) {
	tpls, err := s.initServiceTemplate(ctx, project, app, task, team, serviceName)
	if err != nil {
		return nil, err
	}

	res := make([][]byte, 0, len(tpls))
	for _, tpl := range tpls {
		data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
		if err != nil {
			return nil, err
		}

		res = append(res, []byte(data))
	}

	return res, nil
}

// ApplyService 声明式更新服务
func (s *Service) ApplyService(ctx context.Context, clusterName entity.ClusterName,
	yamlData []byte, env string) (*v1.Service, error) {
	applyService, err := s.decodeServiceYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	oldService, err := s.GetServiceDetail(ctx, clusterName,
		&req.GetServiceDetailReq{
			Namespace: applyService.GetNamespace(),
			Name:      applyService.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		service, e := s.CreateService(ctx, clusterName, applyService, env)
		if e != nil {
			return nil, e
		}
		return service, nil
	}

	log.Infoc(ctx, "start process patch service data:%#v", applyService)
	// ClusterIP类型不分配nodeport
	if applyService.Spec.Type == v1.ServiceTypeClusterIP {
		if oldService.Spec.ClusterIP != "None" {
			applyService.Spec.ClusterIP = oldService.Spec.ClusterIP
		}
	} else {
		// k8s api对于不填NodePort时，会重新分配
		// 为了避免重新分配导致502问题，需要获取上一次的NodePort赋值
		portMap := make(map[int32]v1.ServicePort)
		for _, curPort := range applyService.Spec.Ports {
			portMap[curPort.Port] = curPort
		}
		for _, curPort := range oldService.Spec.Ports {
			if port, ok := portMap[curPort.Port]; ok {
				port.NodePort = curPort.NodePort
				portMap[curPort.Port] = port
			}
		}
		ports := make([]v1.ServicePort, 0)
		for _, port := range portMap {
			ports = append(ports, port)
		}
		applyService.Spec.Ports = ports
	}
	log.Infoc(ctx, "finish process patch service data:%#v", applyService)

	service, err := s.PatchService(ctx, clusterName, applyService, env)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// CreateService 创建服务
func (s *Service) CreateService(ctx context.Context, clusterName entity.ClusterName,
	service *v1.Service, envName string) (*v1.Service, error) {
	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.CoreV1().Services(service.GetNamespace()).
		Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// WaitInClusterDNSChangeReady wait dns ready
// 1. 强制等待TTL周期
func (s *Service) WaitInClusterDNSChangeReady(_ context.Context) {
	time.Sleep(time.Duration(config.Conf.Other.InClusterDNSChangeWaitDuration))
}

// GetShadowServiceName return service with shadow prefix
// 禁用集群内DNS解析 是对于公共库来说的
// TODO: 公共库需要有更灵活的实现和ams配合
func (s *Service) GetShadowServiceName(serviceName string) string {
	return fmt.Sprintf("%s%s", entity.ShadowServicePrefix, serviceName)
}

// GetServiceDetail 获取服务详情
func (s *Service) GetServiceDetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetServiceDetailReq) (*v1.Service, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.CoreV1().Services(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) getServiceLabelSelector(getReq *req.GetServicesReq) string {
	labels := make([]string, 0)
	if getReq.ProjectName != "" {
		labels = append(labels, fmt.Sprintf("project=%s", getReq.ProjectName))
	}

	if getReq.AppName != "" {
		labels = append(labels, fmt.Sprintf("app=%s", getReq.AppName))
	}
	return strings.Join(labels, ",")
}

// PatchService 更新服务
func (s *Service) PatchService(ctx context.Context, clusterName entity.ClusterName,
	service *v1.Service, envName string) (*v1.Service, error) {
	patchData, err := json.Marshal(service)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.CoreV1().Services(service.GetNamespace()).
		Patch(ctx, service.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// GetServices 获取服务列表
func (s *Service) GetServices(ctx context.Context, clusterName entity.ClusterName, envName string,
	getReq *req.GetServicesReq) ([]v1.Service, error) {
	selector := s.getServiceLabelSelector(getReq)

	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.CoreV1().Services(getReq.Namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return list.Items, nil
}

// DeleteService 删除服务
func (s *Service) DeleteService(ctx context.Context, clusterName entity.ClusterName,
	deleteReq *req.DeleteServiceReq) error {
	policy := metav1.DeletePropagationForeground
	c, err := s.GetK8sTypedClient(clusterName, deleteReq.Env)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	err = c.CoreV1().Services(deleteReq.Namespace).
		Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// GetServiceLoadBalancerIP 获取服务的负载均衡IP
func (s *Service) GetServiceLoadBalancerIP(ctx context.Context,
	clusterName entity.ClusterName, name, namespace string) (string, error) {
	svc, err := s.GetServiceDetail(ctx, clusterName,
		&req.GetServiceDetailReq{
			Namespace: namespace,
			Name:      name,
			Env:       namespace,
		})
	if err != nil {
		return "", err
	} else if svc.Spec.Type == v1.ServiceTypeLoadBalancer && len(svc.Status.LoadBalancer.Ingress) == 0 {
		return "", errors.Wrapf(_errcode.NotFindK8sServiceIngressError, "service name:%s namespace:%s", name, namespace)
	} else if svc.Spec.Type == v1.ServiceTypeClusterIP {
		return "", nil
	}
	return svc.Status.LoadBalancer.Ingress[0].IP, nil
}

// GetCurrentServiceName return the name of service
func (s *Service) GetCurrentServiceName(ctx context.Context, clusterName entity.ClusterName,
	envName entity.AppEnvName, app *resp.AppDetailResp) (string, error) {
	if app.ServiceExposeType != entity.AppServiceExposeTypeIngress || app.ServiceType != entity.AppServiceTypeRestful {
		return app.ServiceName, nil
	}

	// todo 多集群迁移:强制跨集群访问的问题
	svc, err := s.GetServiceDetail(ctx, clusterName, &req.GetServiceDetailReq{
		Namespace: s.GetNamespaceByApp(app, envName, clusterName),
		Name:      s.GetShadowServiceName(app.ServiceName),
		Env:       string(envName),
	})

	if err != nil {
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return app.ServiceName, nil
		}
		return "", err
	}

	return svc.GetName(), nil
}

// DescribeService 获取服务 describe 信息
func (s *Service) DescribeService(ctx context.Context, clusterName entity.ClusterName,
	descReq *req.DescribeServiceReq) (*resp.DescribeServiceResp, error) {
	res := new(resp.DescribeServiceResp)
	svc, err := s.GetServiceDetail(ctx, clusterName,
		&req.GetServiceDetailReq{
			Namespace: descReq.Namespace,
			Name:      descReq.Name,
			Env:       descReq.Env,
		})
	if err != nil {
		return nil, err
	}
	for _, lb := range svc.Status.LoadBalancer.Ingress {
		res.Status.LoadBalancerIP = append(res.Status.LoadBalancerIP, lb.IP)
	}
	res.Status.ClusterIP = svc.Spec.ClusterIP

	serviceEvents, err := s.GetK8sResourceEvents(ctx, clusterName,
		&req.GetK8sResourceEventsReq{
			Namespace: descReq.Namespace,
			Resource:  svc,
			Env:       descReq.Env,
		})
	if err != nil {
		return nil, err
	}
	res.Events = serviceEvents.Events

	endpoints, err := s.GetEndpointsDetail(ctx, clusterName,
		&req.GetEndpointsReq{
			Namespace: descReq.Namespace,
			Name:      descReq.Name,
			Env:       descReq.Env,
		})
	if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
		return res, nil
	}

	if err != nil {
		return nil, err
	}

	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			res.Status.Endpoints = append(res.Status.Endpoints, resp.ServiceEndpoints{
				IP:       addr.IP,
				NodeName: *addr.NodeName,
			})
		}
	}

	return res, nil
}

// ApplyServiceFromTpl 依据模板声明式更新服务
func (s *Service) ApplyServiceFromTpl(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp, serviceName string) error {
	// 渲染模版
	tpls, err := s.RenderServiceTemplates(ctx, project, app, task, team, serviceName)
	if err != nil {
		return err
	}

	// 生成Service
	for _, tpl := range tpls {
		_, err = s.ApplyService(ctx, task.ClusterName, tpl, string(task.EnvName))
		if err != nil {
			return err
		}
	}

	return nil
}

// generateServiceAnnotationsForLB 为不同云服务商生成负载均衡相关的 Service annotations
func (s *Service) generateServiceAnnotationsForLB(vendor entity.VendorName, loadBalancerName string, loadBalancerID string, loadBalancerCertID string) (map[string]string, error) {
	switch vendor {
	case entity.VendorAli:
		return map[string]string{
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-name":                     loadBalancerName,
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id":                       loadBalancerID,
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cert-id":                  loadBalancerCertID,
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-protocol-port":            "https:443,http:80",
			"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners": "true",
		}, nil

	case entity.VendorHuawei:
		// 文档: https://support.huaweicloud.com/usermanual-cce/cce_01_0014.html
		// 暂时没有会话保持的需求, 不填写 kubernetes.io/session-affinity-mode 和 kubernetes.io/elb.session-affinity-option
		// 集群所在子网 ID 在 Kubernetes v1.11.7-r0 以上已经支持不填 kubernetes.io/elb.subnet-id
		// 负载均衡算法按照默认值 ROUND_ROBIN (不指定 kubernetes.io/elb.lb-algorithm)
		// ELB 健康检查并不需要开启, 依靠 k8s 的健康检查即可
		return map[string]string{
			// CCE-Turbo 集群必须填写 elb.class, 暂定必定使用独享型 ELB
			"kubernetes.io/elb.class": "performance",
			"kubernetes.io/elb.id":    loadBalancerID,
		}, nil

	default:
		return nil, errors.Wrap(_errcode.UnknownVendorError, string(vendor))
	}
}
