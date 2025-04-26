package v2

import (
	v1 "rulai/models/resp"
)

type AppRunningStatusListResp struct {
	AppID         string                        `json:"app_id"`
	RunningStatus []*v1.RunningStatusDetailResp `json:"running_status"`
}
