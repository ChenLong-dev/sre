package log

import (
	"context"

	_context "gitlab.shanhai.int/sre/library/base/context"
)

// 增加额外参数
func addExtraField(ctx context.Context, fields map[string]interface{}) {
	fields[_appID] = c.AppID
	fields[_uuid] = ctx.Value(_context.ContextUUIDKey)
}
