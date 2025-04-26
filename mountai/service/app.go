package service

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/slice"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/config"
	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
)

// 已支持的 sentry 操作枚举值
const (
	CreativeSentryProjectSlug = "#create"
)

// 名称正则表达式 k8s不能以-开头结尾
var appNameRegexp = regexp.MustCompile(`^([0-9a-z]+[0-9a-z\-]*)?[0-9a-z]$`)

// CreateApp 创建应用
func (s *Service) CreateApp(ctx context.Context, createReq *req.CreateAppReq,
	project *resp.ProjectDetailResp) (*resp.AppDetailResp, error) {
	// 创建实体
	now := time.Now()
	app := &entity.App{
		ID:               primitive.NewObjectID(),
		ServiceName:      createReq.GetServiceName(project.Name),
		AliLogConfigName: createReq.GetAliLogConfigName(project.Name),
		Env:              createReq.GetDefaultEnv(project.Name, createReq.EnableBranchChangeNotification),
		EnableIstio:      createReq.EnableIstio.ValueOrZero(),
		CreateTime:       &now,
		UpdateTime:       &now,
	}

	err := deepcopy.Copy(createReq).To(app)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	if createReq.SentryProjectSlug != "" {
		if createReq.SentryProjectSlug == CreativeSentryProjectSlug {
			createSentryResp, e := s.CreateAppSentry(ctx, project.Team, project, &resp.AppDetailResp{
				Name: app.Name,
			})
			if e != nil {
				return nil, e
			}

			app.SentryProjectPublicDsn = createSentryResp.SentryProjectPublicDsn
			app.SentryProjectSlug = createSentryResp.SentryProjectSlug
		} else {
			sentryProjectPublicDsn, e := s.GetSentryProjectPublicDsn(ctx, &req.GetSentryProjectKeyReq{
				ProjectSlug: createReq.SentryProjectSlug,
			})
			if e != nil {
				return nil, e
			}

			app.SentryProjectPublicDsn = sentryProjectPublicDsn
			app.SentryProjectSlug = createReq.SentryProjectSlug
		}
	}

	// 落库
	_, err = s.dao.CreateSingleApp(ctx, app)
	if err != nil {
		return nil, err
	}

	// 返回
	res := &resp.AppDetailResp{}

	err = deepcopy.Copy(app).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// 生成更新应用的map
func (s *Service) generateUpdateAppMap(_ context.Context, updateReq *req.UpdateAppReq) map[string]interface{} {
	change := make(map[string]interface{})
	change["update_time"] = time.Now()

	if updateReq.Name != "" {
		change["name"] = updateReq.Name
	}

	if len(updateReq.Env) != 0 {
		for name, env := range updateReq.Env {
			if env.ServiceProtocol != "" {
				change[fmt.Sprintf("env.%s.service_protocol", name)] = env.ServiceProtocol
			}
			if env.LogStoreName != "" {
				change[fmt.Sprintf("env.%s.log_store_name", name)] = env.LogStoreName
			}
			if env.AliAlarmName != "" {
				change[fmt.Sprintf("env.%s.ali_alarm_name", name)] = env.AliAlarmName
			}
			if env.LogTailName != "" {
				change[fmt.Sprintf("env.%s.log_tail_name", name)] = env.LogTailName
			}
			if !env.EnableBranchChangeNotification.IsZero() {
				change[fmt.Sprintf("env.%s.enable_branch_change_notification", name)] = env.EnableBranchChangeNotification.ValueOrZero()
			}
			if !env.EnableHotReload.IsZero() {
				change[fmt.Sprintf("env.%s.enable_hot_reload", name)] = env.EnableHotReload.ValueOrZero()
			}
		}
	}

	if updateReq.SentryProjectSlug != "" {
		change["sentry_project_slug"] = updateReq.SentryProjectSlug
	}
	if updateReq.SentryProjectPublicDsn != "" {
		change["sentry_project_public_dsn"] = updateReq.SentryProjectPublicDsn
	}
	if !updateReq.Description.IsZero() {
		change["description"] = updateReq.Description.ValueOrZero()
	}

	change["enable_istio"] = updateReq.EnableIstio.ValueOrZero()

	return change
}

// UpdateApp 更新应用
func (s *Service) UpdateApp(ctx context.Context, id string, updateReq *req.UpdateAppReq) error {
	// 生成更新map
	changeMap := s.generateUpdateAppMap(ctx, updateReq)

	// 更新库
	err := s.dao.UpdateSingleApp(ctx, id, bson.M{
		"$set": changeMap,
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteSingleApp 删除单个应用
func (s *Service) DeleteSingleApp(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	err = s.dao.DeleteSingleApp(ctx, bson.M{
		"_id": objectID,
	})
	if err != nil {
		return err
	}

	return nil
}

// GetAppDetail 获取应用详情
func (s *Service) GetAppDetail(ctx context.Context, id string) (*resp.AppDetailResp, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	app, err := s.GetAppByObjectID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	res := &resp.AppDetailResp{}

	err = deepcopy.Copy(app).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	// 兼容历史数据
	if app.Type == entity.AppTypeService && app.ServiceExposeType == "" {
		res.ServiceExposeType = entity.AppServiceExposeTypeIngress
	}

	return res, nil
}

func (s *Service) GetAppByObjectID(ctx context.Context, objectID primitive.ObjectID) (*entity.App, error) {
	return s.dao.FindSingleApp(ctx, bson.M{"_id": objectID})
}

// GetAppDetailByIDAndIgnoreDeleteStatus 通过主键获取应用详情(忽略删除状态)
func (s *Service) GetAppDetailByIDAndIgnoreDeleteStatus(ctx context.Context, id string) (*resp.AppDetailResp, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	app, err := s.dao.FindSingleAppByObjectID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	res := &resp.AppDetailResp{}

	err = deepcopy.Copy(app).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	// 兼容历史数据
	if app.Type == entity.AppTypeService && app.ServiceExposeType == "" {
		res.ServiceExposeType = entity.AppServiceExposeTypeIngress
	}

	return res, nil
}

// 获取应用过滤器
func (s *Service) getAppsFilter(_ context.Context, getReq *req.GetAppsReq) (bson.M, error) {
	filter := bson.M{}
	if getReq.ProjectID != "" {
		filter["project_id"] = getReq.ProjectID
	}

	if getReq.Name != "" {
		filter["name"] = getReq.Name
	}

	if getReq.Type != "" {
		filter["type"] = getReq.Type
	}

	if getReq.ServiceType != "" {
		filter["service_type"] = getReq.ServiceType
	}

	if getReq.ServiceName != "" {
		filter["service_name"] = getReq.ServiceName
	}

	if getReq.Keyword != "" {
		filter["name"] = bson.M{
			"$regex": getReq.Keyword,
		}
	}

	if len(getReq.ProjectIDs) != 0 {
		filter["project_id"] = bson.M{
			"$in": getReq.ProjectIDs,
		}
	}

	if getReq.AliLogConfigName != "" {
		filter["ali_log_config_name"] = getReq.AliLogConfigName
	}

	if getReq.EnvName != "" {
		filter[fmt.Sprintf("env.%s", getReq.EnvName)] = bson.M{
			"$exists": true,
		}
	}
	if len(getReq.IDs) > 0 {
		objectIDs := make([]primitive.ObjectID, len(getReq.IDs))
		for i, id := range getReq.IDs {
			objectID, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				return nil, errors.Wrap(_errcode.InvalidHexStringError, err.Error())
			}
			objectIDs[i] = objectID
		}
		filter["_id"] = bson.M{
			"$in": objectIDs,
		}
	}

	return filter, nil
}

// GetApps 获取应用列表
func (s *Service) GetApps(ctx context.Context, getReq *req.GetAppsReq) ([]*resp.AppListResp, error) {
	filter, err := s.getAppsFilter(ctx, getReq)
	if err != nil {
		return nil, err
	}

	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit

	apps, err := s.dao.FindApps(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})
	if err != nil {
		return nil, err
	}

	res := make([]*resp.AppListResp, 0)

	err = deepcopy.Copy(&apps).To(&res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// GetAppsDetails 获取应用详情列表
func (s *Service) GetAppsDetails(ctx context.Context, getReq *req.GetAppsReq) ([]*resp.AppDetailResp, error) {
	filter, err := s.getAppsFilter(ctx, getReq)
	if err != nil {
		return nil, err
	}

	limit := int64(getReq.Limit)
	skip := int64(getReq.Page-1) * limit

	apps, err := s.dao.FindApps(ctx, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
		Sort:  dao.MongoSortByIDAsc,
	})
	if err != nil {
		return nil, err
	}

	res := make([]*resp.AppDetailResp, 0)

	err = deepcopy.Copy(&apps).To(&res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// GetAppsCount 获取应用数量
func (s *Service) GetAppsCount(ctx context.Context, getReq *req.GetAppsReq) (int, error) {
	filter, err := s.getAppsFilter(ctx, getReq)
	if err != nil {
		return 0, err
	}

	res, err := s.dao.CountApps(ctx, filter)
	if err != nil {
		return 0, err
	}

	return res, nil
}

// CheckAppNameLegal 校验应用名合法性
func (s *Service) CheckAppNameLegal(ctx context.Context, project *resp.ProjectDetailResp, appType entity.AppType, name string) error {
	// 校验非法字符
	if !appNameRegexp.MatchString(name) {
		return errors.Wrapf(errcode.InvalidParams, "name:%s contains invalid char", name)
	}

	// 总长度不能超过50
	if len(s.GenerateTaskVersion(project.Name, name, appType, "", time.Now())) > 50 {
		return errors.Wrapf(_errcode.TaskVersionNameLengthError, "project:%s app %s ", project.Name, name)
	}

	// 应用名不可重复
	count, err := s.GetAppsCount(ctx, &req.GetAppsReq{
		ProjectID: project.ID,
		Name:      name,
	})
	if err != nil {
		return err
	} else if count != 0 {
		return errors.Wrapf(_errcode.AppNameDuplicateError, "project %s app %s", project.Name, name)
	}

	// 容器名不可重复
	sameContainerCount, err := s.GetAppsCount(ctx, &req.GetAppsReq{
		AliLogConfigName: utils.GetPodContainerName(project.Name, name),
	})
	if err != nil {
		return err
	}

	if sameContainerCount > 0 {
		return errors.Wrapf(_errcode.ContainerNameNotUniqueError, "project:%s app %s", project.Name, name)
	}

	return nil
}

// GetAppLoadBalancersByEnvAndCluster 获取应用在指定环境、集群的 LB 实例
func (s *Service) GetAppLoadBalancersByEnvAndCluster(_ context.Context, app *resp.AppDetailResp, clusterName entity.ClusterName,
	envName entity.AppEnvName) []entity.ServiceLoadBalancer {
	lbs := make([]entity.ServiceLoadBalancer, 0, len(app.LoadBalancerInfo))
	for _, lbInfo := range app.LoadBalancerInfo {
		if lbInfo.Cluster == clusterName && lbInfo.Env == envName {
			lbs = append(lbs, lbInfo)
		}
	}

	return lbs
}

func (s *Service) UpdateAppLoadBalancerInfo(ctx context.Context, app *resp.AppDetailResp, createReq *req.CreateTaskReq) error {
	update := new(req.UpdateAppReq)

	item := entity.ServiceLoadBalancer{
		Env:                createReq.EnvName,
		Cluster:            createReq.ClusterName,
		LoadBalancerID:     createReq.Param.LoadBalancerID,
		LoadbalancerCertId: createReq.Param.LoadBalancerCertID,
	}

	updatedLoadBalancerInfo := append(app.LoadBalancerInfo, item)
	update.LoadBalancerInfo = updatedLoadBalancerInfo

	err := s.UpdateApp(ctx, app.ID, update)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getGrafanaAppPath(envName entity.AppEnvName, clusterName entity.ClusterName) (string, error) {
	clusters := config.Conf.K8sClusters[string(envName)]

	for _, cluster := range clusters {
		if cluster.Name == string(clusterName) {
			return fmt.Sprintf("%s%s", cluster.GrafanaHost, cluster.GrafanaAppPath), nil
		}
	}

	return "", errors.Wrapf(errcode.InvalidParams, "no grafana app path found for env[%s], cluster[%s]", envName, clusterName)
}

// 获取应用监控跳转url
func (s *Service) getAppMonitorURL(project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	team *resp.TeamDetailResp, envName entity.AppEnvName, clusterName entity.ClusterName) (string, error) {
	dataSource := config.Conf.K8s.PrdContextName
	if envName == entity.AppEnvStg || envName == entity.AppEnvFat {
		dataSource = config.Conf.K8s.StgContextName
	}

	dataSource = fmt.Sprintf("%s-%s", clusterName, dataSource)

	grafanaPath, err := s.getGrafanaAppPath(envName, clusterName)
	if err != nil {
		return "", err
	}

	if s.GetApplicationIstioState(context.Background(), envName, clusterName, app) {
		envName = entity.IstioNamespacePrefix + envName
	}

	return fmt.Sprintf("%s?var-datasource=%s&var-namespace=%s&var-team=%s&var-project=%s&var-app=%s",
		grafanaPath, dataSource, envName, team.Label, project.Name, app.Name), nil
}

// GetAppExtraInfo 获取应用额外信息
func (s *Service) GetAppExtraInfo(ctx context.Context, clusterName entity.ClusterName,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	envName entity.AppEnvName) (*resp.AppEnvExtraDetailResp, error) {
	if _, ok := app.Env[envName]; !ok {
		return nil, errors.Wrapf(errcode.InvalidParams, "project: %v app: %v env:%v", project.Name, app.Name, envName)
	}

	clusterInfo, err := s.getClusterInfo(clusterName, string(envName))
	if err != nil {
		return nil, err
	}

	c, err := s.getVendorController(clusterInfo.vendor)
	if err != nil {
		return nil, err
	}

	logStoreURL, logStoreURLBasedProject, err := c.GetLogStoreURL(ctx, clusterName, project, app, envName)
	if err != nil {
		return nil, err
	}

	monitorURL, err := s.getAppMonitorURL(project, app, project.Team, envName, clusterName)
	if err != nil {
		return nil, err
	}

	res := &resp.AppEnvExtraDetailResp{
		LogStoreURL:             logStoreURL,
		LogStoreURLBasedProject: logStoreURLBasedProject,
		AccessHosts:             make([]string, 0),
		MonitorURL:              monitorURL,
	}

	if app.Type == entity.AppTypeService {
		switch app.ServiceExposeType {
		case entity.AppServiceExposeTypeLB:
			_, err := s.GetServiceDetail(ctx, clusterName,
				&req.GetServiceDetailReq{
					Namespace: string(envName),
					Name:      app.ServiceName,
					Env:       string(envName),
				})
			if err == nil {
				res.AccessHosts = append(
					res.AccessHosts,
					s.getAliPrivateZoneK8sFullDomainName(app.ServiceName, envName),
					s.getAliPrivateZoneK8sFullDomainNameWithCluster(app.ServiceName, envName, clusterName),
				)
			} else if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				return nil, err
			}
		case entity.AppServiceExposeTypeIngress:
			projectInJuncle := false
			juncleDomainSuffix := string(entity.ClusterJuncle + ".cluster.local")

			clusterInfo, err := s.getClusterInfo(entity.ClusterJuncle, string(envName))
			if err == nil {
				projectInJuncle = slice.StrSliceContains(clusterInfo.visibleProjectIDs, app.ProjectID)
			}

			if s.GetApplicationIstioState(ctx, envName, clusterName, app) {
				vs, vsErr := s.GetVirtualServiceDetail(ctx, clusterName, envName, &req.VirtualServiceReq{
					Namespace: string(entity.IstioNamespacePrefix + envName),
					Name:      project.Name + "-" + app.Name,
					Env:       string(envName),
				})

				if vsErr != nil {
					if errcode.EqualError(_errcode.K8sResourceNotFoundError, vsErr) {
						return res, nil
					}
					return nil, vsErr
				}

				for i := range vs.Spec.Hosts {
					if !projectInJuncle && strings.LastIndex(vs.Spec.Hosts[i], juncleDomainSuffix) > 0 {
						continue
					}

					res.AccessHosts = append(res.AccessHosts, vs.Spec.Hosts[i])
				}
			} else {
				// 全量 Ingress 已推广，通过 Ingress 方式暴露的服务，域名必定来自于 Ingress 记录
				accessHosts, err := s.getIngressAccessHosts(ctx, clusterName, app.ServiceName, string(envName))
				if err != nil {
					return nil, err
				}

				res.AccessHosts = append(res.AccessHosts, accessHosts...)
			}
		case entity.AppServiceExposeTypeInternal:
			// 内部不暴露，只能通过service name访问
			res.AccessHosts = append(res.AccessHosts, app.ServiceName)
		}
	}

	return res, nil
}

const IngressPubSuffix = "-pub"

func (s *Service) getIngressAccessHosts(ctx context.Context, clusterName entity.ClusterName, serviceName, envName string) (accessHosts []string, err error) {
	serviceNames := []string{
		serviceName,
		serviceName + IngressPubSuffix,
	}

	for _, sn := range serviceNames {
		ingress, err := s.GetIngressDetail(ctx, clusterName,
			&req.GetIngressDetailReq{
				Namespace: string(envName),
				Name:      sn,
				Env:       string(envName),
			})
		if err != nil {
			if errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
				continue
			}

			return nil, err
		}

		for _, i := range ingress.Spec.Rules {
			accessHosts = append(accessHosts, i.Host)
		}
	}

	return
}

// CalculateAppRecommendResource 计算推荐配置
// 推荐的目标是提高资源利用率，而非降低扩缩容概率(ams原版本)
// 取数原则：资源最小使用量取近一天最低点，降低扩容概率；资源最高使用量，取近一周最高点，资源冗余以应对流量波动
func (s *Service) CalculateAppRecommendResource(_ context.Context,
	getReq *req.CalculateAppRecommendReq) (*resp.CalculateAppRecommendResp, error) {
	res := new(resp.CalculateAppRecommendResp)

	// 除以扩容阈值得到真正需要的resource request
	// CPU
	HPACPUThreshold := float64(DefaultHPACpuTarget) / 100
	dailyNeededCPU := getReq.DailyMinTotalCPU / HPACPUThreshold
	weeklyNeededCPU := getReq.WeeklyMaxTotalCPU / HPACPUThreshold
	// Memo
	HPAMemThreshold := float64(DefaultHPAMemTarget) / 100
	dailyNeededMem := getReq.DailyMinTotalMem / HPAMemThreshold
	weeklyNeededMem := getReq.WeeklyMaxTotalMem / HPAMemThreshold

	var minPodCount, maxPodCount int
	// 以cpu为主要计算资源
	cpuResourceList := entity.CPUStgRequestResourceList
	if getReq.EnvName == entity.AppEnvPrd || getReq.EnvName == entity.AppEnvPre {
		cpuResourceList = entity.CPUPrdRequestResourceList
	}

	for _, cpuResource := range cpuResourceList {
		cpu, err := strconv.ParseFloat(string(cpuResource), 64)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		dailyValuation := dailyNeededCPU / cpu
		weeklyValuation := weeklyNeededCPU / cpu
		curMinPodCount := math.Ceil(dailyValuation)
		curMaxPodCount := math.Ceil(weeklyValuation)
		// 所需cpu/最小推荐cpu小于0.2，视为浪费，尝试候选列表下一个规格
		if dailyNeededCPU/(curMinPodCount*cpu) < 0.2 {
			continue
		}

		res.RecommendCPURequest = cpuResource
		minPodCount = int(curMinPodCount)
		maxPodCount = int(curMaxPodCount)
		break
	}

	// 默认给最低规格
	if res.RecommendCPURequest == "" {
		res.RecommendCPURequest = entity.CPUResourceNano
		minPodCount, maxPodCount = 1, 1
	}

	memResourceList := entity.MemStgRequestResourceList
	if getReq.EnvName == entity.AppEnvPrd || getReq.EnvName == entity.AppEnvPre {
		memResourceList = entity.MemPrdRequestResourceList
	}

	for _, memResource := range memResourceList {
		memGB, err := strconv.ParseFloat(
			strings.TrimSuffix(string(memResource), "Gi"),
			64,
		)
		if err != nil {
			return nil, errors.Wrap(errcode.InternalError, err.Error())
		}

		memBytes := memGB * 1024 * 1024 * 1024
		dailyValuation := dailyNeededMem / memBytes
		weeklyValuation := weeklyNeededMem / memBytes
		curMinPodCount := math.Ceil(dailyValuation)
		curMaxPodCount := math.Ceil(weeklyValuation)

		// 最小规格
		if memResource == entity.MemResourceNano && curMinPodCount == 1 {
			res.RecommendMemRequest = memResource
			break
		}

		// 所需memo/最小推荐memo小于0.2，视为浪费，尝试候选列表下一个规格
		if dailyNeededMem/(curMinPodCount*memBytes) < 0.2 {
			continue
		}

		if weeklyNeededMem/(math.Max(float64(maxPodCount), curMaxPodCount)*memBytes) < 0.2 {
			continue
		}

		res.RecommendMemRequest = memResource
		minPodCount = int(math.Max(float64(minPodCount), curMinPodCount))
		maxPodCount = int(math.Max(float64(maxPodCount), curMaxPodCount))
		break
	}

	res.RecommendMinPodCount = strconv.Itoa(minPodCount)
	// 最大加一，避免到达HPA上界
	res.RecommendMaxPodCount = strconv.Itoa(maxPodCount + 1)
	return res, nil
}
