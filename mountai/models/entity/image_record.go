package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// 容器类型
const (
	ContainerTypeNormalContainer    = "normal_container"
	ContainerTypeInitContainer      = "init_container"
	ContainerTypeEphemeralContainer = "ephemeral_container"
)

// EmptyUnexpectedImageRecord 通用空对象
var EmptyUnexpectedImageRecord = new(UnexpectedImageRecord)

// UnexpectedImageRecord 非预期镜像记录
// 目前非预期镜像指的是华为镜像仓库之外的镜像
type UnexpectedImageRecord struct {
	ID                 primitive.ObjectID     `bson:"_id,omitempty" json:"_id"`
	Cluster            ClusterName            `bson:"cluster" json:"cluster"`
	Namespace          string                 `bson:"namespace" json:"namespace"`
	AMSProjectName     string                 `bson:"ams_project_name" json:"ams_project_name"`
	AMSAppName         string                 `bson:"ams_app_name" json:"ams_app_name"`
	OwnerReferenceKind string                 `bson:"owner_reference_kind" json:"owner_reference_kind"`
	OwnerReferenceName string                 `bson:"owner_reference_name" json:"owner_reference_name"`
	PodName            string                 `bson:"pod_name" json:"pod_name"`
	ImageList          []*UnexpectedImageInfo `bson:"image_list" json:"image_list"`
}

// UnexpectedImageInfo 非预期镜像信息
type UnexpectedImageInfo struct {
	ContainerName string `bson:"container_name" json:"container_name"`
	ContainerType string `bson:"container_type" json:"container_type"`
	Image         string `bson:"image" json:"image"`
}

func (r *UnexpectedImageRecord) TableName() string { return "unexpected_image_record" }
