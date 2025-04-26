package req

import "gitlab.shanhai.int/sre/library/base/null"

// ==========================
// 请求模型
// ==========================
type InvokeRestfulReq struct {
	Method            null.String `json:"method" form:"method" binding:"required"`
	Url               null.String `json:"url" form:"url" binding:"required"`
	GoroutineCount    null.Int    `json:"goroutine_count" form:"goroutine_count" binding:"required"`
	Times             null.Int    `json:"times" form:"times" binding:"required"`
	SleepMilliseconds null.Int    `json:"sleep_milliseconds" form:"sleep_milliseconds" binding:"required"`
}
