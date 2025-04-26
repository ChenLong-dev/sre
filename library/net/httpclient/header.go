package httpclient

import "net/http"

// 请求头
type Header struct {
	http.Header
}

// json请求头
func NewJsonHeader() *Header {
	header := http.Header{}
	header.Set("Content-Type", "application/json")

	return &Header{
		Header: header,
	}
}

// form表单请求头
func NewFormURLEncodedHeader() *Header {
	header := http.Header{}
	header.Set("Content-Type", "application/x-www-form-urlencoded")

	return &Header{
		Header: header,
	}
}

// 增加header
func (h *Header) Add(key, value string) *Header {
	h.Header.Add(key, value)
	return h
}

// 设置header
func (h *Header) Set(key, value string) *Header {
	h.Header.Set(key, value)
	return h
}

// 默认json请求头
func GetDefaultHeader() *Header {
	return NewJsonHeader()
}
