// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package gojenkins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// DefaultCrowdCheckPath 默认crowd验证路径
	DefaultCrowdCheckPath = "/j_acegi_security_check"
)

// Request Methods

type APIRequest struct {
	Method   string
	Endpoint string
	Payload  io.Reader
	Headers  http.Header
	Suffix   string
}

func (ar *APIRequest) SetHeader(key string, value string) *APIRequest {
	ar.Headers.Set(key, value)
	return ar
}

func NewAPIRequest(method string, endpoint string, payload io.Reader) *APIRequest {
	var headers = http.Header{}
	var suffix string
	ar := &APIRequest{method, endpoint, payload, headers, suffix}
	return ar
}

type CrowdAuth struct {
	AuthKey   string
	AuthValue string

	// crowd读写锁
	rwMutex sync.RWMutex
}

type Requester struct {
	Client    *http.Client
	CACert    []byte
	SslVerify bool
	CrowdAuth CrowdAuth
	Config    *Config
}

func (r *Requester) SetCrumb(ar *APIRequest) error {
	crumbData := map[string]string{}
	response, _ := r.GetJSON("/crumbIssuer/api/json", &crumbData, nil)

	if response.StatusCode == 200 && crumbData["crumbRequestField"] != "" {
		ar.SetHeader(crumbData["crumbRequestField"], crumbData["crumb"])
		ar.SetHeader("Cookie", response.Header.Get("set-cookie"))
	}

	return nil
}

func (r *Requester) PostJSON(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = "api/json"
	return r.Do(ar, &responseStruct, querystring)
}

func (r *Requester) Post(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	ar.Suffix = ""
	return r.Do(ar, &responseStruct, querystring)
}

func (r *Requester) PostFiles(endpoint string, payload io.Reader, responseStruct interface{}, querystring map[string]string, files []string) (*http.Response, error) {
	ar := NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ar); err != nil {
		return nil, err
	}
	return r.Do(ar, &responseStruct, querystring, files)
}

func (r *Requester) PostXML(endpoint string, xml string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	payload := bytes.NewBuffer([]byte(xml))
	ar := NewAPIRequest("POST", endpoint, payload)
	if err := r.SetCrumb(ar); err != nil {
		return nil, err
	}
	ar.SetHeader("Content-Type", "application/xml")
	ar.Suffix = ""
	return r.Do(ar, &responseStruct, querystring)
}

func (r *Requester) GetJSON(endpoint string, responseStruct interface{}, query map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", endpoint, nil)
	ar.SetHeader("Content-Type", "application/json")
	ar.Suffix = "api/json"
	return r.Do(ar, &responseStruct, query)
}

func (r *Requester) GetXML(endpoint string, responseStruct interface{}, query map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", endpoint, nil)
	ar.SetHeader("Content-Type", "application/xml")
	ar.Suffix = ""
	return r.Do(ar, responseStruct, query)
}

func (r *Requester) Get(endpoint string, responseStruct interface{}, querystring map[string]string) (*http.Response, error) {
	ar := NewAPIRequest("GET", endpoint, nil)
	ar.Suffix = ""
	return r.Do(ar, responseStruct, querystring)
}

func (r *Requester) SetClient(client *http.Client) *Requester {
	r.Client = client
	return r
}

//Add auth on redirect if required.
func (r *Requester) redirectPolicyFunc(req *http.Request, via []*http.Request) error {
	if r.Config.BasicAuth != nil {
		req.SetBasicAuth(r.Config.BasicAuth.UserName, r.Config.BasicAuth.Password)
	}
	return nil
}

func (r *Requester) Do(ar *APIRequest, responseStruct interface{}, options ...interface{}) (*http.Response, error) {
	if !strings.HasSuffix(ar.Endpoint, "/") && ar.Method != "POST" {
		ar.Endpoint += "/"
	}

	fileUpload := false
	var files []string
	URL, err := url.Parse(r.Config.BaseURL + ar.Endpoint + ar.Suffix)
	if err != nil {
		return nil, err
	}

	for _, o := range options {
		switch v := o.(type) {
		case map[string]string:

			querystring := make(url.Values)
			for key, val := range v {
				querystring.Set(key, val)
			}

			URL.RawQuery = querystring.Encode()
		case []string:
			fileUpload = true
			files = v
		}
	}
	var req *http.Request

	if fileUpload {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for _, file := range files {
			fileData, err := os.Open(file)
			if err != nil {
				Error.Println(err.Error())
				return nil, err
			}

			part, err := writer.CreateFormFile("file", filepath.Base(file))
			if err != nil {
				Error.Println(err.Error())
				return nil, err
			}
			if _, err = io.Copy(part, fileData); err != nil {
				return nil, err
			}
			defer fileData.Close()
		}
		var params map[string]string
		json.NewDecoder(ar.Payload).Decode(&params)
		for key, val := range params {
			if err = writer.WriteField(key, val); err != nil {
				return nil, err
			}
		}
		if err = writer.Close(); err != nil {
			return nil, err
		}
		req, err = http.NewRequest(ar.Method, URL.String(), body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		req, err = http.NewRequest(ar.Method, URL.String(), ar.Payload)
		if err != nil {
			return nil, err
		}
	}

	if r.Config.BasicAuth != nil {
		req.SetBasicAuth(r.Config.BasicAuth.UserName, r.Config.BasicAuth.Password)
	}

	for k := range ar.Headers {
		req.Header.Add(k, ar.Headers.Get(k))
	}

	isRetry := false
	for {
		// 添加crowd
		err = r.addCrowdCookies(req)
		if err != nil {
			return nil, err
		}
		response, err := r.Client.Do(req)
		if err != nil {
			return nil, err
		}
		errorText := response.Header.Get("X-Error")
		if errorText != "" {
			return nil, errors.New(errorText)
		}

		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		// ioutil.ReadAll会清空reader
		response.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		// cookie过期，重试，只允许重试一次
		if r.Config.Crowd != nil && response.StatusCode == http.StatusForbidden &&
			strings.Contains(string(bodyBytes), r.Config.Crowd.IdentifyContent) && !isRetry {
			if err := r.freshCrowdAuth(1); err != nil {
				return nil, err
			}
			isRetry = true
			continue
		}

		switch responseStruct.(type) {
		case *string:
			return r.ReadRawResponse(response, responseStruct)
		default:
			if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
				return r.ReadJSONResponse(response, responseStruct)
			} else {
				return response, nil
			}
		}
	}
}

func (r *Requester) ReadRawResponse(response *http.Response, responseStruct interface{}) (*http.Response, error) {
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if str, ok := responseStruct.(*string); ok {
		*str = string(content)
	} else {
		return nil, fmt.Errorf("Could not cast responseStruct to *string")
	}

	return response, nil
}

func (r *Requester) ReadJSONResponse(response *http.Response, responseStruct interface{}) (*http.Response, error) {
	defer response.Body.Close()

	err := json.NewDecoder(response.Body).Decode(responseStruct)
	return response, err
}

// 获取并设置新的cookie
func (r *Requester) freshCrowdAuth(retryCount int) error {
	var rawReq http.Request
	if err := rawReq.ParseForm(); err != nil {
		return err
	}
	formData := map[string]string{
		"Submit":     "Sign+in",
		"j_username": r.Config.Crowd.UserName,
		"j_password": r.Config.Crowd.Password,
		"from":       "/",
	}
	for key, value := range formData {
		rawReq.Form.Add(key, value)
	}
	bodyStr := strings.TrimSpace(rawReq.Form.Encode())
	checkPath := r.Config.Crowd.CheckPath
	if checkPath == "" {
		checkPath = DefaultCrowdCheckPath
	}
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s%s", r.Config.BaseURL, checkPath),
		strings.NewReader(bodyStr),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if r.Config.BasicAuth != nil {
		req.SetBasicAuth(r.Config.BasicAuth.UserName, r.Config.BasicAuth.Password)
	}
	err = r.addCrowdCookies(req)
	if err != nil {
		return err
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			resp := req.Response
			if resp == nil || resp.StatusCode != http.StatusFound {
				return http.ErrUseLastResponse
			}

			// 302重定向取最后一个cookie并重新请求
			err = r.getCrowdCookie(resp)
			if err != nil {
				return err
			}
			err = r.addCrowdCookies(req)
			if err != nil {
				return err
			}

			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// 更新cookie
	err = r.getCrowdCookie(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		if retryCount > 0 {
			return r.freshCrowdAuth(0)
		}
		return fmt.Errorf("crowd auth wrong status code: %v", resp.StatusCode)
	}
	return nil
}

func (r *Requester) addCrowdCookies(req *http.Request) error {
	if r.CrowdAuth.AuthKey != "" && r.CrowdAuth.AuthValue != "" {
		r.CrowdAuth.rwMutex.RLock()
		req.AddCookie(&http.Cookie{
			Name:  r.CrowdAuth.AuthKey,
			Value: r.CrowdAuth.AuthValue,
		})
		r.CrowdAuth.rwMutex.RUnlock()
	}
	return nil
}

func (r *Requester) getCrowdCookie(resp *http.Response) error {
	cookies := resp.Cookies()
	if len(cookies) > 0 {
		r.CrowdAuth.rwMutex.Lock()
		r.CrowdAuth.AuthKey = cookies[len(cookies)-1].Name
		r.CrowdAuth.AuthValue = cookies[len(cookies)-1].Value
		r.CrowdAuth.rwMutex.Unlock()
	}
	return nil
}
