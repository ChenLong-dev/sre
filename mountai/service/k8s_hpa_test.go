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

func TestService_HPA(t *testing.T) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				tpl, err := s.initHPATemplate(
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
				fmt.Printf("apply HPA: %s, err: %v\n", data, err)

				_, err = s.ApplyHPA(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				hpa, err := s.GetHPADetail(context.Background(), ensuredClusterName,
					&req.GetHPADetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				fmt.Printf("err: %v\n", err)
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, hpa.GetNamespace())
				assert.Equal(t, tpl.MaxReplicas, hpa.Spec.MaxReplicas)

				hpaDesc, err := s.DescribeHPA(context.Background(), ensuredClusterName,
					&req.DescribeHPAReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.Nil(t, err)
				assert.NotEmpty(t, hpaDesc.Status)

				tpl.MaxReplicas = int32(6)
				data, err = s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyHPA(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				hpa, err = s.GetHPADetail(context.Background(), ensuredClusterName,
					&req.GetHPADetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, hpa.GetNamespace())
				assert.Equal(t, tpl.MaxReplicas, hpa.Spec.MaxReplicas)

				err = s.DeleteHPA(context.Background(), ensuredClusterName,
					&req.DeleteHPAReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 10)

				_, err = s.GetHPADetail(context.Background(), ensuredClusterName,
					&req.GetHPADetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.NotNil(t, err)

				_, err = s.DescribeHPA(context.Background(), ensuredClusterName,
					&req.DescribeHPAReq{
						Namespace: tpl.Namespace,
						Name:      tpl.Name,
						Env:       string(entity.AppEnvStg),
					})
				assert.NotNil(t, err)
			})
		}
	}
}
