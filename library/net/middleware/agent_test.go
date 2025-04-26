package middleware

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	httpUtil "gitlab.shanhai.int/sre/library/base/net"
)

func TestParseUserAgentMiddleware(t *testing.T) {
	t.Run("android", func(t *testing.T) {
		router := gin.New()
		router.Use(ParseUserAgentMiddleware())
		router.GET("/", func(c *gin.Context) {
			assert.Equal(t, ContextPhoneTypeAndroid, c.GetString(ContextPhoneTypeKey))
			assert.Equal(t, "8.4.2.0", c.GetString(ContextAppVersionKey))
		})

		header := make(http.Header)
		header.Add(HeaderUserAgentKey, "Android-QingtingFM QingTing-Android/8.4.2.0 Dalvik/2.1.0 (Linux; U; Android 9; MHA-AL00 Build/HUAWEIMHA-AL00)")
		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", header, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})

	t.Run("ios", func(t *testing.T) {
		router := gin.New()
		router.Use(ParseUserAgentMiddleware())
		router.GET("/", func(c *gin.Context) {
			assert.Equal(t, ContextPhoneTypeIOS, c.GetString(ContextPhoneTypeKey))
			assert.Equal(t, "8.4.1.3", c.GetString(ContextAppVersionKey))
		})

		header := make(http.Header)
		header.Add(HeaderUserAgentKey, "QingTing-iOS/8.4.1.3 com.Qting.QTTour Mozilla/5.0 (iPhone; CPU iPhone OS 12_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148")
		r, err := httpUtil.TestGinJsonRequest(router, "GET", "/", header, nil, nil)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
