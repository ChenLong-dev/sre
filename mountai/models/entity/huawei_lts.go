package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// EmptyAppLTSStream AppLTSStream 空对象(用于获取表名)
var EmptyAppLTSStream = new(AppLTSStream)

// AppLTSStream 应用对应的华为云 LTS 日志流记录
type AppLTSStream struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	AppID       string             `bson:"app_id"`       // 应用ID
	ClusterName ClusterName        `bson:"cluster_name"` // 集群名称
	EnvName     AppEnvName         `bson:"env_name"`     // 环境名称
	StreamID    string             `bson:"stream_id"`    // LTS 日志流 ID
	StreamName  string             `bson:"stream_name"`  // LTS 日志流名称
	RuleID      string             `bson:"rule_id"`      // AOM 到 LTS 接入规则 ID
	DumpID      string             `bson:"dump_id"`      // OBS 日志转储 ID
}

func (als *AppLTSStream) TableName() string {
	return "app_lts_stream"
}
