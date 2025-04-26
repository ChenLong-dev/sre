package resp

type GetAppConfigResp struct {
	Code    int                     `json:"errcode"`
	Message string                  `json:"errmsg"`
	Data    *GetAppConfigDetailResp `json:"data,omitempty"`
}

type GetAppConfigDetailResp struct {
	CommitID string      `json:"commit_id"`
	Config   interface{} `json:"config"`
}

type GetProjectResourceFromConfigResp struct {
	Kafka []*AliProjectResourceResp `json:"kafka"`
	AMQP  []*AliProjectResourceResp `json:"amqp"`
	Etcd  []*AliProjectResourceResp `json:"etcd"`
	Mysql []*AliProjectResourceResp `json:"mysql"`
	Mongo []*AliProjectResourceResp `json:"mongo"`
	Redis []*AliProjectResourceResp `json:"redis"`
}

type AliProjectResourceResp struct {
	ID         string `json:"id"`
	ConsoleURL string `json:"console_url"`
}
