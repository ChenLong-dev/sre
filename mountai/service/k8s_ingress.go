package service

import (
	"time"

	"context"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
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

	networkingV1 "k8s.io/api/networking/v1"
)

func (s *Service) decodeIngressYamlData(_ context.Context, yamlData []byte) (*networkingV1.Ingress, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	ingress, ok := obj.(*networkingV1.Ingress)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not networkingV1.Ingress")
	}
	return ingress, nil
}

// GetIngressDetail 获取 Ingress 详情
func (s *Service) GetIngressDetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetIngressDetailReq) (*networkingV1.Ingress, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.NetworkingV1().Ingresses(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// ApplyIngress 声明式创建/更新 Ingress
func (s *Service) ApplyIngress(ctx context.Context, clusterName entity.ClusterName,
	yamlData []byte, env string) (ingress *networkingV1.Ingress, err error) {
	applyIngress, err := s.decodeIngressYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetIngressDetail(ctx, clusterName,
		&req.GetIngressDetailReq{
			Namespace: applyIngress.GetNamespace(),
			Name:      applyIngress.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		ingress, err = s.CreateIngress(ctx, clusterName, env, applyIngress)
		return
	}

	ingress, err = s.PatchIngress(ctx, clusterName, env, applyIngress)
	return
}

// CreateIngress 创建 Ingress
func (s *Service) CreateIngress(ctx context.Context, clusterName entity.ClusterName, envName string,
	ingress *networkingV1.Ingress) (*networkingV1.Ingress, error) {
	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.NetworkingV1().Ingresses(ingress.GetNamespace()).
		Create(ctx, ingress, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// PatchIngress 更新 Ingress
func (s *Service) PatchIngress(ctx context.Context, clusterName entity.ClusterName, envName string,
	ingress *networkingV1.Ingress) (*networkingV1.Ingress, error) {
	patchData, err := json.Marshal(ingress)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sTypedClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.NetworkingV1().Ingresses(ingress.GetNamespace()).
		Patch(ctx, ingress.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// RenderIngressTemplate 渲染 Ingress 模板
func (s *Service) RenderIngressTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, backendServiceName string) ([]byte, error) {
	tpl, err := s.initIngressTemplate(ctx, project, app, task, backendServiceName)
	if err != nil {
		return nil, err
	}
	data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// 获取ingress注释
func (s *Service) getIngressAnnotations(_ *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp) map[string]string {
	res := make(map[string]string)
	if task.Param.IsSupportStickySession {
		res[entity.IngressAnnotationNameAffinity] = entity.IngressAnnotationValueAffinityCookie
		res[entity.IngressAnnotationNameAffinityMode] = entity.IngressAnnotationValueAffinityModeBalanced
		res[entity.IngressAnnotationNameSessionCookieMaxAge] = strconv.Itoa(task.Param.SessionCookieMaxAge)
	} else {
		res[entity.IngressAnnotationNameAffinity] = ""
		res[entity.IngressAnnotationNameAffinityMode] = ""
		res[entity.IngressAnnotationNameSessionCookieMaxAge] = ""
	}
	if app.ServiceType == entity.AppServiceTypeGRPC {
		res[entity.IngressAnnotationNameBackendProtocol] = entity.IngressAnnotationValueGRPCS
	}

	if s.k8sClusters[task.EnvName][task.ClusterName].ingressClass != "" {
		res[entity.IngressAnnotationIngressClasss] = s.k8sClusters[task.EnvName][task.ClusterName].ingressClass
	}

	return res
}

func (s *Service) initIngressTemplate(_ context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, backendServiceName string) (*entity.IngressTemplate, error) {
	tpl := &entity.IngressTemplate{
		ProjectName:            project.Name,
		AppName:                app.Name,
		Name:                   app.ServiceName,
		Namespace:              task.Namespace,
		ServiceName:            backendServiceName,
		ServiceHost:            s.getAliPrivateZoneK8sFullDomainName(app.ServiceName, task.GetEnvName()),
		ServiceHostWithCluster: s.getAliPrivateZoneK8sFullDomainNameWithCluster(app.ServiceName, task.GetEnvName(), task.ClusterName),
		Annotations:            s.getIngressAnnotations(project, app, task),
	}

	// 当前kong存在bug, 需要ingress包含集群所有的域名, 做兼容处理
	clusters := config.Conf.K8sClusters[string(task.EnvName)]

	// 暂不删除，为多集群做准备
	for _, cluster := range clusters {
		tpl.ServiceHostsWithCluster = append(tpl.ServiceHostsWithCluster,
			s.getAliPrivateZoneK8sFullDomainNameWithCluster(app.ServiceName, task.GetEnvName(), entity.ClusterName(cluster.Name)))
	}

	if app.ServiceType == entity.AppServiceTypeGRPC {
		clusterInfo, err := s.getClusterInfo(task.ClusterName, string(task.EnvName))
		if err != nil {
			return nil, err
		}
		tpl.SecretName = clusterInfo.tlsSecretName
		tpl.ServicePort = entity.ServiceGRPCDefaultInternalPort
	} else if app.ServiceType == entity.AppServiceTypeRestful {
		tpl.ServicePort = entity.ServiceDefaultInternalPort
	} else {
		return nil, errors.Wrap(errcode.InvalidParams, "unsupported service type")
	}
	return tpl, nil
}

// ApplyIngressFromTpl 从模板声明式创建/更新 Ingress
func (s *Service) ApplyIngressFromTpl(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, backendServiceName string) (err error) {
	// 渲染ingress模版
	data, err := s.RenderIngressTemplate(ctx, project, app, task, backendServiceName)
	if err != nil {
		return
	}
	// 生成Ingress
	_, err = s.ApplyIngress(ctx, task.ClusterName, data, string(task.EnvName))
	return
}

// DeleteIngress 删除 Ingress
func (s *Service) DeleteIngress(ctx context.Context, clusterName entity.ClusterName,
	deleteReq *req.DeleteIngressReq) error {
	policy := metav1.DeletePropagationForeground

	c, err := s.GetK8sTypedClient(clusterName, deleteReq.Env)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	err = c.NetworkingV1().Ingresses(deleteReq.Namespace).
		Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// createIngressPrivateZoneRecordForClusterDomain 为集群专用域名添加Ingress私有域记录
func (s *Service) createIngressPrivateZoneRecordForClusterDomain(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	ingress, err := s.GetIngressDetail(ctx, task.ClusterName,
		&req.GetIngressDetailReq{
			Namespace: task.Namespace,
			Name:      app.ServiceName,
			Env:       string(task.EnvName),
		})
	if err != nil {
		return err
	}

	if len(ingress.Status.LoadBalancer.Ingress) > 0 && ingress.Status.LoadBalancer.Ingress[0].IP != "" {
		var updater string
		updater, err = s.getQDNSUpdater(ctx, task.OperatorID)
		if err != nil {
			return err
		}

		domainName := s.getAliPrivateZoneK8sDomainNameWithCluster(app.ServiceName, task.GetEnvName(), task.ClusterName)
		domainController := s.GetDomainControllerFromProjectOwners(project.Owners)
		err = s.createAliPrivateZoneRecord(ctx, domainName, ingress.Status.LoadBalancer.Ingress[0].IP, updater, domainController, entity.ARecord)
		if err != nil {
			return err
		}
	}

	return nil
}

// createIngressPrivateZoneRecord 为集群所有域名添加Ingress私有域记录
func (s *Service) createIngressPrivateZoneRecord(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp) (err error) {
	ingress, err := s.GetIngressDetail(ctx, task.ClusterName,
		&req.GetIngressDetailReq{
			Namespace: task.Namespace,
			Name:      app.ServiceName,
			Env:       string(task.EnvName),
		})
	if err != nil {
		return err
	}

	if len(ingress.Status.LoadBalancer.Ingress) > 0 && ingress.Status.LoadBalancer.Ingress[0].IP != "" {
		var updater string
		updater, err = s.getQDNSUpdater(ctx, task.OperatorID)
		if err != nil {
			return err
		}

		domainName := s.getAliPrivateZoneK8sDomainName(app.ServiceName, task.GetEnvName())
		domainController := s.GetDomainControllerFromProjectOwners(project.Owners)
		err = s.createAliPrivateZoneRecord(ctx, domainName, ingress.Status.LoadBalancer.Ingress[0].IP, updater, domainController, entity.ARecord)
		if err != nil {
			return err
		}

		domainName = s.getAliPrivateZoneK8sDomainNameWithCluster(app.ServiceName, task.EnvName, task.ClusterName)
		// TODO: 当前以第一个项目负责人替代域名负责人，未来与运维商讨负责人机制
		err = s.createAliPrivateZoneRecord(ctx, domainName, ingress.Status.LoadBalancer.Ingress[0].IP, updater, domainController, entity.ARecord)
		if err != nil {
			return err
		}
	}

	return nil
}

// deleteIngressPrivateZoneRecord 删除Ingress私有域记录
func (s *Service) deleteIngressPrivateZoneRecord(ctx context.Context, clusterName entity.ClusterName,
	serviceName string, task *resp.TaskDetailResp) error {
	ingress, err := s.GetIngressDetail(ctx, clusterName,
		&req.GetIngressDetailReq{
			Namespace: task.Namespace,
			Name:      serviceName,
			Env:       string(task.EnvName),
		})
	if err != nil {
		if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil
		}
		return err
	}

	if len(ingress.Status.LoadBalancer.Ingress) > 0 && ingress.Status.LoadBalancer.Ingress[0].IP != "" {
		var updater string
		updater, err = s.getQDNSUpdater(ctx, task.OperatorID)
		if err != nil {
			return err
		}

		domainName := s.getAliPrivateZoneK8sDomainName(serviceName, task.EnvName)
		err = s.deleteAliPrivateZoneRecord(ctx, domainName, ingress.Status.LoadBalancer.Ingress[0].IP, updater, entity.ARecord)
		if err != nil {
			return err
		}

		domainName = s.getAliPrivateZoneK8sDomainNameWithCluster(serviceName, task.EnvName, clusterName)
		err = s.deleteAliPrivateZoneRecord(ctx, domainName, ingress.Status.LoadBalancer.Ingress[0].IP, updater, entity.ARecord)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckPrivateZoneRecordEntry 检查 private_zone 解析的统一入口
func (s *Service) CheckPrivateZoneRecordEntry(ctx context.Context, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	switch app.ServiceExposeType {
	case entity.AppServiceExposeTypeLB:
		return s.checkLBPrivateZoneRecord(ctx, app, task)
	case entity.AppServiceExposeTypeIngress:
		if s.GetApplicationIstioState(ctx, task.EnvName, task.ClusterName, app) {
			return s.checkVirtualServicePrivateZoneRecord(ctx, app, task)
		}
		return s.checkIngressPrivateZoneRecord(ctx, app, task)
	case entity.AppServiceExposeTypeInternal:
		return nil
	}

	return nil
}

// checkIngressPrivateZoneRecord 检验 Ingress 解析是否注册至 private_zone
func (s *Service) checkIngressPrivateZoneRecord(ctx context.Context, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	ingress, err := s.GetIngressDetail(ctx, task.ClusterName,
		&req.GetIngressDetailReq{
			Namespace: task.Namespace,
			Name:      app.ServiceName,
			Env:       string(task.EnvName),
		})
	if err != nil {
		return err
	}

	ingressAddress := ingress.Status.LoadBalancer.Ingress[0].IP

	multiClusterSupported, err := s.CheckMultiClusterSupport(ctx, task.EnvName, app.ProjectID)
	if err != nil {
		return err
	}

	// 多集群白名单内的应用只检测集群专属域名注册情况
	if multiClusterSupported {
		domainName := s.getAliPrivateZoneK8sDomainNameWithCluster(app.ServiceName, task.EnvName, task.ClusterName)
		return s.checkAliPrivateZoneRegistration(ctx, domainName, ingressAddress)
	}

	err = s.checkAllAliPrivateZoneRegistration(
		ctx, app.ServiceName, task.EnvName, task.ClusterName, ingressAddress)
	if err != nil {
		return err
	}

	return nil
}

// checkLBPrivateZoneRecord 检查阿里云 LB 解析是否注册到 private_zone
func (s *Service) checkLBPrivateZoneRecord(ctx context.Context, app *resp.AppDetailResp, task *resp.TaskDetailResp) error {
	svc, err := s.GetServiceDetail(ctx, task.ClusterName, &req.GetServiceDetailReq{
		Namespace: task.Namespace,
		Name:      app.ServiceName,
		Env:       string(task.EnvName),
	})
	if err != nil {
		return err
	}

	if svc.Spec.Type != v1.ServiceTypeLoadBalancer {
		return errors.Wrapf(errcode.InternalError, "non-LB-type of service(%s)", app.ServiceName)
	}

	if len(svc.Status.LoadBalancer.Ingress) == 0 {
		return errors.Wrapf(_errcode.NotFindK8sServiceIngressError, "service name:%s namespace:%s", app.ServiceName, task.EnvName)
	}

	multiClusterSupported, err := s.CheckMultiClusterSupport(ctx, task.EnvName, app.ProjectID)
	if err != nil {
		return err
	}

	// 多集群白名单内的应用只检测集群专属域名注册情况
	if multiClusterSupported {
		domainName := s.getAliPrivateZoneK8sDomainNameWithCluster(app.ServiceName, task.EnvName, task.ClusterName)
		return s.checkAliPrivateZoneRegistration(ctx, domainName, svc.Status.LoadBalancer.Ingress[0].IP)
	}

	err = s.checkAllAliPrivateZoneRegistration(ctx, app.ServiceName, task.GetEnvName(), task.ClusterName,
		svc.Status.LoadBalancer.Ingress[0].IP)
	if err != nil {
		return err
	}

	return nil
}

// DescribeIngress 获取 Ingress describe 信息
func (s *Service) DescribeIngress(ctx context.Context, clusterName entity.ClusterName,
	descReq *req.DescribeIngressReq) (*resp.DescribeIngressResp, error) {
	res := new(resp.DescribeIngressResp)
	ingress, err := s.GetIngressDetail(ctx, clusterName,
		&req.GetIngressDetailReq{
			Name:      descReq.Name,
			Namespace: descReq.Namespace,
			Env:       descReq.Env,
		})
	if err != nil {
		return nil, err
	}

	err = deepcopy.Copy(ingress).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	ingressEvents, err := s.GetK8sResourceEvents(ctx, clusterName,
		&req.GetK8sResourceEventsReq{
			Namespace: descReq.Namespace,
			Resource:  ingress,
			Env:       descReq.Env,
		})
	if err != nil {
		return nil, err
	}
	res.Events = ingressEvents.Events

	return res, nil
}

// checkAllAliPrivateZoneRegistration 检验集群所有相关的域名和对应地址是否注册至阿里云私有域
func (s *Service) checkAllAliPrivateZoneRegistration(ctx context.Context,
	serviceName string, envName entity.AppEnvName, clusterName entity.ClusterName, address string) error {
	domainName := s.getAliPrivateZoneK8sDomainName(serviceName, envName)
	err := s.checkAliPrivateZoneRegistration(ctx, domainName, address)
	if err != nil {
		return err
	}

	domainName = s.getAliPrivateZoneK8sDomainNameWithCluster(serviceName, envName, clusterName)
	return s.checkAliPrivateZoneRegistration(ctx, domainName, address)
}

// checkAliPrivateZoneRegistration 检验域名和对应地址是否注册至阿里云私有域
func (s *Service) checkAliPrivateZoneRegistration(ctx context.Context, domainName, address string) error {
	records, err := s.getAliPrivateZoneRecord(ctx, &req.GetQDNSRecordsReq{
		DomainType:       entity.ARecord,
		DomainRecordName: domainName,
		DomainValue:      address,
		PrivateZone:      AliK8sPrivateZone,
		PageNumber:       1,
		PageSize:         req.GetQDNSRecordsPageSizeLimit,
	})
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return errors.Wrapf(_errcode.IngressNotRegisteredToPrivateZoneError, "domain: %s.%s", domainName, AliK8sPrivateZone)
	}

	return nil
}

// WaitIngressChangeReady wait ingress ready
// TODO: 需要看看有没有更优雅的实现 或者是否需要
func (s *Service) WaitIngressChangeReady(_ context.Context) {
	time.Sleep(time.Duration(config.Conf.Other.IngressChangeWaitDuration))
}
