package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 资源类型
type ResourceType string

const (
	// mysql
	ResourceTypeRds ResourceType = "rds"
	// redis
	ResourceTypeRedis ResourceType = "redis"
	// mongo
	ResourceTypeMongo ResourceType = "mongo"
	// hbase
	ResourceTypeHbase ResourceType = "hbase"
	// cdn
	ResourceTypeCdn ResourceType = "cdn"
	// ecs
	ResourceTypeEcs ResourceType = "ecs"
)

var AllResourceTypes = []ResourceType{ResourceTypeRds, ResourceTypeRedis, ResourceTypeMongo,
	ResourceTypeHbase, ResourceTypeCdn, ResourceTypeEcs}

// 资源实例
type ResourceInstance struct {
	// 实例id(带provider前缀)
	InstanceID string `json:"instance_id"`
	// 实例id
	ID string `json:"id"`
	// 主机名
	InstanceName string `json:"instance_name"`
	// 实例运行状态
	Status string `json:"status"`
	// 资源类型
	Type ResourceType `json:"type"`
	// 资源来源类型
	Provider ProviderType `json:"provider"`
	// 资源连接地址
	ConnectionStr string `json:"connection_str"`
}

const ResourceTypeCachePrefix = "resource:"

// 资源来源类型
type ProviderType string

const (
	// aliyun
	ProviderTypeAliyun ProviderType = "aliyun"
)

const ProviderTypePrefix = "provider:"

var AllProviderType = []ProviderType{ProviderTypeAliyun}

type ResourceList struct {
	// rds
	Rds []string `bson:"rds" json:"rds"`
	// redis
	Redis []string `bson:"redis" json:"redis"`
	// mongo
	Mongo []string `bson:"mongo" json:"mongo"`
	// hbase
	Hbase []string `bson:"hbase" json:"hbase"`
	// cdn
	Cdn []string `bson:"cdn" json:"cdn"`
	// ecs
	Ecs []string `bson:"ecs" json:"ecs"`
}

// 资源
type Resource struct {
	// object id
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// 项目id
	ProjectID string `bson:"project_id" json:"project_id"`

	*ResourceList `bson:",inline"`
	// 环境
	Env string `bson:"env" json:"env"`
	// git config-center commit id
	CommitID string `bson:"commit_id" json:"commit_id"`

	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
}

func (*Resource) TableName() string {
	return "project_resource"
}

var ResourceConfParseMap = map[ProviderType]map[ResourceType]string{
	ProviderTypeAliyun: {
		ResourceTypeRds:   ".mysql.rds.aliyuncs.com",
		ResourceTypeRedis: ".redis.rds.aliyuncs.com",
		ResourceTypeMongo: ".mongodb.rds.aliyuncs.com",
		ResourceTypeHbase: "-proxy-hbaseue.hbaseue.rds.aliyuncs.com",
	},
}
