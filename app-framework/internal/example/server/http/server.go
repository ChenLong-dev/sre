package http

import (
	"github.com/gin-gonic/gin"
	"gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/app-framework/internal/example/server/http/handler"
	"gitlab.shanhai.int/sre/library/net/middleware"
)

// ====================
// >>>请勿删除<<<
//
// 获取http服务器
// ====================
func GetServer() framework.ServerInterface {
	svr := new(framework.HttpServer)

	svr.Middleware = func(e *gin.Engine) {
		// ====================
		// 根据实际情况，选择性添加
		// ====================
		// 设置中间件
		e.Use(middleware.ParseUserAgentMiddleware())
	}

	// ====================
	// >>>请勿删除<<<
	//
	// 配置路由
	//
	// 健康检查及数据统计接口默认已实现，可通过配置文件改变接口url，默认url分别为
	//	/health 及 /metrics
	// ====================
	svr.Router = func(e *gin.Engine) {
		// ====================
		// 根据实际情况，选择性添加
		// ====================

		// 设置路由
		v1API := e.Group("/v1/api")
		{
			vipItems := v1API.Group("/vip_items")
			{
				// ==========================
				// 获取vip商品列表
				//
				// 包含以下使用示例:
				// mongo数据库;
				// deepcopy拷贝模型;
				// error正确处理;
				// ==========================
				vipItems.GET("", handler.GetVipItemsList)
			}

			vipRecommendItem := v1API.Group("/vip_recommend")
			{
				// ==========================
				// 获取vip推荐商品详情
				//
				// 包含以下使用示例:
				// mysql数据库;
				// mysql事务处理;
				// redis缓存;
				// deepcopy拷贝模型;
				// error正确处理;
				// ==========================
				vipRecommendItem.GET("/:item_id", handler.GetVipRecommendItemDetail)
			}

			vipUser := v1API.Group("/vip_user")
			{
				// ==========================
				// 获取vip用户信息
				//
				// 包含以下使用示例:
				// mongo数据库;
				// deepcopy拷贝模型;
				// error正确处理;
				// ==========================
				vipUser.GET("/:user_id", handler.GetVipUserDetail)
			}

			mediaSource := v1API.Group("/media_source")
			{
				// ==========================
				// 创建媒体资源
				//
				// 包含以下使用示例:
				// kafka发送消息;
				// deepcopy拷贝模型;
				// error正确处理;
				// ==========================
				mediaSource.POST("", handler.CreateMediaSource)
			}

			vipRefund := v1API.Group("/vip_refund")
			{
				// ==========================
				// 创建会员退款
				//
				// 包含以下使用示例:
				// amqp发送消息;
				// deepcopy拷贝模型;
				// error正确处理;
				// ==========================
				vipRefund.POST("", handler.CreateVipRefund)
			}

			userAuth := v1API.Group("/auth")
			{
				// ==========================
				// 用户鉴权
				//
				// 包含以下使用示例:
				// http客户端请求;
				// deepcopy拷贝模型;
				// error正确处理;
				// ==========================
				userAuth.POST("/verify", handler.VerifyUser)
			}

			restful := v1API.Group("/restful")
			{
				// ==========================
				// 调用restful接口
				//
				// 包含以下使用示例:
				// 多协程并行处理;
				// http客户端请求;
				// deepcopy拷贝模型;
				// error正确处理;
				// ==========================
				restful.POST("invoke", handler.InvokeRestful)
			}

			apollo := v1API.Group("/apollo")
			{
				// ==========================
				// 获取apollo配置
				//
				// 包含以下使用示例:
				// 获取apollo配置;
				// ==========================
				apollo.GET("/config", handler.GetApolloConfig)
			}
		}

		// 设置路由
		v2API := e.Group("/v2/api")
		{
			github := v2API.Group("/github")
			{
				// ==========================
				// 获取聚合github用户数据
				//
				// 包含以下使用示例:
				// 多协程并行处理;
				// redlock分布式锁;
				// 服务降级;
				// http客户端请求;
				// ==========================
				github.GET("/aggregation/:owner", handler.GetOwnerGithubAggregation)

				// ==========================
				// 获取github用户仓库
				//
				// 包含以下使用示例:
				// 校验器校验请求参数;
				// http客户端请求;
				// ==========================
				github.GET("/repos/:owner", handler.GetOwnerGithubRepos)
			}
		}
	}

	return svr
}
