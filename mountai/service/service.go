package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/AliyunContainerService/kubernetes-cronhpa-controller/pkg/apis/autoscaling/v1beta1"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/pkg/errors"
	framework "gitlab.shanhai.int/sre/app-framework"
	infraJenkins "gitlab.shanhai.int/sre/gojenkins"
	"gitlab.shanhai.int/sre/library/net/cm"
	"gitlab.shanhai.int/sre/library/net/httpclient"
	istioV1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"rulai/config"
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
	"rulai/vendors"
)

var (
	// SVC 全局服务
	SVC *Service
	// K8sSystemUser 全局系统用户(自定义admin)
	K8sSystemUser *resp.UserProfileResp
)

// ====================
// >>>请勿删除<<<
//
// 自定义服务
// ====================
type Service struct {
	// ====================
	// >>>请勿删除<<<
	//
	// 基础服务
	// ====================
	*framework.DefaultService

	// ====================
	// >>>请勿删除<<<
	//
	// 数据层
	// ====================
	dao *dao.Dao

	// ====================
	// 根据实际情况，选择性保留
	// ====================
	// Http客户端
	httpClient *httpclient.Client

	aliClient     *sdk.Client
	aliLogClient  sls.ClientInterface
	jenkinsClient *infraJenkins.Jenkins

	k8sClusters map[entity.AppEnvName]map[entity.ClusterName]*k8sCluster
	vendors     map[entity.VendorName]vendors.Controller

	// CI流程Jenkins客户端
	jenkinsCIClient *infraJenkins.Jenkins
}

// k8sCluster k8s集群配置
type k8sCluster struct {
	id                             string
	nameInVendor                   string
	name                           entity.ClusterName
	vendor                         entity.VendorName
	envName                        entity.AppEnvName
	version                        *version.Info
	config                         *rest.Config
	typedClient                    kubernetes.Interface
	dynamicClient                  dynamic.Interface
	tlsSecretName                  string
	disableLogConfig               bool
	imageRegistryHostWithNamespace string
	k8sGroupVersions               map[string]*schema.GroupVersion
	ingressClass                   string
	localDNS                       string
	visibleProjectIDs              []string
}

type k8sVendorConfigs struct {
	baseConfig       *config.VendorConfig
	controllerConfig *vendors.ControllerConfig
}

// ====================
// >>>请勿删除<<<
//
// 新建服务
// ====================
func New() *Service {
	// ====================
	// >>>请勿删除<<<
	//
	// 新建数据层
	// ====================
	d, err := dao.New()
	if err != nil {
		panic(err)
	}

	// jenkins
	jenkinsClient, err := infraJenkins.CreateJenkins(nil, config.Conf.Jenkins).Init()
	if err != nil {
		panic(err)
	}

	// CI流程jenkins
	// jenkinsCIClient, err := infraJenkins.CreateJenkins(nil, config.Conf.JenkinsCI.GoJenkins).Init()
	// if err != nil {
	// 	panic(err)
	// }

	vendorConfigsMapping, err := loadVendorConfigsMapping()
	if err != nil {
		panic(err)
	}

	k8sClusters, err := loadK8sClustersAndClusterConfigsInVendorConfigs(vendorConfigsMapping)
	if err != nil {
		panic(err)
	}

	err = checkKongConfigs()
	if err != nil {
		panic(err)
	}

	// 阿里云 SDK, TODO: 迁移至 vendor 包
	aliCfg, ok := vendorConfigsMapping[entity.VendorAli]
	if !ok {
		panic(errors.Errorf("missing ali configs"))
	}

	aliClient, err := sdk.NewClientWithAccessKey(aliCfg.baseConfig.RegionID,
		aliCfg.baseConfig.AccessKeyID, aliCfg.baseConfig.AccessKeySecret)
	if err != nil {
		panic(err)
	}

	// 阿里云日志
	aliLogClient := sls.CreateNormalInterface(aliCfg.baseConfig.LogEndpoint,
		aliCfg.baseConfig.AccessKeyID, aliCfg.baseConfig.AccessKeySecret, "")

	// 改为 return 前替换 SVC 全局变量的方式, 防止 init 过程变复杂后出现 SVC 全局变量部分提前生效问题
	svc := &Service{
		// ====================
		// >>>请勿删除<<<
		// ====================
		DefaultService: framework.GetDefaultService(config.Conf.Config, d),
		dao:            d,
		// ====================
		// 根据实际情况，选择性保留
		// ====================
		httpClient:    httpclient.NewHttpClient(config.Conf.HTTPClient),
		aliClient:     aliClient,
		aliLogClient:  aliLogClient,
		jenkinsClient: jenkinsClient,
		k8sClusters:   k8sClusters,
		vendors:       make(map[entity.VendorName]vendors.Controller, len(vendorConfigsMapping)),
		// jenkinsCIClient: jenkinsCIClient,
	}

	// 向Scheme注册CRD
	// 当前 AMS 使用的 CRD 包括:
	//   1. CronHPA(阿里云开源), 无论是否是阿里云的集群均安装
	//   2. Istio 相关组件
	err = registerCRDsInScheme(v1beta1.AddToScheme, istioV1beta1.AddToScheme)
	if err != nil {
		panic(err)
	}

	// 初始化系统用户信息
	ctx := context.Background()
	user, err := svc.GetUserInfo(ctx, entity.K8sSystemUserID)
	if err != nil {
		panic(err)
	}
	K8sSystemUser = user

	// 运营商控制器
	for vendorName, vendorClusterCfg := range vendorConfigsMapping {
		vendorClusterCfg.controllerConfig.HTTPClient = svc.httpClient
		vendorClusterCfg.controllerConfig.DAO = svc.dao

		svc.vendors[vendorName], err = vendors.NewController(vendorClusterCfg.controllerConfig)
		if err != nil {
			panic(err)
		}
	}

	SVC = svc
	return svc
}

// 新建k8s动态资源客户端
func newK8sDynamicClient(cfg *rest.Config) (dynamic.Interface, error) {
	k8sClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

// 新建k8s静态资源客户端
func newK8sTypedClient(cfg *rest.Config) (*kubernetes.Clientset, error) {
	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

// 获取k8s配置
func getK8sConfig(path, ctxName string) (*rest.Config, error) {
	// 读取配置
	basicCfg, err := getPrivateK8sConfig(path)
	if err != nil {
		return nil, err
	}

	// 生成需要的配置
	cfg, err := clientcmd.NewNonInteractiveClientConfig(
		*basicCfg, ctxName,
		&clientcmd.ConfigOverrides{}, nil,
	).ClientConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// getPrivateK8sConfig 获取私有k8s配置
func getPrivateK8sConfig(path string) (*clientcmdapi.Config, error) {
	var (
		data []byte
		err  error
	)
	if strings.HasPrefix(path, "private/") {
		data, err = cm.DefaultClient().GetOriginFile(path, "")
	} else {
		data, err = ioutil.ReadFile(path)
	}
	if err != nil {
		return nil, err
	}

	// 读取配置
	return clientcmd.Load(data)
}

// GetK8sTypedClient 获取k8s静态资源客户端
func (s *Service) GetK8sTypedClient(clusterName entity.ClusterName, env string) (kubernetes.Interface, error) {
	info, err := s.getClusterInfo(clusterName, env)
	if err != nil {
		return nil, err
	}

	return info.typedClient, nil
}

// GetK8sDynamicClient 获取k8s动态资源客户端
func (s *Service) GetK8sDynamicClient(clusterName entity.ClusterName, env string) (dynamic.Interface, error) {
	info, err := s.getClusterInfo(clusterName, env)
	if err != nil {
		return nil, err
	}

	return info.dynamicClient, nil
}

// getClusterInfo 获取指定k8s集群信息
func (s *Service) getClusterInfo(name entity.ClusterName, env string) (*k8sCluster, error) {
	if env == "" {
		return nil, errors.Wrap(_errcode.NamespaceNotExistsError, env)
	}

	clusterSet, ok := s.k8sClusters[entity.AppEnvName(env)]
	if !ok {
		return nil, errors.Wrapf(_errcode.ClusterNotExistsError, "no cluster with env(%s)", env)
	}

	cluster, ok := clusterSet[name]
	if !ok {
		return nil, errors.Wrapf(_errcode.ClusterNotExistsError, "cluster(%s) with env(%s)", name, env)
	}

	return cluster, nil
}

// getVendorController 获取运营商控制器
func (s *Service) getVendorController(vendorName entity.VendorName) (vendors.Controller, error) {
	c, ok := s.vendors[vendorName]
	if !ok {
		return nil, errors.Wrap(_errcode.UnknownVendorError, string(vendorName))
	}

	return c, nil
}

// getVendorControllerFromClusterAndNamespace 通过集群和命名空间获取运营商控制器
func (s *Service) getVendorControllerFromClusterAndNamespace(clusterName entity.ClusterName,
	namespace string) (vendors.Controller, error) {
	clusterInfo, err := s.getClusterInfo(clusterName, namespace)
	if err != nil {
		return nil, err
	}

	return s.getVendorController(clusterInfo.vendor)
}

// loadVendorConfigsMapping 加载云服务商配置
func loadVendorConfigsMapping() (map[entity.VendorName]*k8sVendorConfigs, error) {
	vendorConfigsMapping := make(map[entity.VendorName]*k8sVendorConfigs)
	for _, vendorCfg := range config.Conf.Vendors {
		if !entity.CheckVendorSupport(entity.VendorName(vendorCfg.Name)) {
			return nil, errors.Errorf("unsupported vendor(%s)", vendorCfg.Name)
		}

		if _, ok := vendorConfigsMapping[entity.VendorName(vendorCfg.Name)]; ok {
			return nil, errors.Errorf("duplicate vendor_name(%s) found", vendorCfg.Name)
		}

		vendorConfigsMapping[entity.VendorName(vendorCfg.Name)] = &k8sVendorConfigs{
			baseConfig: vendorCfg,
			controllerConfig: &vendors.ControllerConfig{
				VendorName:       entity.VendorName(vendorCfg.Name),
				AccessKeyID:      vendorCfg.AccessKeyID,
				AccessKeySecret:  vendorCfg.AccessKeySecret,
				DisableLogConfig: vendorCfg.DisableLogConfig,
				LogEndpoint:      vendorCfg.LogEndpoint,
				RegionID:         vendorCfg.RegionID,
				// HTTPClient 最后直接使用 Service 创建的 HTTP 客户端
				// Clusters 在读取 k8s 集群配置时填充内容
				Clusters: make(map[entity.AppEnvName]map[entity.ClusterName]*vendors.ClusterConfig),
			},
		}
	}

	return vendorConfigsMapping, nil
}

// loadK8sClustersAndClusterConfigsInVendorConfigs 加载 k8s 集群配置及运营商配置中的集群配置
func loadK8sClustersAndClusterConfigsInVendorConfigs(
	vendorConfigsMapping map[entity.VendorName]*k8sVendorConfigs) (map[entity.AppEnvName]map[entity.ClusterName]*k8sCluster, error) {
	k8sClusters := make(map[entity.AppEnvName]map[entity.ClusterName]*k8sCluster)
	defaultClusterFound := false
	for envName, clusterCfgs := range config.Conf.K8sClusters {
		clusterSet := make(map[entity.ClusterName]*k8sCluster)
		for _, clusterCfg := range clusterCfgs {
			vendorConfigs, ok := vendorConfigsMapping[entity.VendorName(clusterCfg.Vendor)]
			if !ok {
				return nil, errors.Errorf("vendor(%s) not found for cluster(%s) in env(%s)", clusterCfg.Vendor, clusterCfg.Name, envName)
			}

			if clusterCfg.Name == "" {
				return nil, errors.Errorf("no cluster_name found in cluster config(%+v)", clusterCfg)
			}

			if clusterCfg.Name == string(entity.DefaultClusterName) {
				defaultClusterFound = true
			}

			cluster, e := initCluster(entity.AppEnvName(envName), clusterCfg, vendorConfigs.baseConfig)
			if e != nil {
				return nil, e
			}

			if _, ok = clusterSet[cluster.name]; ok {
				return nil, errors.Errorf("duplicate cluster_name(%s) found in env(%s)", cluster.name, cluster.envName)
			}

			clusterSet[cluster.name] = cluster

			clusterMapping, ok := vendorConfigs.controllerConfig.Clusters[cluster.envName]
			if !ok {
				clusterMapping = make(map[entity.ClusterName]*vendors.ClusterConfig)
				vendorConfigs.controllerConfig.Clusters[cluster.envName] = clusterMapping
			}

			clusterMapping[cluster.name] = &vendors.ClusterConfig{
				ClusterID:           clusterCfg.ClusterID,
				ClusterNameInVendor: clusterCfg.ClusterNameInVendor,
				LogGroupID:          clusterCfg.LogBucketID,
				LogGroupName:        clusterCfg.LogBucketName,
				K8sTypedClient:      cluster.typedClient,
				K8sDynamicClient:    cluster.dynamicClient,
				Region:              clusterCfg.Region,
			}
		}

		k8sClusters[entity.AppEnvName(envName)] = clusterSet
	}

	if !defaultClusterFound {
		return nil, errors.New("no default cluster found")
	}

	return k8sClusters, nil
}

// checkKongConfigs 校验 Kong 配置项
func checkKongConfigs() error {
	for env, envCfg := range config.Conf.Kong.Envs {
		supported := false
		for _, supportedEnvName := range entity.AppNormalEnvNames {
			if env == string(supportedEnvName) {
				supported = true
				break
			}
		}

		if !supported {
			return errors.Errorf("unsupported env name(%s) found in kong admin host config", env)
		}

		if envCfg.AdminHost == nil {
			return errors.Errorf("no admin host found for env name(%s) in kong admin host config", env)
		}

		if envCfg.Address == "" {
			return errors.Errorf("no load balancer found for env name(%s) in kong admin host config", env)
		}
	}

	return nil
}

// initCluster 注册集群
func initCluster(envName entity.AppEnvName, clusterCfg *config.K8sClusterConfig, vendorCfg *config.VendorConfig) (*k8sCluster, error) {
	supportedK8sObjectKindsByVendor, ok := entity.AMSSupportedK8sObjectKindsByVendors[entity.VendorName(vendorCfg.Name)]
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "missing object_kinds configuration for vendor(%s)", vendorCfg.Name)
	}

	if vendorCfg.ImageRegistryConfig == nil ||
		vendorCfg.ImageRegistryConfig.Host == "" ||
		vendorCfg.ImageRegistryConfig.Namespace == "" {
		return nil, errors.Wrapf(_errcode.K8sInternalError,
			"missing image_registry_host(%#v) configuration for vendor(%s)", vendorCfg.ImageRegistryConfig, vendorCfg.Name)
	}

	k8sConfig, err := getK8sConfig(clusterCfg.KubeConfigPath, clusterCfg.ContextName)
	if err != nil {
		return nil, err
	}
	k8sTypedClient, err := newK8sTypedClient(k8sConfig)
	if err != nil {
		return nil, err
	}
	k8sDynamicClient, err := newK8sDynamicClient(k8sConfig)
	if err != nil {
		return nil, err
	}

	discovery := k8sTypedClient.Discovery()
	serverVersion, err := discovery.ServerVersion()
	if err != nil {
		return nil, err
	}

	// k8s 推荐的 apiVersion 是稳定版优先, 并非最新版优先, AMS 期望的最新版本需要自己筛选
	_, resourceList, err := discovery.ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}

	k8sGroupVersions := make(map[string]*schema.GroupVersion)
	for _, group := range resourceList {
		newVer, e := entity.ParseK8sGroupVersionString(group.GroupVersion)
		if e != nil {
			return nil, e
		}

		var preferred bool
		for i := range group.APIResources {
			kind := group.APIResources[i].Kind
			if _, ok = entity.AMSSupportedK8sObjectKindsByAllVendors[kind]; !ok {
				if _, ok = supportedK8sObjectKindsByVendor[kind]; !ok {
					continue
				}
			}

			oldVer := k8sGroupVersions[kind]
			preferred, e = newVer.IsPreferredThan(oldVer)
			if e != nil {
				return nil, e
			}

			if preferred {
				k8sGroupVersions[kind] = newVer.GroupVersion
			}
		}
	}

	// 确认所有 AMS 支持的资源类型都有版本控制
	var missingKinds []string
	for kind := range entity.AMSSupportedK8sObjectKindsByAllVendors {
		if _, ok = k8sGroupVersions[kind]; !ok {
			missingKinds = append(missingKinds, kind)
		}
	}
	for kind := range supportedK8sObjectKindsByVendor {
		if _, ok = k8sGroupVersions[kind]; !ok {
			missingKinds = append(missingKinds, kind)
		}
	}
	// if len(missingKinds) > 0 {
	// 	return nil, errors.Errorf("missing AMS supported k8s kinds(cluster=%s, env=%s): %+v", clusterCfg.Name, envName, missingKinds)
	// }

	cluster := &k8sCluster{
		id:                             clusterCfg.ClusterID,
		nameInVendor:                   clusterCfg.ClusterNameInVendor,
		name:                           entity.ClusterName(clusterCfg.Name),
		vendor:                         entity.VendorName(vendorCfg.Name),
		envName:                        envName,
		version:                        serverVersion,
		config:                         k8sConfig,
		typedClient:                    k8sTypedClient,
		dynamicClient:                  k8sDynamicClient,
		tlsSecretName:                  clusterCfg.TLSSecretName,
		disableLogConfig:               vendorCfg.DisableLogConfig,
		imageRegistryHostWithNamespace: fmt.Sprintf("%s/%s", vendorCfg.ImageRegistryConfig.Host, vendorCfg.ImageRegistryConfig.Namespace),
		k8sGroupVersions:               make(map[string]*schema.GroupVersion, len(k8sGroupVersions)),
		ingressClass:                   clusterCfg.IngressClass,
		localDNS:                       clusterCfg.LocalDNS,
		visibleProjectIDs:              clusterCfg.VisibleProjectIDs,
	}

	cluster.k8sGroupVersions = k8sGroupVersions

	// HACK: 保证运行稳定, 先采用原先使用的 k8s apiVersion, 完全迁移完新集群后再采用自动选取最新版本的方式
	err = forceK8sObjectKinds(cluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

type SchemeRegisterFunc func(s *runtime.Scheme) error

func registerCRDsInScheme(registerFuncs ...SchemeRegisterFunc) error {
	for _, r := range registerFuncs {
		err := r(scheme.Scheme)
		if err != nil {
			return err
		}
	}
	return nil
}
