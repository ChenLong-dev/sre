package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"

	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
)

func TestService_ReplicaSet(t *testing.T) {
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

				err = s.ApplyDeploymentAndIgnoreResponse(context.Background(), task.ClusterName, task.EnvName, []byte(data))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				rs, err := s.GetReplicaSets(context.Background(), ensuredClusterName, string(entity.AppEnvStg),
					&req.GetReplicaSetsReq{
						Namespace:   tpl.Namespace,
						ProjectName: tpl.ProjectName,
						AppName:     tpl.AppName,
						Version:     tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				assert.Equal(t, 1, len(rs))
				assert.Equal(t, tpl.Namespace, rs[0].GetNamespace())
				assert.Equal(t, tpl.ProjectName, rs[0].GetLabels()["project"])
				assert.Equal(t, tpl.AppName, rs[0].GetLabels()["app"])
				assert.Equal(t, tpl.DeploymentVersion, rs[0].GetLabels()["version"])
				assert.Equal(t, int32(2), *rs[0].Spec.Replicas)

				deploy, err := s.GetDeploymentDetail(context.Background(), ensuredClusterName, task.EnvName,
					&req.GetDeploymentDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.Nil(t, err)

				restartReq := &req.RestartDeploymentReq{
					Namespace: deploy.Namespace,
					Name:      deploy.Name,
				}
				err = s.RestartDeploymentAndIgnoreResponse(context.Background(), ensuredClusterName, task.EnvName, restartReq)
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				oldRS, err := s.GetReplicaSetDetail(context.Background(), ensuredClusterName, string(entity.AppEnvStg),
					&req.GetReplicaSetDetailReq{
						Namespace: tpl.Namespace,
						Name:      rs[0].GetName(),
					})
				assert.Nil(t, err)
				assert.Less(t, *oldRS.Spec.Replicas, int32(2))

				rs, err = s.GetReplicaSets(context.Background(), ensuredClusterName, string(entity.AppEnvStg),
					&req.GetReplicaSetsReq{
						Namespace:   tpl.Namespace,
						ProjectName: tpl.ProjectName,
						AppName:     tpl.AppName,
						Version:     tpl.DeploymentVersion,
					})
				assert.Nil(t, err)
				assert.Equal(t, 2, len(rs))
				assert.Equal(t, tpl.Namespace, rs[1].GetNamespace())
				assert.Greater(t, *rs[1].Spec.Replicas, int32(0))

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
			})
		}
	}
}
