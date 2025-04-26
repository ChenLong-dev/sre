package middleware

import (
	"github.com/gin-gonic/gin"
	"regexp"
)

const (
	HeaderUserAgentKey = "User-Agent"

	ContextPhoneTypeKey  = "phone_type"
	ContextAppVersionKey = "app_version"

	ContextPhoneTypeAndroid = "android"
	ContextPhoneTypeIOS     = "ios"
)

var (
	UserAgentRegex = regexp.MustCompile(`QingTing-(iOS|Android)(-WV)?\/(\d+\.\d+\.\d+(.\d+)?)`)
)

// 解析UserAgent，并将相关信息装入context中
func ParseUserAgentMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userAgent := c.Request.Header.Get(HeaderUserAgentKey)
		userAgentMatch := UserAgentRegex.FindStringSubmatch(userAgent)

		c.Set(ContextPhoneTypeKey, ContextPhoneTypeAndroid)
		if len(userAgentMatch) > 1 && userAgentMatch[1] == "iOS" {
			c.Set(ContextPhoneTypeKey, ContextPhoneTypeIOS)
		}
		if len(userAgentMatch) > 3 {
			c.Set(ContextAppVersionKey, userAgentMatch[3])
		}

		c.Next()
	}
}
