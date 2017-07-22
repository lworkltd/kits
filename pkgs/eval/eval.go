package eval

// Eval 是一个根据表达式获得值的接口
type Eval interface {
	Value(string) (string, error)
}

var defaultEval Eval

// Value 获取一个表达式的值
func Value(s string) (string, error) {
	return defaultEval.Value(s)
}
