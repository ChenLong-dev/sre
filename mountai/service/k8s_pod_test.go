package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
)

func TestService_Pod(t *testing.T) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			configData, err := s.GetAppConfig(context.Background(), &req.GetConfigManagerFileReq{
				ProjectID:  testProject.ID,
				EnvName:    testRestfulServiceStartTask.EnvName,
				CommitID:   testRestfulServiceStartTask.Param.ConfigCommitID,
				IsDecrypt:  true,
				FormatType: req.ConfigManagerFormatTypeJSON,
			})
			assert.Nil(t, err)

			configTpl, err := s.initAppConfigMapTemplate(
				context.Background(),
				testProject,
				testRestfulServiceApp,
				testTeam,
				testRestfulServiceStartTask,
				configData.Config.(map[string]interface{}),
			)
			require.NoError(t, err)

			data, err := s.RenderK8sTemplate(context.Background(),
				testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, configTpl)
			assert.Nil(t, err)
			fmt.Printf("%s\n", data)

			_, err = s.ApplyConfigMap(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
			assert.Nil(t, err)
			time.Sleep(time.Second * 10)

			_, err = s.GetConfigMapDetail(context.Background(), ensuredClusterName,
				&req.GetConfigMapDetailReq{
					Namespace: configTpl.Namespace,
					Name:      configTpl.Name,
				})
			assert.Nil(t, err)

			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				task := new(resp.TaskDetailResp)
				e := deepcopy.Copy(testRestfulServiceStartTask).To(task)
				require.Nil(t, e)
				task.ClusterName = ensuredClusterName

				tpl, initErr := s.initDeploymentTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					task,
					testTeam,
				)
				require.NoError(t, initErr)

				data, e := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, e)
				fmt.Printf("%s\n", data)

				e = s.ApplyDeploymentAndIgnoreResponse(context.Background(), task.ClusterName, task.EnvName, []byte(data))
				assert.Nil(t, e)
				time.Sleep(time.Second * 5)

				pods, e := s.GetPods(context.Background(), ensuredClusterName,
					&req.GetPodsReq{
						Namespace:   tpl.Namespace,
						ProjectName: tpl.ProjectName,
						AppName:     tpl.AppName,
						Version:     tpl.DeploymentVersion,
					})
				assert.Nil(t, e)
				assert.Equal(t, 2, len(pods))
				assert.Equal(t, tpl.Namespace, pods[0].GetNamespace())
				assert.Equal(t, tpl.ProjectName, pods[0].GetLabels()["project"])
				assert.Equal(t, tpl.AppName, pods[0].GetLabels()["app"])
				assert.Equal(t, tpl.DeploymentVersion, pods[0].GetLabels()["version"])
				time.Sleep(time.Second * 30)

				res, e := s.ExecPodCommand(context.Background(), ensuredClusterName,
					&req.ExecPodReq{
						Namespace: tpl.Namespace,
						Name:      pods[0].GetName(),
						Env:       string(entity.AppEnvStg),
						Commands:  []string{"echo", "this is a test"},
					})
				assert.Nil(t, e)
				assert.Equal(t, "this is a test\n", string(res))

				pod, e := s.GetPodDetail(context.Background(), ensuredClusterName,
					&req.GetPodDetailReq{
						Namespace: tpl.Namespace,
						Name:      pods[1].GetName(),
						Env:       string(entity.AppEnvStg),
					})
				assert.Nil(t, e)
				assert.Equal(t, tpl.DeploymentVersion, pod.GetLabels()["version"])

				podDesc, e := s.DescribePod(context.Background(), ensuredClusterName,
					&req.GetRunningPodDescriptionReq{
						EnvName: tpl.Namespace,
						Name:    pods[1].GetName(),
					})
				assert.Nil(t, e)
				assert.NotEmpty(t, podDesc.Status)
				assert.NotEqual(t, 0, len(podDesc.Events))

				e = s.DeleteDeployment(context.Background(), ensuredClusterName, task.EnvName,
					&req.DeleteDeploymentReq{
						Namespace: tpl.Namespace,
						Name:      tpl.DeploymentVersion,
					})
				assert.Nil(t, e)
				time.Sleep(time.Second * 15)

				_, e = s.GetPodDetail(context.Background(), ensuredClusterName,
					&req.GetPodDetailReq{
						Namespace: tpl.Namespace,
						Name:      pods[1].GetName(),
					})
				assert.NotNil(t, e)

				_, e = s.DescribePod(context.Background(), ensuredClusterName,
					&req.GetRunningPodDescriptionReq{
						EnvName: tpl.Namespace,
						Name:    pods[1].GetName(),
					})
				assert.NotNil(t, e)
			})

			err = s.DeleteConfigMap(context.Background(), ensuredClusterName,
				&req.DeleteConfigMapReq{
					Namespace: configTpl.Namespace,
					Name:      configTpl.Name,
				})
			assert.Nil(t, err)
			time.Sleep(time.Second * 10)

			_, err = s.GetConfigMapDetail(context.Background(), ensuredClusterName,
				&req.GetConfigMapDetailReq{
					Namespace: configTpl.Namespace,
					Name:      configTpl.Name,
				})
			assert.NotNil(t, err)
		}
	}
}
