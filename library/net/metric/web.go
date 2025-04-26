package metric

import "github.com/prometheus/client_golang/prometheus"

// 总请求数量
var WebRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "web_request_total",
	},
	[]string{"method", "url"},
)

// 请求时间百分位图
var WebRequestDurationSummary = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "web_request_duration_millisecond_summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.05, 0.95: 0.005, 0.99: 0.005},
	},
	[]string{"method", "url"},
)

// 总响应数量
var WebResponseTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "web_response_total",
	},
	[]string{"method", "url", "status_code", "err_code"},
)
