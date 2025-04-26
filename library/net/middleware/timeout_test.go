package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
	"gitlab.shanhai.int/sre/library/net/errcode"
	_gin "gitlab.shanhai.int/sre/library/net/gin"
	"gitlab.shanhai.int/sre/library/net/response"
)

func TestTimeoutMiddleware(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(ctime.Duration(time.Second)))
		router.GET("/", func(c *gin.Context) {
			time.Sleep(500 * time.Millisecond)
			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("timeout", func(t *testing.T) {
		router := gin.New()
		router.Use(TimeoutMiddleware(ctime.Duration(time.Second)))
		router.GET("/", func(c *gin.Context) {
			time.Sleep(1001 * time.Millisecond)
			response.JSON(c, nil, errcode.UnknownError)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusGatewayTimeout, r.Code)
	})

	t.Run("seq", func(t *testing.T) {
		router := gin.New()
		router.Use(_gin.GetDefaultFormatter(&_gin.Config{
			Config: &render.Config{
				Stdout: true,
			},
			RequestBodyOut: true,
		}))
		router.Use(TimeoutMiddleware(ctime.Duration(time.Second)))
		router.GET("/", func(c *gin.Context) {
			time.Sleep(1001 * time.Millisecond)
			c.JSON(http.StatusOK, nil)
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
