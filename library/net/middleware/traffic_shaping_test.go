package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
	"gitlab.shanhai.int/sre/library/net/trafficshaping"
)

func TestTrafficShapingMiddleware(t *testing.T) {
	t.Run("reject", func(t *testing.T) {
		router := gin.New()
		router.Use(TrafficShapingMiddleware([]*trafficshaping.Rule{
			{
				Type:            trafficshaping.QPS,
				ControlBehavior: trafficshaping.Reject,
				Limit:           10,
			},
			{
				Type:            trafficshaping.Concurrency,
				ControlBehavior: trafficshaping.Reject,
				Limit:           1,
			},
		}))

		var lastPassTime int64
		router.GET("/", func(c *gin.Context) {
			if !atomic.CompareAndSwapInt64(&lastPassTime, 0, time.Now().UnixNano()) {
				assert.Greater(t, time.Now().UnixNano()-atomic.LoadInt64(&lastPassTime), int64(500*time.Millisecond))
			}

			time.Sleep(500 * time.Millisecond)
			fmt.Println(time.Now().Format("2006-01-02 15:04:05.000"), "passed")
			c.JSON(http.StatusOK, nil)
		})

		wg := new(sync.WaitGroup)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
					assert.Nil(t, err)
					assert.Contains(t, []int{http.StatusOK, http.StatusTooManyRequests}, r.Code)
				}
			}(i)
		}
		wg.Wait()
	})

	t.Run("waiting", func(t *testing.T) {
		router := gin.New()
		router.Use(TrafficShapingMiddleware([]*trafficshaping.Rule{
			{
				Type:            trafficshaping.QPS,
				ControlBehavior: trafficshaping.Waiting,
				Limit:           2,
				MaxWaitingTime:  time.Second * 3,
			},
		}))

		var lastPassTime int64
		router.GET("/", func(c *gin.Context) {
			if !atomic.CompareAndSwapInt64(&lastPassTime, 0, time.Now().UnixNano()) {
				assert.Greater(t, time.Now().UnixNano()-atomic.LoadInt64(&lastPassTime), int64(500*time.Millisecond))
			}

			time.Sleep(500 * time.Millisecond)
			fmt.Println(time.Now().Format("2006-01-02 15:04:05.000"), "passed")
			c.JSON(http.StatusOK, nil)
		})

		wg := new(sync.WaitGroup)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for j := 0; j < 2; j++ {
					r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
					assert.Nil(t, err)
					assert.Contains(t, []int{http.StatusOK, http.StatusTooManyRequests}, r.Code)
				}
			}(i)
		}
		wg.Wait()

	})
}
