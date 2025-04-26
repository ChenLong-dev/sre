package entity

// LogStoreIndexKey 日志仓库索引键值类型
type LogStoreIndexKey string

const (
	// LogIndexEnv 环境键值
	LogIndexEnv LogStoreIndexKey = "env"
	// LogIndexApp 应用键值
	LogIndexApp LogStoreIndexKey = "app"
	// LogIndexProject 项目键值
	LogIndexProject LogStoreIndexKey = "project"
	// LogIndexCluster 集群键值
	LogIndexCluster LogStoreIndexKey = "cluster"
	// LogIndexVendor 云服务商键值
	LogIndexVendor LogStoreIndexKey = "vendor"
)

var (
	// LogStoreIndexKeys 索引键值数组
	LogStoreIndexKeys = []LogStoreIndexKey{LogIndexEnv, LogIndexApp, LogIndexProject, LogIndexCluster, LogIndexVendor}
)

const (
	// AliLogIndexTypeText logstore索引类型
	AliLogIndexTypeText = "text"
)

var (
	// DefaultAliLogIndexToken 默认日志内容分隔符
	DefaultAliLogIndexToken = []string{
		" ", "\n", "\t", "\r",
		",", ";", "[", "]", "{", "}", "(", ")",
		"&", "^", "*", "#", "@", "~", "=", "<", ">", "/",
		"\\", "?", ":", "'", "\"",
	}
)
