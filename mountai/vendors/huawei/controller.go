package huawei

import (
	"time"

	"gitlab.shanhai.int/sre/library/net/httpclient"

	"rulai/dao"
	"rulai/models/entity"
	"rulai/vendors/huawei/APIGW-go-sdk-2.0.2/core"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	swr "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/swr/v2/region"
)

// Controller 华为云控制器
type Controller struct {
	VendorName       entity.VendorName
	DisableLogConfig bool
	LogEndpoint      string
	ProjectID        string
	HTTPClient       *httpclient.Client
	APISigner        *core.Signer
	Clusters         map[entity.AppEnvName]map[entity.ClusterName]*ClusterConfig
	// TODO: 华为云支持通过名称查询日志流以及通过名称查询日志流接入规则后移除 DAO
	DAO       *dao.Dao
	SwrClient *swr.SwrClient
}

// ClusterConfig 华为云集群配置
type ClusterConfig struct {
	ClusterID           string
	ClusterNameInVendor string // 云服务商定义的集群名称(区别于AMS集群名称)
	LogGroupID          string
	LogGroupName        string
	Region              string
}

func (c *Controller) Name() entity.VendorName { return c.VendorName }

func (c *Controller) Region() string { return c.ProjectID } // 华为云的区域ID是一级项目ID

// NewHuaweiController 创建华为云控制器实例
func NewHuaweiController(disableLogConfig bool, logEndpoint, regionID, accessKeyID, accessKeySecret string,
	httpClient *httpclient.Client, clusterCfg map[entity.AppEnvName]map[entity.ClusterName]*ClusterConfig, d *dao.Dao) *Controller {
	auth := basic.NewCredentialsBuilder().
		WithAk(accessKeyID).
		WithSk(accessKeySecret).
		Build()
	client := swr.NewSwrClient(
		swr.SwrClientBuilder().
			WithRegion(region.CN_EAST_3).
			WithCredential(auth).
			WithHttpConfig(&config.HttpConfig{
				Timeout: time.Second * 15,
			}).
			Build(),
	)
	return &Controller{
		VendorName:       entity.VendorHuawei,
		DisableLogConfig: disableLogConfig,
		LogEndpoint:      logEndpoint,
		ProjectID:        regionID,
		HTTPClient:       httpClient,
		APISigner:        &core.Signer{Key: accessKeyID, Secret: accessKeySecret},
		Clusters:         clusterCfg,
		DAO:              d,
		SwrClient:        client,
	}
}
