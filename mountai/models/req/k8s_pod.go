package req

import (
	"rulai/models/entity"
)

type GetPodDetailReq struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Env       string `json:"env"`
}

type GetPodsReq struct {
	Namespace   string `json:"namespace"`
	ProjectName string `json:"project_name"`
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
	JobName     string `json:"job_name"`
	Env         string `json:"env"` // 拆分 namespace 和 env
}

type GetPodLogReq struct {
	Namespace     string `json:"namespace"`
	Name          string `json:"name"`
	Env           string `json:"env"`
	ContainerName string `json:"container_name"`
}

type ExecPodReq struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Commands  []string `json:"commands"`
	Env       string   `json:"env"`
	Container string   `json:"container"`
}

type GetRunningPodDescriptionReq struct {
	EnvName       string             `form:"env_name" json:"env_name" binding:"required"`
	ClusterName   entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
	Name          string             `form:"-" json:"-"`
	Namespace     string             `form:"namespace" json:"namespace"`
	Container     string             `form:"-" json:"-"`
	Env           string             `json:"env"`
	ContainerName string             `form:"container_name" json:"container_name" binding:"required"`
}
