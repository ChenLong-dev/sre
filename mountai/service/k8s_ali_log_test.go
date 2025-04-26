package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	_errcode "rulai/utils/errcode"

	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestService_AliLogConfig(t *testing.T) {
	// 分集群测试
	for envName, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			// CRD 不一定所有集群都有
			err := s.isClusterSupportedK8sObjectKind(clusterName, string(envName), entity.K8sObjectKindAliyunLogConfig)
			if err != nil {
				require.True(t, errcode.EqualError(_errcode.UnsupportedK8sObjectKind, err))
				continue
			}

			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				tpl, err := s.initAliLogConfigTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					testRestfulServiceStartTask,
					testTeam,
				)
				assert.Nil(t, err)

				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyAliLogConfig(context.Background(), ensuredClusterName, string(entity.AppEnvStg), []byte(data))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				aliLogConfig, err := s.GetAliLogConfigDetail(context.Background(), ensuredClusterName,
					&req.GetAliLogConfigDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, aliLogConfig.GetNamespace())
				assert.Equal(t, tpl.TeamLabel, aliLogConfig.GetLabels()["team"])
				logStoreName, isExisted, err := unstructured.NestedString(aliLogConfig.Object, "spec", "logstore")
				assert.Nil(t, err)
				assert.True(t, isExisted)
				assert.Equal(t, tpl.LogStoreName, logStoreName)

				list, err := s.ListAliLogConfig(context.Background(), ensuredClusterName, string(entity.AppEnvStg),
					&req.ListAliLogConfigReq{
						Namespace:   tpl.Namespace,
						ProjectName: testProject.Name,
					})
				assert.Nil(t, err)
				assert.Equal(t, len(list.Items), 1)
				assert.Equal(t, list.Items[0].GetLabels()["project"], testProject.Name)

				tpl.TeamLabel = "vip"
				data, err = s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyAliLogConfig(context.Background(), ensuredClusterName, string(entity.AppEnvStg), []byte(data))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				aliLogConfig, err = s.GetAliLogConfigDetail(context.Background(), ensuredClusterName,
					&req.GetAliLogConfigDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, aliLogConfig.GetNamespace())
				assert.Equal(t, tpl.TeamLabel, aliLogConfig.GetLabels()["team"])

				err = s.DeleteAliLogConfig(context.Background(), ensuredClusterName, string(entity.AppEnvStg),
					&req.DeleteAliLogConfigReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				_, err = s.GetAliLogConfigDetail(context.Background(), ensuredClusterName, &req.GetAliLogConfigDetailReq{
					Namespace: tpl.Namespace,
					Name:      tpl.Name,
				})
				assert.NotNil(t, err)

				list, err = s.ListAliLogConfig(context.Background(), ensuredClusterName, string(entity.AppEnvStg), &req.ListAliLogConfigReq{
					Namespace:   tpl.Namespace,
					ProjectName: testProject.Name,
				})
				assert.Nil(t, err)
				assert.Equal(t, len(list.Items), 0)
			})
		}
	}
}
