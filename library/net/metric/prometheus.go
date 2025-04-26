package metric

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_context "gitlab.shanhai.int/sre/library/base/context"
	_gin "gitlab.shanhai.int/sre/library/net/gin"
)

// Web收集器
var WebCollector = []prometheus.Collector{
	WebRequestTotal, WebRequestDurationSummary, WebResponseTotal,
}

// DB收集器
var DBCollector = []prometheus.Collector{
	MongoRequestTotal, MongoRequestDurationSummary,
	RedisRequestTotal, RedisRequestDurationSummary,
	GormRequestTotal, GormRequestDurationSummary,
	RedlockRequestTotal,
}

// 其他收集器
var OtherCollector = []prometheus.Collector{
	HttpRequestTotal, HttpRequestDurationSummary, HttpResponseTotal,
	GoroutineRequestTotal, GoroutineRequestDurationSummary, GoroutineResponseTotal,
}

// 初始化
func Init() {
	collector := make([]prometheus.Collector, 0)
	collector = append(collector, WebCollector...)
	collector = append(collector, DBCollector...)
	collector = append(collector, OtherCollector...)

	CustomInit(collector...)
}

// 手动初始化
func CustomInit(cs ...prometheus.Collector) {
	prometheus.MustRegister(cs...)
}

// Gin统计指标处理器
func GinMetricsHandler(ctx *gin.Context) {
	promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
}

// 普罗米修斯中间件
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		relativePath := _gin.GetGinRelativePath(c)

		start := time.Now()
		WebRequestTotal.With(
			prometheus.Labels{
				"method": c.Request.Method,
				"url":    relativePath,
			},
		).Inc()

		c.Next()

		errCode := -1
		if value, ok := c.Get(_context.ContextErrCode); ok {
			if code, ok := value.(int); ok {
				errCode = code
			}
		}
		WebResponseTotal.With(prometheus.Labels{
			"method":      c.Request.Method,
			"url":         relativePath,
			"status_code": strconv.Itoa(c.Writer.Status()),
			"err_code":    strconv.Itoa(errCode),
		}).Inc()

		duration := time.Since(start)
		WebRequestDurationSummary.With(
			prometheus.Labels{
				"method": c.Request.Method,
				"url":    relativePath,
			},
		).Observe(duration.Seconds() * 1000)
	}
}
