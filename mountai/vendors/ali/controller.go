package ali

import (
	"rulai/models/entity"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	sls "github.com/aliyun/aliyun-log-go-sdk"
)

// Controller 阿里云控制器
// TODO: 迁移日志相关逻辑至此
type Controller struct {
	VendorName       entity.VendorName
	DisableLogConfig bool
	LogEndpoint      string
	RegionID         string
	SDKClient        *sdk.Client
	LogSDKClient     sls.ClientInterface
}

func (c *Controller) Name() entity.VendorName { return c.VendorName }

func (c *Controller) Region() string { return c.RegionID }

// NewAliController 创建阿里云控制器实例
func NewAliController(disableLogConfig bool, logEndpoint, regionID, accessKeyID, accessKeySecret string) (*Controller, error) {
	sdkClient, err := sdk.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}

	return &Controller{
		VendorName:       entity.VendorAli,
		DisableLogConfig: disableLogConfig,
		LogEndpoint:      logEndpoint,
		RegionID:         regionID,
		SDKClient:        sdkClient,
		LogSDKClient:     sls.CreateNormalInterface(logEndpoint, accessKeyID, accessKeySecret, ""),
	}, nil
}
