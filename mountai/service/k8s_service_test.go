package service

import (
	"rulai/models/entity"
	"rulai/models/req"

	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

// 测试的相对路径有区别
const testTemplateFileDir = "../template/k8s/"

func TestService_Service(t *testing.T) {
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
				assert.Equal(t, tpl.Namespace, service.GetNamespace())
				assert.Equal(t, "ams-app-framework-http", service.GetName())
				assert.Equal(t, tpl.ProjectName, service.GetLabels()["project"])
				assert.Equal(t, tpl.AppName, service.GetLabels()["app"])
				assert.Equal(t, v1.ServiceTypeClusterIP, service.Spec.Type)
				assert.Equal(t, strconv.Itoa(int(tpl.Ports[0].TargetPort)), service.Spec.Ports[0].TargetPort.String())
				beforeNodePort := service.Spec.Ports[0].NodePort

				data, err = s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyService(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)

				time.Sleep(time.Second * 10)

				// 检查node port是否变化
				service, err = s.GetServiceDetail(context.Background(), ensuredClusterName,
					&req.GetServiceDetailReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.Nil(t, err)
				assert.Equal(t, beforeNodePort, service.Spec.Ports[0].NodePort)

				// 测试service改targetPort
				tpl.Ports[0].TargetPort = 8080
				data, err = s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyService(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				list, err := s.GetServices(context.Background(), ensuredClusterName, string(entity.AppEnvStg),
					&req.GetServicesReq{
						Namespace:   "stg",
						ProjectName: "ams-app-framework",
						AppName:     "http",
						Env:         "stg",
					})
				assert.Nil(t, err)
				assert.Equal(t, 1, len(list))
				assert.Equal(t, "ams-app-framework-http", list[0].GetName())
				assert.Equal(t, strconv.Itoa(int(tpl.Ports[0].TargetPort)), list[0].Spec.Ports[0].TargetPort.String())

				svcDesc, err := s.DescribeService(context.Background(), ensuredClusterName,
					&req.DescribeServiceReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.Nil(t, err)
				assert.NotNil(t, svcDesc.Status)
				assert.Len(t, svcDesc.Events, 0)

				err = s.DeleteService(context.Background(), ensuredClusterName,
					&req.DeleteServiceReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				_, err = s.GetServiceDetail(context.Background(), ensuredClusterName,
					&req.GetServiceDetailReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.NotNil(t, err)

				_, err = s.DescribeService(context.Background(), ensuredClusterName,
					&req.DescribeServiceReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				assert.NotNil(t, err)
			})
		}
	}
}

func BenchmarkDescribeService(b *testing.B) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			b.Run(string(ensuredClusterName)+"::describe_service", func(b *testing.B) {
				tpl, err := s.initServiceTemplate(
					context.Background(),
					testProject,
					testRestfulServiceApp,
					testRestfulServiceStartTask,
					testTeam,
					testRestfulServiceApp.ServiceName,
				)
				assert.Nil(b, err)

				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				if err != nil {
					b.Error(err)
				}
				fmt.Printf("%s\n", data)

				_, err = s.ApplyService(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				if err != nil {
					b.Error(err)
				}
				time.Sleep(time.Second * 15)

				_, err = s.GetServiceDetail(context.Background(), ensuredClusterName,
					&req.GetServiceDetailReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				if err != nil {
					b.Error(err)
				}

				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					_, e := s.DescribeService(context.Background(), ensuredClusterName,
						&req.DescribeServiceReq{
							Namespace: "stg",
							Name:      "ams-app-framework-http",
							Env:       "stg",
						})
					if e != nil {
						b.Error(e)
					}
				}
				b.StopTimer()

				err = s.DeleteService(context.Background(), ensuredClusterName,
					&req.DeleteServiceReq{
						Namespace: "stg",
						Name:      "ams-app-framework-http",
						Env:       "stg",
					})
				if err != nil {
					b.Error(err)
				}
			})
		}
	}
}
