package context

import "context"

// 获取指定key的value
func GetString(ctx context.Context, key string) string {
	return GetStringOrDefault(ctx, key, "")
}

// 获取指定key的value，没有获取到则返回默认
func GetStringOrDefault(ctx context.Context, key string, defaultValue string) string {
	v, ok := ctx.Value(key).(string)
	if !ok || v == "" {
		return defaultValue
	}

	return v
}
