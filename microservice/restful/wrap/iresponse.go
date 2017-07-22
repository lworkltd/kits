package wrap

// IResponse 是基于API响应通用封装结构所需字段的一个通用接口
type IResponse interface {
	Result() bool
	Message() string
	Mcode() string
	Data() interface{}
}

// WrappedResponse 是基于API响应通用封装结构（或者说是本司rest接口的一个普遍共识）的一个实现
type WrappedResponse struct {
	result  bool
	message string
	mcode   string
	data    interface{}
}

// Result 返回执行成功或失败
func (me *WrappedResponse) Result() bool {
	return me.result
}

// Message 返回错误描述
func (me *WrappedResponse) Message() string {
	return me.message
}

// Mcode 返回错误代码
func (me *WrappedResponse) Mcode() string {
	return me.mcode
}

// Data 返回响应的数据
func (me *WrappedResponse) Data() interface{} {
	return me.data
}
