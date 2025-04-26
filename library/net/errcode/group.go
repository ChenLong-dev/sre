package errcode

// 新建错误组
func NewGroup(code Codes) Group {
	return Group{
		Codes:    code,
		children: make([]error, 0),
	}
}

// 错误组
type Group struct {
	Codes
	children []error
}

// 子错误
func (g Group) Children() []error {
	return g.children
}

func (g Group) Details() []interface{} {
	details := make([]interface{}, len(g.children))
	for i, err := range g.children {
		details[i] = err.Error()
	}
	return details
}

// 增加子错误
func (g Group) AddChildren(errs ...error) Group {
	g.children = append(g.children, errs...)
	return g
}
