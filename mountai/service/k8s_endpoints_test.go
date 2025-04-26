package service

import (
	"rulai/models/entity"
	"rulai/models/req"

	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestService_Endpoints(t *testing.T) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				tpl, err := s.initServiceTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					testRestfulServiceStartTask,
					testTeam,
					testRestfulServiceApp.ServiceName,
				)
				assert.Nil(t, err)

				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyService(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				service, err := s.GetServiceDetail(context.Background(), ensuredClusterName,
					&req.GetServiceDetailReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.Nil(t, err)

				endpoints, err := s.GetEndpointsDetail(context.Background(), ensuredClusterName,
					&req.GetEndpointsReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.Nil(t, err)
				assert.Equal(t, service.Name, endpoints.ObjectMeta.Name)

				err = s.DeleteService(context.Background(), ensuredClusterName,
					&req.DeleteServiceReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.Nil(t, err)
			})
		}
	}
}
