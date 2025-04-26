package metric

import "github.com/prometheus/client_golang/prometheus"

// 总请求数量
var MongoRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "mongo_request_total",
	},
	[]string{"web_url", "web_method", "func_name", "db_name", "collection_name"},
)

// 请求时间百分位图
var MongoRequestDurationSummary = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "mongo_request_duration_millisecond_summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.05, 0.95: 0.005, 0.99: 0.005},
	},
	[]string{"func_name", "db_name", "collection_name"},
)

// 总请求数量
var RedisRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redis_request_total",
	},
	[]string{"web_url", "web_method", "endpoint"},
)

// 请求时间百分位图
var RedisRequestDurationSummary = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "redis_request_duration_millisecond_summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.05, 0.95: 0.005, 0.99: 0.005},
	},
	[]string{"endpoint"},
)

// 总请求数量
var GormRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "gorm_request_total",
	},
	[]string{"web_url", "web_method", "dsn", "operation"},
)

// 请求时间百分位图
var GormRequestDurationSummary = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "gorm_request_duration_millisecond_summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.05, 0.95: 0.005, 0.99: 0.005},
	},
	[]string{"dsn", "operation"},
)

// 总请求数量
var RedlockRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redlock_request_total",
	},
	[]string{"web_url", "web_method"},
)
