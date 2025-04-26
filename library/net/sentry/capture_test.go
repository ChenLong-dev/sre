package sentry

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func TestCaptureWithBreadAndTags(t *testing.T) {
	t.Skip("Should have sentry environment")
	Init(&Config{
		DSN:         "your_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})
	ctx := context.Background()

	t.Run("CaptureWithBreadAndTags", func(t *testing.T) {
		CaptureWithBreadAndTags(
			ctx,
			errors.New("error"),
			&Breadcrumb{
				Category: "redis",
				Data: map[string]interface{}{
					"d1": "d1",
				},
			},
			Tag{
				Key:   "tag1",
				Value: "value1",
			})
	})

	t.Run("CaptureWithCommonCtx", func(t *testing.T) {
		ctx := context.Background()
		ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub())
		CaptureWithBreadAndTags(ctx, errors.New("new error1"), &Breadcrumb{
			Category: "redis",
			Data: map[string]interface{}{
				"d1": "d1",
			},
		})

		CaptureWithBreadAndTags(ctx, errors.New("new error2"), &Breadcrumb{
			Category: "redis",
			Data: map[string]interface{}{
				"d1": "d1",
			},
		})
	})

	t.Run("CaptureWithDiffCtx", func(t *testing.T) {
		CaptureWithBreadAndTags(context.Background(), errors.New("new error1"), &Breadcrumb{
			Category: "redis",
			Data: map[string]interface{}{
				"d1": "d1",
			},
		})

		CaptureWithBreadAndTags(context.Background(), errors.New("new error2"), &Breadcrumb{
			Category: "redis",
			Data: map[string]interface{}{
				"d1": "d1",
			},
		})
	})

	t.Run("CaptureNil", func(t *testing.T) {
		CaptureWithBreadAndTags(
			ctx, nil, &Breadcrumb{})
	})
}

func TestAddBreadcrumb(t *testing.T) {
	t.Skip("Should have sentry environment")
	Init(&Config{
		DSN:         "your_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})

	ctx := context.Background()
	ctx = AddBreadcrumb(ctx, &Breadcrumb{
		Category: "redis1",
		Data: map[string]interface{}{
			"d1": "d1",
		},
	})
	ctx = AddBreadcrumb(ctx, &Breadcrumb{
		Category: "redis",
		Data: map[string]interface{}{
			"d2": "d2",
		},
	})
	CaptureWithBreadAndTags(ctx, errors.Wrap(errors.New("new error1"), "a"), &Breadcrumb{
		Category: "redis",
		Data: map[string]interface{}{
			"d3": "d3",
		},
	})
	CaptureWithBreadAndTags(context.Background(), errors.Wrap(errors.New("new error2"), "b"), &Breadcrumb{
		Category: "redis",
		Data: map[string]interface{}{
			"d4": "d4",
		},
	})
}

func TestCaptureWithTags(t *testing.T) {
	t.Skip("Should have sentry environment")
	Init(&Config{
		DSN:         "your_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})

	t.Run("CaptureWithTags", func(t *testing.T) {
		CaptureWithTags(context.Background(), errors.New("error"), Tag{
			Key:   "tag1",
			Value: "value1",
		})
	})
}

func TestCaptureMessage(t *testing.T) {
	t.Skip("Should have sentry environment")
	Init(&Config{
		DSN:         "your_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})

	ctx := context.Background()
	t.Run("CaptureMessage", func(t *testing.T) {
		CaptureMessage(ctx, "msg")
	})
}

func TestCapturePanic(t *testing.T) {
	t.Skip("Should have sentry environment")
	Init(&Config{
		DSN:         "your_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})

	assert.Panics(t, func() {
		CapturePanic(context.Background(), func() {
			panic("panic")
		})
	})
}

func TestGinMiddleware(t *testing.T) {
	t.Skip("Should have sentry environment")
	Init(&Config{
		DSN:         "your_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})

	t.Run("Recover", func(t *testing.T) {
		app := gin.New()
		app.Use(GinMiddleware(&GinOption{
			Recover: true,
		}))

		app.GET("/", func(c *gin.Context) {
			panic("this is a panic")
		})

		r, err := httpUtil.TestGinJsonRequest(app, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("RecoverRePanic", func(t *testing.T) {
		app := gin.New()
		app.Use(GinMiddleware(nil))

		app.GET("/", func(c *gin.Context) {
			panic("this is a panic")
		})

		assert.Panics(t, func() {
			_, _ = httpUtil.TestGinJsonRequest(app, "GET", "/", nil, nil, nil)
		})
	})
}

func BenchmarkGinMiddleware(b *testing.B) {
	b.Skip("Should have sentry environment")
	Init(&Config{
		DSN:         "your_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})

	b.Run("safe", func(b *testing.B) {
		app := gin.New()
		app.Use(GinMiddleware(nil))
		app.GET("/", func(c *gin.Context) {
			time.Sleep(time.Millisecond * 100)
			c.JSON(200, nil)
		})

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = httpUtil.TestGinJsonRequest(app, "GET", "/", nil, nil, nil)
			}
		})
	})
}

func BenchmarkGlobalTagsMiddleware(b *testing.B) {
	b.Skip("Should have sentry environment")
	Init(&Config{
		DSN:         "your_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})

	b.Run("safe", func(b *testing.B) {
		app := gin.New()
		app.Use(GinMiddleware(nil))
		app.Use(GlobalTagsMiddleware(nil))
		app.GET("/", func(c *gin.Context) {
			time.Sleep(time.Millisecond * 100)
			c.JSON(200, nil)
		})

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = httpUtil.TestGinJsonRequest(app, "GET", "/", nil, nil, nil)
			}
		})
	})
}
