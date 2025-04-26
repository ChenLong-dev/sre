package resp

// LTSErrorCode 华为云日志服务(LTS)错误码
type LTSErrorCode string

// 需要关注的华为云日志服务(LTS)错误码
const (
	LTSDuplicateLogStreamName        LTSErrorCode = "LTS.0205" // 日志流名称已存在
	LTSLogStreamAssociatedByTransfer LTSErrorCode = "LTS.0207" // 日志流关联日志转储(不允许删除)
	LTSLogStreamNotExist             LTSErrorCode = "LTS.0208" // 日志流不存在
	LTSDuplicateAOMMappingRuleName   LTSErrorCode = "LTS.0740" // 日志流 AOM 接入规则名称已存在
	LTSInvalidAOMMappingRuleID       LTSErrorCode = "LTS.0745" // 日志流 AOM 接入规则 ID 非法(不存在时报该错误)
)

// HuaweiLTSResp 华为云日志服务响应值接口类型
type HuaweiLTSResp interface {
	GetRequestID() string
	GetErrorCode() LTSErrorCode
	GetErrorMsg() string
}

// HuaweiLTSStandardResp 华为云日志服务(LTS)标准响应值
type HuaweiLTSStandardResp struct {
	RequestID string       `json:"request_id"` // 并非所有响应值都有 request_id
	ErrorCode LTSErrorCode `json:"error_code"`
	ErrorMsg  string       `json:"error_msg"`
}

func (r *HuaweiLTSStandardResp) GetRequestID() string { return r.RequestID }

func (r *HuaweiLTSStandardResp) GetErrorCode() LTSErrorCode { return r.ErrorCode }

func (r *HuaweiLTSStandardResp) GetErrorMsg() string { return r.ErrorMsg }

// HuaweiLTSCodeDetailsInMessageResp 华为云日志服务 code-details 封装在 message 中格式的响应值
type HuaweiLTSCodeDetailsInMessageResp struct {
	RequestID string               `json:"request_id"` // 并非所有响应值都有 request_id, 也并非一定在这一层
	Message   HuaweiLTSCodeDetails `json:"message"`
}
type HuaweiLTSCodeDetails struct {
	RequestID string       `json:"request_id"` // 并非所有响应值都有 request_id, 也并非一定在这一层
	Code      LTSErrorCode `json:"code"`
	Details   string       `json:"details"`
}

func (res *HuaweiLTSCodeDetailsInMessageResp) GetRequestID() string {
	if res.RequestID != "" {
		return res.RequestID
	}
	return res.Message.RequestID
}

func (res *HuaweiLTSCodeDetailsInMessageResp) GetErrorCode() LTSErrorCode { return res.Message.Code }

func (res *HuaweiLTSCodeDetailsInMessageResp) GetErrorMsg() string { return res.Message.Details }

// CreateHuaweiLTSStreamResp 创建华为云日志服务(LTS)日志流响应参数
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0016.html
type CreateHuaweiLTSStreamResp struct {
	LogStreamID string `json:"log_stream_id"`
}

// CreateHuaweiAOMToLTSStreamMappingResp 创建华为云 AOM 到 LTS 日志流的接入规则响应参数
// FIXME: 官方文档问题比较多: https://support.huaweicloud.com/api-lts/lts_api_0064.html
// 目前响应是个数组
type CreateHuaweiAOMToLTSStreamMappingResp []*CreateHuaweiAOMToLTSStreamMappingRespItem

// CreateHuaweiAOMToLTSStreamMappingRespItem 创建华为云 AOM 到 LTS 日志流的接入规则响应参数数组内对象结构
type CreateHuaweiAOMToLTSStreamMappingRespItem struct {
	ProjectID string                           `json:"project_id"`
	RuleID    string                           `json:"rule_id"`
	RuleName  string                           `json:"rule_name"`
	RuleInfo  *HuaweiAOMToLTSStreamMappingInfo `json:"rule_info"`
}

// DeleteHuaweiAOMToLTSStreamMappingResp 删除华为云 AOM 到 LTS 日志流的接入规则响应参数
// FIXME: 官方文档有问题, 实际响应是个字符串数组, 当前选择不解析这种不规范的响应: https://support.huaweicloud.com/api-lts/lts_api_0066.html
// 先占位, 未来华为云调整后再接入
// type DeleteHuaweiAOMToLTSStreamMappingResp []string

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
	TargetLogGroupID    string `json:"target_log_group_id"`
	TargetLogGroupName  string `json:"target_log_group_name"`
	TargetLogStreamID   string `json:"target_log_stream_id"`
	TargetLogStreamName string `json:"target_log_stream_name"`
}

// CreateHuaweiLTSTemplateResp 创建华为云日志服务(LTS)结构化配置响应参数
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0060.html
type CreateHuaweiLTSTemplateResp struct {
	TemplateID string `json:"result"` // 该接口响应字段名为 "result"
}

// CreateHuaweiLTSDumpResp 创建华为云 LTS 日志 OBS 转储响应参数
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0027.html
type CreateHuaweiLTSDumpResp struct {
	LogDumpOBSID string `json:"log_dump_obs_id"`
}

// ListLTSStreamResp 查询指定日志组下的所有日志流响应参数
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0017.html
type ListLTSStreamResp struct {
	LogStreams []*LtsLogStreamInfo `json:"log_streams"`
}

// LtsLogStreamInfo 日志流信息结构
type LtsLogStreamInfo struct {
	CreationTime  int64  `json:"creation_time"`
	LogStreamName string `json:"log_stream_name"`
	IsFavorite    bool   `json:"is_favorite"`
	FilterCount   int    `json:"filter_count"`
	LogStreamID   string `json:"log_stream_id"`
}

// ListAOMToLTSRuleResp 查询 AOMToLTSRule 规则列表
// 官方文档: https://support.huaweicloud.com/api-lts/lts_api_0067.html
type ListAOMToLTSRuleResp []*CreateHuaweiAOMToLTSStreamMappingRespItem
