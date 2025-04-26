package entity

type MaxTotalCPUTemplate struct {
	// 环境名
	EnvName AppEnvName
	// 容器标签名
	// prd集群为 container_name
	// 其他集群为 container
	ContainerLabelName string
	// 容器名
	ContainerName string
	// 统计时间
	CountTime string
}

type MaxTotalMemTemplate struct {
	// 环境名
	EnvName AppEnvName
	// 容器标签名
	// prd集群为 container_name
	// 其他集群为 container
	ContainerLabelName string
	// 容器名
	ContainerName string
	// 统计时间
	CountTime string
}

type WastedMaxCPUUsageRateTemplate struct {
	// 环境名
	EnvName AppEnvName
	// 容器标签名
	// prd集群为 container_name
	// 其他集群为 container
	ContainerLabelName string
	// 容器名
	ContainerName string
	// 统计时间
	CountTime string
	// 使用率限制
	UsageRateLimit float64
	// 最小cpu资源规格
	MinCPUResource string
}

type WastedMaxMemUsageRateTemplate struct {
	// 环境名
	EnvName AppEnvName
	// 容器标签名
	// prd集群为 container_name
	// 其他集群为 container
	ContainerLabelName string
	// 容器名
	ContainerName string
	// 统计时间
	CountTime string
	// 使用率限制
	UsageRateLimit float64
	// 最小内存资源规格
	MinMemResource string
}
