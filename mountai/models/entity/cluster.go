package entity

// ClusterName 集群名称
type ClusterName string

// 已支持集群名称
const (
	ClusterJuncle ClusterName = "juncle"
	ClusterZeus   ClusterName = "zeus"
)

// 历史集群名称
const (
	ClusterBrox ClusterName = "brox" // 阿里云集群, 2022 年迁移至华为云集群后移除
)

// 特殊集群名称
const (
	EmptyClusterName ClusterName = ""
)

// DefaultClusterName 当前默认集群名称(未来集群切换时修改)
const DefaultClusterName = ClusterZeus

// ValidateClusterName 校验集群名称
func ValidateClusterName(name ClusterName) bool {
	return name == ClusterJuncle || name == ClusterZeus
}
