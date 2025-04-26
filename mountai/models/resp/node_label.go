package resp

import "rulai/models/entity"

// NodeLabelListResp : 节点标签列表(某种节点标签支持的所有标签值)
type NodeLabelListResp struct {
	Type   entity.NodeLabelKeyType `json:"type"`
	Values []string                `json:"values"`
}
