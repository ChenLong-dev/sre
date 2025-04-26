package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"gitlab.shanhai.int/sre/library/net/errcode"
	batchv1 "k8s.io/api/batch/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
)

// SortCronJobReverseSlice : CronJob的创建时间倒序数组
type SortCronJobReverseSlice []batchv1.CronJob

func (s SortCronJobReverseSlice) Len() int      { return len(s) }
func (s SortCronJobReverseSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortCronJobReverseSlice) Less(i, j int) bool {
	return s[i].GetCreationTimestamp().After(s[j].GetCreationTimestamp().Time)
}

func (s *Service) initCronJobTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) (*entity.CronJobTemplate, error) {
	podAnnotations, err := s.getWorkloadAnnotationsByVendor(ctx, task.ClusterName, task.EnvName, project, app, team)
	if err != nil {
		return nil, err
	}

	if task.Param.MetricsPort != 0 {
		podAnnotations[QingTingAnnotationMetricsPort] = strconv.Itoa(task.Param.MetricsPort)
	}

	configName, labels := "", s.getLabels(project, app, task, team)
	if task.Param.ConfigCommitID != "" {
		cm, e := s.GetCompatibleConfigMapDetail(ctx, project, app, task)
		if e != nil {
			return nil, e
		}
		configName = cm.Name

		hash := cm.GetLabels()[ConfigMapLabelHash]
		podAnnotations[AnnotationConfigHash] = hash
		labels[LabelConfigHash] = hash
	}

	envVars, err := s.getSystemEnv(project, app, task, team)
	if err != nil {
		return nil, err
	}

	for _, key := range utils.SortedMapKeys(task.Param.Vars) {
		envVars = append(envVars, &entity.EnvTemplate{
			Name:  key,
			Value: task.Param.Vars[key],
		})
	}

	template := &entity.CronJobTemplate{
		Namespace:                     task.Namespace,
		ProjectName:                   project.Name,
		AppName:                       app.Name,
		CronJobVersion:                task.Version,
		Labels:                        labels,
		Env:                           envVars,
		ImageName:                     entity.ImageName(task.Param.ImageVersion),
		Tolerations:                   task.Param.NodeSelector,
		ConfigName:                    configName,
		ConfigMountPath:               task.Param.ConfigMountPath,
		LogStoreName:                  project.LogStoreName,
		ContainerName:                 utils.GetPodContainerName(project.Name, app.Name),
		Schedule:                      task.Param.CronParam,
		CronCommand:                   task.Param.CronCommand,
		PreStopCommand:                task.Param.PreStopCommand,
		TerminationGracePeriodSeconds: task.Param.TerminationGracePeriodSeconds,
		RestartPolicy:                 task.Param.RestartPolicy,
		ConcurrencyPolicy:             task.Param.ConcurrencyPolicy,
		StartingDeadlineSeconds:       entity.DefaultStartingDeadlineSeconds,
		SuccessfulJobsHistoryLimit:    int32(task.Param.SuccessfulHistoryLimit),
		FailedJobsHistoryLimit:        int32(task.Param.FailedHistoryLimit),
		ActiveDeadlineSeconds:         task.Param.ActiveDeadlineSeconds,
		Suspend:                       false,
		CPULimit:                      task.Param.CPULimit,
		MemoryLimit:                   task.Param.MemLimit,
		CPURequest:                    task.Param.CPURequest,
		MemoryRequest:                 task.Param.MemRequest,
		PodAnnotations:                podAnnotations,
		NodeAffinity: entity.GenerateNodeAffinityTemplate(
			app.Type,
			task.Param.NodeAffinityLabelConfig,
		),
		BackoffLimit: task.Param.BackoffLimit,
	}

	return template, nil
}

// RenderCronJobTemplate : 渲染定时任务模板
func (s *Service) RenderCronJobTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) ([]byte, error) {
	tpl, err := s.initCronJobTemplate(ctx, project, app, task, team)
	if err != nil {
		return nil, err
	}

	data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

func (s *Service) decodeCronJobYamlData(_ context.Context, yamlData []byte) (*batchv1.CronJob, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	cronJob, ok := obj.(*batchv1.CronJob)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not batchv1.CronJob")
	}
	return cronJob, nil
}

// ApplyCronJob : 声明式创建/更新定时任务
func (s *Service) ApplyCronJob(ctx context.Context,
	clusterName entity.ClusterName, yamlData []byte, env string) (*batchv1.CronJob, error) {
	applyCronJob, err := s.decodeCronJobYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetCronJobDetail(ctx, clusterName,
		&req.GetCronJobDetailReq{
			Namespace: applyCronJob.GetNamespace(),
			Name:      applyCronJob.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		cronJob, e := s.CreateCronJob(ctx, clusterName, applyCronJob, env)
		if err != nil {
			return nil, e
		}
		return cronJob, nil
	}

	cronJob, err := s.PatchCronJob(ctx, clusterName, applyCronJob, env)
	if err != nil {
		return nil, err
	}
	return cronJob, nil
}

// CreateCronJob : 创建定时任务
func (s *Service) CreateCronJob(ctx context.Context,
	clusterName entity.ClusterName, cronJob *batchv1.CronJob, env string) (*batchv1.CronJob, error) {
	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.BatchV1().CronJobs(cronJob.GetNamespace()).
		Create(ctx, cronJob, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// GetCronJobDetail : 获取定时任务详情
func (s *Service) GetCronJobDetail(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetCronJobDetailReq) (*batchv1.CronJob, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// todo 这里等全部替换完成,可以删除
	if getReq.Namespace == "" {
		getReq.Namespace = getReq.Env
	}

	res, err := c.BatchV1().CronJobs(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) getCronJobsLabelSelector(getReq *req.GetCronJobsReq) string {
	labels := make([]string, 0)
	if getReq.ProjectName != "" {
		labels = append(labels, fmt.Sprintf("project=%s", getReq.ProjectName))
	}

	if getReq.AppName != "" {
		labels = append(labels, fmt.Sprintf("app=%s", getReq.AppName))
	}

	if getReq.Version != "" {
		labels = append(labels, fmt.Sprintf("version=%s", getReq.Version))
	}
	return strings.Join(labels, ",")
}

// GetCronJobs : 批量获取定时任务列表
func (s *Service) GetCronJobs(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetCronJobsReq) ([]batchv1.CronJob, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	selector := s.getCronJobsLabelSelector(getReq)

	list, err := c.BatchV1().CronJobs(getReq.Namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// 时间倒序
	res := SortCronJobReverseSlice(list.Items)
	sort.Sort(res)
	return res, nil
}

// PatchCronJob ： 更新定时任务
func (s *Service) PatchCronJob(ctx context.Context,
	clusterName entity.ClusterName, cronJob *batchv1.CronJob, env string) (*batchv1.CronJob, error) {
	patchData, err := json.Marshal(cronJob)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.BatchV1().CronJobs(cronJob.GetNamespace()).
		Patch(ctx, cronJob.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// SuspendCronJob : 暂停定时任务
func (s *Service) SuspendCronJob(ctx context.Context,
	clusterName entity.ClusterName, cronJob *batchv1.CronJob, env string) (*batchv1.CronJob, error) {
	suspend := true
	cronJob.Spec.Suspend = &suspend

	return s.PatchCronJob(ctx, clusterName, cronJob, env)
}

// ResumeCronJob : 恢复定时任务
func (s *Service) ResumeCronJob(ctx context.Context,
	clusterName entity.ClusterName, cronJob *batchv1.CronJob, env string) (*batchv1.CronJob, error) {
	suspend := false
	cronJob.Spec.Suspend = &suspend

	return s.PatchCronJob(ctx, clusterName, cronJob, env)
}

// DeleteCronJobs : 批量删除定时任务
func (s *Service) DeleteCronJobs(ctx context.Context,
	clusterName entity.ClusterName, deleteReq *req.DeleteCronJobsReq) error {
	res, err := s.GetCronJobs(ctx, clusterName,
		&req.GetCronJobsReq{
			Namespace:   deleteReq.Namespace,
			ProjectName: deleteReq.ProjectName,
			AppName:     deleteReq.AppName,
			Env:         deleteReq.Env,
		})
	if err != nil {
		return err
	}

	for i := range res {
		deploy := res[i]

		if deleteReq.InverseName != "" && deploy.GetName() == deleteReq.InverseName {
			continue
		}

		err = s.DeleteCronJob(ctx, clusterName,
			&req.DeleteCronJobReq{
				Namespace: deploy.GetNamespace(),
				Name:      deploy.GetName(),
				Env:       deleteReq.Env,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteCronJob : 删除定时任务
func (s *Service) DeleteCronJob(ctx context.Context,
	clusterName entity.ClusterName, deleteReq *req.DeleteCronJobReq) error {
	c, err := s.GetK8sTypedClient(clusterName, deleteReq.Env)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// todo 这里在全部替换完以后可以删除
	if deleteReq.Namespace == "" {
		deleteReq.Namespace = deleteReq.Env
	}

	policy := metav1.DeletePropagationForeground
	err = c.BatchV1().CronJobs(deleteReq.Namespace).
		Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// DescribeCronJob : 获取定时任务的 Describe 信息
func (s *Service) DescribeCronJob(ctx context.Context, clusterName entity.ClusterName,
	descReq *req.DescribeCronJobReq) (*resp.DescribeCronJobResp, error) {
	cronjob, err := s.GetCronJobDetail(ctx, clusterName, &req.GetCronJobDetailReq{
		Namespace: descReq.Namespace,
		Name:      descReq.Name,
		Env:       descReq.Env,
	})
	if err != nil {
		return nil, err
	}

	res := new(resp.DescribeCronJobResp)
	res.Status.Name = cronjob.Name
	res.Status.LastScheduleTime = utils.FormatK8sTime(cronjob.Status.LastScheduleTime)

	cronjobEvents, err := s.GetK8sResourceEvents(ctx, clusterName,
		&req.GetK8sResourceEventsReq{
			Namespace: descReq.Namespace,
			Resource:  cronjob,
		})
	if err != nil {
		return nil, err
	}
	res.Events = cronjobEvents.Events

	return res, nil
}

// GetJobSpecFromCronJob 基于cronjob创建job
func (s *Service) GetJobSpecFromCronJob(task *resp.TaskDetailResp, cronJob *batchv1.CronJob) (*batchv1.Job, error) {
	// spec注释生成
	annotations := make(map[string]string)
	for k, v := range cronJob.Spec.JobTemplate.Annotations {
		annotations[k] = v
	}
	annotations[entity.AnnoatationCronJobInstantiateType] = string(entity.LaunchTypeManual)

	// spec标签生成
	labels := cronJob.Spec.JobTemplate.Labels
	labels[entity.LabelKeyLaunchType] = string(entity.LaunchTypeManual)

	// job结构体构建
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: batchv1.SchemeGroupVersion.String(),
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   task.Namespace,
			Name:        task.Param.ManualJobName,
			Annotations: annotations,
			Labels:      labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(cronJob, batchv1.SchemeGroupVersion.WithKind("CronJob")),
			},
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}
	return job, nil
}

func (s *Service) getCronjobNextSchedTime(schedule string, lastSchedTime time.Time) (string, error) {
	sched, err := cron.ParseStandard(schedule)
	if err != nil {
		return "", errors.Wrapf(errcode.InvalidParams,
			"unparseable schedule: %s : %s", schedule, err)
	}

	return sched.Next(lastSchedTime).Format(utils.DefaultTimeFormatLayout), nil
}
