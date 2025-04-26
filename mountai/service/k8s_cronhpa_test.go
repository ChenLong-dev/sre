package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"rulai/models/entity"
	"rulai/models/req"
)

var (
	minInterval = 30 * time.Minute
	maxInterval = 24 * time.Hour
	testCases   = []struct {
		name   string
		param  *req.CreateTaskParamReq
		errStr string
	}{
		{
			name: "less 0.5h in everyday",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "less",
					TargetSize:   2,
					UpSchedule:   "0 1 * * *",
					DownSchedule: "25 1 * * *",
					RunOnce:      false,
				}},
			},
			errStr: "interval of schedule invalid",
		},
		{
			name: "overlap between two everyday",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 * * *",
					DownSchedule: "0 3 * * *",
					RunOnce:      false,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * * *",
					DownSchedule: "0 4 * * *",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "less 0.5h in everyday of certain month",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "less",
					TargetSize:   2,
					UpSchedule:   "0 1 * 2,3,4 *",
					DownSchedule: "25 1 * 2,3,4 *",
					RunOnce:      false,
				}},
			},
			errStr: "interval of schedule invalid",
		},
		{
			name: "overlap between two everyday in month",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 * 7-9 *",
					DownSchedule: "0 3 * 7-9 *",
					RunOnce:      false,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * 7-9 *",
					DownSchedule: "0 4 * 7-9 *",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "overlap between everyday and everyday in month",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 * * *",
					DownSchedule: "0 3 * * *",
					RunOnce:      false,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * 7-9 *",
					DownSchedule: "0 4 * 7-9 *",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "normal certain week",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 * * 1",
					DownSchedule: "0 3 * * 3",
					RunOnce:      false,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * * 4",
					DownSchedule: "0 4 * * 5",
					RunOnce:      false,
				}},
			},
		},
		{
			name: "overlap certain week",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 * * 1",
					DownSchedule: "0 3 * * 3",
					RunOnce:      false,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * * 1",
					DownSchedule: "0 4 * * 5",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "overlap between certain week and everyday",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 * * 1",
					DownSchedule: "0 3 * * 3",
					RunOnce:      false,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * * *",
					DownSchedule: "0 4 * * *",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "overlap between certain week and everyday in certain month",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 * * 1",
					DownSchedule: "0 3 * * 3",
					RunOnce:      false,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * 3 *",
					DownSchedule: "0 4 * 3 *",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "invalid runOnce",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 1 8 *",
					DownSchedule: "0 3 3 * *",
					RunOnce:      true,
				}},
			},
			errStr: "unsupported schedule",
		},
		{
			name: "overlong runOnce",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 1 8 *",
					DownSchedule: "0 3 1 9 *",
					RunOnce:      true,
				}},
			},
			errStr: "schedule of once job greater than 7 days",
		},
		{
			name: "overlap between runOnce and everyday",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 1 8 *",
					DownSchedule: "0 3 2 8 *",
					RunOnce:      true,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * * *",
					DownSchedule: "0 4 * * *",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "overlap between runOnce and everyday in certain month",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 1 8 *",
					DownSchedule: "0 3 2 8 *",
					RunOnce:      true,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * 8 *",
					DownSchedule: "0 4 * 8 *",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "overlap between runOnce and certain week",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 1 8 *",
					DownSchedule: "0 3 2 8 *",
					RunOnce:      true,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 * * 0",
					DownSchedule: "0 4 * * 6",
					RunOnce:      false,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "overlap between runOnce",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 1 8 *",
					DownSchedule: "0 3 2 8 *",
					RunOnce:      true,
				}, {
					Name:         "snd",
					TargetSize:   2,
					UpSchedule:   "0 2 1 8 *",
					DownSchedule: "0 4 3 8 *",
					RunOnce:      true,
				}},
			},
			errStr: "cron schedule groups overlap",
		},
		{
			name: "share the same name",
			param: &req.CreateTaskParamReq{
				MinPodCount: 1,
				MaxPodCount: 5,
				CronScaleJobGroups: []*entity.CronScaleJobGroup{{
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 1 1 8 *",
					DownSchedule: "0 3 2 8 *",
					RunOnce:      true,
				}, {
					Name:         "fst",
					TargetSize:   2,
					UpSchedule:   "0 2 4 8 *",
					DownSchedule: "0 4 5 8 *",
					RunOnce:      true,
				}},
			},
			errStr: "groups sharing the same name",
		},
	}
)

func TestValidateCronHPA(t *testing.T) {
	for _, testCase := range testCases {
		tt := testCase
		t.Run(tt.name, func(t *testing.T) {
			err := s.ValidateCronHPA(tt.param, minInterval, maxInterval)
			if tt.errStr != "" {
				if assert.NotNil(t, err) {
					assert.Contains(t, err.Error(), tt.errStr)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
