package cm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"gitlab.shanhai.int/sre/library/net/errcode"
)

type ConfigManagerResponse struct {
	Code    int    `json:"errcode"`
	Message string `json:"errmsg"`
}

// 获取源文件
func (c *Client) GetOriginFile(path string, commitID string) ([]byte, error) {
	response, err := http.Get(fmt.Sprintf("%s%s?path=%s&commitID=%s",
		c.Config.Host, c.Config.FileAPI, path, commitID))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		errResp := new(ConfigManagerResponse)
		err = json.Unmarshal(body, errResp)
		if err != nil {
			return nil, err
		}
		return nil, errcode.GetOrNewErrCode(errResp.Code, errResp.Message)
	}

	return body, nil
}
