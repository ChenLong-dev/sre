package handlers

import (
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func GetResources(c *gin.Context) {
	getReq := new(req.GetResourcesReq)
	err := c.ShouldBindQuery(getReq)

	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	providers := make([]entity.ProviderType, 0)
	if getReq.Provider == "" {
		providers = entity.AllProviderType
	} else {
		providers = append(providers, entity.ProviderType(getReq.Provider))
	}

	instances := make([]*entity.ResourceInstance, 0)

	for _, provider := range providers {
		res, err := service.SVC.GetResourceListFromCache(c, provider, entity.ResourceType(getReq.Type))
		if err != nil {
			response.JSON(c, nil, err)
			return
		}

		instances = append(instances, res...)
	}

	response.JSON(c, map[string]interface{}{
		"list":  instances,
		"count": len(instances),
	}, nil)
}

func GetProjectResources(c *gin.Context) {
	projectID := c.Param("project_id")
	getReq := new(req.GetProjectResourceReq)
	err := c.ShouldBindQuery(getReq)

	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	res, err := service.SVC.GetProjectResources(c, projectID, getReq.EnvName)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")

	response.JSON(c, res, nil)
}

func UpdateProjectResources(c *gin.Context) {
	projectID := c.Param("project_id")
	updateReq := new(req.UpdateProjectResourcesReq)
	err := c.ShouldBindJSON(updateReq)

	if err != nil {
		response.JSON(c, nil, errors.Wrapf(errcode.InvalidParams, "%s", err))
		return
	}

	err = service.SVC.UpdateProjectResources(c, projectID, updateReq)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, updateReq, nil)
}
