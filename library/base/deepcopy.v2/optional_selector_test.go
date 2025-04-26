package deepcopy

import (
	"github.com/stretchr/testify/assert"
)

// to tag
type toTagSrc struct {
	Int        int           `deepcopy:"to:ToInt"`
	DiffStruct toTagChildSrc `deepcopy:"to:ToStruct"`
}
type toTagChildSrc struct {
	String string `deepcopy:"to:ToString"`
}
type toTagDst struct {
	ToInt    int
	ToStruct toTagChildDst
}
type toTagChildDst struct {
	ToString string
}

var toTagTest = []TestCase{
	{
		Name: "to-base",
		Src: &toTagSrc{
			Int: 1,
		},
		Dst: new(toTagDst),
		Config: &Config{
			EnableOptionalTags: []string{
				FieldToTagName,
			},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*toTagSrc)
			dst := dc.dst.(*toTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.Int, dst.ToInt)
		},
	},
	{
		Name: "to-struct",
		Src: &toTagSrc{
			DiffStruct: toTagChildSrc{
				String: "abc",
			},
		},
		Dst: new(toTagDst),
		Config: &Config{
			EnableOptionalTags: []string{
				FieldToTagName,
			},
		},
		Assert: func(t assert.TestingT, dc *DeepCopier, err error) {
			src := dc.src.(*toTagSrc)
			dst := dc.dst.(*toTagDst)
			assert.Nil(t, err)
			assert.Equal(t, src.DiffStruct.String, dst.ToStruct.ToString)
		},
	},
}
