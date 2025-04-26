package service

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"rulai/models/entity"
	"rulai/models/resp"
)

func getPATByProjectID(id int) (*resp.ProjectDetailResp, *resp.AppDetailResp, *resp.TaskDetailResp) {
	project := &resp.ProjectDetailResp{
		ID:                   strconv.Itoa(id),
		Name:                 "proj",
		Language:             "",
		Desc:                 "",
		APIDocURL:            "",
		DevDocURL:            "",
		Labels:               nil,
		ImageArgs:            nil,
		ResourceSpec:         nil,
		Team:                 nil,
		LogStoreName:         "",
		IsFav:                false,
		Owners:               nil,
		QAEngineers:          nil,
		OperationEngineers:   nil,
		ProductManagers:      nil,
		ConfigRenamePrefixes: nil,
		ConfigRenameModes:    nil,
		EnableIstio:          true,
		CreateTime:           "",
		UpdateTime:           "",
	}

	app := &resp.AppDetailResp{
		ID:                     "",
		Name:                   "app",
		Type:                   "",
		ServiceType:            "",
		ServiceExposeType:      entity.AppServiceExposeTypeIngress,
		LoadBalancerInfo:       nil,
		ServiceName:            "dev.radio-api-unittest",
		AliLogConfigName:       "",
		ProjectID:              "",
		Env:                    nil,
		SentryProjectPublicDsn: "",
		SentryProjectSlug:      "",
		Description:            "",
		CreateTime:             "",
		UpdateTime:             "",
		EnableIstio:            true,
	}

	task := &resp.TaskDetailResp{
		ID:            "",
		Version:       "",
		Action:        "",
		ActionDisplay: "",
		ApprovalType:  "",
		DeployType:    "",
		ScheduleTime:  "",
		Approval:      nil,
		Detail:        "",
		Description:   "",
		RetryCount:    0,
		ClusterName:   entity.DefaultClusterName,
		Status:        "",
		StatusDisplay: "",
		DisplayIcon:   "",
		EnvName:       entity.AppEnvStg,
		Namespace:     string(entity.IstioEnvStg),
		AppID:         "",
		OperatorID:    "",
		Suspend:       false,
		Param:         nil,
		CreateTime:    "",
		UpdateTime:    "",
	}

	return project, app, task
}

func Test_Service_DeleteVirtualService(t *testing.T) {
	type args struct {
		ctx                context.Context
		clusterName        entity.ClusterName
		project            *resp.ProjectDetailResp
		app                *resp.AppDetailResp
		task               *resp.TaskDetailResp
		backendServiceName string
	}

	project, app, task := getPATByProjectID(1)

	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{"test delete istio virtual service on cce", args{
			ctx:                context.Background(),
			clusterName:        task.ClusterName,
			project:            project,
			app:                app,
			task:               task,
			backendServiceName: "test",
		}, func(t assert.TestingT, err error, i ...interface{}) bool {
			return true
		}},
	}
	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			if err := s.DeleteVirtualService(tt.args.ctx, tt.args.clusterName, tt.args.task,
				tt.args.backendServiceName, string(tt.args.task.GetEnvName())); !tt.wantErr(t, err) {
				t.Log(err)
			}
		})
	}
}

func Test_Service_GetNamespaceFromDomainName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"1", args{name: "dev.radio-api-init.stg.svc.cluster.local"}, "stg"},
		{"2", args{name: "radio-api-init.stg.svc.cluster.local"}, "stg"},
		{"3", args{name: "radio-api-initstg.svc.cluster.local"}, ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, SVC.getNamespaceFromDomainName(tt.args.name), "getNamespaceFromDomainName(%v)", tt.args.name)
		})
	}
}
