// 运维 QDNS 系统打通
// QDNS API文档：https://git2.qingtingfm.com/devops/qdns/-/blob/develop/doc/api.md
package resp

import (
	"net/http"
	"time"

	"gitlab.shanhai.int/sre/library/base/null"

	"rulai/models/entity"
)

// QDNS API 成功响应的 status 值
const (
	QDNSRecordSuccessStatus = 0   // 统一接入域名记录 API 的成功响应 status
	QDNSSuccessStatus       = 200 // 统一接入路由规则 API 的成功响应 status
)

// QDNS 特殊错误响应
var (
	QDNSStatusDuplicateRecordResp = &QDNSRespCommonFields{Status: 400, Msg: "域名记录重复"}
	QDNSStatusRecordNotFoundResp  = &QDNSRespCommonFields{Status: 500, Msg: "该记录不存在"}
)

// QDNS 格式化样式
const QDNSTimeFormatLayout = "\"2006-01-02 15:04:05.999999999\""

// QDNSTime QDNS 时间格式
type QDNSTime time.Time

func (qt *QDNSTime) UnmarshalJSON(data []byte) error {
	t, err := time.ParseInLocation(QDNSTimeFormatLayout, string(data), time.Local)
	if err != nil {
		return err
	}

	*qt = QDNSTime(t)
	return nil
}

type QDNSResp interface {
	GetStatus() int
	GetMessage() string
}

// EqualQDNSErrorCase 判断响应是否等同于特殊错误响应(忽略data)
func EqualQDNSErrorCase(exp, act QDNSResp) bool {
	if act == nil {
		return exp == nil
	}

	if exp == nil {
		return false
	}

	return act.GetStatus() == exp.GetStatus() && act.GetMessage() == exp.GetMessage()
}

// QDNSRespCommonFields QDNS响应参数公共字段
type QDNSRespCommonFields struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
}

func (f *QDNSRespCommonFields) GetStatus() int { return f.Status }

func (f *QDNSRespCommonFields) GetMessage() string { return f.Msg }

// QDNSStandardResp QDNS标准响应参数
type QDNSStandardResp struct {
	QDNSRespCommonFields
	Data interface{} `json:"data"`
}

// QDNSStandardListResp QDNS标准列表响应参数
type QDNSStandardListResp struct {
	QDNSRespCommonFields
	Total      int         `json:"total"`
	PageNumber int         `json:"page_number"`
	PageSize   int         `json:"page_size"`
	Data       interface{} `json:"data"`
}

// QDNSFrontendInfoResp QDNS 统一接入 Kong Route 中的路由信息
type QDNSFrontendInfoResp struct {
	AccessHost string               `json:"access_host"`
	RouteInfo  []*KongRouteInfoResp `json:"route_info"`
}

// GetQDNSUpdateBusinessBackendResp QDNS统一接入信息详情(Kong Upstream 聚合)
type GetQDNSBusinessDetailResp struct {
	Env                entity.QDNSEnvName              `json:"env"`
	BindID             string                          `json:"bind_id"`
	ID                 int                             `json:"id"`
	Name               string                          `json:"name"`
	ActiveHealthCheck  *KongUpstreamActiveHealthCheck  `json:"active_health_check"`
	PassiveHealthCheck *KongUpstreamPassiveHealthCheck `json:"passive_health_check"`
	Service            []*KongService                  `json:"service"`
	Route              []*KongRoute                    `json:"route"`
	Targets            []*KongTarget                   `json:"targets"`
	Status             int                             `json:"status"`
	Tags               []string                        `json:"tags"`
	CreateTime         *QDNSTime                       `json:"create_time"`
	UpdateTime         *QDNSTime                       `json:"update_time"`
	LastModify         string                          `json:"last_modify"`
}

// KongService Kong 网关 Service 信息
type KongService struct {
	Env            entity.QDNSEnvName  `json:"env"`
	BindID         string              `json:"bind_id"`
	ID             int                 `json:"id"`
	Name           string              `json:"name"`
	Retries        int                 `json:"retries"`
	Protocol       entity.KongProtocol `json:"protocol"`
	Host           string              `json:"host"`
	Port           int                 `json:"port"`
	ServicePath    null.String         `json:"service_path"`
	ConnectTimeout int64               `json:"connect_timeout"`
	WriteTimeout   int64               `json:"write_timeout"`
	ReadTimeout    int64               `json:"read_timeout"`
	Tags           []string            `json:"tags"`
	Business       string              `json:"business"`
	BusinessName   string              `json:"business_name"`
	CreateTime     *QDNSTime           `json:"create_time"`
	UpdateTime     *QDNSTime           `json:"update_time"`
	Status         int                 `json:"status"`
	SLALabel       string              `json:"slaLabel"`
	Note           string              `json:"note"`
	Visible        int                 `json:"visible"`
	LastModify     string              `json:"last_modify"`
}

// KongRoute Kong 网关 Route 信息
type KongRoute struct {
	Env                     entity.QDNSEnvName            `json:"env"`
	BindID                  string                        `json:"bind_id"`
	ID                      int                           `json:"id"`
	Name                    string                        `json:"name"`
	Methods                 []string                      `json:"methods"`
	Host                    []string                      `json:"host"`
	Path                    []string                      `json:"path"`
	Headers                 http.Header                   `json:"headers"`
	HTTPSRedirectStatusCode int                           `json:"https_redirect_status_code"`
	RegexPriority           int                           `json:"regex_priority"`
	StripPath               int                           `json:"strip_path"`
	PathHandling            entity.KongPathHandleBehavior `json:"path_handling"`
	PreserveHost            int                           `json:"preserve_host"`
	Protocol                []entity.KongProtocol         `json:"protocol"`
	Tags                    []string                      `json:"tags"`
	ServiceID               string                        `json:"service_id"`
	CreateTime              *QDNSTime                     `json:"create_time"`
	UpdateTime              *QDNSTime                     `json:"update_time"`
	Status                  string                        `json:"status"`
	LastModify              string                        `json:"last_modify"`
}

// KongTarget Kong 网关 Target 信息
type KongTarget struct {
	ID        string         `json:"id"` // Kong Target ID
	CreatedAt float64        `json:"created_at"`
	Upstream  *entity.KongID `json:"upstream"`
	Target    string         `json:"target"`
	Weight    int            `json:"weight"`
	Tags      []string       `json:"tags"`
}

// KongUpstreamActiveHealthCheck Kong 网关主动健康检查配置(Kong定时发起请求进行健康检查)
type KongUpstreamActiveHealthCheck struct {
	Healthy                KongUpstreamActiveHealthCheckHealthyConfig   `json:"healthy"`
	Unhealthy              KongUpstreamActiveHealthCheckUnhealthyConfig `json:"unhealthy"`
	HTTPSVerifyCertificate bool                                         `json:"https_verify_certificate"`
	HTTPPath               string                                       `json:"http_path"`
	Timeout                int64                                        `json:"timeout"`
	HTTPSSNI               null.String                                  `json:"https_sni"`
	Concurrency            int                                          `json:"concurrency"`
	Type                   entity.KongUpstreamHealthCheckType           `json:"type" binding:"oneof=http https tcp"`
}

// KongUpstreamActiveHealthCheckHealthyConfig Kong 网关主动健康检查健康状态配置
type KongUpstreamActiveHealthCheckHealthyConfig struct {
	HTTPStatuses []int `json:"http_statuses"`
	Interval     int64 `json:"interval"`
	Successes    int   `json:"successes"`
}

// KongUpstreamActiveHealthCheckUnhealthyConfig Kong 网关主动健康检查非健康状态配置
type KongUpstreamActiveHealthCheckUnhealthyConfig struct {
	HTTPStatuses []int `json:"http_statuses"`
	TCPFailures  int   `json:"tcp_failures"`
	Timeouts     int   `json:"timeouts"`
	HTTPFailures int   `json:"http_failures"`
	Interval     int64 `json:"interval"`
}

// KongUpstreamPassiveHealthCheck Kong 网关被动健康检查配置(按照响应进行健康检查)
type KongUpstreamPassiveHealthCheck struct {
	Healthy   KongUpstreamPassiveHealthCheckHealthyConfig   `json:"healthy"`
	Unhealthy KongUpstreamPassiveHealthCheckUnhealthyConfig `json:"unhealthy"`
	Type      entity.KongUpstreamHealthCheckType            `json:"type" binding:"oneof=http https tcp"`
}

// KongUpstreamPassiveHealthCheckHealthyConfig Kong 网关被动健康检查健康状态配置
type KongUpstreamPassiveHealthCheckHealthyConfig struct {
	Successes    int   `json:"successes"`
	HTTPStatuses []int `json:"http_statuses"`
}

// KongUpstreamPassiveHealthCheckUnhealthyConfig Kong 网关被动健康检查非健康状态配置
type KongUpstreamPassiveHealthCheckUnhealthyConfig struct {
	HTTPFailures int   `json:"http_failures"`
	HTTPStatuses []int `json:"http_statuses"`
	TCPFailures  int   `json:"tcp_failures"`
	Timeouts     int   `json:"timeouts"`
}

// Routes QDNS get_tag 返回的路由对象
type Routes struct {
	CreatedAt               int      `json:"created_at"`
	Hosts                   []string `json:"hosts"`
	HTTPSRedirectStatusCode int      `json:"https_redirect_status_code"`
	ID                      string   `json:"id"`
	Name                    string   `json:"name"`
	PathHandling            string   `json:"path_handling"`
	Paths                   []string `json:"paths"`
	PreserveHost            bool     `json:"preserve_host"`
	Protocols               []string `json:"protocols"`
	RegexPriority           int      `json:"regex_priority"`
	StripPath               bool     `json:"strip_path"`
	Tags                    []string `json:"tags"`
	UpdatedAt               int      `json:"updated_at"`
}

// Services QDNS get_tag 返回的服务对象
type Services struct {
	ConnectTimeout    int                            `json:"connect_timeout"`
	CreatedAt         int                            `json:"created_at"`
	Host              string                         `json:"host"`
	ID                string                         `json:"id"`
	Name              string                         `json:"name"`
	Path              interface{}                    `json:"path"`
	Port              int                            `json:"port"`
	Protocol          string                         `json:"protocol"`
	ReadTimeout       int                            `json:"read_timeout"`
	Retries           int                            `json:"retries"`
	Tags              []string                       `json:"tags"`
	UpdatedAt         int                            `json:"updated_at"`
	WriteTimeout      int                            `json:"write_timeout"`
	ActiveHealthCheck *KongUpstreamActiveHealthCheck `json:"active_health_check"`
}

// Upstreams QDNS get_tag 返回的后端服务对象
type Upstreams struct {
	Algorithm        string        `json:"algorithm"`
	CreatedAt        int           `json:"created_at"`
	HashFallback     string        `json:"hash_fallback"`
	HashOn           string        `json:"hash_on"`
	HashOnCookiePath string        `json:"hash_on_cookie_path"`
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Slots            int           `json:"slots"`
	Tags             []string      `json:"tags"`
	HealthChecks     *HealthChecks `json:"healthchecks"`
}

type HealthChecks struct {
	Active  *KongUpstreamActiveHealthCheck  `json:"active"`
	Passive *KongUpstreamPassiveHealthCheck `json:"passive"`
}

// GetTagResponse QDNS get_tag 返回的结果集对象
type GetTagResponse struct {
	QDNSRespCommonFields
	Data *struct {
		Routes    []*Routes    `json:"routes"`
		Services  []*Services  `json:"services"`
		Upstreams []*Upstreams `json:"upstreams"`
	} `json:"data"`
}
