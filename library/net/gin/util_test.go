package gin

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func Test_getGinRelativePath(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		router := gin.New()
		router.GET("/:id", func(c *gin.Context) {
			path := GetGinRelativePath(c)
			assert.Equal(t, "/:id", path)
			c.JSON(http.StatusOK, nil)
		})

		_, err := httpUtil.TestGinJsonRequest(router, "GET", "/test", nil, nil, nil)
		assert.Nil(t, err)
	})

	t.Run("unknown", func(t *testing.T) {
		router := gin.New()
		router.NoRoute(func(c *gin.Context) {
			path := GetGinRelativePath(c)
			assert.Equal(t, "unknown", path)
		})
		router.GET("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, nil)
		})

		_, err := httpUtil.TestGinJsonRequest(router, "GET", "/test/123", nil, nil, nil)
		assert.Nil(t, err)
	})
}
