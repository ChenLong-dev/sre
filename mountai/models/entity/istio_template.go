package entity

const (
	IstioVirtualService  = "VirtualService"
	DefaultIstioGateway  = "istio-system/gateway"
	DefaultIstioNodePort = 31820
)

type VirtualServiceTemplate struct {
	// 服务版本
	APIVersion string
	// 服务名称
	Name                    string
	Namespace               string
	ProjectName             string
	AppName                 string
	Annotations             map[string]string
	IstioGateway            string
	ServiceHost             string
	ServiceHostsWithCluster []string
	ServiceName             string
	HTTPRoutes              []*MatchRule
}

type HTTPRoute struct {
	MatchRule []*MatchRule
	Host      []string
}

type MatchRule struct {
	Name         string
	MatchType    MatchType
	MatchValue   string
	Rewrite      bool
	RewriteValue string
}

type MatchType string

const (
	MatchTypePrefix MatchType = "prefix"
	MatchTypeRegex  MatchType = "regex"
)

func (t *VirtualServiceTemplate) Kind() string { return IstioVirtualService }

func (t *VirtualServiceTemplate) SetAPIVersion(ver string) { t.APIVersion = ver }
