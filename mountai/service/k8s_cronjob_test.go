package service

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"

	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
)

func TestService_CronJob(t *testing.T) {
	// 分集群测试
	for _, clusterSet := range s.k8sClusters {
		for clusterName := range clusterSet {
			ensuredClusterName := clusterName
			t.Run(string(ensuredClusterName)+"::app-framework", func(t *testing.T) {
				// TODO: 初始化必要的 ConfigMap
				task := new(resp.TaskDetailResp)
				err := deepcopy.Copy(testCronJobStartTask).To(task)
				require.Nil(t, err)

				task.ClusterName = ensuredClusterName

				tpl, err := s.initCronJobTemplate(
					context.Background(),
					testProject,
					testCronJobApp,
					task,
					testTeam,
				)
				require.NoError(t, err)

				data, err := s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyCronJob(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				job, err := s.GetCronJobDetail(context.Background(), ensuredClusterName,
					&req.GetCronJobDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.CronJobVersion,
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.Namespace, job.GetNamespace())
				assert.Equal(t, tpl.CronJobVersion, job.GetName())
				assert.Equal(t, tpl.ProjectName, job.GetLabels()["project"])
				assert.Equal(t, tpl.AppName, job.GetLabels()["app"])
				assert.Equal(t, tpl.CronJobVersion, job.GetLabels()["version"])
				assert.Equal(t, tpl.Schedule, job.Spec.Schedule)
				assert.False(t, *job.Spec.Suspend)

				tpl.Schedule = "0 1 * * *"
				data, err = s.RenderK8sTemplate(context.Background(),
					testTemplateFileDir, ensuredClusterName, entity.AppEnvStg, tpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", data)

				_, err = s.ApplyCronJob(context.Background(), ensuredClusterName, []byte(data), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				list, err := s.GetCronJobs(context.Background(), ensuredClusterName,
					&req.GetCronJobsReq{
						Namespace:   tpl.Namespace,
						ProjectName: tpl.ProjectName,
						AppName:     tpl.AppName,
					})
				assert.Nil(t, err)
				assert.Equal(t, 1, len(list))
				assert.Equal(t, tpl.Schedule, list[0].Spec.Schedule)

				_, err = s.SuspendCronJob(context.Background(), ensuredClusterName, &list[0], string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				job, err = s.GetCronJobDetail(context.Background(), ensuredClusterName,
					&req.GetCronJobDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.CronJobVersion,
					})
				assert.Nil(t, err)
				assert.True(t, *job.Spec.Suspend)

				_, err = s.ResumeCronJob(context.Background(), ensuredClusterName, job, string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				job, err = s.GetCronJobDetail(context.Background(), ensuredClusterName,
					&req.GetCronJobDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.CronJobVersion,
					})
				assert.Nil(t, err)
				assert.False(t, *job.Spec.Suspend)

				testCronJobOtherTask := new(resp.TaskDetailResp)
				err = deepcopy.Copy(task).To(testCronJobOtherTask)
				assert.Nil(t, err)

				testCronJobOtherApp := new(resp.AppDetailResp)
				err = deepcopy.Copy(testCronJobApp).To(testCronJobOtherApp)
				assert.Nil(t, err)

				testCronJobOtherTask.Param.CronParam = "* 10 * * *"
				testCronJobOtherApp.Name = "other-cronjob"
				testCronJobOtherTask.Version = "ams-app-framework-other-cronjob"
				otherTpl, err := s.initCronJobTemplate(
					context.Background(),
					testProject,
					testCronJobOtherApp,
					testCronJobOtherTask,
					testTeam,
				)
				require.NoError(t, err)

				otherData, err := s.RenderTemplate(context.Background(),
					testTemplateFileDir, otherTpl)
				assert.Nil(t, err)
				fmt.Printf("%s\n", otherData)

				_, err = s.ApplyCronJob(context.Background(), ensuredClusterName, []byte(otherData), string(entity.AppEnvStg))
				assert.Nil(t, err)
				time.Sleep(time.Second * 5)

				err = s.DeleteCronJobs(context.Background(), ensuredClusterName,
					&req.DeleteCronJobsReq{
						Namespace:   otherTpl.Namespace,
						ProjectName: otherTpl.ProjectName,
						InverseName: tpl.CronJobVersion,
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 15)

				_, err = s.GetCronJobDetail(context.Background(), ensuredClusterName,
					&req.GetCronJobDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.CronJobVersion,
					})
				assert.Nil(t, err)
				_, err = s.GetCronJobDetail(context.Background(), ensuredClusterName,
					&req.GetCronJobDetailReq{
						Namespace: tpl.Namespace,
						Name:      otherTpl.CronJobVersion,
					})
				assert.NotNil(t, err)

				cronjob, err := s.DescribeCronJob(context.Background(), ensuredClusterName,
					&req.DescribeCronJobReq{
						Namespace: tpl.Namespace,
						Name:      tpl.CronJobVersion,
					})
				assert.Nil(t, err)
				assert.Equal(t, tpl.CronJobVersion, cronjob.Status.Name)

				_, err = s.DescribeCronJob(context.Background(), ensuredClusterName,
					&req.DescribeCronJobReq{
						Namespace: tpl.Namespace,
						Name:      otherTpl.CronJobVersion,
					})
				assert.NotNil(t, err)

				err = s.DeleteCronJob(context.Background(), ensuredClusterName,
					&req.DeleteCronJobReq{
						Namespace: tpl.Namespace,
						Name:      tpl.CronJobVersion,
					})
				assert.Nil(t, err)
				time.Sleep(time.Second * 15)

				_, err = s.GetCronJobDetail(context.Background(), ensuredClusterName,
					&req.GetCronJobDetailReq{
						Namespace: tpl.Namespace,
						Name:      tpl.CronJobVersion,
					})
				assert.NotNil(t, err)
			})
		}
	}
}

func Test_GetCronjobNextSchedTime(t *testing.T) {
	t.Run("getCronjobNextSchedTime", func(t *testing.T) {
		now := "2021-02-02 18:21:36"
		testSchedData := []struct {
			expr     string
			expected string
			err      string
		}{
			{
				expr:     "*/5 * * * *",
				expected: "2021-02-02 18:25:00",
			},
			{
				expr:     "* */1 * * *",
				expected: "2021-02-02 18:22:00",
			},
			{
				expr:     "15 8 * * *",
				expected: "2021-02-03 08:15:00",
			},
			{
				expr:     "@every 5m",
				expected: "2021-02-02 18:26:36",
			},
			{
				expr: "5 j * * *",
				err:  "failed to parse int from",
			},
			{
				expr: "* * * *",
				err:  "expected exactly 5 fields",
			},
		}

		for _, c := range testSchedData {
			lastTime, err := time.Parse(utils.DefaultTimeFormatLayout, now)
			if err != nil {
				t.Errorf("time parse err: %v", err)
			}
			schedTime, err := s.getCronjobNextSchedTime(c.expr, lastTime)
			if c.err != "" && (err == nil || !strings.Contains(err.Error(), c.err)) {
				t.Errorf("%s => expected %v, got %v", c.expr, c.err, err)
			}
			if c.err == "" && err != nil {
				t.Errorf("%s => unexpected error %v", c.expr, err)
			}
			if !reflect.DeepEqual(schedTime, c.expected) {
				t.Errorf("%s => expected %s, got %s", c.expr, c.expected, schedTime)
			}
		}
	})
}
