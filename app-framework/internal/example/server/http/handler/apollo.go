package handler

import (
	"github.com/gin-gonic/gin"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	"gitlab.shanhai.int/sre/library/net/response"
)

// 获取apollo配置
func GetApolloConfig(c *gin.Context) {
	// 需要获取配置的的namespace
	namespace := c.Query("namespace")
	allConf := service.SVC.GetApolloConf(namespace)

	// 响应
	response.StandardJSON(c, allConf, nil)
}
