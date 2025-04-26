package req

import "gitlab.shanhai.int/sre/library/base/null"

// 华为云日志服务(LTS)特殊常量
const (
	HuaweiLTSDeploymentTagAll     = "__ALL_DEPLOYMENTS__"
	HuaweiLTSFileNameTagAll       = "__ALL_FILES__"
	HuaweiLTSDumpTypeCycle        = "cycle"
	HuaweiLTSDumpFormatRaw        = "RAW"
	HuaweiLTSDumpFormatJSON       = "JSON"
	HuaweiLTSDumpPeriodUnitMinute = "min"
	HuaweiLTSDumpPeriodUnitHour   = "hour"
)

// CreateHuaweiLTSStreamReq 创建华为云 LTS 日志流请求参数
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0016.html
type CreateHuaweiLTSStreamReq struct {
	LogStreamName string   `json:"log_stream_name"`
	TTLInDays     null.Int `json:"ttl_in_days"`
}

// CreateHuaweiAOMToLTSStreamMappingReq 创建华为云 AOM 到 LTS 日志流的接入规则请求参数
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0064.html
type CreateHuaweiAOMToLTSStreamMappingReq struct {
	RuleName  string                           `json:"rule_name"`
	RuleInfo  *HuaweiAOMToLTSStreamMappingInfo `json:"rule_info"`
	ProjectID string                           `json:"project_id"`
}

// CreateHuaweiLTSDumpReq 创建华为云 LTS 日志 OBS 转储请求参数
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0027.html
type CreateHuaweiLTSDumpReq struct {
	LogGroupID    string   `json:"log_group_id"`
	LogStreamIDs  []string `json:"log_stream_ids"`
	OBSBucketName string   `json:"obs_bucket_name"`
	Type          string   `json:"type,omitempty"`      // RAW/JSON, 默认为 RAW
	StorageFormat string   `json:"storage_format"`      // 必须填 "cycle"
	SwitchOn      bool     `json:"switch_on,omitempty"` // 注意不填默认值是 true
	PrefixName    string   `json:"prefix_name,omitempty"`
	DirPrefixName string   `json:"dir_prefix_name,omitempty"`
	// 仅支持下列组合: ["2min","5min","30min","1hour","3hour","6hour","12hour"]
	Period     int64  `json:"period"`
	PeriodUnit string `json:"period_unit"`
}

// HuaweiAOMToLTSStreamMappingInfo 华为云 AOM 到 LTS 日志流的接入规则详情
// FIXME: 官方文档问题比较多: https://support.huaweicloud.com/api-lts/lts_api_0064.html
type HuaweiAOMToLTSStreamMappingInfo struct {
	ClusterID     string                                 `json:"cluster_id"`
	ClusterName   string                                 `json:"cluster_name"`
	Deployments   []string                               `json:"deployments,omitempty"`
	ContainerName string                                 `json:"container_name,omitempty"`
	Namespace     string                                 `json:"namespace"`
	Files         []*HuaweiAOMToLTSStreamMappingFileInfo `json:"files"`
}

// HuaweiAOMToLTSStreamMappingFileInfo 华为云 AOM 到 LTS 日志流的接入规则文件日志详情
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0066.html
type HuaweiAOMToLTSStreamMappingFileInfo struct {
	FileName      string                  `json:"file_name"`
	LogStreamInfo *HuaweiLTSLogStreamInfo `json:"log_stream_info"`
}

// HuaweiLTSLogStreamInfo 华为云日志服务(LTS)日志流详细信息
type HuaweiLTSLogStreamInfo struct {
	// TargetLogGroupID 与 TargetLogGroupName 选填一个即可, 注意不填的项不能为空字符串
	TargetLogGroupID   string `json:"target_log_group_id,omitempty"`
	TargetLogGroupName string `json:"target_log_group_name,omitempty"`
	// TargetLogStreamID 与 TargetLogStreamName 选填一个即可, 注意不填的项不能为空字符串
	// FIXME: 华为云 API 目前有 bug, 只填写日志流名称时可能会报重名错误, 而不填写日志流名称会关联不到 LTS
	// FIXME: 故当前只支持全部填写 target_log_stream_id 和 target_log_stream_name
	TargetLogStreamID   string `json:"target_log_stream_id,omitempty"`
	TargetLogStreamName string `json:"target_log_stream_name,omitempty"`
}

// CreateHuaweiLTSTemplateReq 创建华为云日志服务(LTS)结构化配置请求参数
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0060.html
type CreateHuaweiLTSTemplateReq struct {
	LogGroupID  string                        `json:"log_group_id"`
	LogStreamID string                        `json:"log_stream_id"`
	ProjectID   string                        `json:"project_id"`
	Content     string                        `json:"content"`
	DemoFields  []*HuaweiLTSTemplateDemoField `json:"demo_fields"`
	ParseType   string                        `json:"parse_type"`
	RegexRules  string                        `json:"regex_rules,omitempty"`
	Layers      int                           `json:"layers,omitempty"`
	Tokenizer   string                        `json:"tokenizer,omitempty"`
	LogFormat   string                        `json:"log_format,omitempty"`
}

// HuaweiLTSTemplateDemoField 华为云日志服务(LTS)结构化配置示例字段
type HuaweiLTSTemplateDemoField struct {
	FieldName       string `json:"fieldName,omitempty"`
	UserDefinedName string `json:"userDefinedName,omitempty"`
	Type            string `json:"type"`
}

// ListLTSStreamReq 查询指定日志组下的所有日志流请求
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0017.html
type ListLTSStreamReq struct {
	ProjectID  string `json:"project_id"`
	LogGroupID string `json:"log_group_id"`
}
