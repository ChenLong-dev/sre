package deepcopy

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// from tag
type fromTagSrc struct {
	Int        int
	DiffStruct fromTagChildSrc
	Map        map[string]string
}
type fromTagChildSrc struct {
	String string
}
type fromTagDst struct {
	FromInt    int             `deepcopy:"from:Int"`
	FromStruct fromTagChildDst `deepcopy:"from:DiffStruct"`
}
type fromTagChildDst struct {
	FromString string `deepcopy:"from:String"`
}
type fromTagMapDst struct {
	FromMap fromTagChildDst `deepcopy:"from:Map"`
}
type fromTagErrorDst struct {
	FromError fromTagChildDst `deepcopy:"from:Error"`
}

var fromTagTest = []TestCase{
	{
		Name: "from-base",
		Src: &fromTagSrc{
			Int: 1,
		},
		Dst: new(fromTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*fromTagSrc)
			dst := dc.dst.(*fromTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Int, dst.FromInt)
		},
	},
	{
		Name: "from-struct",
		Src: &fromTagSrc{
			DiffStruct: fromTagChildSrc{
				String: "abc",
			},
		},
		Dst: new(fromTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*fromTagSrc)
			dst := dc.dst.(*fromTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.DiffStruct.String, dst.FromStruct.FromString)
		},
	},
	{
		Name: "from-map",
		Src: &fromTagSrc{
			Map: map[string]string{
				"String": "abc",
			},
		},
		Dst: new(fromTagMapDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*fromTagSrc)
			dst := dc.dst.(*fromTagMapDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Map["String"], dst.FromMap.FromString)
		},
	},
	{
		Name: "from-error_normal",
		Src:  &fromTagSrc{},
		Dst:  new(fromTagErrorDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			// 找不到，暂时不报错
			assert.Nil(t, err)
		},
	},
	{
		Name: "from-error_strict",
		Src:  &fromTagSrc{},
		Dst:  new(fromTagErrorDst),
		Config: &Config{
			StrictMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
}

// skip tag
type skipTagSrc struct {
	Int    int
	Bool   bool
	Float  float64
	String string
}
type skipTagDst struct {
	Int    int `deepcopy:"skip"`
	Bool   bool
	Float  float64
	String string `deepcopy:"skip"`
}

var skipTagTest = []TestCase{
	{
		Name: "skip-base",
		Src: &skipTagSrc{
			Int:    1,
			Bool:   true,
			Float:  2.3,
			String: "abc",
		},
		Dst: new(skipTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*skipTagSrc)
			dst := dc.dst.(*skipTagDst)
			assert.Nil(t, err)
			assert.NotEqual(t, src.Int, dst.Int)
			assert.Equal(t, src.Bool, dst.Bool)
			assert.Equal(t, src.Float, dst.Float)
			assert.NotEqual(t, src.String, dst.String)
		},
	},
}

// method tag
type methodTagSrc struct{}

func (s *methodTagSrc) Normal(args map[string]interface{}) int {
	return 1
}
func (s *methodTagSrc) Struct(args map[string]interface{}) methodTagChildSrc {
	return methodTagChildSrc{}
}
func (s *methodTagSrc) Args(args map[string]interface{}) int {
	i, ok := args["Int"].(int)
	if ok {
		return i
	}
	return 0
}
func (s *methodTagSrc) Error(args map[string]interface{}) (int, error) {
	return 0, errors.New("error")
}

type methodTagChildSrc struct {
	String string
}

func (s *methodTagChildSrc) Normal(args map[string]interface{}) string {
	return "b"
}

type methodTagDst struct {
	MethodNormal int               `deepcopy:"method:Normal"`
	MethodStruct methodTagChildDst `deepcopy:"method:Struct"`
	MethodArgs   int               `deepcopy:"method:Args"`
}
type methodTagChildDst struct {
	MethodString string `deepcopy:"method:Normal"`
}
type methodTagErrorDst struct {
	MethodError int `deepcopy:"method:Error"`
}
type methodTagNotFoundDst struct {
	MethodNotFound int `deepcopy:"method:NotFound"`
}

var methodTagTest = []TestCase{
	{
		Name: "method-base",
		Src:  &methodTagSrc{},
		Dst:  new(methodTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*methodTagSrc)
			dst := dc.dst.(*methodTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Normal(make(map[string]interface{})), dst.MethodNormal)
		},
	},
	{
		Name: "method-struct",
		Src:  &methodTagSrc{},
		Dst:  new(methodTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*methodTagSrc)
			dst := dc.dst.(*methodTagDst)
			assert.Nil(t, err)
			childSrc := src.Struct(make(map[string]interface{}))
			assert.Equal(t, childSrc.Normal(make(map[string]interface{})), dst.MethodStruct.MethodString)
		},
	},
	{
		Name: "method-args",
		Src:  &methodTagSrc{},
		Dst:  new(methodTagDst),
		Args: map[string]interface{}{
			"Int": 1,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*methodTagSrc)
			dst := dc.dst.(*methodTagDst)
			assert.Nil(t, err)
			args := map[string]interface{}{
				"Int": 1,
			}
			assert.Equal(t, src.Args(args), dst.MethodArgs)
		},
	},
	{
		Name: "method-error",
		Src:  &methodTagSrc{},
		Dst:  new(methodTagErrorDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*methodTagSrc)
			assert.NotNil(t, err)
			_, e1 := src.Error(make(map[string]interface{}))
			assert.NotNil(t, e1)
			assert.Equal(t, e1.Error(), err.Error())
		},
	},
	{
		Name: "method-not_found_normal",
		Src:  &methodTagSrc{},
		Dst:  new(methodTagNotFoundDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.Nil(t, err)
		},
	},
	{
		Name: "method-not_found_strict",
		Src:  &methodTagSrc{},
		Dst:  new(methodTagNotFoundDst),
		Config: &Config{
			StrictMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
}

// default tag
type defaultTagSrc struct {
	DefaultExist int
}
type defaultTagChildSrc struct{}
type defaultTagDst struct {
	DefaultNormal int                `deepcopy:"default:Normal"`
	DefaultStruct defaultTagChildDst `deepcopy:"default:Struct"`
	DefaultArgs   int                `deepcopy:"default:Args"`

	DefaultInt    int     `deepcopy:"default:2"`
	DefaultInt64  int64   `deepcopy:"default:-1"`
	DefaultUInt   uint    `deepcopy:"default:4"`
	DefaultUInt64 uint64  `deepcopy:"default:5"`
	DefaultBool   bool    `deepcopy:"default:true"`
	DefaultString string  `deepcopy:"default:this is a string"`
	DefaultFloat  float64 `deepcopy:"default:1.23"`

	DefaultExist int `deepcopy:"default:Normal"`
}

func (d *defaultTagDst) Normal(args map[string]interface{}) int {
	return 1
}
func (d *defaultTagDst) Struct(args map[string]interface{}) defaultTagChildSrc {
	return defaultTagChildSrc{}
}
func (d *defaultTagDst) Args(args map[string]interface{}) int {
	i, ok := args["Int"].(int)
	if ok {
		return i
	}
	return 0
}

type defaultTagErrorDst struct {
	MethodError int `deepcopy:"default:Error"`
}

func (d *defaultTagErrorDst) Error(args map[string]interface{}) (int, error) {
	return 0, errors.New("error")
}

type defaultTagChildDst struct {
	String string `deepcopy:"default:Normal"`
}

func (d *defaultTagChildDst) Normal(args map[string]interface{}) string {
	return "b"
}

type defaultTagConvertErrorDst struct {
	DefaultUInt uint `deepcopy:"default:-1"`
}

type defaultTagNotFoundDst struct {
	UInt uintptr `deepcopy:"default:this is a string"`
}

var defaultTagTest = []TestCase{
	{
		Name: "default-normal",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, dst.Normal(make(map[string]interface{})), dst.DefaultNormal)
		},
	},
	{
		Name: "default-struct",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			childDst := new(defaultTagChildDst)
			assert.Equal(t, childDst.Normal(make(map[string]interface{})), dst.DefaultStruct.String)
		},
	},
	{
		Name: "default-args",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Args: map[string]interface{}{
			"Int": 1,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			args := map[string]interface{}{
				"Int": 1,
			}
			assert.Equal(t, dst.Args(args), dst.DefaultArgs)
		},
	},
	{
		Name: "default-error",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagErrorDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagErrorDst)
			assert.NotNil(t, err)
			_, e1 := dst.Error(make(map[string]interface{}))
			assert.NotNil(t, e1)
			assert.Equal(t, e1.Error(), err.Error())
		},
	},
	{
		Name: "default-int",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, 2, dst.DefaultInt)
		},
	},
	{
		Name: "default-int64",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, int64(-1), dst.DefaultInt64)
		},
	},
	{
		Name: "default-uint",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, uint(4), dst.DefaultUInt)
		},
	},
	{
		Name: "default-uint64",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, uint64(5), dst.DefaultUInt64)
		},
	},
	{
		Name: "default-bool",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, true, dst.DefaultBool)
		},
	},
	{
		Name: "default-float",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, 1.23, dst.DefaultFloat)
		},
	},
	{
		Name: "default-string",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, "this is a string", dst.DefaultString)
		},
	},
	{
		Name: "default-exist",
		Src: &defaultTagSrc{
			DefaultExist: 10,
		},
		Dst: new(defaultTagDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*defaultTagSrc)
			dst := dc.dst.(*defaultTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.DefaultExist, dst.DefaultExist)
			assert.NotEqual(t, dst.Normal(make(map[string]interface{})), dst.DefaultExist)
		},
	},
	{
		Name: "default-convert_error",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagConvertErrorDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
	{
		Name: "default-not_found_normal",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagNotFoundDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.Nil(t, err)
		},
	},
	{
		Name: "default-not_found_strict",
		Src:  &defaultTagSrc{},
		Dst:  new(defaultTagNotFoundDst),
		Config: &Config{
			StrictMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			assert.NotNil(t, err)
		},
	},
}
