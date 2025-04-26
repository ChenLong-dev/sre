package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	"time"
)

// 镜像信息
type JenkinsBuildImage struct {
	ID                  primitive.ObjectID `bson:"_id" json:"_id"`
	ProjectID           string             `bson:"project_id" json:"project_id"`
	ProjectName         string             `bson:"project_name" json:"project_name"`
	ImageTag            string             `bson:"image_tag" json:"image_tag"`
	BuildID             string             `bson:"build_id" json:"build_id"`
	JobName             string             `bson:"job_name" json:"job_name"`
	JobURL              string             `bson:"job_url" json:"job_url"`
	ImageRepoURL        string             `bson:"image_repo_url" json:"image_repo_url"`
	BuildArg            string             `bson:"build_arg" json:"build_arg"`
	BuildArgWithMask    string             `bson:"build_arg_with_mask" json:"build_arg_with_mask"`
	BuildArgsTemplateID string             `bson:"build_args_template_id" json:"build_args_template_id"`
	BranchName          string             `bson:"branch_name" json:"branch_name"`
	CommitID            string             `bson:"commit_id" json:"commit_id"`
	Description         string             `bson:"description" json:"description"`
	UserID              string             `bson:"user_id" json:"user_id"`
	CreateTime          *time.Time         `bson:"create_time" json:"create_time"`
	UpdateTime          *time.Time         `bson:"update_time" json:"update_time"`
}

func (*JenkinsBuildImage) TableName() string {
	return "jenkins_build_image"
}

type JenkinsBuild struct {
	BuildID      string `json:"build_id"`
	JobName      string `json:"job_name"`
	JobURL       string `json:"job_url"`
	ImageRepoURL string `json:"image_repo_url"`
	BuildArg     string `json:"build_arg"`
	BranchName   string `json:"branch_name"`
	CommitID     string `json:"commit_id"`
	CreateTime   string `json:"create_time"`
	Description  string `json:"description"`
	UserID       string `json:"user_id"`
}
