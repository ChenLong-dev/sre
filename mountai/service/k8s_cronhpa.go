package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AliyunContainerService/kubernetes-cronhpa-controller/pkg/apis/autoscaling/v1beta1"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	"gitlab.shanhai.int/sre/library/net/errcode"
)

var (
	CronHPAGVR = schema.GroupVersionResource{
		Group:    "autoscaling.alibabacloud.com",
		Version:  "v1beta1",
		Resource: "cronhorizontalpodautoscalers",
	}
)

const (
	// MinInterval 每日最小调度区间间隔和 MaxInterval 每日最大调度区间间隔。
	// cronhpa控制的hpa对pod数量的调整是一个渐进的过程，高频调整pod下限会
	// 造成pod数量不稳定
	MinInterval = 30 * time.Minute
	MaxInterval = 24 * time.Hour
)

// RenderCronHPATemplate 渲染 CronHPA 模板
func (s *Service) RenderCronHPATemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, team *resp.TeamDetailResp) ([]byte, error) {
	tpl, err := s.initCronHPATemplate(ctx, project, app, task, team)
	if err != nil {
		return nil, err
	}

	data, err := s.RenderK8sTemplate(ctx, entity.DefaultTemplateFileDir, task.ClusterName, task.EnvName, tpl)
	if err != nil {
		return nil, err
	}

	return []byte(data), nil
}

// initCronHPATemplate 初始化cronHPA模板
func (s *Service) initCronHPATemplate(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, _ *resp.TeamDetailResp) (*entity.CronHPATemplate, error) {
	_, ok := app.Env[task.EnvName]
	if !ok {
		return nil, errors.Wrapf(errcode.InternalError, "couldn't find env:%s", task.EnvName)
	}

	// 当前 AMS 的 CronHPA 控制目标是 HPA
	scaleTargetRefKind := entity.K8sObjectKindHPA
	scaleTargetRefGroupVersion, err := s.getK8sResourceGroupVersion(ctx, task.ClusterName, task.EnvName, scaleTargetRefKind)
	if err != nil {
		return nil, errors.Wrapf(errcode.InternalError, "get apiVersion for %s(cluster=%s, env=%s) failed: %s",
			scaleTargetRefKind, task.ClusterName, task.EnvName, err)
	}

	template := &entity.CronHPATemplate{
		Name:        task.Version,
		Namespace:   task.Namespace,
		ProjectName: project.Name,
		AppName:     app.Name,
		ScaleTargetRef: entity.ScaleTargetRefTemplate{
			Kind:       scaleTargetRefKind,
			Name:       task.Version,
			APIVersion: scaleTargetRefGroupVersion.String(),
		},
		Labels:                   map[string]string{"cluster": string(task.ClusterName)},
		CronScaleJobs:            []*entity.CronScaleJob{},
		CronScaleJobExcludeDates: []string{},
	}

	// 初始化跳过日期
	for _, date := range task.Param.CronScaleJobExcludeDates {
		template.CronScaleJobExcludeDates = append(template.CronScaleJobExcludeDates, "* "+date)
	}

	// 初始化任务
	pairs := buildCronHPAScaleJobPairs(task.Param.CronScaleJobGroups, task.Param.MinPodCount)
	for _, pair := range pairs {
		template.CronScaleJobs = append(template.CronScaleJobs,
			pair.ScaleUp,
			pair.ScaleDown)
	}

	return template, nil
}

// ApplyCronHPA 声明式创建/更新 CronHPA
func (s *Service) ApplyCronHPA(ctx context.Context, clusterName entity.ClusterName,
	yamlData []byte, env string) (*v1beta1.CronHorizontalPodAutoscaler, error) {
	applyCronHPA, err := s.decodeCronHPAYamlData(ctx, yamlData)
	if err != nil {
		return nil, err
	}

	_, err = s.GetCronHPADetail(ctx, clusterName,
		&req.GetCronHPADetailReq{
			Namespace: applyCronHPA.GetNamespace(),
			Name:      applyCronHPA.GetName(),
			Env:       env,
		})
	if err != nil {
		if !errcode.EqualError(_errcode.K8sResourceNotFoundError, err) {
			return nil, err
		}

		return s.CreateCronHPA(ctx, clusterName, env, applyCronHPA)
	}

	return s.PatchCronHPA(ctx, clusterName, env, applyCronHPA)
}

// decodeCronHPAYamlData 解析cronHPA配置
func (s *Service) decodeCronHPAYamlData(_ context.Context, yamlData []byte) (*v1beta1.CronHorizontalPodAutoscaler, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(yamlData, nil, nil)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}
	cronHPA, ok := obj.(*v1beta1.CronHorizontalPodAutoscaler)

	if !ok {
		return nil, errors.Wrapf(_errcode.K8sInternalError, "object is not v2beta2.HorizontalPodAutoscaler")
	}
	return cronHPA, nil
}

// GetCronHPADetail 获取 HPA 详情
func (s *Service) GetCronHPADetail(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetCronHPADetailReq) (*v1beta1.CronHorizontalPodAutoscaler, error) {
	c, err := s.GetK8sDynamicClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	obj, err := c.Resource(CronHPAGVR).Namespace(getReq.Namespace).Get(ctx, getReq.Name, v1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil, errors.Wrap(_errcode.K8sResourceNotFoundError, err.Error())
		}
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	cronHPA, err := utils.ConvertUnstructToCronHPA(obj)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return cronHPA, nil
}

// CreateCronHPA 创建 CronHPA
func (s *Service) CreateCronHPA(ctx context.Context, clusterName entity.ClusterName, envName string,
	cronHPA *v1beta1.CronHorizontalPodAutoscaler) (*v1beta1.CronHorizontalPodAutoscaler, error) {
	c, err := s.GetK8sDynamicClient(clusterName, envName)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	u, err := utils.ConvertCronHPAToUnstruct(cronHPA)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.Resource(CronHPAGVR).Namespace(cronHPA.Namespace).Create(ctx, u, v1.CreateOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	ret, err := utils.ConvertUnstructToCronHPA(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return ret, nil
}

// PatchCronHPA 更新 CronHPA
func (s *Service) PatchCronHPA(ctx context.Context, clusterName entity.ClusterName, env string,
	cronHPA *v1beta1.CronHorizontalPodAutoscaler) (*v1beta1.CronHorizontalPodAutoscaler, error) {
	patchData, err := json.Marshal(cronHPA)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	c, err := s.GetK8sDynamicClient(clusterName, env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	res, err := c.Resource(CronHPAGVR).
		Namespace(cronHPA.Namespace).
		Patch(ctx, cronHPA.Name, types.MergePatchType, patchData, v1.PatchOptions{})
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	ret, err := utils.ConvertUnstructToCronHPA(res)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return ret, nil
}

// DeleteCronHPAs 批量删除 CronHPA
func (s *Service) DeleteCronHPAs(ctx context.Context, clusterName entity.ClusterName,
	deleteReq *req.DeleteCronHPAsReq) error {
	res, err := s.GetCronHPAs(ctx, clusterName,
		&req.GetCronHPAsReq{
			Namespace:   deleteReq.Namespace,
			ProjectName: deleteReq.ProjectName,
			AppName:     deleteReq.AppName,
			Env:         deleteReq.Env,
		})
	if err != nil {
		return err
	}

	for i := range res {
		cronHPA := res[i]

		if deleteReq.InverseVersion != "" && cronHPA.GetName() == deleteReq.InverseVersion {
			continue
		}

		err = s.DeleteCronHPA(ctx, clusterName, deleteReq.Env,
			&req.DeleteCronHPAReq{
				Namespace: cronHPA.GetNamespace(),
				Name:      cronHPA.GetName(),
			})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCronHPAs 获取 CronHPA 列表
func (s *Service) GetCronHPAs(ctx context.Context, clusterName entity.ClusterName,
	getReq *req.GetCronHPAsReq) ([]*v1beta1.CronHorizontalPodAutoscaler, error) {
	selector := s.getCronHPAsLabelSelector(getReq)

	c, err := s.GetK8sDynamicClient(clusterName, getReq.Env)
	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	list, err := c.Resource(CronHPAGVR).Namespace(getReq.Namespace).List(ctx, v1.ListOptions{
		LabelSelector: selector})

	if err != nil {
		return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	ret := make([]*v1beta1.CronHorizontalPodAutoscaler, 0)
	for i := range list.Items {
		r, e := utils.ConvertUnstructToCronHPA(&list.Items[i])
		if e != nil {
			return nil, errors.Wrap(_errcode.K8sInternalError, err.Error())
		}
		ret = append(ret, r)
	}
	return ret, nil
}

func (s *Service) getCronHPAsLabelSelector(getReq *req.GetCronHPAsReq) string {
	labels := make([]string, 0)
	if getReq.ProjectName != "" {
		labels = append(labels, fmt.Sprintf("project=%s", getReq.ProjectName))
	}
	if getReq.Version != "" {
		labels = append(labels, fmt.Sprintf("version=%s", getReq.Version))
	}
	if getReq.AppName != "" {
		labels = append(labels, fmt.Sprintf("app=%s", getReq.AppName))
	}
	return strings.Join(labels, ",")
}

// DeleteCronHPA 删除 CronHPA
func (s *Service) DeleteCronHPA(ctx context.Context,
	clusterName entity.ClusterName, envName string, deleteReq *req.DeleteCronHPAReq) error {
	policy := v1.DeletePropagationForeground

	c, err := s.GetK8sDynamicClient(clusterName, envName)
	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	err = c.Resource(CronHPAGVR).Namespace(deleteReq.Namespace).Delete(ctx,
		deleteReq.Name,
		v1.DeleteOptions{
			PropagationPolicy: &policy,
		})

	if err != nil {
		return errors.Wrap(_errcode.K8sInternalError, err.Error())
	}

	return nil
}

// ValidateCronHPA 校验 CronHPA
func (s *Service) ValidateCronHPA(param *req.CreateTaskParamReq, minInterval, maxInterval time.Duration) error {
	var (
		groups       = param.CronScaleJobGroups
		excludeDates = param.CronScaleJobExcludeDates
		minSize      = param.MinPodCount
		maxSize      = param.MaxPodCount
		// 全年每日执行
		everyDay []*entity.CronScaleJobGroup
		// 特定月份每日执行
		everyDayInCertainMonth []*entity.CronScaleJobGroup
		// 一周内执行一次
		certainWeek []*entity.CronScaleJobGroup
		// 单次执行
		onceDay []*entity.CronScaleJobGroup
		err     error
	)

	// 重名校验
	err = checkDuplicateCronHPAName(groups)
	if err != nil {
		return err
	}

	// 任务分组
	everyDay, everyDayInCertainMonth, certainWeek, onceDay, err = groupCronAutoScaleSchedules(groups, minSize, maxSize)
	if err != nil {
		return err
	}

	// 校验每日执行任务
	err = checkCronHPAEverydayJobs(everyDay)
	if err != nil {
		return err
	}

	// 校验特定月份每日执行任务
	err = checkCronHPAEveryDayInCertainMonthJobs(everyDayInCertainMonth, everyDay,
		minInterval, maxInterval, excludeDates)
	if err != nil {
		return err
	}

	// 校验星期任务
	err = checkCronHPACertainWeekJobs(certainWeek, everyDayInCertainMonth, everyDay)
	if err != nil {
		return err
	}

	// 校验一次性任务
	err = checkCronHPAOnceDayJobs(onceDay, certainWeek, everyDayInCertainMonth, everyDay)
	if err != nil {
		return err
	}

	return nil
}

// checkDuplicateCronHPAName 校验同名任务组
func checkDuplicateCronHPAName(groups []*entity.CronScaleJobGroup) error {
	m := make(map[string]struct{})
	for _, g := range groups {
		_, ok := m[g.Name]
		if ok {
			return errors.Wrap(errcode.InvalidParams, "groups sharing the same name: "+g.Name)
		}
		m[g.Name] = struct{}{}
	}
	return nil
}

func groupCronAutoScaleSchedules(groups []*entity.CronScaleJobGroup, minSize, maxSize int) (
	everyDay, everyDayInCertainMonth, certainWeek, onceDay []*entity.CronScaleJobGroup, err error) {
	for _, group := range groups {
		// 目标值大于HPA上限
		if group.TargetSize > maxSize {
			return nil, nil, nil, nil, errors.Wrap(
				errcode.InvalidParams,
				"the target size of the cronHPA greater than the max size of the HPA: "+group.Name)
		}
		// 目标值小于HPA下限
		if group.TargetSize < minSize {
			return nil, nil, nil, nil, errors.Wrap(
				errcode.InvalidParams,
				"the target size of the cronHPA less than the min size of the HPA: "+group.Name)
		}
		switch utils.TypeOfCronHPAScheduleGroup(group) {
		case utils.Unsupported:
			return nil, nil, nil, nil,
				errors.Wrap(errcode.InvalidParams, "unsupported schedule: "+group.Name)
		case utils.Everyday:
			everyDay = append(everyDay, group)
		case utils.EveryDayInCertainMonth:
			everyDayInCertainMonth = append(everyDayInCertainMonth, group)
		case utils.CertainWeek:
			certainWeek = append(certainWeek, group)
		case utils.OnceDay:
			onceDay = append(onceDay, group)
		}
	}
	return everyDay, everyDayInCertainMonth, certainWeek, onceDay, nil
}

func checkCronHPAEverydayJobs(everyday []*entity.CronScaleJobGroup) error {
	startTime := time.Now()
	period := 24 * time.Hour
	minInterval := 30 * time.Minute
	maxInterval := 24 * time.Hour
	return checkCronHPAJobsInPeriod(everyday, startTime, period, minInterval, maxInterval)
}

func checkCronHPAEveryDayInCertainMonthJobs(everyDayInMonth, everyDay []*entity.CronScaleJobGroup,
	minInterval, maxInterval time.Duration, excludeDates []string) error {
	if len(everyDayInMonth) == 0 {
		return nil
	}
	period := 24 * time.Hour
	now := time.Now()
	var skipDates []time.Time
	// 计算跳过日期
	for _, exclude := range excludeDates {
		initTime := now
		for i := 0; i < 100; i++ {
			s, e := utils.GetScheduleNextTime(exclude, initTime)
			if e != nil {
				return errors.Wrap(errcode.InvalidParams, e.Error())
			}
			skipDates = append(skipDates, s)
			initTime = s
		}
	}
	// 与每日任务比较
	for _, d := range everyDayInMonth {
		startTime, err := utils.GetScheduleNextTime(d.UpSchedule, now)
		if err != nil {
			return errors.Wrap(errcode.InvalidParams, err.Error())
		}
		err = checkCronHPAJobsInPeriod(append(everyDay, d), startTime.Add(-1*time.Hour), period, minInterval, maxInterval)
		if err != nil {
			return err
		}
	}

	// 内部比较
	for i := 0; i < len(everyDayInMonth)-1; i++ {
		err := checkCronHPAJobsInPeriod([]*entity.CronScaleJobGroup{everyDayInMonth[i]}, now, period, minInterval, maxInterval)
		if err != nil {
			return err
		}
		for j := i + 1; j < len(everyDayInMonth); j++ {
			// 获取重合日期
			overlapDates, err := getCronHPAOverlapTimeInterval(now, everyDayInMonth[i], everyDayInMonth[j])
			if err != nil {
				return errors.Wrap(errcode.InvalidParams, err.Error())
			}
			for _, o := range overlapDates {
				flg := true
				// 过滤跳过日期
				for _, d := range skipDates {
					if o.Start.Year() == d.Year() &&
						o.Start.Month() == d.Month() &&
						o.Start.Day() == d.Day() {
						flg = false
						break
					}
				}
				if flg {
					// 校验同一天是否生效
					err = checkCronHPAJobsInPeriod(
						[]*entity.CronScaleJobGroup{everyDayInMonth[i], everyDayInMonth[j]},
						o.Start.Add(-1*time.Hour), period, minInterval, maxInterval)
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return nil
}

func checkCronHPACertainWeekJobs(certainWeek, everyDayInMonth, everyDay []*entity.CronScaleJobGroup) error {
	if len(certainWeek) == 0 {
		return nil
	}
	now := time.Now()
	// 与每日任务
	if len(everyDay) > 0 {
		for _, d := range certainWeek {
			up, err := utils.GetScheduleNextTime(d.UpSchedule, now)
			if err != nil {
				return errors.Wrap(errcode.InvalidParams, err.Error())
			}

			down, err := utils.GetScheduleNextTime(d.DownSchedule, up)
			if err != nil {
				return errors.Wrap(errcode.InvalidParams, err.Error())
			}

			err = checkCronHPAJobsInPeriod(append(everyDay, d), up.Add(-1*time.Hour),
				down.Sub(up), time.Duration(0), 7*24*time.Hour)
			if err != nil {
				return err
			}
		}
	}

	// 与特定月份每日任务
	if len(everyDayInMonth) > 0 {
		for _, c := range everyDayInMonth {
			// 获取有效起始时间
			next, err := utils.GetScheduleNextTime(c.UpSchedule, now)
			if err != nil {
				return errors.Wrap(errcode.InvalidParams, err.Error())
			}
			next = next.Add(-1 * time.Hour)

			for _, d := range certainWeek {
				up, err := utils.GetScheduleNextTime(d.UpSchedule, next)
				if err != nil {
					return errors.Wrap(errcode.InvalidParams, err.Error())
				}

				down, err := utils.GetScheduleNextTime(d.DownSchedule, up)
				if err != nil {
					return errors.Wrap(errcode.InvalidParams, err.Error())
				}
				err = checkCronHPAJobsInPeriod([]*entity.CronScaleJobGroup{c, d}, up.Add(-1*time.Hour),
					down.Sub(up), time.Duration(0), 7*24*time.Hour)
				if err != nil {
					return err
				}
			}
		}
	}

	// 内部比较
	err := checkCronHPAJobsInPeriod(certainWeek, now,
		7*24*time.Hour, time.Duration(0), 7*24*time.Hour)
	if err != nil {
		return err
	}

	return nil
}

func checkCronHPAOnceDayJobs(onceDay, certainWeek, everyDayInMonth, everyDay []*entity.CronScaleJobGroup) error {
	now := time.Now()
	// 与每日任务，特定月份每日任务，星期任务比较
	for _, d := range onceDay {
		up, err := utils.GetScheduleNextTime(d.UpSchedule, now)
		if err != nil {
			return errors.Wrap(errcode.InvalidParams, err.Error())
		}

		down, err := utils.GetScheduleNextTime(d.DownSchedule, up)
		if err != nil {
			return errors.Wrap(errcode.InvalidParams, err.Error())
		}

		if down.Sub(up) > 7*24*time.Hour {
			return errors.Wrap(errcode.InvalidParams, "schedule of once job greater than 7 days: "+d.Name)
		}
		local := []*entity.CronScaleJobGroup{d}
		local = append(local, everyDay...)
		local = append(local, everyDayInMonth...)
		local = append(local, certainWeek...)
		err = checkCronHPAJobsInPeriod(local, up.Add(-1*time.Hour), down.Sub(up), time.Duration(0), 7*24*time.Hour)
		if err != nil {
			return err
		}
	}

	// 内部比较
	err := checkCronHPAJobsInPeriod(onceDay, now, 365*24*time.Hour, time.Duration(0), 7*24*time.Hour)
	if err != nil {
		return err
	}
	return nil
}

func checkCronHPAJobsInPeriod(groups []*entity.CronScaleJobGroup, startTime time.Time,
	period, minInterval, maxInterval time.Duration) error {
	var timeRanges []*utils.TimeRange
	var err error
	if len(groups) == 0 {
		return nil
	}
	// 每日执行任务
	for _, d := range groups {
		up, e := utils.GetScheduleNextTime(d.UpSchedule, startTime)
		if e != nil {
			return errors.Wrap(errcode.InvalidParams, e.Error())
		}

		down, e := utils.GetScheduleNextTime(d.DownSchedule, up)
		if e != nil {
			return errors.Wrap(errcode.InvalidParams, e.Error())
		}

		if down.Sub(up) < minInterval || down.Sub(up) > maxInterval {
			return errors.Wrap(errcode.InvalidParams,
				"interval of schedule invalid: "+d.Name)
		}
	}
	timeRanges, err = buildCronHPATimeRanges(groups, startTime, startTime.Add(period))
	if err != nil {
		return err
	}
	if utils.IsCronHPAJobsOverlap(timeRanges) {
		return errors.Wrap(
			errcode.InvalidParams,
			"cron schedule groups overlap")
	}
	return nil
}

func buildCronHPATimeRanges(groups []*entity.CronScaleJobGroup, initTime, deadline time.Time) ([]*utils.TimeRange, error) {
	timeRanges := make([]*utils.TimeRange, 0)
	for _, g := range groups {
		startTime := initTime
		for startTime.Before(deadline) {
			start, err := utils.GetScheduleNextTime(g.UpSchedule, startTime)
			if err != nil {
				return nil, errors.Wrap(errcode.InternalError, err.Error())
			}

			end, err := utils.GetScheduleNextTime(g.DownSchedule, start)
			if err != nil {
				return nil, errors.Wrap(errcode.InternalError, err.Error())
			}

			timeRanges = append(timeRanges, &utils.TimeRange{
				Name:  g.Name,
				Start: start,
				End:   end,
			})
			startTime = end
		}
	}

	return timeRanges, nil
}

type CronHPAJobTimeInterval struct {
	Start time.Time
	End   time.Time
}

func getCronHPAOverlapTimeInterval(initTime time.Time, first, second *entity.CronScaleJobGroup) ([]*CronHPAJobTimeInterval, error) {
	var intervals []*CronHPAJobTimeInterval
	var firstDays []*CronHPAJobTimeInterval
	var secondDays []*CronHPAJobTimeInterval
	intervalHour := 23 * time.Hour
	startTime := initTime
	deadline := initTime.Add(12 * 30 * 24 * time.Hour)

	for startTime.Before(deadline) || len(firstDays) > 100 {
		s, err := utils.GetScheduleNextTime(first.UpSchedule, startTime)
		if err != nil {
			return nil, err
		}
		e, err := utils.GetScheduleNextTime(first.UpSchedule, s)
		if err != nil {
			return nil, err
		}
		firstDays = append(firstDays, &CronHPAJobTimeInterval{
			Start: s,
			End:   e,
		})
		startTime = startTime.Add(intervalHour)
		if startTime.Before(e) {
			startTime = e
		}
	}

	startTime = initTime
	for startTime.Before(deadline) || len(secondDays) > 100 {
		s, err := utils.GetScheduleNextTime(second.UpSchedule, startTime)
		if err != nil {
			return nil, err
		}
		e, err := utils.GetScheduleNextTime(second.UpSchedule, s)
		if err != nil {
			return nil, err
		}
		secondDays = append(secondDays, &CronHPAJobTimeInterval{
			Start: s,
			End:   e,
		})
		startTime = startTime.Add(intervalHour)
		if startTime.Before(e) {
			startTime = e
		}
	}

	if len(firstDays) > len(secondDays) {
		firstDays, secondDays = secondDays, firstDays
	}

	for _, fst := range firstDays {
		for _, snd := range secondDays {
			if fst.Start.Year() == snd.Start.Year() &&
				fst.Start.Month() == snd.Start.Month() &&
				fst.Start.Day() == snd.Start.Day() {
				interval := fst
				if interval.Start.After(snd.Start) {
					interval.Start = snd.Start.Add(-1 * time.Hour)
				}
				if interval.End.Before(snd.End) {
					interval.End = snd.End
				}
				intervals = append(intervals, interval)
			}
		}
	}

	return intervals, nil
}

func buildCronHPAScaleJobPairs(groups []*entity.CronScaleJobGroup, minSize int) []*entity.CronScaleJobPair {
	pairs := make([]*entity.CronScaleJobPair, len(groups))
	for i, g := range groups {
		pairs[i] = &entity.CronScaleJobPair{}
		pairs[i].Name = g.Name
		pairs[i].ScaleUp = &entity.CronScaleJob{
			Name:       g.Name + "-up",
			Schedule:   "0 " + g.UpSchedule,
			TargetSize: g.TargetSize,
			RunOnce:    g.RunOnce,
		}
		pairs[i].ScaleDown = &entity.CronScaleJob{
			Name:       g.Name + "-down",
			Schedule:   "0 " + g.DownSchedule,
			TargetSize: minSize,
			RunOnce:    g.RunOnce,
		}
	}
	return pairs
}
