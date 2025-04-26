package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetScheduleNextTime(t *testing.T) {
	now := time.Now()
	t.Run("six bit", func(t *testing.T) {
		d, err := GetScheduleNextTime("* * 5 * * *", now)
		assert.Nil(t, err)
		assert.NotEmpty(t, d)
	})

	t.Run("five bit", func(t *testing.T) {
		d, err := GetScheduleNextTime("* 5 * * *", now)
		assert.Nil(t, err)
		assert.NotEmpty(t, d)
	})

	t.Run("error", func(t *testing.T) {
		d, err := GetScheduleNextTime("5 * * *", now)
		assert.NotNil(t, err)
		assert.Empty(t, d)
	})
}

func TestIsTimeRangesOverlapByCrontab(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name          string
		schedulePairs [][]string
		expected      bool
	}{
		{
			name:     "empty",
			expected: false,
		},
		{
			name: "single",
			schedulePairs: [][]string{
				{
					"0 0 8 * * *",
					"0 0 10 * * *",
					"oper 1",
				},
			},
			expected: false,
		},
		{
			name: "split-range normal",
			schedulePairs: [][]string{
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(-1*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(1*time.Hour).Hour()),
					"oper 1",
				},
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(2*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(3*time.Hour).Hour()),
					"oper 2",
				},
			},
			expected: false,
		},
		{
			name: "edge overlap",
			schedulePairs: [][]string{
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(-1*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(1*time.Hour).Hour()),
					"oper 1",
				},
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(1*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(2*time.Hour).Hour()),
					"oper 2",
				},
			},
			expected: true,
		},
		{
			name: "part overlap",
			schedulePairs: [][]string{
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(1*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(3*time.Hour).Hour()),
					"oper 1",
				},
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(2*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(4*time.Hour).Hour()),
					"oper 2",
				},
			},
			expected: true,
		},
		{
			name: "same name range",
			schedulePairs: [][]string{
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(1*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(3*time.Hour).Hour()),
					"oper 1",
				},
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(2*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(4*time.Hour).Hour()),
					"oper 1",
				},
			},
			expected: true,
		},
		{
			name: "completely including overlap",
			schedulePairs: [][]string{
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(1*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(4*time.Hour).Hour()),
					"oper 1",
				},
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(2*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(3*time.Hour).Hour()),
					"oper 2",
				},
			},
			expected: true,
		},
		{
			name: "start and end with down",
			schedulePairs: [][]string{
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(-3*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(1*time.Hour).Hour()),
					"oper 1",
				},
				{
					fmt.Sprintf("0 0 %d * * *", now.Add(-2*time.Hour).Hour()),
					fmt.Sprintf("0 0 %d * * *", now.Add(-1*time.Hour).Hour()),
					"oper 2",
				},
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		tt := testCase
		t.Run(tt.name, func(t *testing.T) {
			timeRanges := make([]*TimeRange, len(tt.schedulePairs))
			for i, pair := range tt.schedulePairs {
				u, err := GetScheduleNextTime(pair[0], now)
				assert.Nil(t, err)
				d, err := GetScheduleNextTime(pair[1], now)
				assert.Nil(t, err)
				timeRanges[i] = &TimeRange{
					Name:  pair[2],
					Start: u,
					End:   d,
				}
			}
			b := IsCronHPAJobsOverlap(timeRanges)
			assert.Equal(t, tt.expected, b)
		})
	}
}
