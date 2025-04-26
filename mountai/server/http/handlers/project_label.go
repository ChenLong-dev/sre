package handlers

import (
	"rulai/service"
	"rulai/utils/response"

	"github.com/gin-gonic/gin"
)

func GetProjectLabels(c *gin.Context) {
	labels, err := service.SVC.GetProjectLabels(c)
	if err != nil {
		response.JSON(c, nil, err)
		return
	}

	response.JSON(c, labels, nil)
}
