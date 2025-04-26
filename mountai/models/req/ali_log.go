package req

import (
	sls "github.com/aliyun/aliyun-log-go-sdk"

	"rulai/models/entity"
)

// AliGetLogStoreDetailReq 获取logstore请求
type AliGetLogStoreDetailReq struct {
	ProjectName string `json:"project_name"`
	StoreName   string `json:"store_name"`
}

// AliDeleteLogStoreReq 删除logstore请求
type AliDeleteLogStoreReq struct {
	ProjectName string `json:"project_name"`
	StoreName   string `json:"store_name"`
}

// AliGetLogStoreIndexReq 获取logstore索引请求
type AliGetLogStoreIndexReq struct {
	// ProjectName 集群名称
	ProjectName string `json:"project_name"`
	// StoreName 日志仓库名称
	StoreName string `json:"log_store_name"`
}

// AliAddLogStoreIndexesReq 添加logstore索引请求
type AliAddLogStoreIndexesReq struct {
	// ProjectName 集群名称
	ProjectName string                  `json:"project_name"`
	StoreName   string                  `json:"log_store_name"`
	Index       map[string]sls.IndexKey `json:"index"`
}

// AliUpdateStoreLogIndexReq 更新logstore索引请求
type AliUpdateStoreLogIndexReq struct {
	// ProjectName 集群名称
	ProjectName string    `json:"project_name"`
	StoreName   string    `json:"log_store_name"`
	Index       sls.Index `json:"index"`
}

// AliCreateStoreLogIndexReq the request struct of creating new logstore index
type AliCreateStoreLogIndexReq struct {
	Namespace        string `json:"namespace"`
	AliLogConfigName string `json:"ali_log_config_name"`
	EnvName          string `json:"env_name"`
}

// Create ali logStore shipper request
type CreateAliLogStoreOSSShipperReq struct {
	ShipperName    string `json:"shipperName"`
	OSSBucket      string `json:"ossBucket"`
	OSSPrefix      string `json:"ossPrefix"`
	RoleArn        string `json:"roleArn"`
	BufferInterval int    `json:"bufferInterval,omitempty"`
	BufferSize     int    `json:"bufferSize,omitempty"`
	CompressType   string `json:"compressType,omitempty"`
	PathFormat     string `json:"pathFormat,omitempty"`
	Format         string `json:"format,omitempty"`
}

type CreateAliLogStoreShipperReq struct {
	TargetType          string `json:"target_type"`
	ProjectName         string `json:"project_name"`
	LogStoreName        string `json:"log_store_name"`
	LogStoreProject     string `json:"log_store_project"`
	ShipperName         string `json:"shipper_name"`
	CreateOSSShipperReq *CreateAliLogStoreOSSShipperReq
}

type DeleteAliLogStoreShipperReq struct {
	LogStoreProject string `json:"log_store_project"`
	LogStoreName    string `json:"log_store_name"`
	ShipperName     string `json:"shipper_name"`
}

type SyncAliLogStoreColdStorageShipperReq struct {
	EnvName      entity.AppEnvName `json:"env_name"`
	ProjectName  string            `json:"project_name"`
	LogStoreName string            `json:"log_store_name"`
}
