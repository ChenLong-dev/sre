package entity

// VendorName 云服务商名称
type VendorName string

// 云服务商列表
const (
	VendorAli    VendorName = "ali"
	VendorHuawei VendorName = "huawei"
)

// 云服务商特殊 annotations 标签
const (
	// AnnotationsAOMLogRelabel 华为云容器标准输出日志收集自定义标签(日志中显示)
	AnnotationsAOMLogRelabel = "kubernetes.AOM.log.relabel"
	// AnnotationsAOMLogStdout 华为云容器标准输出日志收集标准输出列表, 不指定该 key 代表收集 pod 内全部容器的标准输出
	// 填写容器名, 例如: '["container_1", "container_2"]', 如果不采任何日志填 '[]'
	AnnotationsAOMLogStdout = "kubernetes.AOM.log.stdout"
)

// 华为云日志相关 annotations relabel 标签限制
const (
	AnnotationsAOMLogRelabelLimits              = 16 // 截止 2021-08-19 内部文档支持的最大 kv 对为 16 个
	AnnotationsAOMLogRelabelKeyValueLengthLimit = 64 // 截止 2021-08-19 内部文档支持的 kv值 最大长度为 64, FIXME: 超过时会被截取, 千万小心
)

// AnnotationsAOMLogRelabelKeyReverseSet 华为云容器日志收集标签内部 key 不可使用的默认标签集合
// NOTE: 注意判断时不分大小写
var AnnotationsAOMLogRelabelKeyReverseSet = map[string]struct{}{
	"podname": {}, "appname": {}, "containername": {}, "clusterid": {}, "clustername": {}, "serverlesspkg": {}, "serverlessfunc": {},
	"projectid": {}, "serviceid": {}, "namespace": {}, "pid": {}, "hostid": {}, "hostname": {}, "hostip": {}, "hostipv6": {}}

// supportedVendors 当前支持的云服务商
var supportedVendors = map[VendorName]struct{}{
	VendorAli:    {},
	VendorHuawei: {},
}

// CheckVendorSupport 校验云服务商是否被支持
func CheckVendorSupport(name VendorName) bool {
	_, ok := supportedVendors[name]
	return ok
}
