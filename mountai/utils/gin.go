package utils

import (
	"bytes"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

// Check whether body is empty
func IsRequestBodyNotEmpty(c *gin.Context) (bool, error) {
	bodyData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return false, err
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyData))

	return len(bodyData) > 0, nil
}
