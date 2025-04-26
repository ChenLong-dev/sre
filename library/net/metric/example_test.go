package metric

import "github.com/gin-gonic/gin"

func ExampleInit() {
	Init()

	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/metrics", GinMetricsHandler)
}
