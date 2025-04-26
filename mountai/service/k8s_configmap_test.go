package service

import (
	"rulai/models/entity"
	"rulai/models/req"

	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestService_ConfigMap(t *testing.T) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				configData, err := s.GetAppConfig(context.Background(), &req.GetConfigManagerFileReq{
					ProjectID:  testProject.ID,
					EnvName:    entity.AppEnvStg,
					CommitID:   "c91da6fae5653c09b84e46d5438de3be234f1057",
					IsDecrypt:  true,
					FormatType: req.ConfigManagerFormatTypeJSON,
				})
				assert.Nil(t, err)

				tpl, err := s.initAppConfigMapTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					testTeam,
					testRestfulServiceStartTask,
					configData.Config.(map[string]interface{}),
				)
				require.NoError(t, err)

				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyConfigMap(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				cm, err := s.GetConfigMapDetail(context.Background(), ensuredClusterName,
					&req.GetConfigMapDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, cm.GetNamespace())
				assert.Equal(t, tpl.Labels[ConfigMapLabelCommit], cm.GetLabels()[ConfigMapLabelCommit])
				assert.Equal(
					t,
					"abcdefg",
					cm.Data["config.txt"],
				)

				updateData := map[string]interface{}{
					"data": map[string]interface{}{
						"config.txt": "test",
					},
				}
				yamlData, err := yaml.Marshal(updateData)
				assert.Nil(t, err)
				tpl.Data = string(yamlData)

				data, err = s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyConfigMap(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				cm, err = s.GetConfigMapDetail(context.Background(), ensuredClusterName,
					&req.GetConfigMapDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, cm.GetNamespace())
				assert.Equal(
					t,
					"test",
					cm.Data["config.txt"],
				)

				err = s.DeleteConfigMap(context.Background(), ensuredClusterName,
					&req.DeleteConfigMapReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				_, err = s.GetConfigMapDetail(context.Background(), ensuredClusterName,
					&req.GetConfigMapDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.NotNil(t, err)
			})
		}
	}
}
