package goroutine

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func TestErrGroup_Wait(t *testing.T) {
	Init(&Config{
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [status: %s] [mode: %m] %S  Group: %N:%I , Current: %n:%i , %E",
		},
	})

	t.Run("normal", func(t *testing.T) {
		ctx := context.Background()
		eg := WithContext(ctx, "Test")

		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(2 * time.Second)

			return nil
		})
		eg.Go(ctx, "Test2", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			return errors.New("this is a error")
		})

		err := eg.Wait()

		assert.NotNil(t, err)
	})
}

func TestErrGroup_Go(t *testing.T) {
	Init(&Config{
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [status: %s] [mode: %m] %S  Group: %N:%I , Current: %n:%i , %E",
		},
	})

	t.Run("normal", func(t *testing.T) {
		ctx := context.Background()
		eg := WithContext(ctx, "Test")

		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			return nil
		})
		eg.Go(ctx, "Test2", func(c context.Context) error {
			time.Sleep(2 * time.Second)

			return errors.New("this is a error")
		})

		err := eg.Wait()

		assert.NotNil(t, err)
	})

	t.Run("panic", func(t *testing.T) {
		ctx := context.Background()
		eg := WithContext(ctx, "Test")

		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			panic(errors.New("this is panic"))
		})
		eg.Go(ctx, "Test2", func(c context.Context) error {
			time.Sleep(2 * time.Second)

			return nil
		})

		err := eg.Wait()

		assert.NotNil(t, err)
	})

	t.Run("duplicate", func(t *testing.T) {
		ctx := context.Background()
		eg := WithContext(ctx, "Test")

		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			return nil
		})
		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(2 * time.Second)

			return errors.New("this is a error")
		})

		err := eg.Wait()

		assert.Nil(t, err)
	})
}

func TestErrGroup_GetGoroutineInfo(t *testing.T) {
	Init(&Config{
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [status: %s] [mode: %m] %S  Group: %N:%I , Current: %n:%i , %E",
		},
	})

	t.Run("normal", func(t *testing.T) {
		ctx := context.Background()
		eg := WithContext(ctx, "Test")

		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			return nil
		})
		eg.Go(ctx, "Test2", func(c context.Context) error {
			time.Sleep(2 * time.Second)

			return errors.New("this is a error")
		})

		e := eg.Wait()

		info, err := eg.GetGoroutineInfo("Test2")
		assert.Nil(t, err)
		assert.Equal(t, fmt.Sprintf("%s", e), fmt.Sprintf("%s", info.Error))
		assert.Equal(t, stateError, info.State)
		info, err = eg.GetGoroutineInfo("Test1")
		assert.Nil(t, err)
		assert.Nil(t, info.Error)
		_, err = eg.GetGoroutineInfo("Test3")
		assert.NotNil(t, err)
	})
}

func TestSetMaxWorker(t *testing.T) {
	Init(&Config{
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [status: %s] [mode: %m] %S  Group: %N:%I , Current: %n:%i , %E",
		},
	})

	t.Run("remain", func(t *testing.T) {
		ctx := context.Background()
		eg := New("Test", SetCancelMode(ctx), SetMaxWorker(2, false))

		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			return nil
		})

		err := eg.Wait()
		assert.Nil(t, err)
	})

	t.Run("full wait", func(t *testing.T) {
		ctx := context.Background()
		eg := New("Test", SetCancelMode(ctx), SetMaxWorker(2, true))

		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			return nil
		})
		eg.Go(ctx, "Test2", func(c context.Context) error {
			time.Sleep(500 * time.Millisecond)

			return nil
		})
		eg.Go(ctx, "Test3", func(c context.Context) error {
			time.Sleep(3 * time.Second)

			return nil
		})
		eg.Go(ctx, "Test4", func(c context.Context) error {
			time.Sleep(500 * time.Millisecond)

			return nil
		})
		eg.Go(ctx, "Test5", func(c context.Context) error {
			time.Sleep(100 * time.Millisecond)

			return nil
		})

		err := eg.Wait()
		assert.Nil(t, err)
	})

	t.Run("full not-wait", func(t *testing.T) {
		ctx := context.Background()
		eg := New("Test", SetCancelMode(ctx), SetMaxWorker(2, false))

		eg.Go(ctx, "Test1", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			return nil
		})
		eg.Go(ctx, "Test2", func(c context.Context) error {
			time.Sleep(500 * time.Millisecond)

			return nil
		})
		eg.Go(ctx, "Test3", func(c context.Context) error {
			time.Sleep(1 * time.Second)

			return nil
		})
		eg.Go(ctx, "Test4", func(c context.Context) error {
			time.Sleep(500 * time.Millisecond)

			return nil
		})
		eg.Go(ctx, "Test5", func(c context.Context) error {
			time.Sleep(800 * time.Millisecond)

			return nil
		})

		err := eg.Wait()
		assert.NotNil(t, err)
	})
}
