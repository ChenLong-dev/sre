package req

import (
	"rulai/models/entity"

	"time"
)

type CleanAliServiceDependencyReq struct {
	OperatorID    string            `json:"operator_id"`
	EnvName       entity.AppEnvName `json:"env_name"`
	ServiceName   string            `json:"service_name"`
	AliAlarmName  string            `json:"ali_alarm_name"`
	IngressRecord bool              `json:"ingress_record"`
}

type AliGetRepoTagsReq struct {
	// 项目名
	ProjectName string `json:"project_name"`
	Page        int    `json:"page"`
	Size        int    `json:"size"`
}

type AliDescribeContactGroupListReq struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

type AliDescribeLoadBalancersReq struct {
	// 服务ip地址
	Address string `json:"address"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

type AliSetLoadBalancerNameReq struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AliDescribeLoadBalancerAttributeReq struct {
	ID string `json:"id"`
}

type AliDescribeMetricDataReq struct {
	Namespace  string    `json:"namespace"`
	MetricName string    `json:"metric_name"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Dimensions string    `json:"dimensions"`
}

type AliDescribeMetricLastReq struct {
	Namespace  string `json:"namespace"`
	MetricName string `json:"metric_name"`
	Dimensions string `json:"dimensions"`
}

type AliPutResourceMetricRuleReq struct {
	SLBID             string `json:"slb_id"`
	Name              string `json:"name"`
	ContractGroupName string `json:"contract_group_name"`
}

type AliDescribeMetricRuleListReq struct {
	Name string `json:"name"`
	Page int    `json:"page"`
	Size int    `json:"size"`
}

type AliDescribeContactListReq struct {
	ContactGroupName string `json:"contact_group_name"`
}

type AliDescribeZonesReq struct {
	Keyword string `json:"keyword"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

type AliDescribeZoneRecordsReq struct {
	ZoneID  string `json:"zone_id"`
	Keyword string `json:"keyword"`
	Page    int    `json:"page"`
	Size    int    `json:"size"`
}

type AliAddZoneRecordReq struct {
	ZoneID     string `json:"zone_id"`
	ServiceIP  string `json:"service_ip"`
	DomainName string `json:"domain_name"`
}

type AliUpdateZoneRecordReq struct {
	RecordID   int    `json:"record_id"`
	ServiceIP  string `json:"service_ip"`
	DomainName string `json:"domain_name"`
}

type AliDescribeResourceInstancesReq struct {
	Page int `json:"page"`
	Size int `json:"size"`
}
