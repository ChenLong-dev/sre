package vendors

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/httpclient"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"rulai/dao"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	_errcode "rulai/utils/errcode"
	"rulai/vendors/ali"
	"rulai/vendors/huawei"
)

// Controller 云服务商控制器接口类型
type Controller interface {
	Name() entity.VendorName
	Region() string
	LogConfigController
	ImageController
}

// LogConfigController 云服务商日志配置控制器接口类型
type LogConfigController interface {
	ApplyLogConfig(ctx context.Context,
		project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp, team *resp.TeamDetailResp) error
	LogConfigExistanceCheck(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, app *resp.AppDetailResp) error
	DeleteLogConfig(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error
	EnsureLogIndex(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, app *resp.AppDetailResp) error
	ApplyLogDump(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error
	DeleteLogDump(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error
	GetLogStoreURL(ctx context.Context, clusterName entity.ClusterName, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
		envName entity.AppEnvName) (logStoreURL, logStoreURLBasedProject string, err error)
}

type ImageController interface {
	GetRepoTags(ctx context.Context, getReq *req.GetRepoTagsReq) ([]*resp.GetDockerTagsResp, int, error)
}

// ClusterConfig 云服务商集群配置
type ClusterConfig struct {
	ClusterID           string
	ClusterNameInVendor string // 云服务商定义的集群名称(区别于AMS集群名称)
	LogGroupID          string
	LogGroupName        string
	K8sTypedClient      kubernetes.Interface
	K8sDynamicClient    dynamic.Interface
	Region              string
}

// ControllerConfig 云服务商控制器配置
type ControllerConfig struct {
	VendorName       entity.VendorName
	AccessKeyID      string
	AccessKeySecret  string
	DisableLogConfig bool
	LogEndpoint      string
	RegionID         string
	HTTPClient       *httpclient.Client
	Clusters         map[entity.AppEnvName]map[entity.ClusterName]*ClusterConfig
	// TODO: 华为云支持通过名称查询日志流以及通过名称查询日志流接入规则后移除 DAO
	DAO *dao.Dao
}

// NewController 创建云服务商控制器
func NewController(cfg *ControllerConfig) (Controller, error) {
	switch cfg.VendorName {
	case entity.VendorAli:
		return ali.NewAliController(cfg.DisableLogConfig, cfg.LogEndpoint, cfg.RegionID, cfg.AccessKeyID, cfg.AccessKeySecret)

	case entity.VendorHuawei:
		envClusterMapping := make(map[entity.AppEnvName]map[entity.ClusterName]*huawei.ClusterConfig, len(cfg.Clusters))
		for envName, clusterConfigs := range cfg.Clusters {
			clusterMapping := make(map[entity.ClusterName]*huawei.ClusterConfig, len(clusterConfigs))
			for clusterName, clusterConfig := range clusterConfigs {
				clusterMapping[clusterName] = &huawei.ClusterConfig{
					ClusterID:           clusterConfig.ClusterID,
					ClusterNameInVendor: clusterConfig.ClusterNameInVendor,
					LogGroupID:          clusterConfig.LogGroupID,
					LogGroupName:        clusterConfig.LogGroupName,
					Region:              clusterConfig.Region,
				}
			}
			envClusterMapping[envName] = clusterMapping
		}
		return huawei.NewHuaweiController(cfg.DisableLogConfig, cfg.LogEndpoint, cfg.RegionID,
			cfg.AccessKeyID, cfg.AccessKeySecret, cfg.HTTPClient, envClusterMapping, cfg.DAO), nil
	}

	return nil, errors.Wrap(_errcode.UnknownVendorError, string(cfg.VendorName))
}
