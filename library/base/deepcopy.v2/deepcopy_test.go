package deepcopy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/base/deepcopy"
	"gitlab.shanhai.int/sre/library/base/null"
)

// 测试用例
type TestCase struct {
	Name   string
	Src    interface{}
	Dst    interface{}
	Args   map[string]interface{}
	Config *Config
	Assert func(t assert.TestingT, dc *DeepCopier, err error)
}

// 基础类型
type baseSrc struct {
	Int    int
	Bool   bool
	Float  float64
	String string
}
type baseDst struct {
	Int    int
	Bool   bool
	Float  float64
	String string
}

// 指针测试
type basePtrSrc struct {
	BasePtrToInt   *int
	BasePtrToBool  *bool
	BasePtrToFloat *float64
	BasePtrToStr   *string

	BaseIntToPtr   int
	BaseBoolToPtr  bool
	BaseFloatToPtr float64
	BaseStrToPtr   string
}
type basePtrDst struct {
	BasePtrToInt   int
	BasePtrToBool  bool
	BasePtrToFloat float64
	BasePtrToStr   string

	BaseIntToPtr   *int
	BaseBoolToPtr  *bool
	BaseFloatToPtr *float64
	BaseStrToPtr   *string
}

var (
	baseInt   = 2
	baseBool  = false
	baseFloat = 3.4
	baseStr   = "def"
)
var baseTest = []TestCase{
	{
		Name: "base-base",
		Src: &baseSrc{
			Int:    1,
			Bool:   true,
			Float:  2.3,
			String: "abc",
		},
		Dst: new(baseDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*baseSrc)
			dst := dc.dst.(*baseDst)
			assert.Nil(t, err)

			assert.Equal(t, src.Int, dst.Int)
			assert.Equal(t, src.Bool, dst.Bool)
			assert.Equal(t, src.Float, dst.Float)
			assert.Equal(t, src.String, dst.String)
		},
	},
	{
		Name: "base-override",
		Src:  &baseSrc{},
		Dst: &baseDst{
			Int:    1,
			Bool:   true,
			Float:  2.3,
			String: "abc",
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*baseDst)
			assert.Nil(t, err)
			assert.Zero(t, dst.Int)
			assert.Zero(t, dst.Bool)
			assert.Zero(t, dst.Float)
			assert.Zero(t, dst.String)
		},
	},
	{
		Name: "base-ptrConvert",
		Src: &basePtrSrc{
			BasePtrToInt:   &baseInt,
			BasePtrToFloat: &baseFloat,
			BasePtrToStr:   &baseStr,
			BasePtrToBool:  &baseBool,

			BaseIntToPtr:   3,
			BaseBoolToPtr:  true,
			BaseFloatToPtr: 4.5,
			BaseStrToPtr:   "ghi",
		},
		Dst: new(basePtrDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*basePtrSrc)
			dst := dc.dst.(*basePtrDst)
			assert.Nil(t, err)

			assert.Equal(t, *src.BasePtrToInt, dst.BasePtrToInt)
			assert.Equal(t, *src.BasePtrToFloat, dst.BasePtrToFloat)
			assert.Equal(t, *src.BasePtrToStr, dst.BasePtrToStr)
			assert.Equal(t, *src.BasePtrToBool, dst.BasePtrToBool)

			assert.Equal(t, src.BaseIntToPtr, *dst.BaseIntToPtr)
			assert.Equal(t, src.BaseBoolToPtr, *dst.BaseBoolToPtr)
			assert.Equal(t, src.BaseFloatToPtr, *dst.BaseFloatToPtr)
			assert.Equal(t, src.BaseStrToPtr, *dst.BaseStrToPtr)
		},
	},
}

// 嵌套结构体/指针
type nestedSrc struct {
	DiffPtr    *nestedChildSrc
	DiffStruct nestedChildSrc

	SamePtr    *nestedChildSrc
	SameStruct nestedChildSrc

	PtrToStruct *nestedChildSrc
	StructToPtr nestedChildSrc
}
type nestedChildSrc struct {
	String string
}
type nestedDst struct {
	DiffPtr    *nestedChildDst
	DiffStruct nestedChildDst

	SamePtr    *nestedChildSrc
	SameStruct nestedChildSrc

	PtrToStruct nestedChildDst
	StructToPtr *nestedChildDst
}
type nestedChildDst struct {
	String string
}

var nestedTest = []TestCase{
	{
		Name: "nested-diff_type",
		Src: &nestedSrc{
			DiffStruct: nestedChildSrc{
				String: "a",
			},
			DiffPtr: &nestedChildSrc{
				String: "b",
			},
		},
		Dst: new(nestedDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*nestedSrc)
			dst := dc.dst.(*nestedDst)
			assert.Nil(t, err)
			assert.Equal(t, src.DiffStruct.String, dst.DiffStruct.String)
			assert.Equal(t, src.DiffPtr.String, dst.DiffPtr.String)
		},
	},
	{
		Name: "nested-same_type",
		Src: &nestedSrc{
			SameStruct: nestedChildSrc{
				String: "a",
			},
			SamePtr: &nestedChildSrc{
				String: "b",
			},
		},
		Dst: new(nestedDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*nestedSrc)
			dst := dc.dst.(*nestedDst)
			assert.Nil(t, err)
			assert.Equal(t, src.SameStruct.String, dst.SameStruct.String)
			assert.Equal(t, src.SamePtr.String, dst.SamePtr.String)
		},
	},
	{
		Name: "nested-mix",
		Src: &nestedSrc{
			StructToPtr: nestedChildSrc{
				String: "a",
			},
			PtrToStruct: &nestedChildSrc{
				String: "b",
			},
		},
		Dst: new(nestedDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*nestedSrc)
			dst := dc.dst.(*nestedDst)
			assert.Nil(t, err)
			assert.Equal(t, src.StructToPtr.String, dst.StructToPtr.String)
			assert.Equal(t, src.PtrToStruct.String, dst.PtrToStruct.String)
		},
	},
	{
		Name: "nested-override",
		Src:  &nestedSrc{},
		Dst: &nestedDst{
			SameStruct: nestedChildSrc{
				String: "a",
			},
			SamePtr: &nestedChildSrc{
				String: "b",
			},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*nestedDst)
			assert.Nil(t, err)
			assert.Zero(t, dst.SameStruct.String)
			assert.Zero(t, dst.SamePtr)
		},
	},
}

// map
type mapSrc struct {
	PtrToMap    *mapChildSrc
	StructToMap mapChildSrc

	MapToStruct map[string]interface{}
	MapToPtr    map[string]interface{}

	MapToMap map[string]interface{}
}
type mapChildSrc struct {
	Int int
}
type mapDst struct {
	PtrToMap    map[string]interface{}
	StructToMap map[string]interface{}

	MapToStruct mapChildSrc
	MapToPtr    *mapChildSrc

	MapToMap map[string]interface{}
}

var mapTest = []TestCase{
	{
		Name: "map-base",
		Src: &mapSrc{
			MapToMap: map[string]interface{}{
				"Int": 1,
			},
		},
		Dst: new(mapDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*mapSrc)
			dst := dc.dst.(*mapDst)
			assert.Nil(t, err)
			assert.Equal(t, src.MapToMap["Int"].(int), dst.MapToMap["Int"].(int))
		},
	},
	{
		Name: "map-dst",
		Src: &mapSrc{
			PtrToMap: &mapChildSrc{
				Int: 1,
			},
			StructToMap: mapChildSrc{
				Int: 2,
			},
		},
		Dst: new(mapDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*mapSrc)
			dst := dc.dst.(*mapDst)
			assert.Nil(t, err)
			assert.Equal(t, src.PtrToMap.Int, dst.PtrToMap["Int"].(int))
			assert.Equal(t, src.StructToMap.Int, dst.StructToMap["Int"].(int))
		},
	},
	{
		Name: "map-src",
		Src: &mapSrc{
			MapToPtr: map[string]interface{}{
				"Int": 1,
			},
			MapToStruct: map[string]interface{}{
				"Int": 2,
			},
		},
		Dst: new(mapDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*mapSrc)
			dst := dc.dst.(*mapDst)
			assert.Nil(t, err)
			assert.Equal(t, src.MapToPtr["Int"], dst.MapToPtr.Int)
			assert.Equal(t, src.MapToStruct["Int"], dst.MapToStruct.Int)
		},
	},
	{
		Name: "map-override",
		Src:  &mapSrc{},
		Dst: &mapDst{
			MapToMap: map[string]interface{}{
				"Int": 1,
			},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*mapDst)
			assert.Nil(t, err)
			assert.Zero(t, dst.MapToMap)
		},
	},
}

// 切片/数组
type sliceSrc struct {
	Base []string

	DiffPtr    []*sliceChildSrc
	DiffStruct []sliceChildSrc

	SamePtr    []*sliceChildSrc
	SameStruct []sliceChildSrc

	PtrToStruct []*sliceChildSrc
	StructToPtr []sliceChildSrc
}
type sliceChildSrc struct {
	String string
}
type sliceDst struct {
	Base []string

	DiffPtr    []*sliceChildDst
	DiffStruct []sliceChildDst

	SamePtr    []*sliceChildSrc
	SameStruct []sliceChildSrc

	PtrToStruct []sliceChildDst
	StructToPtr []*sliceChildDst
}
type sliceChildDst struct {
	String string
}

var sliceTest = []TestCase{
	{
		Name: "slice-base",
		Src: &sliceSrc{
			Base: []string{"a"},
		},
		Dst: new(sliceDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sliceSrc)
			dst := dc.dst.(*sliceDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Base[0], dst.Base[0])
		},
	},
	{
		Name: "slice-diff_type",
		Src: &sliceSrc{
			DiffStruct: []sliceChildSrc{
				{
					String: "a",
				},
			},
			DiffPtr: []*sliceChildSrc{
				{
					String: "b",
				},
			},
		},
		Dst: new(sliceDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sliceSrc)
			dst := dc.dst.(*sliceDst)
			assert.Nil(t, err)
			assert.Equal(t, src.DiffStruct[0].String, dst.DiffStruct[0].String)
			assert.Equal(t, src.DiffPtr[0].String, dst.DiffPtr[0].String)
		},
	},
	{
		Name: "slice-same_type",
		Src: &sliceSrc{
			SameStruct: []sliceChildSrc{
				{
					String: "a",
				},
			},
			SamePtr: []*sliceChildSrc{
				{
					String: "b",
				},
			},
		},
		Dst: new(sliceDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sliceSrc)
			dst := dc.dst.(*sliceDst)
			assert.Nil(t, err)
			assert.Equal(t, src.SameStruct[0].String, dst.SameStruct[0].String)
			assert.Equal(t, src.SamePtr[0].String, dst.SamePtr[0].String)
		},
	},
	{
		Name: "slice-mix",
		Src: &sliceSrc{
			StructToPtr: []sliceChildSrc{
				{
					String: "a",
				},
			},
			PtrToStruct: []*sliceChildSrc{
				{
					String: "b",
				},
			},
		},
		Dst: new(sliceDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*sliceSrc)
			dst := dc.dst.(*sliceDst)
			assert.Nil(t, err)
			assert.Equal(t, src.StructToPtr[0].String, dst.StructToPtr[0].String)
			assert.Equal(t, src.PtrToStruct[0].String, dst.PtrToStruct[0].String)
		},
	},
	{
		Name: "slice-override",
		Src:  &sliceSrc{},
		Dst: &sliceDst{
			Base: []string{"a"},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*sliceDst)
			assert.Nil(t, err)
			assert.Zero(t, dst.Base)
		},
	},
}

// 匿名结构体
type anonymousStructSrc struct {
	S2 string
	AnonymousChildStructSrc

	DuplicateString string
	AnonymousDuplicateStructChild
	*AnonymousDuplicatePtrChild
}
type AnonymousDuplicateStructChild struct {
	DuplicateString string
}
type AnonymousDuplicatePtrChild struct {
	DuplicateString string
}
type anonymousUnexportedSrc struct {
	S2 string
	anonymousChildStructSrc
}
type anonymousPtrSrc struct {
	S2 string
	*AnonymousChildPtrSrc
}
type AnonymousChildStructSrc struct {
	S1 string
}
type AnonymousChildPtrSrc struct {
	S1 string
}
type anonymousChildStructSrc struct {
	S1 string
}
type AnonymousChildTraversalStructSrc struct {
	S1 string
	S2 string
}
type anonymousTraversalPtrDst struct {
	*AnonymousChildTraversalStructSrc
}
type anonymousTraversalStructDst struct {
	AnonymousChildTraversalStructSrc
}
type anonymousDuplicateDst struct {
	DuplicateString string
	AnonymousDuplicateStructChild
	*AnonymousDuplicatePtrChild
}
type anonymousStructDst struct {
	S1 string
	AnonymousChildStructDst
	AnonymousChildStructSrc
}
type anonymousUnexportedDst struct {
	S1 string
	anonymousChildStructDst
}
type anonymousPtrDst struct {
	S1 string
	*AnonymousChildPtrDst
}
type AnonymousChildStructDst struct {
	S2 string
}
type AnonymousChildPtrDst struct {
	S2 string
}
type anonymousChildStructDst struct {
	S2 string
}

var anonymousTest = []TestCase{
	{
		Name: "anonymous-unexported_src",
		Src: &anonymousUnexportedSrc{
			anonymousChildStructSrc: anonymousChildStructSrc{
				S1: "a",
			},
		},
		Dst: new(anonymousUnexportedDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousUnexportedSrc)
			dst := dc.dst.(*anonymousUnexportedDst)
			// 源结构体内未导出字段允许拷贝
			assert.Nil(t, err)
			assert.Equal(t, src.anonymousChildStructSrc.S1, dst.S1)
		},
	},
	{
		Name: "anonymous-unexported_dst",
		Src: &anonymousUnexportedSrc{
			S2: "a",
		},
		Dst: new(anonymousUnexportedDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousUnexportedSrc)
			dst := dc.dst.(*anonymousUnexportedDst)
			// 目标结构体内未导出字段不允许拷贝
			assert.Nil(t, err)
			assert.NotEqual(t, src.S2, dst.anonymousChildStructDst.S2)
		},
	},
	{
		Name: "anonymous-same_type",
		Src: &anonymousStructSrc{
			AnonymousChildStructSrc: AnonymousChildStructSrc{
				S1: "a",
			},
		},
		Dst: new(anonymousStructDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousStructSrc)
			dst := dc.dst.(*anonymousStructDst)
			assert.Nil(t, err)
			assert.Equal(t, src.AnonymousChildStructSrc.S1, dst.AnonymousChildStructSrc.S1)
		},
	},
	{
		Name: "anonymous-src_struct",
		Src: &anonymousStructSrc{
			AnonymousChildStructSrc: AnonymousChildStructSrc{
				S1: "a",
			},
		},
		Dst: new(anonymousStructDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousStructSrc)
			dst := dc.dst.(*anonymousStructDst)
			assert.Nil(t, err)
			assert.Equal(t, src.AnonymousChildStructSrc.S1, dst.S1)
		},
	},
	{
		Name: "anonymous-dst_struct",
		Src: &anonymousStructSrc{
			S2: "a",
		},
		Dst: new(anonymousStructDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousStructSrc)
			dst := dc.dst.(*anonymousStructDst)
			assert.Nil(t, err)
			assert.Equal(t, src.S2, dst.AnonymousChildStructDst.S2)
		},
	},
	{
		Name: "anonymous-src_ptr",
		Src: &anonymousPtrSrc{
			AnonymousChildPtrSrc: &AnonymousChildPtrSrc{
				S1: "a",
			},
		},
		Dst: new(anonymousPtrDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousPtrSrc)
			dst := dc.dst.(*anonymousPtrDst)
			assert.Nil(t, err)
			assert.Equal(t, src.AnonymousChildPtrSrc.S1, dst.S1)
		},
	},
	{
		Name: "anonymous-dst_ptr",
		Src: &anonymousPtrSrc{
			S2: "a",
		},
		Dst: new(anonymousPtrDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousPtrSrc)
			dst := dc.dst.(*anonymousPtrDst)
			assert.Nil(t, err)
			assert.Equal(t, src.S2, dst.AnonymousChildPtrDst.S2)
		},
	},

	{
		Name: "anonymous-duplicate_struct",
		Src: &anonymousStructSrc{
			DuplicateString: "outside",
			AnonymousDuplicateStructChild: AnonymousDuplicateStructChild{
				DuplicateString: "inside",
			},
		},
		Dst: new(anonymousDuplicateDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousStructSrc)
			dst := dc.dst.(*anonymousDuplicateDst)
			assert.Nil(t, err)
			assert.Equal(t, src.DuplicateString, dst.DuplicateString)
			assert.Equal(t, src.AnonymousDuplicateStructChild.DuplicateString, dst.AnonymousDuplicateStructChild.DuplicateString)
		},
	},
	{
		Name: "anonymous-duplicate_ptr",
		Src: &anonymousStructSrc{
			DuplicateString: "outside",
			AnonymousDuplicatePtrChild: &AnonymousDuplicatePtrChild{
				DuplicateString: "inside",
			},
		},
		Dst: new(anonymousDuplicateDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*anonymousStructSrc)
			dst := dc.dst.(*anonymousDuplicateDst)
			assert.Nil(t, err)
			assert.Equal(t, src.DuplicateString, dst.DuplicateString)
			assert.Equal(t, src.AnonymousDuplicatePtrChild.DuplicateString, dst.AnonymousDuplicatePtrChild.DuplicateString)
		},
	},
	{
		Name: "anonymous-override",
		Src:  &anonymousStructSrc{},
		Dst: &anonymousStructDst{
			AnonymousChildStructSrc: AnonymousChildStructSrc{
				S1: "a",
			},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*anonymousStructDst)
			assert.Nil(t, err)
			assert.Zero(t, dst.AnonymousChildStructSrc.S1)
		},
	},
	{
		Name: "anonymous-ptr_full_traversal",
		Src: &AnonymousChildTraversalStructSrc{
			S1: "s1",
			S2: "",
		},
		Config: &Config{
			NotZeroMode:       true,
			FullTraversalMode: true,
		},
		Dst: &anonymousTraversalPtrDst{
			AnonymousChildTraversalStructSrc: &AnonymousChildTraversalStructSrc{
				S2: "s2",
				S1: "",
			},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*AnonymousChildTraversalStructSrc)
			dst := dc.dst.(*anonymousTraversalPtrDst)
			assert.Nil(t, err)
			assert.Equal(t, src.S1, dst.S1)
			assert.NotEqual(t, src.S2, dst.S2)
		},
	},
	{
		Name: "anonymous-struct_full_traversal",
		Src: &AnonymousChildTraversalStructSrc{
			S1: "s1",
			S2: "",
		},
		Config: &Config{
			NotZeroMode:       true,
			FullTraversalMode: true,
		},
		Dst: &anonymousTraversalStructDst{
			AnonymousChildTraversalStructSrc: AnonymousChildTraversalStructSrc{
				S2: "s2",
				S1: "",
			},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*AnonymousChildTraversalStructSrc)
			dst := dc.dst.(*anonymousTraversalStructDst)
			assert.Nil(t, err)
			assert.Equal(t, src.S1, dst.S1)
			assert.NotEqual(t, src.S2, dst.S2)
		},
	},
}

// interface
type interfaceSrc struct {
	Int    int
	Bool   bool
	Float  float64
	String string
	Ptr    *interfaceChildSrc
	Struct interfaceChildSrc
}
type interfaceChildSrc struct {
	String string
}
type interfaceDst struct {
	Int    interface{}
	Bool   interface{}
	Float  interface{}
	String interface{}
	Ptr    interface{}
	Struct interface{}
}

var interfaceTest = []TestCase{
	{
		Name: "interface-base",
		Src: &interfaceSrc{
			Int:    1,
			Bool:   true,
			Float:  2.3,
			String: "abc",
		},
		Dst: new(interfaceDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*interfaceSrc)
			dst := dc.dst.(*interfaceDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Int, dst.Int.(int))
			assert.Equal(t, src.Bool, dst.Bool.(bool))
			assert.Equal(t, src.Float, dst.Float.(float64))
			assert.Equal(t, src.String, dst.String.(string))
		},
	},
	{
		Name: "interface-struct",
		Src: &interfaceSrc{
			Struct: interfaceChildSrc{
				String: "abc",
			},
		},
		Dst: new(interfaceDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*interfaceSrc)
			dst := dc.dst.(*interfaceDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Struct.String, dst.Struct.(interfaceChildSrc).String)
		},
	},
	{
		Name: "interface-ptr",
		Src: &interfaceSrc{
			Ptr: &interfaceChildSrc{
				String: "abc",
			},
		},
		Dst: new(interfaceDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*interfaceSrc)
			dst := dc.dst.(*interfaceDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Ptr.String, dst.Ptr.(*interfaceChildSrc).String)
		},
	},
}

// mk tag
type mkTagSrc struct {
	PtrToMap    *mkTagChildSrc
	StructToMap mkTagChildSrc
}
type mkTagChildSrc struct {
	Int int `deepcopy:"mk:int"`
}
type mkTagDst struct {
	PtrToMap    map[string]interface{}
	StructToMap map[string]interface{}
}

var mkTagTest = []TestCase{
	{
		Name: "mk-struct",
		Src: &mkTagSrc{
			StructToMap: mkTagChildSrc{
				Int: 2,
			},
		},
		Dst: new(mkTagDst),
		Config: &Config{
			EnableOptionalTags: []string{MapKeyTagName},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*mkTagSrc)
			dst := dc.dst.(*mkTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.StructToMap.Int, dst.StructToMap["int"].(int))
			_, ok := dst.StructToMap["Int"]
			assert.Equal(t, false, ok)
		},
	},
	{
		Name: "mk-ptr",
		Src: &mkTagSrc{
			PtrToMap: &mkTagChildSrc{
				Int: 2,
			},
		},
		Dst: new(mkTagDst),
		Config: &Config{
			EnableOptionalTags: []string{MapKeyTagName},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*mkTagSrc)
			dst := dc.dst.(*mkTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.PtrToMap.Int, dst.PtrToMap["int"].(int))
			_, ok := dst.PtrToMap["Int"]
			assert.Equal(t, false, ok)
		},
	},
}

// 非零值模式
type notZeroSrc struct {
	Int       int
	Float     float64
	Bool      bool
	String    string
	Interface interface{}

	NullString null.String
	NullInt    null.Int

	ChildStruct notZeroChildSrc
	ChildPtr    *notZeroChildSrc

	notZeroChildSrc

	ZeroMap     map[string]interface{}
	StructToMap notZeroChildSrc

	ZeroSlice     []string
	OverrideSlice []string
}
type notZeroChildSrc struct {
	Int        int
	ZeroInt    int
	String     string
	ZeroString string
}
type notZeroDst struct {
	Int       int
	Float     float64
	Bool      bool
	String    string
	Interface interface{}

	NullString null.String
	NullInt    null.Int

	ChildStruct notZeroChildSrc
	ChildPtr    *notZeroChildSrc

	notZeroChildSrc

	ZeroMap     map[string]interface{}
	StructToMap map[string]interface{}

	ZeroSlice     []string
	OverrideSlice []string
}

var notZeroTest = []TestCase{
	{
		Name: "not_zero-base",
		Src:  &notZeroSrc{},
		Dst: &notZeroDst{
			Int:       1,
			Float:     1.2,
			Bool:      true,
			String:    "123",
			Interface: "string",
		},
		Config: &Config{
			NotZeroMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*notZeroDst)
			assert.Nil(t, err)
			assert.NotZero(t, dst.Int)
			assert.NotZero(t, dst.Float)
			assert.NotZero(t, dst.Bool)
			assert.NotZero(t, dst.String)
			assert.NotZero(t, dst.Interface)
		},
	},
	{
		Name: "not_zero-sql",
		Src: &notZeroSrc{
			NullInt: null.NewInt(1, false),
		},
		Dst: &notZeroDst{
			NullInt:    null.IntFrom(2),
			NullString: null.StringFrom("null"),
		},
		Config: &Config{
			NotZeroMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*notZeroSrc)
			dst := dc.dst.(*notZeroDst)
			assert.Nil(t, err)
			assert.NotZero(t, dst.NullString.ValueOrZero())
			assert.Equal(t, "null", dst.NullString.ValueOrZero())
			assert.NotEqual(t, src.NullInt.ValueOrZero(), dst.NullInt.ValueOrZero())
			assert.Equal(t, 2, dst.NullInt.ValueOrZero())
		},
	},
	{
		Name: "not_zero-struct",
		Src:  &notZeroSrc{},
		Dst: &notZeroDst{
			ChildStruct: notZeroChildSrc{
				Int:    1,
				String: "abc",
			},
			ChildPtr: &notZeroChildSrc{
				Int:    2,
				String: "bcd",
			},
		},
		Config: &Config{
			NotZeroMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*notZeroDst)
			assert.Nil(t, err)

			assert.NotZero(t, dst.ChildStruct.Int)
			assert.NotZero(t, dst.ChildStruct.String)

			assert.NotZero(t, dst.ChildPtr)
			assert.NotZero(t, dst.ChildPtr.Int)
			assert.NotZero(t, dst.ChildPtr.String)
		},
	},
	{
		Name: "not_zero-anonymous",
		Src:  &notZeroSrc{},
		Dst: &notZeroDst{
			notZeroChildSrc: notZeroChildSrc{
				Int:    1,
				String: "abc",
			},
		},
		Config: &Config{
			NotZeroMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*notZeroDst)
			assert.Nil(t, err)
			assert.NotZero(t, dst.notZeroChildSrc.Int)
			assert.NotZero(t, dst.notZeroChildSrc.String)
		},
	},
	{
		Name: "not_zero-map",
		Src: &notZeroSrc{
			StructToMap: notZeroChildSrc{
				Int:        2,
				ZeroInt:    0,
				String:     "abc",
				ZeroString: "",
			},
		},
		Dst: &notZeroDst{
			ZeroMap: map[string]interface{}{
				"1": "a",
			},
			StructToMap: map[string]interface{}{
				"Int":        "a",
				"ZeroInt":    "b",
				"String":     1,
				"ZeroString": 2,
			},
		},
		Config: &Config{
			NotZeroMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*notZeroSrc)
			dst := dc.dst.(*notZeroDst)
			assert.Nil(t, err)

			assert.NotZero(t, dst.ZeroMap)

			assert.Equal(t, src.StructToMap.Int, dst.StructToMap["Int"].(int))
			assert.Equal(t, src.StructToMap.String, dst.StructToMap["String"].(string))

			_, ok := dst.StructToMap["ZeroInt"]
			assert.Equal(t, false, ok)
			_, ok = dst.StructToMap["ZeroString"]
			assert.Equal(t, false, ok)
		},
	},
	{
		Name: "not_zero-slice",
		Src: &notZeroSrc{
			OverrideSlice: []string{"a", "", "b"},
		},
		Dst: &notZeroDst{
			ZeroSlice:     []string{"a", "b"},
			OverrideSlice: []string{"1", "2", "3"},
		},
		Config: &Config{
			NotZeroMode: true,
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*notZeroSrc)
			dst := dc.dst.(*notZeroDst)
			assert.Nil(t, err)

			assert.NotZero(t, dst.ZeroSlice)
			assert.Equal(t, 2, len(dst.ZeroSlice))
			assert.NotZero(t, dst.ZeroSlice[0])
			assert.NotZero(t, dst.ZeroSlice[1])

			assert.Equal(t, len(src.OverrideSlice), len(dst.OverrideSlice))
			assert.Equal(t, src.OverrideSlice[0], dst.OverrideSlice[0])
			assert.Equal(t, src.OverrideSlice[1], dst.OverrideSlice[1])
			assert.Equal(t, src.OverrideSlice[2], dst.OverrideSlice[2])
		},
	},
}

// 组合
type mixSrc struct {
	TimePtr  *time.Time
	TimeNull null.Time
}

func (s *mixSrc) GetNullTime(args map[string]interface{}) null.Time {
	return s.TimeNull
}

type mixDst struct {
	Time            string `deepcopy:"from:TimePtr;timeformat:2006-01-02 15:04:05"`
	MethodNullTime  string `deepcopy:"method:GetNullTime;timeformat:2006-01-02 15:04:05"`
	FromNullTime    string `deepcopy:"from:TimeNull;timeformat:2006-01-02 15:04:05"`
	DefaultNullTime string `deepcopy:"default:GetDefaultNullTime;timeformat:2006-01-02 15:04:05"`
	SomeNullTime    null.Time
}

func (d *mixDst) GetDefaultNullTime(args map[string]interface{}) null.Time {
	return d.SomeNullTime
}

var mixTime = time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local)
var mixTest = []TestCase{
	{
		Name: "mix-from&timeformat",
		Src: &mixSrc{
			TimePtr: &mixTime,
		},
		Dst: new(mixDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*mixSrc)
			dst := dc.dst.(*mixDst)
			assert.Nil(t, err)
			assert.Equal(t, src.TimePtr.Format("2006-01-02 15:04:05"), dst.Time)
		},
	},
	{
		Name: "mix-method&sql&timeformat",
		Src: &mixSrc{
			TimeNull: null.TimeFrom(mixTime),
		},
		Dst: new(mixDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*mixSrc)
			dst := dc.dst.(*mixDst)
			assert.Nil(t, err)
			assert.Equal(t, src.GetNullTime(make(map[string]interface{})).ValueOrZero().Format("2006-01-02 15:04:05"), dst.MethodNullTime)
		},
	},
	{
		Name: "mix-from&sql&timeformat",
		Src: &mixSrc{
			TimeNull: null.TimeFrom(mixTime),
		},
		Dst: new(mixDst),
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*mixSrc)
			dst := dc.dst.(*mixDst)
			assert.Nil(t, err)
			assert.Equal(t, src.TimeNull.ValueOrZero().Format("2006-01-02 15:04:05"), dst.FromNullTime)
		},
	},
	{
		Name: "mix-default&sql&timeformat",
		Src:  &mixSrc{},
		Dst: &mixDst{
			SomeNullTime: null.TimeFrom(mixTime),
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			dst := dc.dst.(*mixDst)
			assert.Nil(t, err)
			assert.Equal(t, dst.GetDefaultNullTime(make(map[string]interface{})).ValueOrZero().Format("2006-01-02 15:04:05"), dst.DefaultNullTime)
		},
	},
}

func TestCopy(t *testing.T) {
	list := make([]TestCase, 0)

	// 基本测试
	list = append(list, baseTest...)
	list = append(list, nestedTest...)
	list = append(list, mapTest...)
	list = append(list, sliceTest...)
	list = append(list, anonymousTest...)
	list = append(list, interfaceTest...)

	// 系统selector测试
	list = append(list, fromTagTest...)
	list = append(list, skipTagTest...)
	list = append(list, methodTagTest...)
	list = append(list, defaultTagTest...)

	// 可选selector测试
	list = append(list, toTagTest...)

	// 系统transformer测试
	list = append(list, timeFormatTagTest...)
	list = append(list, sqlTest...)
	list = append(list, stringTagTest...)
	list = append(list, boolTagTest...)
	list = append(list, objectIDTagTest...)

	// 可选transformer测试

	// 其他测试
	list = append(list, mkTagTest...)
	list = append(list, notZeroTest...)
	list = append(list, mixTest...)

	for _, item := range list {
		t.Run(item.Name, func(t *testing.T) {
			dc := Copy(item.Src).SetConfig(item.Config)
			err := dc.WithArgs(item.Args).To(item.Dst)
			item.Assert(t, dc, err)
		})
	}
}

// benchmark
type benchmarkSimpleSrc struct {
	Int    int
	Bool   bool
	Float  float64
	String string
}
type benchmarkSimpleDst struct {
	Int    int
	Bool   bool
	Float  float64
	String string
}

type benchmarkComplexSrc struct {
	SomeInt     int
	ToBool      bool `deepcopy:"to:Bool"`
	TimeFormat  time.Time
	String      null.String
	Skip        int
	Interface   string
	StructToPtr BenchmarkComplexChildSrc
	PtrToStruct *BenchmarkComplexChildSrc
	*BenchmarkComplexChildSrc
}

func (s *benchmarkComplexSrc) SomeMethod(args map[string]interface{}) string {
	return "some method"
}

type BenchmarkComplexChildSrc struct {
	String string
}

type benchmarkComplexDst struct {
	Int           int `deepcopy:"from:SomeInt"`
	Bool          bool
	TimeFormat    string `deepcopy:"timeformat:2006-01-02 15:04:05"`
	String        string
	Method        string      `deepcopy:"method:SomeMethod"`
	Skip          int         `deepcopy:"skip"`
	Interface     interface{} `deepcopy:"force"`
	DefaultMethod string      `deepcopy:"default:DefaultSomeMethod"`
	DefaultString string      `deepcopy:"default:some string"`
	StructToPtr   *BenchmarkComplexChildSrc
	PtrToStruct   BenchmarkComplexChildSrc
	*BenchmarkComplexChildSrc
}

func (s *benchmarkComplexDst) DefaultSomeMethod(args map[string]interface{}) string {
	return "default some method"
}

func BenchmarkCopy(b *testing.B) {
	b.Run("simple", func(b *testing.B) {
		b.Run("direct", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				src := &benchmarkSimpleSrc{
					Int:    1,
					Bool:   true,
					Float:  2.34,
					String: "this is a test",
				}
				_ = &benchmarkSimpleDst{
					Int:    src.Int,
					Bool:   src.Bool,
					Float:  src.Float,
					String: src.String,
				}
			}
		})

		b.Run("v1", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = deepcopy.Copy(&benchmarkSimpleSrc{
					Int:    1,
					Bool:   true,
					Float:  2.34,
					String: "this is a test",
				}).To(&benchmarkSimpleDst{})
			}
		})

		b.Run("v2-system", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Copy(&benchmarkSimpleSrc{
					Int:    1,
					Bool:   true,
					Float:  2.34,
					String: "this is a test",
				}).To(&benchmarkSimpleDst{})
			}
		})

		b.Run("v2-optional", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Copy(&benchmarkSimpleSrc{
					Int:    1,
					Bool:   true,
					Float:  2.34,
					String: "this is a test",
				}).
					SetConfig(&Config{
						EnableOptionalTags: []string{
							FieldToTagName,
						},
					}).
					To(&benchmarkSimpleDst{})
			}
		})
	})

	b.Run("complex", func(b *testing.B) {
		b.Run("direct", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				src := &benchmarkComplexSrc{
					SomeInt:    1,
					ToBool:     true,
					TimeFormat: time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local),
					String:     null.StringFrom("this is a test"),
					Skip:       1,
					StructToPtr: BenchmarkComplexChildSrc{
						String: "struct",
					},
					PtrToStruct: &BenchmarkComplexChildSrc{
						String: "ptr",
					},
					BenchmarkComplexChildSrc: &BenchmarkComplexChildSrc{
						String: "anonymous",
					},
				}
				_ = &benchmarkComplexDst{
					Int:           src.SomeInt,
					Bool:          src.ToBool,
					TimeFormat:    src.TimeFormat.Format("2006-01-02 15:04:05"),
					String:        src.String.ValueOrZero(),
					Method:        src.SomeMethod(make(map[string]interface{})),
					Skip:          0,
					Interface:     src.Interface,
					DefaultMethod: new(benchmarkComplexDst).DefaultSomeMethod(make(map[string]interface{})),
					DefaultString: "some string",
					StructToPtr: &BenchmarkComplexChildSrc{
						String: src.StructToPtr.String,
					},
					PtrToStruct: BenchmarkComplexChildSrc{
						String: src.PtrToStruct.String,
					},
					BenchmarkComplexChildSrc: &BenchmarkComplexChildSrc{
						String: src.BenchmarkComplexChildSrc.String,
					},
				}
			}
		})

		b.Run("v1", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = deepcopy.Copy(&benchmarkComplexSrc{
					SomeInt:    1,
					ToBool:     true,
					TimeFormat: time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local),
					String:     null.StringFrom("this is a test"),
					Skip:       1,
					StructToPtr: BenchmarkComplexChildSrc{
						String: "struct",
					},
					PtrToStruct: &BenchmarkComplexChildSrc{
						String: "ptr",
					},
					BenchmarkComplexChildSrc: &BenchmarkComplexChildSrc{
						String: "anonymous",
					},
				}).To(&benchmarkComplexDst{})
			}
		})

		b.Run("v2-system", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Copy(&benchmarkComplexSrc{
					SomeInt:    1,
					ToBool:     true,
					TimeFormat: time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local),
					String:     null.StringFrom("this is a test"),
					Skip:       1,
					StructToPtr: BenchmarkComplexChildSrc{
						String: "struct",
					},
					PtrToStruct: &BenchmarkComplexChildSrc{
						String: "ptr",
					},
					BenchmarkComplexChildSrc: &BenchmarkComplexChildSrc{
						String: "anonymous",
					},
				}).
					To(&benchmarkComplexDst{})
			}
		})

		b.Run("v2-optional", func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = Copy(&benchmarkComplexSrc{
					SomeInt:    1,
					ToBool:     true,
					TimeFormat: time.Date(2015, 4, 25, 22, 30, 15, 0, time.Local),
					String:     null.StringFrom("this is a test"),
					Skip:       1,
					StructToPtr: BenchmarkComplexChildSrc{
						String: "struct",
					},
					PtrToStruct: &BenchmarkComplexChildSrc{
						String: "ptr",
					},
					BenchmarkComplexChildSrc: &BenchmarkComplexChildSrc{
						String: "anonymous",
					},
				}).
					SetConfig(&Config{
						EnableOptionalTags: []string{
							FieldToTagName,
						},
					}).
					To(&benchmarkComplexDst{})
			}
		})
	})

	b.Run("concurrent-safe", func(b *testing.B) {
		b.Run("v2", func(b *testing.B) {
			list := make([]TestCase, 0)

			// 基本测试
			list = append(list, baseTest...)
			list = append(list, nestedTest...)
			list = append(list, mapTest...)
			list = append(list, sliceTest...)
			list = append(list, anonymousTest...)
			list = append(list, interfaceTest...)

			// 系统selector测试
			list = append(list, fromTagTest...)
			list = append(list, skipTagTest...)
			list = append(list, methodTagTest...)
			list = append(list, defaultTagTest...)

			// 可选selector测试
			list = append(list, toTagTest...)

			// 系统transformer测试
			list = append(list, timeFormatTagTest...)
			list = append(list, sqlTest...)
			list = append(list, stringTagTest...)
			list = append(list, boolTagTest...)
			list = append(list, objectIDTagTest...)

			// 可选transformer测试

			// 其他测试
			list = append(list, mkTagTest...)
			list = append(list, notZeroTest...)
			list = append(list, mixTest...)

			for _, item := range list {
				b.Run(item.Name, func(b *testing.B) {
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							dc := Copy(item.Src).SetConfig(item.Config)
							err := dc.WithArgs(item.Args).To(item.Dst)
							item.Assert(b, dc, err)
						}
					})
				})
			}
		})
	})
}
