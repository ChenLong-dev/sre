package httpclient

import "net/url"

// 请求Form
type Form struct {
	url.Values
}

// 新建Form参数
func NewForm() *Form {
	v := url.Values{}

	return &Form{
		Values: v,
	}
}

// 添加Form参数
func (f *Form) Add(key, value string) *Form {
	f.Values.Add(key, value)
	return f
}
