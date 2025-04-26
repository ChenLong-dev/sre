package utils

import (
	"fmt"
	"strings"
)

// UnifyK8sMinorVersion 输出k8s标准小版本
// TODO: 调研小版本的+号是否是标准 (e.g.: 18+)
func UnifyK8sMinorVersion(v string) string {
	return strings.TrimRight(v, "+")
}

// GetPodContainerName 获取pod容器名称
func GetPodContainerName(projectName, appName string) string {
	return fmt.Sprintf("%s-%s", projectName, appName)
}
