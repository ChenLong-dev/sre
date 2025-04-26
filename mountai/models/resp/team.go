package resp

import (
	"regexp"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

type TeamDetailResp struct {
	ID             string            `json:"id" deepcopy:"method:GenerateObjectIDString"`
	Name           string            `json:"name,omitempty"`
	DingHook       string            `json:"ding_hook,omitempty"`
	Label          string            `json:"label,omitempty"`
	AliAlarmName   string            `json:"ali_alarm_name,omitempty"`
	SentrySlug     string            `json:"sentry_slug,omitempty"`
	ExtraDingHooks map[string]string `json:"extra_ding_hooks,omitempty"`
}

func (t *TeamDetailResp) GetDingAccessToken() (string, error) {
	res := regexp.MustCompile(`access_token=(.+)$`).FindAllStringSubmatch(t.DingHook, -1)
	if len(res) == 0 {
		return "", errors.Wrapf(errcode.InternalError, "ding hook invalid:%s", t.DingHook)
	}
	if len(res[0]) < 2 {
		return "", errors.Wrapf(errcode.InternalError, "ding hook invalid:%s", t.DingHook)
	}
	return res[0][1], nil
}

type TeamListResp struct {
	ID             string            `json:"id" deepcopy:"method:GenerateObjectIDString"`
	Name           string            `json:"name,omitempty"`
	DingHook       string            `json:"ding_hook,omitempty"`
	Label          string            `json:"label,omitempty"`
	AliAlarmName   string            `json:"ali_alarm_name,omitempty"`
	ExtraDingHooks map[string]string `json:"extra_ding_hooks,omitempty"`
}
