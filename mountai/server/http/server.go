package http

import (
	"github.com/gin-gonic/gin"

	framework "gitlab.shanhai.int/sre/app-framework"

	"rulai/server/http/handlers"
	handlerV2 "rulai/server/http/handlers/v2"

	"rulai/service"
	"rulai/utils"
)

// ====================
// >>>请勿删除<<<
//
// 获取http服务器
// ====================
func GetServer() framework.ServerInterface {
	svr := new(framework.HttpServer)

	svr.Middleware = func(e *gin.Engine) {
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
		addV1Router(e.Group("/api/v1"))
		addV2Router(e.Group("/api/v2"))

		addGrafanaV1Router(e.Group("/grafana/v1"))
	}

	return svr
}

// AddV1Router v1 router
func addV1Router(v1 *gin.RouterGroup) {
	v1.POST("/login", handlers.Login)

	addAuthV1Router(v1.Group("", utils.ParseJWTTokenMiddleware(service.K8sSystemUser),
		utils.ValidateInternalUserMiddleware(service.SVC)))
}

func addV2Router(v2 *gin.RouterGroup) {
	addAuthV2Router(v2.Group("", utils.ParseJWTTokenMiddleware(service.K8sSystemUser),
		utils.ValidateInternalUserMiddleware(service.SVC)))
}

func addAuthV2Router(authV2 *gin.RouterGroup) {
	addStatusV2Router(authV2.Group("/running_status"))
}

func addStatusV2Router(status *gin.RouterGroup) {
	status.GET("", handlerV2.GetRunningStatusList)
}

// AddAuthV1Router add auth v1 router
func addAuthV1Router(authV1 *gin.RouterGroup) {
	addProjectRouter(authV1.Group("/projects", handlers.CheckProject))
	addProjectLabelsRouter(authV1.Group("/project_labels"))
	addAppRouter(authV1.Group("/apps", handlers.CheckApp))
	addJobRouter(authV1.Group("/jobs"))
	addTaskRouter(authV1.Group("/tasks", handlers.CheckTask))
	addStatusRouter(authV1.Group("/running_status"))
	addTeamRouter(authV1.Group("/teams", handlers.CheckTeam))
	addUserRouter(authV1.Group("/users"))
	addActivityRouter(authV1.Group("/activities"))
	addGitRouter(authV1.Group("/git"))
	addLabelRouter(authV1.Group("/node_labels"))
	addResourceRouter(authV1.Group("/resources"))
	addFavProjectRouter(authV1.Group("/fav_projects"))
	addVariableRouter(authV1.Group("/variables", handlers.CheckVariable))
	addImageArgsTemplateRouter(authV1.Group("/image_args_templates", handlers.CheckImageArgsTemplate))
	addUserSubscriptionRouter(authV1.Group("/user_subscriptions"))
	addClusterRouter(authV1.Group("/clusters"))
	addNamespaceRouter(authV1.Group("/namespaces"))
	addConfigRenamePrefixesRouter(authV1.Group("/config_rename_prefixes"))
	addUpstreamV1Router(authV1.Group("/upstream"))
}

func addGrafanaV1Router(grafanaV1 *gin.RouterGroup) {
	grafanaV1.GET("/projects/:project_id/resources", handlers.GetProjectResources)
}

func addProjectRouter(projects *gin.RouterGroup) {
	projects.POST("", handlers.CreateProject)
	projects.GET("", handlers.GetProjects)
	projects.GET("/:project_id", handlers.GetProjectDetail)
	projects.GET("/:project_id/config", handlers.GetProjectConfig)
	projects.GET("/:project_id/resource", handlers.GetProjectResource)
	// TODO: 需要把 path 中的 members 改为 users，保持风格统一
	projects.GET("/:project_id/members/:user_id/role", handlers.CheckUser, handlers.GetProjectUserRole)
	projects.PUT("/:project_id", handlers.UpdateProject)
	projects.DELETE("/:project_id", handlers.DeleteProject)

	addProjectImageRouter(projects.Group("/:project_id/images"))
	addProjectResourceRouter(projects.Group("/:project_id/resources"))
	addProjectCIJobRouter(projects.Group("/:project_id/ci_job"))
	addProjectClusterRouter(projects.Group("/:project_id/clusters"))
	addClustersWithWorkloadRouter(projects.Group("/:project_id/clusters_with_workload"))
}

func addVariableRouter(variable *gin.RouterGroup) {
	variable.GET("", handlers.GetVariables)
	variable.PUT("/:variable_id", handlers.UpdateVariable)
	variable.DELETE("/:variable_id", handlers.DeleteVariable)
	variable.POST("", handlers.CreateVariable)
}

func addImageArgsTemplateRouter(imageArgsTemplate *gin.RouterGroup) {
	imageArgsTemplate.GET("", handlers.GetImageArgsTemplates)
	imageArgsTemplate.PUT("/:image_args_template_id", handlers.UpdateImageArgsTemplate)
	imageArgsTemplate.DELETE("/:image_args_template_id", handlers.DeleteImageArgsTemplate)
	imageArgsTemplate.POST("", handlers.CreateImageArgsTemplate)
}

func addProjectResourceRouter(resource *gin.RouterGroup) {
	resource.GET("", handlers.GetProjectResources)
	resource.PUT("", handlers.UpdateProjectResources)
}

func addProjectImageRouter(image *gin.RouterGroup) {
	image.GET("/last_args", handlers.GetLastImageArgs)

	addProjectImageJobRouter(image.Group("/jobs"))
	addProjectImageTagRouter(image.Group("/tags"))
}

func addProjectImageJobRouter(job *gin.RouterGroup) {
	job.POST("", handlers.CreateImageJob)
	job.GET("", handlers.GetImageJobs)
	job.GET("/:build_id", handlers.GetImageJobDetail)
	job.DELETE("/:build_id", handlers.DeleteImageJob)
	job.POST("/:build_id", handlers.CacheImageJob)
}

func addProjectImageTagRouter(tag *gin.RouterGroup) {
	tag.GET("", handlers.GetImageTags)
}

func addProjectCIJobRouter(ciJob *gin.RouterGroup) {
	ciJob.GET("", handlers.GetProjectCIJob)
	ciJob.POST("", handlers.CreateProjectCIJob)
	ciJob.PUT("", handlers.UpdateProjectCIJob)
}

func addProjectClusterRouter(cluster *gin.RouterGroup) {
	cluster.GET("", handlers.GetProjectSupportedClusters)
}

func addClustersWithWorkloadRouter(cluster *gin.RouterGroup) {
	cluster.GET("", handlers.GetProjectAppsClustersWithWorkload)
}

func addProjectLabelsRouter(label *gin.RouterGroup) {
	label.GET("", handlers.GetProjectLabels)
}

// AddAppRouter app router
func addAppRouter(apps *gin.RouterGroup) {
	apps.POST("", handlers.CreateApp)
	apps.GET("", handlers.GetApps)
	apps.GET("/:id", handlers.GetAppDetail)
	apps.PUT("/:id", handlers.UpdateApp)
	apps.DELETE("/:id", handlers.DeleteApp)
	apps.POST("/:id/correct_name", handlers.CorrectAppName)
	apps.GET("/:id/tips", handlers.GetAppTips)
	apps.POST("/:id/sentry", handlers.CreateAppSentry)
	apps.POST("/:id/cluster_weights", handlers.SetAppClusterWeights)
	apps.GET("/:id/cluster_weights", handlers.GetAppClusterWeights)
	apps.GET("/:id/clusters_with_workload", handlers.GetAppClustersWithWorkload)
}

// AddAppRouter app router
func addJobRouter(apps *gin.RouterGroup) {
	apps.DELETE("", handlers.DeleteJob)
}

func addTaskRouter(task *gin.RouterGroup) {
	task.POST("", handlers.CreateTask)
	task.POST("/batch", handlers.BatchCreateTask)
	task.PUT("/:id", handlers.UpdateTask)
	task.GET("", handlers.GetTasks)
	task.DELETE("/:id", handlers.DeleteTask)

	// 避免路由冲突
	task.GET("/:id", func(c *gin.Context) {
		switch c.Param("id") {
		case "latest":
			handlers.GetLatestDeployTaskDetail(c)
		default:
			handlers.GetTaskDetail(c)
		}
	})
}

func addStatusRouter(status *gin.RouterGroup) {
	status.GET("", handlers.GetRunningStatusList)
	status.GET("/:version", handlers.GetRunningStatusDetail)
	status.GET("/:version/description", handlers.GetRunningStatusDescription)
	status.GET("/:version/pods/:pod_name/logs", handlers.GetRunningPodLogs)
	status.POST("/:version/pods/:pod_name/pprof", handlers.CreateRunningPodPProf)
	status.GET("/:version/pods/:pod_name/description", handlers.GetRunningPodDescription)
}

func addFavProjectRouter(favProject *gin.RouterGroup) {
	favProject.GET("", handlers.GetFavProjects)
	favProject.POST("", handlers.CreateFavProject)
	favProject.DELETE("/:projectID", handlers.DeleteFavProject)
}

func addTeamRouter(team *gin.RouterGroup) {
	team.POST("", handlers.CreateTeam)
	team.GET("", handlers.GetTeams)
	team.GET("/:id", handlers.GetTeamDetail)
	team.PUT("/:id", handlers.UpdateTeam)
	team.DELETE("/:id", handlers.DeleteTeam)
}

func addUserRouter(user *gin.RouterGroup) {
	user.GET("", handlers.GetUsers)
	user.GET("/:user_id/used_projects", handlers.CheckUser, handlers.GetRecentlyProjects)
}

func addActivityRouter(activity *gin.RouterGroup) {
	activity.GET("", handlers.GetActivities)
}

func addGitRouter(git *gin.RouterGroup) {
	git.GET("/projects/:id", handlers.GetGitProjectDetail)
	git.GET("/projects/:id/branch", handlers.GetGitProjectBranches)
}

func addResourceRouter(resource *gin.RouterGroup) {
	resource.GET("", handlers.GetResources)
}

func addLabelRouter(label *gin.RouterGroup) {
	label.GET("", handlers.GetNodeLabelLists)
}

func addUserSubscriptionRouter(subscription *gin.RouterGroup) {
	subscription.POST("", handlers.UserSubscribe)
	subscription.DELETE("", handlers.UserUnsubscribe)
}

func addClusterRouter(cluster *gin.RouterGroup) {
	cluster.GET("", handlers.GetClusters)
}

func addNamespaceRouter(ns *gin.RouterGroup) {
	addNamespaceClusterRouter(ns.Group("/:namespace/clusters"))
}

func addNamespaceClusterRouter(cluster *gin.RouterGroup) {
	cluster.GET("/:name", handlers.GetClusterDetail)
}

func addConfigRenamePrefixesRouter(crp *gin.RouterGroup) {
	crp.POST("", handlers.CreateConfigRenamePrefix)
	crp.DELETE("/:prefix", handlers.DeleteConfigRenamePrefix)
}

func addUpstreamV1Router(upstream *gin.RouterGroup) {
	upstream.POST("/info", handlers.DetermineBackendHost)
}
