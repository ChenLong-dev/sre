package service

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/sentry"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
)

// expectedImagePrefixes 预期的镜像前缀列表
var expectedImagePrefixes = []string{
	"registry.cn-shanghai.aliyuncs.com",
	"crpi-592g7buyguepbrqd.cn-shanghai.personal.cr.aliyuncs.com",
	"crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com",
	"swr.cn-east-3.myhuaweicloud.com", // 华为云镜像仓库华东-上海区

}

// SortPodReverseSlice pod的创建时间倒序数组
type SortPodReverseSlice []v1.Pod

func (s SortPodReverseSlice) Len() int      { return len(s) }
func (s SortPodReverseSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortPodReverseSlice) Less(i, j int) bool {
	return s[i].GetCreationTimestamp().After(s[j].GetCreationTimestamp().Time)
}

// GetPodDetail 获取 Pod 详情
func (s *Service) GetPodDetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetPodDetailReq) (*v1.Pod, error) {
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.CoreV1().Pods(getReq.Namespace).
		Get(ctx, getReq.Name, metav1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return res, nil
}

func (s *Service) getPodsLabelSelector(getReq *req.GetPodsReq) string {
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

	if getReq.JobName != "" {
		labels = append(labels, fmt.Sprintf("job-name=%s", getReq.JobName))
	}
	return strings.Join(labels, ",")
}

// GetPods 获取 Pod 列表
func (s *Service) GetPods(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetPodsReq) ([]v1.Pod, error) {
	selector := s.getPodsLabelSelector(getReq)

	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.CoreV1().Pods(getReq.Namespace).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// 时间倒序
	res := SortPodReverseSlice(list.Items)
	sort.Sort(res)
	return res, nil
}

// GetPodLog 获取 Pod 日志
func (s *Service) GetPodLog(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetPodLogReq) (string, error) {
	var maxTailLines int64 = 300
	c, err := s.GetK8sTypedClient(clusterName, getReq.Env)
	if err != nil {
		return "", errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.CoreV1().Pods(getReq.Namespace).
		GetLogs(getReq.Name, &v1.PodLogOptions{
			Timestamps: true,
			TailLines:  &maxTailLines,
			Container:  getReq.ContainerName,
		}).
		DoRaw(ctx)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return "", errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return "", errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return string(res), nil
}

// ExecPodCommand 在 Pod 中执行指令
func (s *Service) ExecPodCommand(ctx context.Context, clusterName entity.ClusterName,
	execReq *req.ExecPodReq) ([]byte, error) {
	k8sCfg, err := s.GetClusterConfig(ctx, clusterName, execReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	c, err := s.GetK8sTypedClient(clusterName, execReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	exec, err := remotecommand.NewSPDYExecutor(
		k8sCfg,
		"POST",
		c.CoreV1().
			RESTClient().
			Post().
			Resource("pods").
			Name(execReq.Name).
			Namespace(execReq.Namespace).
			SubResource("exec").
			VersionedParams(
				&v1.PodExecOptions{
					Stdin:     false,
					Stdout:    true,
					Stderr:    true,
					TTY:       false,
					Command:   execReq.Commands,
					Container: execReq.Container,
				},
				scheme.ParameterCodec,
			).
			URL(),
	)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	stdOut := new(bytes.Buffer)
	stdErr := new(bytes.Buffer)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: stdOut,
		Stderr: stdErr,
		Tty:    false,
	})
	if err != nil {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "error:%s detail:%v", err.Error(), stdErr.String())
	}

	return stdOut.Bytes(), nil
}

// DescribePod 获取 Pod describe 信息
func (s *Service) DescribePod(ctx context.Context, clusterName entity.ClusterName,
	descPodReq *req.GetRunningPodDescriptionReq) (*resp.DescribePodResp, error) {
	res := new(resp.DescribePodResp)
	pod, err := s.GetPodDetail(ctx, clusterName,
		&req.GetPodDetailReq{
			Name:      descPodReq.Name,
			Namespace: descPodReq.Namespace,
			Env:       descPodReq.Env,
		})
	if err != nil {
		return nil, err
	}

	err = deepcopy.Copy(pod).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	conditions := make([]resp.PodCondition, len(pod.Status.Conditions))
	for i, condition := range res.Status.Conditions {
		condition.LastProbeTime = utils.FormatK8sTime(&pod.Status.Conditions[i].LastProbeTime)
		condition.LastTransitionTime = utils.FormatK8sTime(&pod.Status.Conditions[i].LastTransitionTime)
		conditions[i] = condition
	}
	res.Status.Conditions = conditions

	podEvents, err := s.GetK8sResourceEvents(ctx, clusterName,
		&req.GetK8sResourceEventsReq{
			Namespace: descPodReq.Namespace,
			Resource:  pod,
			Env:       descPodReq.Env,
		})
	if err != nil {
		return nil, err
	}
	res.Events = podEvents.Events

	return res, nil
}

// GetPodsLatestReadyTime 获取pod全部初始化完毕时间
func (s *Service) GetPodsLatestReadyTime(ctx context.Context,
	clusterName entity.ClusterName, getReq *req.GetPodsReq) (time.Time, error) {
	pods, err := s.GetPods(ctx, clusterName, getReq)
	if err != nil {
		return time.Time{}, err
	}

	var latestTime time.Time
	for i := range pods {
		pod := pods[i]

		// 获取当前pod的就绪时间
		var curReadyTime time.Time
		for j := range pod.Status.Conditions {
			condition := pod.Status.Conditions[j]
			if condition.Type != v1.PodReady || condition.Status != v1.ConditionTrue {
				continue
			}
			curReadyTime = condition.LastTransitionTime.Time
		}
		if curReadyTime.IsZero() {
			return curReadyTime, errors.Wrapf(_errcode.K8sResourceNotReadyError, "pod not ready: %s", pod.GetName())
		}
		// 更新最后时间
		if curReadyTime.After(latestTime) {
			latestTime = curReadyTime
		}
	}

	if latestTime.IsZero() {
		return latestTime, errors.Wrap(_errcode.K8sResourceNotReadyError, "no pods is ready")
	}

	return latestTime, nil
}

// WatchClusterPods 监听集群所有 pod
func (s *Service) WatchClusterPods(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName) error {
	return until(ctx, func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				if err == nil {
					err = errors.Wrapf(errcode.InternalError, "WatchClusterPods recovered from panic(%#v)", r)
				} else {
					err = errors.Wrapf(errcode.InternalError, "WatchClusterPods recovered from panic(%#v) with err(%s)", r, err)
				}

				return
			}
		}()

		// 获取所有 namespace 下的 pod 监听
		w, e := s.getPodsWatchInterface(ctx, clusterName, envName, "")
		if e != nil {
			return e
		}

		for {
			select {
			case event, ok := <-w.ResultChan():
				if !ok { // 监听正常中断或结束
					return nil
				}

				switch event.Type {
				case watch.Added, watch.Modified: // pod 新增或变更时均重新检测所有镜像, 更新对应 pod 的镜像情况
					pod, isPod := event.Object.(*v1.Pod)
					if !isPod {
						log.Errorc(ctx, "event(%s) Object(%#v) is not *v1.Pod in cluster(%s)", event.Type, event.Object, clusterName)
						break
					}

					records := s.filterUnexpectedImageRecords(clusterName, pod)
					for _, record := range records {
						if len(record.ImageList) == 0 { // 不存在非预期镜像时删除该 pod 记录
							e = s.dao.DeleteUnexpectedImageRecord(ctx, record)
							if e != nil {
								log.Errorc(ctx, "DeleteUnexpectedImageRecord failed: %s", e)
								sentry.CaptureWithBreadAndTags(ctx, e,
									&sentry.Breadcrumb{
										Category: "DeleteUnexpectedImageRecord",
										Data: map[string]interface{}{
											"Cluster":            record.Cluster,
											"Namespace":          record.Namespace,
											"AMSProjectName":     record.AMSProjectName,
											"AMSAppName":         record.AMSAppName,
											"OwnerReferenceKind": record.OwnerReferenceKind,
											"OwnerReferenceName": record.OwnerReferenceName,
											"PodName":            record.PodName,
										},
									},
								)
							}

							continue
						}

						e = s.dao.UpsertUnexpectedImageRecord(ctx, record)
						if e != nil {
							log.Errorc(ctx, "UpsertUnexpectedImageRecord(cluster=%s) failed: %s", clusterName, e)
							sentry.CaptureWithBreadAndTags(ctx, e,
								&sentry.Breadcrumb{
									Category: "UpsertUnexpectedImageRecord",
									Data: map[string]interface{}{
										"Cluster":   clusterName,
										"Namespace": record.Namespace,
										"PodName":   record.PodName,
										"ImageList": record.ImageList,
									},
								},
							)
						}
					}

				case watch.Deleted: // pod 删除时关联删除其非预期镜像列表记录
					pod, isPod := event.Object.(*v1.Pod)
					if !isPod {
						log.Errorc(ctx, "event(%s) Object(%#v) is not *v1.Pod in cluster(%s)", event.Type, event.Object, clusterName)
						break
					}

					records := s.filterUnexpectedImageRecords(clusterName, pod)
					for _, record := range records {
						e = s.dao.DeleteUnexpectedImageRecord(ctx, record)
						if e != nil {
							log.Errorc(ctx, "DeleteUnexpectedImageRecord(cluster=%s) failed: %s", clusterName, e)
							sentry.CaptureWithBreadAndTags(ctx, e,
								&sentry.Breadcrumb{
									Category: "DeleteUnexpectedImageRecord",
									Data: map[string]interface{}{
										"Cluster":            record.Cluster,
										"Namespace":          record.Namespace,
										"AMSProjectName":     record.AMSProjectName,
										"AMSAppName":         record.AMSAppName,
										"OwnerReferenceKind": record.OwnerReferenceKind,
										"OwnerReferenceName": record.OwnerReferenceName,
										"PodName":            record.PodName,
									},
								},
							)
						}
					}

				case watch.Bookmark:
					log.Infoc(ctx, "ignored pod %s event in cluster(%s)", event.Type, clusterName)

				case watch.Error:
					log.Errorc(ctx, "ignored pod %s event: %#v in cluster(%s)", event.Type, event.Object, clusterName)

				default: // 当前没有其他类型, 但防止未来版本升级出现新增类型, 发送 sentry 报警
					log.Errorc(ctx, "unexpected event_type(%s) in cluster(%s)", event.Type, clusterName)
					sentry.CaptureWithBreadAndTags(
						ctx,
						errors.Wrapf(errcode.InternalError, "unexpected event_type(%s)", event.Type),
						&sentry.Breadcrumb{
							Category: "WatchClusterPods-event",
							Data: map[string]interface{}{
								"Cluster": clusterName,
								"Type":    event.Type,
								"Object":  event.Object,
							},
						})
				}

			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
}

// getPodsWatchInterface 获取pod监听
func (s *Service) getPodsWatchInterface(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, namespace string) (watch.Interface, error) {
	c, err := s.GetK8sTypedClient(clusterName, string(envName))
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	// 不区分 namespace
	return c.CoreV1().Pods(namespace).Watch(ctx, metav1.ListOptions{Watch: true})
}

// IsPodReadyConditionTrue return True if pod is ready
func IsPodReadyConditionTrue(status *v1.PodStatus) bool {
	_, condition := GetPodCondition(status, v1.PodReady)
	return condition != nil && condition.Status == v1.ConditionTrue
}

// GetPodCondition return condition && index
func GetPodCondition(status *v1.PodStatus, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if status == nil {
		return -1, nil
	}
	return GetPodConditionFromList(status.Conditions, conditionType)
}

// GetPodConditionFromList return pod condition from pod conditions
func GetPodConditionFromList(conditions []v1.PodCondition, conditionType v1.PodConditionType) (int, *v1.PodCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

// GetServiceNameFromVersion extracts service name from version name without pub date
func (s *Service) GetServiceNameFromVersion(version string) string {
	if version == "" {
		return ""
	}

	index := strings.LastIndex(version, "-")
	if index > 0 {
		if version[index:] == entity.TaskCanaryVersionSuffix {
			return version[:index]
		}

		if _, err := strconv.ParseFloat(version[index:], 64); err == nil {
			return version[:index]
		}
	}

	return version
}

func (s *Service) filterUnexpectedImageRecords(clusterName entity.ClusterName, pod *v1.Pod) []*entity.UnexpectedImageRecord {
	var (
		expected bool
		prefix   string
		records  []*entity.UnexpectedImageRecord
	)

	idx := 0
	imageList := make(
		[]*entity.UnexpectedImageInfo,
		len(pod.Spec.InitContainers)+len(pod.Spec.Containers)+len(pod.Spec.EphemeralContainers),
	)

	for i := range pod.Spec.InitContainers {
		expected = false
		for _, prefix = range expectedImagePrefixes {
			if strings.HasPrefix(pod.Spec.InitContainers[i].Image, prefix) {
				expected = true
				break
			}
		}

		if !expected {
			imageList[idx] = &entity.UnexpectedImageInfo{
				ContainerName: pod.Spec.InitContainers[i].Name,
				ContainerType: entity.ContainerTypeInitContainer,
				Image:         pod.Spec.InitContainers[i].Image,
			}
			idx++
		}
	}

	for i := range pod.Spec.Containers {
		expected = false
		for _, prefix = range expectedImagePrefixes {
			if strings.HasPrefix(pod.Spec.Containers[i].Image, prefix) {
				expected = true
				break
			}
		}

		if !expected {
			imageList[idx] = &entity.UnexpectedImageInfo{
				ContainerName: pod.Spec.Containers[i].Name,
				ContainerType: entity.ContainerTypeNormalContainer,
				Image:         pod.Spec.Containers[i].Image,
			}
			idx++
		}
	}

	for i := range pod.Spec.EphemeralContainers {
		expected = false
		for _, prefix = range expectedImagePrefixes {
			if strings.HasPrefix(pod.Spec.EphemeralContainers[i].Image, prefix) {
				expected = true
				break
			}
		}

		if !expected {
			imageList[idx] = &entity.UnexpectedImageInfo{
				ContainerName: pod.Spec.EphemeralContainers[i].Name,
				ContainerType: entity.ContainerTypeEphemeralContainer,
				Image:         pod.Spec.EphemeralContainers[i].Image,
			}
			idx++
		}
	}

	imageList = imageList[:idx]

	if pod.Labels["project"] != "" {
		records = []*entity.UnexpectedImageRecord{
			{
				AMSProjectName: pod.Labels["project"],
				AMSAppName:     pod.Labels["app"],
			},
		}
	} else if len(pod.OwnerReferences) > 0 {
		records = make([]*entity.UnexpectedImageRecord, len(pod.OwnerReferences))
		for i := range records {
			records[i] = &entity.UnexpectedImageRecord{
				OwnerReferenceKind: pod.OwnerReferences[i].Kind,
				OwnerReferenceName: pod.OwnerReferences[i].Name,
			}
		}
	} else {
		records = []*entity.UnexpectedImageRecord{
			{
				PodName: pod.Name,
			},
		}
	}

	for i := range records {
		records[i].Cluster = clusterName
		records[i].Namespace = pod.Namespace
		records[i].ImageList = imageList
	}

	return records
}

func until(ctx context.Context, f func() error) error {
	var (
		err    error
		period time.Duration
	)

	for {
		err = f()
		// 执行正常终止时重试立即进行, 否则采取指数退避重试策略
		if err == nil {
			period = 0
		} else if period == 0 {
			period = 1
		} else if period <<= 1; period > 15 {
			period = 15
		}

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(period * time.Second):
			if period > 0 {
				log.Infoc(ctx, "retrying because of error(%s) after %d seconds", err, period)
			} else {
				log.Infoc(ctx, "recovered from watch disconnection")
			}
		}
	}
}
