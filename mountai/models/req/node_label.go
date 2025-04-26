package req

import "rulai/models/entity"

// GetNodeLabelListReq : 获取支持的节点标签列表请求参数
type GetNodeLabelListReq struct {
	Type entity.NodeLabelKeyType `json:"type" binding:"omitempty"`
}
