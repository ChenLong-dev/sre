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
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
)

func TestService_Ingress(t *testing.T) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				currentTask := new(resp.TaskDetailResp)
				err := deepcopy.Copy(testRestfulServiceStartTask).To(currentTask)
				assert.Nil(t, err)
				currentTask.ClusterName = ensuredClusterName

				tpl, err := s.initIngressTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					currentTask,
					testRestfulServiceApp.ServiceName,
				)
				assert.Nil(t, err)
				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyIngress(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				fmt.Printf("cluster: %s, apply ingress: %v\n", ensuredClusterName, err)

				ingress, err := s.GetIngressDetail(context.Background(), ensuredClusterName,
					&req.GetIngressDetailReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})

				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, ingress.GetNamespace())
				assert.Equal(t, "ams-app-framework-http", ingress.GetName())
				assert.Equal(t, tpl.ProjectName, ingress.GetLabels()["project"])
				assert.Equal(t, tpl.AppName, ingress.GetLabels()["app"])
				assert.Equal(t,
					s.getAliPrivateZoneK8sFullDomainName(testRestfulServiceApp.ServiceName, currentTask.EnvName),
					ingress.Spec.Rules[0].Host,
				)
				assert.Equal(t,
					s.getAliPrivateZoneK8sFullDomainNameWithCluster(
						testRestfulServiceApp.ServiceName, currentTask.EnvName, currentTask.ClusterName),
					ingress.Spec.Rules[1].Host,
				)
				assert.Equal(t, testRestfulServiceApp.ServiceName,
					ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName,
				)
				assert.Equal(t, testRestfulServiceApp.ServiceName,
					ingress.Spec.Rules[1].IngressRuleValue.HTTP.Paths[0].Backend.ServiceName,
				)

				tpl.ServiceHost = "test.ams-app-framework-http"
				tpl.ServiceHostWithCluster = "test.ams-app-framework-http." + string(ensuredClusterName)
				data, err = s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyIngress(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				ingress, err = s.GetIngressDetail(context.Background(), ensuredClusterName,
					&req.GetIngressDetailReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.ServiceHost, ingress.Spec.Rules[0].Host)
				assert.Equal(t, tpl.ServiceHostWithCluster, ingress.Spec.Rules[1].Host)

				err = s.DeleteIngress(context.Background(), ensuredClusterName,
					&req.DeleteIngressReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				_, err = s.GetIngressDetail(context.Background(), ensuredClusterName,
					&req.GetIngressDetailReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.NotNil(t, err)
			})
		}
	}
}
