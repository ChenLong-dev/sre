package req

import "rulai/models"

type CreateTeamReq struct {
	Name           string            `json:"name" binding:"required"`
	DingHook       string            `json:"ding_hook" binding:"required"`
	Label          string            `json:"label" binding:"required"`
	AliAlarmName   string            `json:"ali_alarm_name" binding:"required"`
	ExtraDingHooks map[string]string `json:"extra_ding_hooks"`
}

type UpdateTeamReq struct {
	DingHook       string             `json:"ding_hook"`
	Label          string             `json:"label"`
	AliAlarmName   string             `json:"ali_alarm_name"`
	ExtraDingHooks *map[string]string `json:"extra_ding_hooks"`
}

type GetTeamsReq struct {
	models.BaseListRequest
	Keyword      string `form:"keyword" json:"keyword"`
	Label        string
	TeamIDs      string   `form:"team_ids" json:"team_ids"`
	IDs          []string `form:"ids" json:"ids"`
	KeywordField string   `form:"keyword_field" json:"keyword_field"`
}
