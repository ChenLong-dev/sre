package hook

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func TestHook_AddArg(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		et := time.Now().Add(time.Hour)

		hk := NewManager().
			CreateHook(context.Background()).
			AddArg(render.EndTimeArgKey, et)
		assert.Equal(t, et, hk.Arg(render.EndTimeArgKey))
	})
}

func TestHook_Do(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		st := time.Now()
		et := time.Now().Add(time.Hour)
		var sts, ets string

		NewManager().
			RegisterHook(func(hk *Hook) {
				sts = render.PatternStartTime(hk.args).StringValue()
			}, func(hk *Hook) {
				ets = render.PatternEndTime(hk.args).StringValue()
			}).
			CreateHook(context.Background()).
			AddArg(render.StartTimeArgKey, st).
			AddArg(render.EndTimeArgKey, et).
			Do(func() {
				assert.Equal(t, st.Format("2006/01/02 15:04:05.000"), sts)
				assert.NotEqual(t, et.Format("2006/01/02 15:04:05.000"), ets)
			})
		assert.Equal(t, et.Format("2006/01/02 15:04:05.000"), ets)
	})
}

func TestHook_ProcessAfterHook(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		et := time.Now().Add(time.Hour)
		var ets string

		hk := NewManager().
			RegisterAfterHook(func(hk *Hook) {
				ets = render.PatternEndTime(hk.args).StringValue()
			}).
			CreateHook(context.Background()).
			AddArg(render.EndTimeArgKey, et)

		hk.ProcessAfterHook()
		assert.Equal(t, et.Format("2006/01/02 15:04:05.000"), ets)
	})
}

func TestHook_ProcessPreHook(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		st := time.Now()
		var sts string

		hk := NewManager().
			RegisterPreHook(func(hk *Hook) {
				sts = render.PatternStartTime(hk.args).StringValue()
			}).
			CreateHook(context.Background()).
			AddArg(render.StartTimeArgKey, st)

		hk.ProcessPreHook()
		assert.Equal(t, st.Format("2006/01/02 15:04:05.000"), sts)
	})
}
