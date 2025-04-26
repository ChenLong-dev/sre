package utils

const (
	// DefaultTimeFormatLayout 默认时间格式化样式
	DefaultTimeFormatLayout = "2006-01-02 15:04:05"
	// 镜像格式化样式
	ImageTimeFormatLayout = "2006/01/02 15:04:05"
)

// k8s自定义time转换接口
type K8sTime interface {
	// time是否是nil或零值判断方法
	IsZero() bool
	// time转换方法
	Format(string) string
}

func FormatK8sTime(k8sTime K8sTime) string {
	if !k8sTime.IsZero() {
		return k8sTime.Format(DefaultTimeFormatLayout)
	}

	return ""
}
