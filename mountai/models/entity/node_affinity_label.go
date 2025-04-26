package entity

// NodeLabelKeyType : 节点标签键类型
type NodeLabelKeyType string

// 标签类型枚举值
const (
	NodeLabelSpec                    NodeLabelKeyType = "spec"
	NodeLabelCPU                     NodeLabelKeyType = "cpu"
	NodeLabelMemory                  NodeLabelKeyType = "mem"
	NodeLabelExclusiveTypeDeployment NodeLabelKeyType = "exclusive-deployment"
	NodeLabelExclusiveTypeJob        NodeLabelKeyType = "exclusive-job"
	// exclusive-cronjob 标签类型暂不启用
)

// NodeLabelValueType : 节点标签值类型
type NodeLabelValueType string

// 规格标签类型枚举值
// 除了 special 标签特殊用于配合污点使用外
// 其他规格标签代表携带该标签的节点允许等于或大于该标签规格等级的容器进行部署
// 即：
//
//	importance=low 的容器仅会选择 spec=small 的节点
//	importance=medium 的容器会选择 spec=small/medium 的节点
//	importance=high 的容器会选择 spec=small/medium/large 的节点
const (
	// 应用重要性标签：低、中、高
	ApplicationImportanceTypeLow     string = "low"
	ApplicationImportanceTypeMedium  string = "medium"
	ApplicationImportanceTypeHigh    string = "high"
	ApplicationImportanceTypeSpecial string = "special"

	// 普通规格标签：小、中、大
	NodeLabelSpecTypeSmall  NodeLabelValueType = "small"
	NodeLabelSpecTypeMedium NodeLabelValueType = "medium"
	NodeLabelSpecTypeLarge  NodeLabelValueType = "large"
	// 特殊规则标签：固定配合污点使用
	NodeLabelSpecTypeSpecial NodeLabelValueType = "special"
)

// NodeAffinityLabelConfig : 节点亲和性标签
type NodeAffinityLabelConfig struct {
	Importance string             `json:"importance" bson:"importance"`
	CPU        NodeLabelValueType `json:"cpu" bson:"cpu"`
	Mem        NodeLabelValueType `json:"mem" bson:"mem"`
	Exclusive  NodeLabelValueType `json:"exclusive" bson:"exclusive"`
}

// NodeLabelList : 节点标签列表(某种节点标签支持的所有标签值)
type NodeLabelList struct {
	Type   NodeLabelKeyType `json:"type" bson:"type"`
	Values []string         `json:"values" bson:"values"`
}

// TableName : 节点标签数据库表名
func (*NodeLabelList) TableName() string {
	return "node_label"
}
