package middleware

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	_context "gitlab.shanhai.int/sre/library/base/context"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func TestSetDefaultContextValueMiddleware(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		router := gin.New()
		router.Use(SetDefaultContextValueMiddleware())
		router.GET("/", func(c *gin.Context) {
			assert.NotEmpty(t, c.GetString(_context.ContextUUIDKey))
			assert.NotEmpty(t, c.GetString(_context.ContextRequestMethodKey))
			assert.NotEmpty(t, c.GetString(_context.ContextRequestPathKey))
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestGetUUIDMiddleware(t *testing.T) {
	t.Run("header", func(t *testing.T) {
		router := gin.New()
		router.Use(GetUUIDMiddleware("QT-User-ID"))
		router.GET("/", func(c *gin.Context) {
			assert.Equal(t, "123456", c.GetString(_context.ContextUUIDKey))
		})

		header := make(http.Header)
		header.Add("QT-User-ID", "123456")
		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", header, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("query", func(t *testing.T) {
		router := gin.New()
		router.Use(GetUUIDMiddleware("QT-User-ID"))
		router.GET("/", func(c *gin.Context) {
			assert.Equal(t, "123456", c.GetString(_context.ContextUUIDKey))
		})

		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/?QT-User-ID=123456", nil, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
