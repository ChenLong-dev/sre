package gin

import (
	"github.com/gin-gonic/gin"
)

// 获取gin抽象路由
func GetGinRelativePath(c *gin.Context) string {
	path := c.FullPath()
	if path == "" {
		return "unknown"
	}
	return path
}
