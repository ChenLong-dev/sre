package req

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 获取Deployment详情请求参数
type GetDeploymentDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"` // 拆分 env 和 namespace
}

// GetDeploymentsReq 批量获取Deployment列表请求参数
type GetDeploymentsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
	Env         string `json:"env"`
}

// DeleteDeploymentsReq 批量删除Deployment请求参数
type DeleteDeploymentsReq struct {
	Namespace      string                 `json:"namespace"`
	ProjectName    string                 `json:"project_name"`
	AppName        string                 `json:"app_name"`
	Policy         v1.DeletionPropagation `json:"policy"`
	InverseVersion string                 `json:"inverse_version"`
	Env            string                 `json:"env"`
}

// DeleteDeploymentReq 删除Deployment请求参数
type DeleteDeploymentReq struct {
	Namespace string                 `json:"namespace"`
	Name      string                 `json:"name"`
	Policy    v1.DeletionPropagation `json:"policy"`
	Env       string                 `json:"env"`
}

// DescribeDeploymentReq 获取Deployment的Describe信息请求参数
type DescribeDeploymentReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

// RestartDeploymentReq 重启Deployment请求参数
type RestartDeploymentReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

// EnableDeploymentHPAReq 启用Deployment hpa请求参数
type EnableDeploymentHPAReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}
