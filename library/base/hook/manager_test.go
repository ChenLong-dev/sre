package hook

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func TestManager_RegisterLogHook(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		st := time.Now()
		et := time.Now().Add(time.Hour)
		hk := NewManager().
			RegisterLogHook(&render.Config{
				Stdout:        true,
				StdoutPattern: "%S - %E",
			}, map[string]render.PatternFunc{
				"S": render.PatternStartTime,
				"E": render.PatternEndTime,
			}).
			CreateHook(context.Background()).
			AddArg(render.StartTimeArgKey, st).
			AddArg(render.EndTimeArgKey, et)

		hk.ProcessPreHook()
		hk.ProcessAfterHook()
	})

	t.Run("duplicate", func(t *testing.T) {
		hkm := NewManager().
			RegisterLogHook(&render.Config{
				Stdout:        true,
				StdoutPattern: "%S - %E",
			}, map[string]render.PatternFunc{
				"S": render.PatternStartTime,
				"E": render.PatternEndTime,
			})

		assert.Panics(t, func() {
			hkm.RegisterLogHook(&render.Config{
				Stdout:        true,
				StdoutPattern: "%S - %E",
			}, map[string]render.PatternFunc{
				"S": render.PatternStartTime,
				"E": render.PatternEndTime,
			})
		})
	})
}

func TestManager_AddArg(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		st := time.Now()
		hk := NewManager().
			CreateHook(context.Background()).
			AddArg(render.StartTimeArgKey, st)

		assert.Equal(t, st, hk.Arg(render.StartTimeArgKey).(time.Time))
	})
}

func BenchmarkManager_AddArg(b *testing.B) {
	b.Run("safe", func(b *testing.B) {
		st := time.Now()
		hkm := NewManager()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				hkm.AddArg(render.StartTimeArgKey, st).
					CreateHook(context.Background())
			}
		})
	})
}

func TestManager_RegisterHook(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		st := time.Now()
		et := time.Now().Add(time.Hour)
		var sts, ets string

		hk := NewManager().
			RegisterHook(func(hk *Hook) {
				sts = render.PatternStartTime(hk.args).StringValue()
			}, func(hk *Hook) {
				ets = render.PatternEndTime(hk.args).StringValue()
			}).
			CreateHook(context.Background()).
			AddArg(render.StartTimeArgKey, st).
			AddArg(render.EndTimeArgKey, et)

		hk.ProcessPreHook()
		assert.Equal(t, st.Format("2006/01/02 15:04:05.000"), sts)
		assert.NotEqual(t, et.Format("2006/01/02 15:04:05.000"), ets)

		hk.ProcessAfterHook()
		assert.Equal(t, et.Format("2006/01/02 15:04:05.000"), ets)
	})
}

func BenchmarkManager_RegisterHook(b *testing.B) {
	b.Run("safe", func(b *testing.B) {
		st := time.Now()
		et := time.Now().Add(time.Hour)
		hkm := NewManager().
			AddArg(render.StartTimeArgKey, st).
			AddArg(render.EndTimeArgKey, et)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				hkm.RegisterHook(func(hk *Hook) {
					render.PatternStartTime(hk.args).StringValue()
				}, func(hk *Hook) {
					render.PatternEndTime(hk.args).StringValue()
				}).
					CreateHook(context.Background())
			}
		})
	})
}

func TestManager_CreateHook(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		NewManager().CreateHook(context.Background())
	})

	t.Run("same args", func(t *testing.T) {
		m := make(map[string]interface{})
		for i := 0; i < 10; i++ {
			m[strconv.Itoa(i)] = fmt.Sprintf("this is %d", i)
		}

		hkm := NewManager()
		for k, v := range m {
			hkm.AddArg(k, v)
		}
		args := hkm.CreateHook(context.Background()).Args()

		for k, v := range m {
			arg, ok := args[k]
			assert.Equal(t, true, ok)
			assert.Equal(t, v, arg)
		}
	})
}
