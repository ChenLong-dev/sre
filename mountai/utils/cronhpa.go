package utils

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/AliyunContainerService/kubernetes-cronhpa-controller/pkg/apis/autoscaling/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"rulai/models/entity"

	"github.com/robfig/cron/v3"
)

type CronScaleJobType string

const (
	Unsupported            CronScaleJobType = "Unsupported"
	Everyday               CronScaleJobType = "EveryDay"
	EveryDayInCertainMonth CronScaleJobType = "EveryDayInCertainMonth"
	CertainWeek            CronScaleJobType = "CertainWeek"
	OnceDay                CronScaleJobType = "OnceDay"
)

var (
	minutePattern       = `[1-5]?\d`
	hourPattern         = `[1,2]?\d`
	dayPattern          = `[1-3]?\d`
	monthPattern        = `1?\d`
	certainMonthPattern = fmt.Sprintf(`(%v,|%v-|\*/)*%v`, monthPattern, monthPattern, monthPattern)
	weekPattern         = `[0-7]`

	// everyDayReg 表示 `a b * * *`
	// 每日执行任务
	everyDayReg = regexp.MustCompile(fmt.Sprintf(`^%v %v \* \* \*$`, minutePattern, hourPattern))
	// everyDayOfCertainMonthReg 表示 `a b * c *` 或 `a b * c,d,e *` 或 `a b * c-e *`
	// 在特定月份内的每日执行任务
	everyDayOfCertainMonthReg = regexp.MustCompile(fmt.Sprintf(`^%v %v \* %v \*$`,
		minutePattern, hourPattern, certainMonthPattern))
	// certainWeekReg 表示 `a b * * c`
	// 每周最多执行一次，需要确定的星期号码
	certainWeekReg = regexp.MustCompile(fmt.Sprintf(`^%v %v \* \* %v$`,
		minutePattern, hourPattern, weekPattern))
	// onceDayReg 表示 `a b c d *`
	// 一次性任务，需要制定明确起始于终止时间，且时间跨度小于等于7天
	onceDayReg = regexp.MustCompile(fmt.Sprintf(`^%v %v %v %v \*$`,
		minutePattern, hourPattern, dayPattern, monthPattern))
)

// GetScheduleNextTime 获取crontab模板时间startTime以后第一次的调度时间
func GetScheduleNextTime(schedule string, startTime time.Time) (time.Time, error) {
	cp := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sch, err := cp.Parse(schedule)
	if err != nil {
		return time.Time{}, err
	}
	return sch.Next(startTime), nil
}

type TimeRange struct {
	Name  string
	Start time.Time
	End   time.Time
}

const (
	OperationDown = -1
	OperationAny  = 0
	OperationUp   = 1
)

// CronHPAJobOperation cronhpa操作
type CronHPAJobOperation struct {
	Name      string
	Time      time.Time
	Operation int
}

// CronHPAJobOperationSlice cronhpa操作切片
type CronHPAJobOperationSlice struct {
	Operation []*CronHPAJobOperation
}

func (o *CronHPAJobOperationSlice) Less(i, j int) bool {
	return o.Operation[i].Time.Before(o.Operation[j].Time)
}

func (o *CronHPAJobOperationSlice) Len() int {
	return len(o.Operation)
}

func (o *CronHPAJobOperationSlice) Swap(i, j int) {
	o.Operation[i], o.Operation[j] = o.Operation[j], o.Operation[i]
}

// IsCronHPAJobsOverlap 判断cronhpa任务组是否重叠
func IsCronHPAJobsOverlap(timeRanges []*TimeRange) bool {
	lenOfRanges := len(timeRanges)
	if lenOfRanges <= 1 {
		return false
	}

	s := &CronHPAJobOperationSlice{Operation: []*CronHPAJobOperation{}}

	for _, r := range timeRanges {
		s.Operation = append(s.Operation, &CronHPAJobOperation{
			Name:      r.Name,
			Time:      r.Start,
			Operation: OperationUp,
		}, &CronHPAJobOperation{
			Name:      r.Name,
			Time:      r.End,
			Operation: OperationDown,
		})
	}

	sort.Stable(s)

	expectOper := CronHPAJobOperation{
		Name:      "Any",
		Operation: OperationAny,
	}

	for i, o := range s.Operation {
		if i < len(s.Operation)-1 && o.Time.Equal(s.Operation[i+1].Time) {
			return true
		}
		// 满足上次预期
		if expectOper.Name != "Any" && expectOper.Name != o.Name {
			return true
		}
		if expectOper.Operation != OperationAny && expectOper.Operation != o.Operation {
			return true
		}

		// 提出下次预期
		if o.Operation == OperationUp {
			expectOper = CronHPAJobOperation{
				Name:      o.Name,
				Operation: OperationDown,
			}
		} else if o.Operation == OperationDown {
			// 下一操作只能为扩容
			expectOper = CronHPAJobOperation{
				Name:      "Any",
				Operation: OperationUp,
			}
			if i == 0 {
				// 起始为缩容，最后一次操作为同一组的扩容
				lastIdx := len(s.Operation) - 1
				if s.Operation[lastIdx].Name != o.Name || s.Operation[lastIdx].Operation != OperationUp {
					return true
				}
			}
		}
	}
	return false
}

// ConvertUnstructToCronHPA unstruct转化为cronHPA结构体
func ConvertUnstructToCronHPA(u *unstructured.Unstructured) (*v1beta1.CronHorizontalPodAutoscaler, error) {
	cronHPA := new(v1beta1.CronHorizontalPodAutoscaler)
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), cronHPA)
	if err != nil {
		return nil, errors.Wrap(err, "ConvertUnstructToCronHPA failed")
	}
	return cronHPA, nil
}

// ConvertCronHPAToUnstruct cronHPA转化为unstruct结构体
func ConvertCronHPAToUnstruct(c *v1beta1.CronHorizontalPodAutoscaler) (*unstructured.Unstructured, error) {
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(c)
	if err != nil {
		return nil, errors.Wrap(err, "ConvertCronHPAToUnstruct failed")
	}
	return &unstructured.Unstructured{Object: m}, nil
}

// TypeOfCronHPAScheduleGroup 判断cronHPA调度组类型
func TypeOfCronHPAScheduleGroup(group *entity.CronScaleJobGroup) CronScaleJobType {
	if everyDayReg.FindString(group.UpSchedule) != "" &&
		everyDayReg.FindString(group.DownSchedule) != "" {
		return Everyday
	}

	if everyDayOfCertainMonthReg.FindString(group.UpSchedule) != "" &&
		everyDayOfCertainMonthReg.FindString(group.DownSchedule) != "" {
		return EveryDayInCertainMonth
	}

	if certainWeekReg.FindString(group.UpSchedule) != "" &&
		certainWeekReg.FindString(group.DownSchedule) != "" {
		return CertainWeek
	}

	if group.RunOnce && onceDayReg.FindString(group.UpSchedule) != "" &&
		onceDayReg.FindString(group.DownSchedule) != "" {
		return OnceDay
	}
	return Unsupported
}
