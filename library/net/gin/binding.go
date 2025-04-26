package gin

import (
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"net/textproto"
	"reflect"
)

const defaultMemory = 32 << 20

// 默认绑定方法
func BindingDefault(method, contentType string) binding.Binding {
	if method == http.MethodGet {
		return Form
	}

	switch contentType {
	case binding.MIMEJSON:
		return binding.JSON
	case binding.MIMEXML, binding.MIMEXML2:
		return binding.XML
	case binding.MIMEPROTOBUF:
		return binding.ProtoBuf
	case binding.MIMEMSGPACK, binding.MIMEMSGPACK2:
		return binding.MsgPack
	case binding.MIMEYAML:
		return binding.YAML
	case binding.MIMEMultipartPOSTForm:
		return binding.FormMultipart
	default:
		return Form
	}
}

var (
	Form     = formBinding{}
	Query    = queryBinding{}
	Header   = headerBinding{}
	FormPost = formPostBinding{}
)

type formBinding struct{}

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		if err != http.ErrNotMultipart {
			return err
		}
	}
	if err := mapForm(obj, req.Form); err != nil {
		return err
	}
	return validate(obj)
}

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *http.Request, obj interface{}) error {
	values := req.URL.Query()
	if err := mapForm(obj, values); err != nil {
		return err
	}
	return validate(obj)
}

type formPostBinding struct{}

func (formPostBinding) Name() string {
	return "form-urlencoded"
}

func (formPostBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := mapForm(obj, req.PostForm); err != nil {
		return err
	}
	return validate(obj)
}

type headerBinding struct{}

func (headerBinding) Name() string {
	return "header"
}

func (headerBinding) Bind(req *http.Request, obj interface{}) error {
	if err := mapHeader(obj, req.Header); err != nil {
		return err
	}

	return validate(obj)
}

func mapHeader(ptr interface{}, h map[string][]string) error {
	return mappingByPtr(ptr, headerSource(h), "header")
}

type headerSource map[string][]string

var _ setter = headerSource(nil)

func (hs headerSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (isSetted bool, err error) {
	return setByForm(value, field, hs, textproto.CanonicalMIMEHeaderKey(tagValue), opt)
}
