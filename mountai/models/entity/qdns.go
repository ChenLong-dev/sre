package entity

import "regexp"

// DomainRecordType 域名解析记录类型
type DomainRecordType string

// QDNS支持的域名解析记录类型
const (
	ARecord     DomainRecordType = "A"
	CNAMERecord DomainRecordType = "CNAME"
	TXTRecord   DomainRecordType = "TXT"
	SOARecord   DomainRecordType = "SOA"
	AAAARecord  DomainRecordType = "AAAA"
	NXRecord    DomainRecordType = "NX"
)

// QDNSEnvName QDNS 环境名称
type QDNSEnvName string

// QDNS支持的环境名称
const (
	QDNSEnvNameDev    QDNSEnvName = "dev"
	QDNSEnvNameStg    QDNSEnvName = "stg"
	QDNSEnvNamePortal QDNSEnvName = "portal"
	QDNSEnvNameMain   QDNSEnvName = "main"
	QDNSEnvNameInt    QDNSEnvName = "int"
	// 特殊
	QDNSEnvNameUnknown QDNSEnvName = "unknown"
)

// QDNSRecord QDNS解析记录
type QDNSRecord struct {
	ApproveStatus    int              `json:"approve_status"`
	CreateTime       string           `json:"create_time"`
	DomainController string           `json:"domain_controller"`
	DomainID         int              `json:"domain_id"`
	DomainName       string           `json:"domain_name"`
	ID               int              `json:"id"`
	InstanceID       string           `json:"instance_id"`
	LastModify       string           `json:"last_modify"`
	LineType         string           `json:"line_type"`
	Name             string           `json:"name"`
	Remark           string           `json:"remark"`
	Status           int              `json:"status"`
	TTL              string           `json:"ttl"`
	Type             DomainRecordType `json:"type"`
	UpdateTime       string           `json:"update_time"`
	Value            string           `json:"value"`
	Version          string           `json:"version"`
}

var (
	// RegularGetClusterName 提取域名中的集群名称
	RegularGetClusterName = regexp.MustCompile(`\.svc\.([a-zA-Z]+)\.cluster\.local`)
	// RegularDetermineMatchMethod 判断Kong Path 是否是正则匹配
	RegularDetermineMatchMethod = regexp.MustCompile(`[a-zA-Z/]`)
)
