package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.shanhai.int/sre/library/net/trafficshaping"
)

// 流量控制中间件
// 当规则错误时，会panic
func TrafficShapingMiddleware(rules []*trafficshaping.Rule) gin.HandlerFunc {
	p, err := trafficshaping.NewPipeline(rules)
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		err := p.Do(func() {
			c.Next()
		})
		if err != nil {
			c.Writer.WriteHeader(http.StatusTooManyRequests)
			c.Abort()
		}
	}
}
