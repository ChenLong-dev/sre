package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"

	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
)

func TestService_Deployment(t *testing.T) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				task := new(resp.TaskDetailResp)
				e := deepcopy.Copy(testRestfulServiceStartTask).To(task)
				require.Nil(t, e)
				task.ClusterName = ensuredClusterName

				tpl, err := s.initDeploymentTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					task,
					testTeam,
				)
				require.NoError(t, err)

				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Println(err)
				fmt.Printf("%s\n", data)

				err = s.ApplyDeploymentAndIgnoreResponse(context.Background(), task.ClusterName, task.EnvName, []byte(data))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				deploy, err := s.GetDeploymentDetail(context.Background(), ensuredClusterName, task.EnvName,
					&req.GetDeploymentDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, deploy.GetNamespace())
				assert.Equal(t, tpl.DeploymentVersion, deploy.GetName())
				assert.Equal(t, tpl.ProjectName, deploy.GetLabels()["project"])
				assert.Equal(t, tpl.AppName, deploy.GetLabels()["app"])
				assert.Equal(t, tpl.DeploymentVersion, deploy.GetLabels()["version"])
				assert.Equal(t, int32(2), *deploy.Spec.Replicas)

				tpl.Replicas = 1
				data, err = s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				err = s.ApplyDeploymentAndIgnoreResponse(context.Background(), task.ClusterName, task.EnvName, []byte(data))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				list, err := s.GetDeployments(context.Background(), ensuredClusterName, task.EnvName,
					&req.GetDeploymentsReq{
						Namespace:   tpl.Namespace,
						ProjectName: tpl.ProjectName,
						AppName:     tpl.AppName,
						Env:         string(entity.AppEnvStg),
					})
				assert.Nil(t, err)
				assert.Equal(t, 1, len(list))
				assert.Equal(t, tpl.Namespace, list[0].GetNamespace())
				assert.Equal(t, tpl.DeploymentVersion, list[0].GetName())
				assert.Equal(t, tpl.ProjectName, list[0].GetLabels()["project"])
				assert.Equal(t, tpl.AppName, list[0].GetLabels()["app"])
				assert.Equal(t, tpl.DeploymentVersion, list[0].GetLabels()["version"])
				assert.Equal(t, int32(1), *list[0].Spec.Replicas)

				err = s.UpdateDeploymentScaleAndIgnoreResponse(context.Background(), ensuredClusterName, task.EnvName,
					tpl.Namespace, tpl.DeploymentVersion, 2)
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				deploy, err = s.GetDeploymentDetail(context.Background(), ensuredClusterName, task.EnvName,
					&req.GetDeploymentDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, deploy.GetNamespace())
				assert.Equal(t, tpl.DeploymentVersion, deploy.GetName())
				assert.Equal(t, tpl.ProjectName, deploy.GetLabels()["project"])
				assert.Equal(t, tpl.AppName, deploy.GetLabels()["app"])
				assert.Equal(t, tpl.DeploymentVersion, deploy.GetLabels()["version"])
				assert.Equal(t, int32(2), *deploy.Spec.Replicas)

				restartReq := &req.RestartDeploymentReq{
					Namespace: deploy.Namespace,
					Name:      deploy.Name,
				}
				err = s.RestartDeploymentAndIgnoreResponse(context.Background(), ensuredClusterName, task.EnvName, restartReq)
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				deploy, err = s.GetDeploymentDetail(context.Background(), ensuredClusterName, task.EnvName,
					&req.GetDeploymentDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, deploy.GetNamespace())
				assert.Equal(t, tpl.DeploymentVersion, deploy.GetName())
				assert.Equal(t, tpl.ProjectName, deploy.GetLabels()["project"])
				assert.Equal(t, tpl.AppName, deploy.GetLabels()["app"])
				assert.Equal(t, tpl.DeploymentVersion, deploy.GetLabels()["version"])
				assert.Equal(t, int32(2), *deploy.Spec.Replicas)
				assert.NotEmpty(t, deploy.Spec.Template.ObjectMeta.Annotations[K8sAnnotationRestart])

				time.Sleep(time.Second * 10)
				deployment, err := s.DescribeDeployment(context.Background(), ensuredClusterName, task.EnvName,
					&req.DescribeDeploymentReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				assert.NotEmpty(t, deployment.Status)
				assert.NotZero(t, len(deployment.Events))

				grpcTpl, err := s.initDeploymentTemplate(
					context.Background(),
					testProject,
					testGRPCServiceApp,
					testGRPCServiceStartTask,
					testTeam,
				)
				require.NoError(t, err)

				grpcData, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, grpcTpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", grpcData)

				err = s.ApplyDeploymentAndIgnoreResponse(context.Background(), task.ClusterName, task.EnvName, []byte(grpcData))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				err = s.DeleteDeployments(context.Background(), ensuredClusterName, task.EnvName,
					&req.DeleteDeploymentsReq{
						Namespace:      grpcTpl.Namespace,
						ProjectName:    grpcTpl.ProjectName,
						InverseVersion: tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 15)

				_, err = s.GetDeploymentDetail(context.Background(), ensuredClusterName, task.EnvName,
					&req.GetDeploymentDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				_, err = s.GetDeploymentDetail(context.Background(), ensuredClusterName, task.EnvName,
					&req.GetDeploymentDetailReq{
						Namespace: grpcTpl.Namespace,
						Name:      grpcTpl.DeploymentVersion,
					})
				assert.NotNil(t, err)

				err = s.DeleteDeployment(context.Background(), ensuredClusterName, task.EnvName,
					&req.DeleteDeploymentReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 15)

				_, err = s.GetDeploymentDetail(context.Background(), ensuredClusterName, task.EnvName,
					&req.GetDeploymentDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.NotNil(t, err)

				_, err = s.DescribeDeployment(context.Background(), ensuredClusterName, task.EnvName,
					&req.DescribeDeploymentReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.NotNil(t, err)
			})
		}
	}
}
