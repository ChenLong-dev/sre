package sentry

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func ExampleCaptureException() {
	Init(&Config{
		DSN: "sentry_DSN",
		Tags: map[string]string{
			"tag1": "tag1",
			"tag2": "tag2",
		},
		Environment:   "prd",
		ErrCodeFilter: []string{"code1", "code2"},
	},
	)

	wrapErr := errors.Wrap(errcode.BreakerDegradedError, errcode.BreakerTimeoutError.Error())
	ctx := context.Background()
	CaptureWithBreadAndTags(
		ctx,
		wrapErr,
		&Breadcrumb{
			Category: "category1",
			Data: map[string]interface{}{
				"data1": "v1",
				"data2": "v2",
			},
		},
		Tag{
			Key:   "tag1",
			Value: "t1",
		},
	)

	CaptureWithTags(ctx, wrapErr, Tag{
		Key:   "tag1",
		Value: "t1",
	})

	CaptureMessage(ctx, "this is a error")
}

func ExampleRecover() {
	Init(&Config{
		DSN: "sentry_DSN",
		Tags: map[string]string{
			"tag1": "tag1",
			"tag2": "tag2",
		},
		Environment:   "dev",
		ErrCodeFilter: []string{"code1", "code2"},
	},
	)
	ctx := context.Background()

	// recover
	CapturePanic(ctx, func() {
		// do something fearful
		panic("panic")
	})
}

func ExampleNewGinMiddleware() {
	Init(&Config{
		DSN:         "sentry_DSN",
		Environment: "prd",
		Tags: map[string]string{
			"tag": "value",
		},
	})

	app := gin.Default()

	app.Use(
		GinMiddleware(&GinOption{
			Recover: true,
		}),
	)
	app.Use(
		GlobalTagsMiddleware(map[string]string{
			"tag1": "value1",
		}),
	)

	app.GET("/", func(c *gin.Context) {
		// sentry会catch，因为我们设置了rePanic为true
		panic("err")
	})

	_ = app.Run(":3000")
}
