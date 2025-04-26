package httpclient

import "net/url"

// 请求参数
type UrlValue struct {
	url.Values
}

// 新建请求参数
func NewUrlValue() *UrlValue {
	v := url.Values{}

	return &UrlValue{
		Values: v,
	}
}

// 添加请求参数
func (v *UrlValue) Add(key, value string) *UrlValue {
	v.Values.Add(key, value)
	return v
}
