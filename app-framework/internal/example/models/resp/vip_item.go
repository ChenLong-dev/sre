package resp

// ==========================
// 响应模型
//
// deepcopy标记使用deepcopy包: https://gitlab.shanhai.int/sre/library/tree/master/base/deepcopy.v2
// ==========================
type VipItemListResponse struct {
	Name             string  `json:"name"`
	Duration         string  `json:"duration"`
	AutoPeriod       string  `json:"auto_period"`
	OriginalFee      float64 `json:"original_fee" deepcopy:"method:GenerateFixedOriginalFee"`
	Fee              float64 `json:"fee" deepcopy:"method:GenerateFixedFee"`
	Type             string  `json:"type"`
	VipID            string  `json:"vip_id"`
	Disabled         bool    `json:"disabled"`
	Benefit          bool    `json:"benefit"`
	AutoRenew        bool    `json:"autorenew"`
	State            string  `json:"state"`
	AutoType         string  `json:"auto_type"`
	DiscountType     string  `json:"discount_type"`
	RenewFee         float64 `json:"renew_fee,omitempty" deepcopy:"method:GenerateFixedRenewFee"`
	TrialPeriod      string  `json:"trial_period,omitempty"`
	PhoneType        string  `json:"phonetype"`
	ProductID        string  `json:"product_id"`
	UnionSource      string  `json:"union_source,omitempty"`
	ID               string  `json:"id" deepcopy:"method:GenerateObjectIDString"`
	CreateTime       string  `json:"create_time" deepcopy:"method:GenerateCreateTimeFormatString"`
	UpdateTime       string  `json:"update_time" deepcopy:"method:GenerateUpdateTimeFormatString"`
	Single           bool    `json:"single"`
	AppleProductType string  `json:"apple_product_type,omitempty"`
	URL              string  `json:"url"`
	UnionURL         string  `json:"union_url,omitempty"`
	CreateUser       string  `json:"create_user,omitempty"`
	UpdateUser       string  `json:"update_user,omitempty"`
}
