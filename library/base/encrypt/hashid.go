package encrypt

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/speps/go-hashids"
)

func EncodeID(salt string, id int) (data string, err error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 8
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", errors.New(fmt.Sprintf("加密错误:%v", err))
	}
	data, err = h.Encode([]int{id})
	if err != nil {
		return "", errors.New(fmt.Sprintf("加密错误:%v", err))
	}
	return
}

func DecodeID(salt string, data string) (id int, err error) {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = 8
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("解密错误:%v", err))
	}
	e, err := h.DecodeWithError(data)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("解密错误:%v", err))
	}
	return e[0], nil
}
