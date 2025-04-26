package middleware

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func TestCatchPanicMiddleware(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		router := gin.New()
		router.Use(CatchPanicMiddleware())
		router.GET("/", func(c *gin.Context) {
			panic(errors.New("this is a panic"))
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})

	t.Run("string", func(t *testing.T) {
		router := gin.New()
		router.Use(CatchPanicMiddleware())
		router.GET("/", func(c *gin.Context) {
			panic("this is a panic")
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})

	t.Run("middleware panic", func(t *testing.T) {
		router := gin.New()
		router.Use(CatchPanicMiddleware())
		router.Use(func(c *gin.Context) {
			panic("this is a panic")
		})

		router.GET("/", func(c *gin.Context) {
			assert.FailNow(t, "not pass through here")
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})
}
