package resp

type AliGetDockerTagsResp struct {
	Data *struct {
		Total    int `json:"total"`
		PageSize int `json:"pageSize"`
		Page     int `json:"page"`
		Tags     []*struct {
			ImageUpdate int    `json:"imageUpdate"`
			ImageCreate int    `json:"imageCreate"`
			ImageID     string `json:"imageId"`
			Digest      string `json:"digest"`
			ImageSize   int    `json:"imageSize"`
			Tag         string `json:"tag"`
			Status      string `json:"status"`
		} `json:"tags"`
	} `json:"data"`
	RequestID string `json:"requestId"`
}
