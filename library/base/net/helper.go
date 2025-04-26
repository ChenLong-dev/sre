package net

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// 用于进行gin请求测试
func TestGinJsonRequest(e *gin.Engine, method, path string, headers http.Header, requestBody interface{}, requestForm url.Values) (*httptest.ResponseRecorder, error) {
	var bodyReader io.Reader
	if requestBody != nil {
		requestJSON, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(requestJSON)
	}

	req := httptest.NewRequest(method, path, bodyReader)

	if headers == nil {
		headers = make(http.Header)
	}
	req.Header = headers

	if requestForm == nil {
		requestForm = make(url.Values)
	}
	req.PostForm = requestForm

	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)

	return w, nil
}
