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
)

func TestService_Job(t *testing.T) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				tpl, err := s.initJobTemplate(
					context.Background(),
					testProject,
					testJobApp,
					testJobStartTask,
					testTeam,
				)
				require.Nil(t, err)
				assert.Equal(t, tpl.BackoffLimit, testJobStartTask.Param.BackoffLimit)

				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				require.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyJob(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				require.Nil(t, err)

				time.Sleep(time.Second * 5)

				job, err := s.GetJobDetail(context.Background(), ensuredClusterName,
					&req.GetJobDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.JobVersion,
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, job.GetNamespace())
				assert.Equal(t, tpl.ProjectName, job.GetLabels()["project"])
				assert.Equal(t, tpl.AppName, job.GetLabels()["app"])

				jobDesc, err := s.DescribeJob(context.Background(), ensuredClusterName,
					&req.DescribeJobReq{
						Namespace: tpl.Namespace,
						Name:      tpl.JobVersion,
					})
				assert.Nil(t, err)
				assert.NotEmpty(t, jobDesc.Status)
				assert.NotEqual(t, 0, len(jobDesc.Events))

				jobs, err := s.GetJobs(context.Background(), ensuredClusterName,
					&req.GetJobsReq{
						Namespace:   tpl.Namespace,
						ProjectName: tpl.ProjectName,
						AppName:     tpl.AppName,
					})
				assert.Nil(t, err)
				assert.Equal(t, 1, len(jobs))
				assert.Equal(t, tpl.Namespace, jobs[0].GetNamespace())
				assert.Equal(t, tpl.ProjectName, jobs[0].GetLabels()["project"])
				assert.Equal(t, tpl.AppName, jobs[0].GetLabels()["app"])

				time.Sleep(time.Second * 30)

				pods, err := s.GetPods(context.Background(), ensuredClusterName,
					&req.GetPodsReq{
						Namespace:   tpl.Namespace,
						ProjectName: tpl.ProjectName,
						AppName:     tpl.AppName,
						JobName:     job.GetName(),
					})
				assert.Nil(t, err)
				assert.Equal(t, 1, len(pods))

				err = s.DeleteJob(context.Background(), ensuredClusterName,
					&req.DeleteJobReq{
						Namespace: tpl.Namespace,
						Name:      tpl.JobVersion,
					})
				assert.Nil(t, err)

				time.Sleep(time.Second * 15)

				_, err = s.GetJobDetail(context.Background(), ensuredClusterName,
					&req.GetJobDetailReq{
						Namespace: tpl.Namespace,
						Name:      job.GetName(),
					})
				assert.NotNil(t, err)

				_, err = s.DescribeJob(context.Background(), ensuredClusterName,
					&req.DescribeJobReq{
						Namespace: tpl.Namespace,
						Name:      tpl.JobVersion,
					})
				assert.NotNil(t, err)
			})
		}
	}
}
