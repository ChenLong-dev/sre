package entity

import (
	"gitlab.shanhai.int/sre/library/base/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ==========================
// 数据库映射实体
//
// deepcopy标记使用deepcopy包: https://gitlab.shanhai.int/sre/library/tree/master/base/deepcopy.v2
// ==========================
type VipItem struct {
	ID               primitive.ObjectID `bson:"_id" json:"id" deepcopy:"method:GenerateObjectID"`
	CreateTime       *time.Time         `bson:"create_time" json:"create_time"`
	UpdateTime       *time.Time         `bson:"update_time" json:"update_time"`
	Name             string             `bson:"name" json:"name"`
	Disabled         bool               `bson:"disabled" json:"disabled"`
	Fee              float64            `bson:"fee" json:"fee"`
	RenewFee         float64            `bson:"renew_fee,omitempty" json:"renew_fee,omitempty"`
	Single           bool               `bson:"single" json:"single"`
	OriginalFee      float64            `bson:"original_fee" json:"original_fee"`
	VipID            string             `bson:"vip_id" json:"vip_id"`
	Type             string             `bson:"type" json:"type"`
	AppleProductType string             `bson:"apple_product_type,omitempty" json:"apple_product_type,omitempty"`
	Duration         string             `bson:"duration" json:"duration"`
	PhoneType        string             `bson:"phonetype,omitempty" json:"phonetype,omitempty"`
	ProductID        string             `bson:"product_id" json:"product_id"`
	AutoRenew        bool               `bson:"autorenew" json:"autorenew"`
	AutoType         string             `bson:"auto_type,omitempty" json:"auto_type,omitempty"`
	AutoPeriod       string             `bson:"auto_period,omitempty" json:"auto_period,omitempty"`
	Benefit          bool               `bson:"benefit" json:"benefit"`
	DiscountType     string             `bson:"discount_type,omitempty" json:"discount_type,omitempty"`
	State            string             `bson:"state" json:"state"`
	TrialPeriod      string             `bson:"trial_period,omitempty" json:"trial_period,omitempty"`
	URL              string             `bson:"url" json:"url"`
	UnionSource      string             `bson:"union_source,omitempty" json:"union_source,omitempty"`
	UnionURL         string             `bson:"union_url,omitempty" json:"union_url,omitempty"`
	CreateUser       string             `bson:"create_user,omitempty" json:"create_user,omitempty"`
	UpdateUser       string             `bson:"update_user,omitempty" json:"update_user,omitempty"`
}

// ==========================
// 表名方法，该方法名不可更改
// ==========================
func (*VipItem) TableName() string {
	return "vipitems"
}

func (item *VipItem) GenerateObjectIDString(args map[string]interface{}) string {
	return item.ID.Hex()
}

func (item *VipItem) GenerateCreateTimeFormatString(args map[string]interface{}) (string, error) {
	if item.CreateTime == nil {
		return "", nil
	}

	formatString, ok := args["createTimeFormat"].(string)
	if ok {
		return item.CreateTime.Format(formatString), nil
	}

	return item.CreateTime.Format("2006-01-02 15:04:05"), nil
}

func (item *VipItem) GenerateUpdateTimeFormatString(args map[string]interface{}) (string, error) {
	if item.UpdateTime == nil {
		return "", nil
	}

	formatString, ok := args["updateTimeFormat"].(string)
	if ok {
		return item.UpdateTime.Format(formatString), nil
	}

	return item.UpdateTime.Format("2006-01-02 15:04:05"), nil
}

func (item *VipItem) GenerateFixedFee(args map[string]interface{}) (float64, error) {
	feeDigit, ok := args["feeDigit"].(int)
	if !ok {
		feeDigit = 2
	}

	return decimal.ToFixed(feeDigit, item.Fee)
}

func (item *VipItem) GenerateFixedRenewFee(args map[string]interface{}) (float64, error) {
	renewFeeDigit, ok := args["renewFeeDigit"].(int)
	if !ok {
		renewFeeDigit = 2
	}

	return decimal.ToFixed(renewFeeDigit, item.RenewFee)
}

func (item *VipItem) GenerateFixedOriginalFee(args map[string]interface{}) (float64, error) {
	originalFeeDigit, ok := args["originalFeeDigit"].(int)
	if !ok {
		originalFeeDigit = 2
	}

	return decimal.ToFixed(originalFeeDigit, item.OriginalFee)
}
