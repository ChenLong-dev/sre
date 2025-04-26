package resp

// ==========================
// 响应模型
//
// deepcopy标记使用deepcopy包: https://gitlab.shanhai.int/sre/library/tree/master/base/deepcopy.v2
// ==========================
type VipRecommendItemDetail struct {
	ID         string `json:"id" deepcopy:"method:GenerateIDString"`
	VipItemID  string `json:"vip_item_id"`
	CreateTime string `json:"create_time" deepcopy:"method:GenerateCreateTimeFormatString"`
	UpdateTime string `json:"update_time" deepcopy:"method:GenerateUpdateTimeFormatString"`
	Title      string `json:"title"`
	Desc       string `json:"desc"`
}
