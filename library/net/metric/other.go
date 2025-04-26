package metric

import "github.com/prometheus/client_golang/prometheus"

// 总请求数量
var HttpRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "httpclient_request_total",
	},
	[]string{"web_url", "web_method", "host", "method_name"},
)

// 请求时间百分位图
var HttpRequestDurationSummary = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "httpclient_request_duration_millisecond_summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.05, 0.95: 0.005, 0.99: 0.005},
	},
	[]string{"host", "method_name"},
)

// 总响应数量
var HttpResponseTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "httpclient_response_total",
	},
	[]string{"web_url", "web_method", "host", "method_name", "status_code"},
)

// 总请求数量
var GoroutineRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "goroutine_request_total",
	},
	[]string{"web_url", "web_method", "group_name"},
)

// 请求时间百分位图
var GoroutineRequestDurationSummary = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name:       "goroutine_request_duration_millisecond_summary",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.05, 0.95: 0.005, 0.99: 0.005},
	},
	[]string{"group_name", "state"},
)

// 总响应数量
var GoroutineResponseTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "goroutine_response_total",
	},
	[]string{"web_url", "web_method", "group_name", "state"},
)
