package metric

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/base/ctime"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/middleware"
	"gitlab.shanhai.int/sre/library/net/response"
)

func TestPrometheusMiddleware(t *testing.T) {
	Init()

	t.Run("normal", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.GET("/", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.Use(middleware.TimeoutMiddleware(ctime.Duration(time.Second)))
		router.GET("/", func(c *gin.Context) {
			time.Sleep(1001 * time.Millisecond)
			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("response-normal", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.GET("/", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
			response.JSON(c, nil, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("response-error", func(t *testing.T) {
		router := gin.New()
		router.Use(PrometheusMiddleware())
		router.GET("/", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
			response.JSON(c, nil, errcode.InternalError)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
