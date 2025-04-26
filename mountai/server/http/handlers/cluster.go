package handlers

import (
	"strings"

	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"gitlab.shanhai.int/sre/library/net/errcode"
)

// GetClusterDetail 获取集群详情
func GetClusterDetail(c *gin.Context) {
	name := entity.ClusterName(c.Param("name"))
	namespace := c.Param("namespace")
	res, err := service.SVC.GetClusterDetail(c, name, namespace)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, res, nil)
}

// GetClusters 获取集群列表
// 数量不会多，不需要分页参数
func GetClusters(c *gin.Context) {
	getReq := new(req.GetClustersReq)
	err := c.ShouldBindQuery(getReq)
	if err != nil {
		response.JSON(c, nil, errors.Wrap(errcode.InvalidParams, err.Error()))
		return
	}

	// 处理 names 列表
	splitedNames := strings.Split(getReq.Names, ",")
	getReq.ClusterNames = make([]entity.ClusterName, len(splitedNames))
	for i := range splitedNames {
		getReq.ClusterNames[i] = entity.ClusterName(splitedNames[i])
	}

	clusters, err := service.SVC.GetClusters(c, getReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{List: clusters}, nil)
}

// GetProjectSupportedClusters 获取项目支持的集群列表
// 数量不会多，不需要分页参数
func GetProjectSupportedClusters(c *gin.Context) {
	id := c.Param("project_id")

	clusters, err := service.SVC.GetProjectSupportedClusters(c, id)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, models.BaseListResponse{List: clusters}, nil)
}
