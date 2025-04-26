package service

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"time"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"
)

var k8sServerVersionPattern = regexp.MustCompile(`v\d+\.\d+\.\d+`)

// SetMultiClusterSupportForProject 为项目添加多集群支持
func (s *Service) SetMultiClusterSupportForProject(ctx context.Context, projectID string) error {
	return s.dao.SetMultiClusterSupportForProject(ctx, projectID)
}

// GetClusterConfig 获取集群配置
func (s *Service) GetClusterConfig(_ context.Context, name entity.ClusterName, namespace string) (*rest.Config, error) {
	info, err := s.getClusterInfo(name, namespace)
	if err != nil {
		return nil, err
	}

	return info.config, nil
}

// GetClusterDetail 获取集群详情
func (s *Service) GetClusterDetail(_ context.Context, name entity.ClusterName, namespace string) (*resp.ClusterDetailResp, error) {
	info, err := s.getClusterInfo(name, namespace)
	if err != nil {
		return nil, err
	}

	return &resp.ClusterDetailResp{
		Name:          info.name,
		IsDefault:     info.name == entity.DefaultClusterName,
		ServerVersion: info.version.String(),
		Env:           info.envName,
	}, nil
}

// GetClusters 获取集群列表
func (s *Service) GetClusters(_ context.Context, getReq *req.GetClustersReq) ([]*resp.ClusterDetailResp, error) {
	var clusterEnv entity.AppEnvName
	if getReq.Namespace != "" {
		clusterEnv = entity.AppEnvName(getReq.Namespace)
	}

	// 当前算上测试的可能性，集群数量应该不会超过 8 个
	res := make([]*resp.ClusterDetailResp, 0, 8)

	var prefix string
	if getReq.Version != "" {
		prefix = k8sServerVersionPattern.FindString(getReq.Version)
		if prefix == "" {
			return res, nil
		}
	}

	nameMapping := make(map[entity.ClusterName]struct{}, len(getReq.ClusterNames))
	for _, inputName := range getReq.ClusterNames {
		nameMapping[inputName] = struct{}{}
	}
	needNameCheck := len(nameMapping) > 0

	for envName, clusterSet := range s.k8sClusters {
		if clusterEnv != "" && envName != clusterEnv {
			continue
		}

		for clusterName, info := range clusterSet {
			if needNameCheck {
				if _, ok := nameMapping[clusterName]; !ok {
					continue
				}
			}

			ver := info.version.String()
			if getReq.Version != "" || !strings.Contains(ver, prefix) {
				continue
			}

			res = append(res, &resp.ClusterDetailResp{
				Name:          info.name,
				IsDefault:     info.name == entity.DefaultClusterName,
				ServerVersion: ver,
				Env:           info.envName,
			})
		}
	}

	sort.SliceStable(res, func(i, j int) bool {
		if res[i].Env == res[j].Env {
			return res[i].Name < res[j].Name
		}

		return res[i].Env < res[j].Env
	})

	return res, nil
}

// GetProjectAppsClustersWithWorkload 批量获取项目下多个应用各自在指定环境下有工作负载的集群列表
func (s *Service) GetProjectAppsClustersWithWorkload(ctx context.Context, envName entity.AppEnvName,
	projectID string, appIDs []string) ([]*resp.ProjectAppsClustersWithWorkloadResp, error) {
	// 当前未开放多集群支持的如果曾经有过多集群操作需要隐藏
	supportedClusters, err := s.GetProjectSupportedClusters(ctx, projectID, envName)
	if err != nil {
		return nil, err
	}

	// 同步过数据的应用通过 redis 可以直接查询到现存的部署
	appsDeployedTasksMapping, err := s.dao.GetAppsRunningTasks(ctx, envName, appIDs)
	if err != nil {
		return nil, err
	}

	clustersMapping := make(map[entity.ClusterName]*resp.ClusterDetailResp)
	for _, cluster := range supportedClusters {
		if cluster.Env == envName {
			clustersMapping[cluster.Name] = cluster
		}
	}

	// 处理 redis 已有记录的数据, 不在 redis 中的应用需要查询数据库并同步
	var recordedClusters map[entity.ClusterName]struct{}
	clustersCap := len(supportedClusters)
	appsClustersWithWorkload := make([]*resp.ProjectAppsClustersWithWorkloadResp, len(appIDs))
	for i := range appIDs {
		appsClustersWithWorkload[i] = &resp.ProjectAppsClustersWithWorkloadResp{
			AppID:    appIDs[i],
			Clusters: make([]*resp.ClusterDetailResp, 0, clustersCap),
		}

		tasks, ok := appsDeployedTasksMapping[appIDs[i]]
		if !ok {
			continue
		}

		recordedClusters = make(map[entity.ClusterName]struct{}, clustersCap) // 每个集群只添加一条记录, 需要排重
		for _, task := range tasks {
			if cluster, ok := clustersMapping[task.ClusterName]; ok {
				if _, ok = recordedClusters[task.ClusterName]; !ok {
					appsClustersWithWorkload[i].Clusters = append(appsClustersWithWorkload[i].Clusters, cluster)
					recordedClusters[task.ClusterName] = struct{}{}
				}
			}
		}

		// app_id 置为空字符串标记已从 redis 获得
		appIDs[i] = ""
	}

	var deployedTasks []*resp.TaskDetailResp
	for i := range appIDs {
		if appIDs[i] == "" {
			continue
		}

		deployedTasks, err = s.getAndSyncAppRunningTasks(ctx, envName, appIDs[i])
		if err != nil {
			return nil, err
		}

		recordedClusters = make(map[entity.ClusterName]struct{}, clustersCap) // 每个集群只添加一条记录, 需要排重
		for _, task := range deployedTasks {
			if cluster, ok := clustersMapping[task.ClusterName]; ok {
				if _, ok = recordedClusters[task.ClusterName]; !ok {
					appsClustersWithWorkload[i].Clusters = append(appsClustersWithWorkload[i].Clusters, cluster)
					recordedClusters[task.ClusterName] = struct{}{}
				}
			}
		}
	}

	return appsClustersWithWorkload, nil
}

// GetAppClustersWithWorkload 获取应用在指定环境下有工作负载的所有集群
func (s *Service) GetAppClustersWithWorkload(ctx context.Context,
	envName entity.AppEnvName, project *resp.ProjectDetailResp, app *resp.AppDetailResp) ([]*resp.ClusterDetailResp, error) {
	// 当前未开放多集群支持的如果曾经有过多集群操作需要隐藏
	supportedClusters, err := s.GetProjectSupportedClusters(ctx, project.ID, envName)
	if err != nil {
		return nil, err
	}

	// 同步过数据的应用通过 redis 可以直接查询现存的部署
	deployedTasks, err := s.dao.GetAppRunningTasks(ctx, envName, app.ID)
	if err != nil {
		return nil, err
	}

	clustersMapping := make(map[entity.ClusterName]*resp.ClusterDetailResp)
	for _, cluster := range supportedClusters {
		if cluster.Env == envName {
			clustersMapping[cluster.Name] = cluster
		}
	}

	// 如果 redis 同步过数据, 则 deployedTasks 不会是 nil(没有记录时是空数组)
	if deployedTasks != nil {
		clustersWithWorkload := make([]*resp.ClusterDetailResp, 0, len(supportedClusters))
		for _, task := range deployedTasks {
			if cluster, ok := clustersMapping[task.ClusterName]; ok {
				clustersWithWorkload = append(clustersWithWorkload, cluster)
				delete(clustersMapping, task.ClusterName) // 同一个集群只添加一条记录
			}
		}

		return clustersWithWorkload, nil
	}

	deployedTasks, err = s.getAndSyncAppRunningTasks(ctx, envName, app.ID)
	if err != nil {
		return nil, err
	}

	clustersWithWorkload := make([]*resp.ClusterDetailResp, 0, len(supportedClusters))
	for _, task := range deployedTasks {
		if cluster, ok := clustersMapping[task.ClusterName]; ok {
			clustersWithWorkload = append(clustersWithWorkload, cluster)
			delete(clustersMapping, task.ClusterName) // 同一个集群只添加一条记录
		}
	}

	return clustersWithWorkload, nil
}

// GetProjectSupportedClusters 获取项目支持的集群列表
func (s *Service) GetProjectSupportedClusters(ctx context.Context,
	projectID string, envNames ...entity.AppEnvName) ([]*resp.ClusterDetailResp, error) {
	isStgSupported, err := s.CheckMultiClusterSupport(ctx, entity.AppEnvStg, projectID)
	if err != nil {
		return nil, err
	}

	isPrdSupported, err := s.CheckMultiClusterSupport(ctx, entity.AppEnvPrd, projectID)
	if err != nil {
		return nil, err
	}

	clusters, err := s.GetClusters(ctx, new(req.GetClustersReq))
	if err != nil {
		return nil, err
	}

	projectClusters := make([]*resp.ClusterDetailResp, 0)

	for _, cluster := range clusters {
		if _, ok := s.k8sClusters[cluster.Env]; !ok {
			continue
		}

		if _, ok := s.k8sClusters[cluster.Env][cluster.Name]; !ok {
			continue
		}

		if len(s.k8sClusters[cluster.Env][cluster.Name].visibleProjectIDs) > 0 {
			isVisible := false

			for _, visibleProjectID := range s.k8sClusters[cluster.Env][cluster.Name].visibleProjectIDs {
				if visibleProjectID == projectID {
					isVisible = true
					break
				}
			}

			if !isVisible {
				continue
			}
		}

		projectClusters = append(projectClusters, cluster)
	}

	if len(projectClusters) == 0 || isStgSupported && isPrdSupported {
		return projectClusters, nil
	}

	slow := 0
	for fast := range projectClusters {
		if len(envNames) > 0 {
			ok := false
			for _, envName := range envNames {
				if envName == projectClusters[fast].Env {
					ok = true
					break
				}
			}

			if !ok {
				continue
			}
		}

		if projectClusters[fast].Name == entity.DefaultClusterName ||
			(isStgSupported && (projectClusters[fast].Env == entity.AppEnvStg || projectClusters[fast].Env == entity.AppEnvFat)) ||
			(isPrdSupported && (projectClusters[fast].Env == entity.AppEnvPre || projectClusters[fast].Env == entity.AppEnvPrd)) {
			if slow < fast {
				projectClusters[slow] = projectClusters[fast]
			}
			slow++
		}
	}

	return projectClusters[:slow], nil
}

// CheckAndUnifyClusterName 校验集群名
func (s *Service) CheckAndUnifyClusterName(_ context.Context,
	name entity.ClusterName, envName entity.AppEnvName) (entity.ClusterName, error) {
	if name == entity.EmptyClusterName {
		return entity.DefaultClusterName, nil
	}

	_, err := s.getClusterInfo(name, string(envName))
	if err != nil {
		return entity.EmptyClusterName, errors.Wrapf(errcode.InvalidParams, "cluster(%s) not exists", name)
	}

	return name, nil
}

// CheckMultiClusterSupport 校验项目对应环境是否已支持多集群
func (s *Service) CheckMultiClusterSupport(ctx context.Context, envName entity.AppEnvName, projectID string) (bool, error) {
	// TODO: 多集群迁移已经完成, 目前单集群的域名解析逻辑还没有修改, 所以先临时认为所有项目支持多集群
	return true, nil
}

// getClusterVersion 获取集群版本
func (s *Service) getClusterVersion(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName) (*version.Info, error) {
	clusterName, err := s.CheckAndUnifyClusterName(ctx, clusterName, envName)
	if err != nil {
		return nil, err
	}

	info, err := s.getClusterInfo(clusterName, string(envName))
	if err != nil {
		return nil, err
	}

	return info.version, nil
}

// getK8sResourceGroupVersion 获取集群下对应 k8s 资源的 GroupVersion
func (s *Service) getK8sResourceGroupVersion(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, kind string) (*schema.GroupVersion, error) {
	clusterName, err := s.CheckAndUnifyClusterName(ctx, clusterName, envName)
	if err != nil {
		return nil, err
	}

	clusterInfo, err := s.getClusterInfo(clusterName, string(envName))
	if err != nil {
		return nil, err
	}

	return s.getK8sResourceGroupVersionFromClusterInfo(clusterInfo, kind)
}

// getK8sResourceGroupVersionFromClusterInfo 从集群配置中获取集群下对应 k8s 资源的 GroupVersion
func (s *Service) getK8sResourceGroupVersionFromClusterInfo(clusterInfo *k8sCluster, kind string) (*schema.GroupVersion, error) {
	gv, ok := clusterInfo.k8sGroupVersions[kind]
	if !ok {
		return nil, errors.Wrapf(errcode.InternalError, "invalid k8s API kind(%s)", kind)
	}

	return gv, nil
}

// getAndSyncAppRunningTasks 获取应用在指定环境下正在运行的所有部署
func (s *Service) getAndSyncAppRunningTasks(ctx context.Context, envName entity.AppEnvName, appID string) ([]*resp.TaskDetailResp, error) {
	clusterNames, err := s.getEnvClusterNames(ctx, envName)
	if err != nil {
		return nil, err
	}

	var deployedTasks []*resp.TaskDetailResp
	for _, clusterName := range clusterNames {
		clusterTasks, e := s.getAppClusterRunningTasks(ctx, envName, clusterName, appID)
		if e != nil {
			return nil, e
		}

		deployedTasks = append(deployedTasks, clusterTasks...)
	}

	// 同步至 redis, 忽略失败
	err = s.dao.SetAppRunningTasks(ctx, appID, envName, deployedTasks, true)
	if err != nil {
		log.Warnc(ctx, "set %d app running tasks for app(%s) and env(%s) failed: %s", len(deployedTasks), appID, envName, err)
	}

	return deployedTasks, nil
}

// getAppClusterRunningTasks 分集群获取应用在指定环境下正在运行的部署
func (s *Service) getAppClusterRunningTasks(ctx context.Context,
	envName entity.AppEnvName, clusterName entity.ClusterName, appID string) ([]*resp.TaskDetailResp, error) {
	// 先找到最后一次会造成清理的成功的部署操作的 task(之前的 task 都会被它清理, 不需要关注), 确认需要搜索的范围
	// pre 和 prd 环境会造成清理的部署操作是 全量金丝雀部署
	// 其他环境会造成清理的部署操作是 全量部署
	getReq := &req.GetTasksReq{
		AppID:       appID,
		EnvName:     envName,
		ClusterName: clusterName,
		StatusList:  entity.TaskStatusSuccessStateList,
	}
	if envName == entity.AppEnvPre || envName == entity.AppEnvPrd {
		getReq.Action = entity.TaskActionFullCanaryDeploy
	} else {
		getReq.Action = entity.TaskActionFullDeploy
	}

	lastSuccessfulClearedDeployedTask, err := s.GetSingleTask(ctx, getReq)
	if err != nil && !errcode.EqualError(_errcode.NoRequiredTaskError, err) {
		return nil, err
	}

	// 筛选该 app 在该集群最后一次会造成清理的成功部署操作之后发生的所有删除操作和初始化部署类操作
	getReq.Page = 0
	getReq.Limit = 0
	getReq.StatusList = nil
	getReq.Action = entity.TaskActionDelete // entity.TaskActionClean 的情况不应该会触发本查询
	if lastSuccessfulClearedDeployedTask != nil {
		minCreateTime, e := time.ParseInLocation(utils.DefaultTimeFormatLayout, lastSuccessfulClearedDeployedTask.CreateTime, time.Local)
		if e != nil {
			return nil, errors.Wrapf(errcode.InternalError,
				"parse tasks create_time(%s) failed: %s", lastSuccessfulClearedDeployedTask.CreateTime, e)
		}

		getReq.MinTimestamp = int(minCreateTime.Unix())
	}
	deletedTasks, err := s.GetTasks(ctx, getReq)
	if err != nil {
		return nil, err
	}

	getReq.Action = ""
	getReq.ActionList = entity.TaskActionInitDeployList
	deployedTasks, err := s.GetTasks(ctx, getReq)
	if err != nil {
		return nil, err
	}

	tasksMapping := make(map[string]*resp.TaskDetailResp)
	for _, task := range deployedTasks {
		tasksMapping[task.Version] = task
	}

	if lastSuccessfulClearedDeployedTask != nil {
		// 时间范围的搜索条件被修正到了秒级, 且生产环境的两次搜索条件不一样, 不一定能涵盖最后成功的这次部署
		tasksMapping[lastSuccessfulClearedDeployedTask.Version] = lastSuccessfulClearedDeployedTask
	}

	for _, task := range deletedTasks {
		delete(tasksMapping, task.Version)
	}

	deployedTasks = deployedTasks[:0]
	for _, task := range tasksMapping {
		deployedTasks = append(deployedTasks, task)
	}

	return deployedTasks, nil
}

// getEnvClusterNames 获取指定环境下所有集群名
func (s *Service) getEnvClusterNames(_ context.Context, envName entity.AppEnvName) ([]entity.ClusterName, error) {
	clusters, ok := s.k8sClusters[envName]
	if !ok {
		return nil, errors.Wrapf(errcode.InternalError, "invalid env(%s)", envName)
	}

	clusterNames := make([]entity.ClusterName, 0, len(clusters))
	for clusterName := range clusters {
		clusterNames = append(clusterNames, clusterName)
	}

	return clusterNames, nil
}
