package sentry

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/getsentry/sentry-go"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/base/net"
	"gitlab.shanhai.int/sre/library/base/runtime"
	"gitlab.shanhai.int/sre/library/base/slice"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

const (
	// 单个hub最大面包屑数量
	MaxBreadcrumbs = 100
)

var (
	_logger logger
)

// 初始化sentry
func Init(config *Config) {
	if config == nil {
		panic("sentry config is nil")
	}
	if config.DSN == "" {
		panic("sentry DSN can not be empty")
	}
	if config.Config == nil {
		config.Config = &render.Config{}
	}
	if config.Config.StdoutPattern == "" {
		config.Config.StdoutPattern = defaultPattern
	}
	if config.Config.OutPattern == "" {
		config.Config.OutPattern = defaultPattern
	}
	if config.Config.OutFile == "" {
		config.Config.OutFile = _infoFile
	}
	if config.Tags == nil {
		config.Tags = make(map[string]string)
	}
	_logger = getDefaultLogger(config)

	// 创建hub
	err := sentry.Init(sentry.ClientOptions{
		Dsn: config.DSN,
		// TODO: 每个hub的面包屑最大值只能是100
		MaxBreadcrumbs:   MaxBreadcrumbs,
		AttachStacktrace: true,
		Environment:      config.Environment,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			beforeSend(event, hint)

			// 过滤状态码
			for key, value := range event.Tags {
				if key == TagErrorCode && slice.StrSliceContains(config.ErrCodeFilter, value) {
					return nil
				}
			}

			// 日志打印
			_logger.Print(map[string]interface{}{
				"tags":    event.Tags,
				"env":     event.Environment,
				"DSN":     config.DSN,
				"eventID": event.EventID,
				"source": runtime.GetFilterCallers(
					append(
						runtime.DefaultFilterCallerRegexp,
						regexp.MustCompile(`github.com/getsentry(@.*)?/.*.go`),
					),
				),
			})

			return event
		},
	})
	if err != nil {
		panic(err)
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		internalIP, err := net.GetInternalIP()
		if err == nil {
			config.Tags[TagInternalIP] = internalIP
		}

		scope.SetTags(config.Tags)
	})
}

// 关闭sentry
func Close() {
	_logger.Close()
}

// 判断sentry是否init
func IsInit() bool {
	return sentry.CurrentHub().Client() != nil
}

// 修改issue类型,聚合方式,添加标签和额外信息
func beforeSend(event *sentry.Event, hint *sentry.EventHint) {
	var hintErr error
	if hint.OriginalException != nil {
		hintErr = hint.OriginalException
	} else {
		re, ok := hint.RecoveredException.(error)
		if ok {
			hintErr = re
		} else {
			event.Extra[ExtraErrorMsg] = hint.RecoveredException
			return
		}
	}

	var c errcode.Codes
	if !errors.As(hintErr, &c) {
		c = errcode.Cause(hintErr)
	}
	// 修改issue类型
	if len(event.Exception) != 0 {
		event.Exception[len(event.Exception)-1].Type = c.Error()
	}
	// 自定义聚合方式
	event.Fingerprint = []string{strconv.Itoa(c.Code()), c.Message()}
	// 打标签，加额外信息
	event.Tags[TagErrorCode] = strconv.Itoa(c.Code())
	event.Extra[ExtraErrorMsg] = c.Message()
}
