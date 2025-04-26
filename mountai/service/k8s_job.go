package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/batch/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
)

// SortJobReverseSlice : 任务的创建时间倒序数组
type SortJobReverseSlice []v1.Job

func (s SortJobReverseSlice) Len() int      { return len(s) }
func (s SortJobReverseSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortJobReverseSlice) Less(i, j int) bool {
	return s[i].GetCreationTimestamp().After(s[j].GetCreationTimestamp().Time)
}

// GetJobDetail : 获取一次性任务详情
func (s *Service) GetJobDetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetJobDetailReq) (*v1.Job, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.BatchV1().Jobs(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) getJobLabelSelector(getReq *req.GetJobsReq) string {
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

// GetJobs : 批量获取一次性任务列表
func (s *Service) GetJobs(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetJobsReq) ([]v1.Job, error) {
	selector := s.getJobLabelSelector(getReq)

	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.BatchV1().Jobs(getReq.Namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})

	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// 时间倒序
	res := SortJobReverseSlice(list.Items)
	sort.Sort(res)
	return res, nil
}

// RenderJobTemplate : 渲染一次性任务模板
func (s *Service) RenderJobTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) ([]byte, error) {
	tpl, err := s.initJobTemplate(ctx, project, app, task, team)
	if err != nil {
		return nil, err
	}

	data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// ApplyJob : 声明式创建/更新一次性任务
func (s *Service) ApplyJob(ctx context.Context, clusterName entity.ClusterName,
	yamlData []byte, env string) (*v1.Job, error) {
	applyJob, err := s.decodeJobYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetJobDetail(ctx, clusterName,
		&req.GetJobDetailReq{
			Namespace: applyJob.GetNamespace(),
			Name:      applyJob.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		job, e := s.CreateJob(ctx, clusterName, applyJob, env)
		if err != nil {
			return nil, e
		}
		return job, nil
	}

	job, err := s.PatchJob(ctx, clusterName, applyJob, env)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// PatchJob : 更新一次性任务
func (s *Service) PatchJob(ctx context.Context, clusterName entity.ClusterName,
	job *v1.Job, env string) (*v1.Job, error) {
	patchData, err := json.Marshal(job)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.BatchV1().Jobs(job.GetNamespace()).
		Patch(ctx, job.GetName(), types.MergePatchType, patchData, metav1.PatchOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

// CreateJob : 创建一次性任务
func (s *Service) CreateJob(ctx context.Context, clusterName entity.ClusterName,
	job *v1.Job, env string) (*v1.Job, error) {
	c, err := s.GetK8sTypedClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.BatchV1().Jobs(job.GetNamespace()).
		Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) decodeJobYamlData(_ context.Context, yamlData []byte) (*v1.Job, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}
	job, ok := obj.(*v1.Job)
	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not v1.Job")
	}
	return job, nil
}

func (s *Service) initJobTemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) (*entity.JobTemplate, error) {
	podAnnotations, err := s.getWorkloadAnnotationsByVendor(ctx, task.ClusterName, task.EnvName, project, app, team)
	if err != nil {
		return nil, err
	}

	var configName string
	if task.Param.ConfigCommitID != "" {
		configName = s.GetAppNewConfigMapName(task)
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

	if task.Param.MetricsPort != 0 {
		podAnnotations[QingTingAnnotationMetricsPort] = strconv.Itoa(task.Param.MetricsPort)
	}

	template := &entity.JobTemplate{
		Namespace:                     task.Namespace,
		ProjectName:                   project.Name,
		AppName:                       app.Name,
		Labels:                        s.getLabels(project, app, task, team),
		JobVersion:                    task.Version,
		Env:                           envVars,
		ImageName:                     entity.ImageName(task.Param.ImageVersion),
		Tolerations:                   task.Param.NodeSelector,
		ConfigName:                    configName,
		ConfigMountPath:               task.Param.ConfigMountPath,
		LogStoreName:                  project.LogStoreName,
		ContainerName:                 utils.GetPodContainerName(project.Name, app.Name),
		JobCommand:                    task.Param.JobCommand,
		ActiveDeadlineSeconds:         task.Param.ActiveDeadlineSeconds,
		TerminationGracePeriodSeconds: task.Param.TerminationGracePeriodSeconds,
		BackoffLimit:                  task.Param.BackoffLimit,
		CPULimit:                      task.Param.CPULimit,
		MemoryLimit:                   task.Param.MemLimit,
		CPURequest:                    task.Param.CPURequest,
		MemoryRequest:                 task.Param.MemRequest,
		PodAnnotations:                podAnnotations,
		NodeAffinity: entity.GenerateNodeAffinityTemplate(
			app.Type,
			task.Param.NodeAffinityLabelConfig,
		),
	}

	return template, nil
}

// DeleteJobs : 批量删除一次性任务
func (s *Service) DeleteJobs(ctx context.Context, clusterName entity.ClusterName,
	deleteReq *req.DeleteJobsReq) error {
	res, err := s.GetJobs(ctx, clusterName,
		&req.GetJobsReq{
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

		err = s.DeleteJob(ctx, clusterName,
			&req.DeleteJobReq{
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

// DeleteJob : 删除一次性任务
func (s *Service) DeleteJob(ctx context.Context, clusterName entity.ClusterName,
	deleteReq *req.DeleteJobReq) error {
	policy := metav1.DeletePropagationForeground
	c, err := s.GetK8sTypedClient(clusterName, deleteReq.Env)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	if deleteReq.Namespace == "" {
		deleteReq.Namespace = deleteReq.Env
	}

	err = c.BatchV1().Jobs(deleteReq.Namespace).
		Delete(ctx, deleteReq.Name, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		})
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// DescribeJob : 获取一次性任务的 Describe 信息
func (s *Service) DescribeJob(ctx context.Context, clusterName entity.ClusterName,
	descReq *req.DescribeJobReq) (*resp.DescribeJobResp, error) {
	res := new(resp.DescribeJobResp)
	job, err := s.GetJobDetail(ctx, clusterName, &req.GetJobDetailReq{
		Namespace: descReq.Namespace,
		Name:      descReq.Name,
	})
	if err != nil {
		return nil, err
	}

	jobEvents, err := s.GetK8sResourceEvents(ctx, clusterName,
		&req.GetK8sResourceEventsReq{
			Namespace: descReq.Namespace,
			Resource:  job,
		})
	if err != nil {
		return nil, err
	}

	res.Events = jobEvents.Events

	err = deepcopy.Copy(job).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	res.Status.StartTime = utils.FormatK8sTime(job.Status.StartTime)

	conditions := make([]resp.JobCondition, len(job.Status.Conditions))
	for i, condition := range res.Status.Conditions {
		condition.LastProbeTime = utils.FormatK8sTime(&job.Status.Conditions[i].LastProbeTime)
		condition.LastTransitionTime = utils.FormatK8sTime(&job.Status.Conditions[i].LastTransitionTime)
		conditions[i] = condition
	}
	res.Status.Conditions = conditions

	return res, nil
}
