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

func TestService_GetK8sResourceEvents(t *testing.T) {
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
						Env:       "stg",
						Name:      "ams-app-framework-http",
					})

				assert.Nil(t, err)

				events, err := s.GetK8sResourceEvents(context.Background(), ensuredClusterName,
					&req.GetK8sResourceEventsReq{
						Namespace: "stg",
						Resource:  service,
					})
				assert.Nil(t, err)
				assert.NotEmpty(t, events.Events)
				assert.NotZero(t, len(events.Events))

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
