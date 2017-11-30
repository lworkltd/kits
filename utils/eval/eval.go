package eval

import (
	"bytes"
	"fmt"
	"strings"
)

// Eval 是一个根据表达式获得值的接口
type Eval interface {
	Value(string) (string, error)
}

var defaultEval Eval = &evalImpl{}

// Value 获取一个表达式的值
func Value(s string) (string, error) {
	return defaultEval.Value(s)
}

type evalImpl struct {
}

// 获取结果
func (impl evalImpl) Value(desc string) (str string, err error) {
	index := strings.Index(desc, "${")
	if index < 0 {
		return desc, nil
	}

	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); ok {
				err = fmt.Errorf(s)
				return
			}
			panic(err)
		}
	}()

	var buffer bytes.Buffer
	executeDesc(&buffer, desc[0:])

	return buffer.String(), nil
}

// 解析语句
// 符合统配的语句格式为`${method,arg1[,arg2,...]}`
// 在匹配的过程中，会调用注册好的method，传递后缀的参数，得到一个字符串结果
func executeDesc(buffer *bytes.Buffer, desc string) {
	index := strings.Index(desc, "${")
	if index < 0 {
		buffer.WriteString(desc)
		return
	}
	if index > 0 {
		buffer.WriteString(desc[:index])
		desc = desc[index:]
	}
	end := strings.Index(desc, "}")
	if end < 0 {
		panic(fmt.Sprintf("bad syntax,expect need comma, from %v", desc))
	}

	input := desc[2:end]
	name, args, err := parseDesc(input)
	if err != nil {
		panic(err.Error())
	}
	execFunc, exist := executors[name]
	if !exist {
		panic(fmt.Sprintf("exector %s not found", name))
	}

	result, _, err := execFunc(args...)
	if err != nil {
		panic(err.Error())
	}
	buffer.WriteString(result)
	if len(desc)-1 > end {
		executeDesc(buffer, desc[end+1:])
	}
}

// 调用器
type ExecutorFunc func(...string) (string, bool, error)

var executors map[string]ExecutorFunc

func init() {
	executors = make(map[string]ExecutorFunc, 10)
}

// 注册调用器
func RegisterExecutor(name string, executor ExecutorFunc) {
	executors[name] = executor
}

// 注册键值调用
func RegisterKeyValueExecutor(name string, f func(string) (string, bool, error)) {
	executors[name] = SingleArgsExecutor(f)
}

type executor struct {
	exec ExecutorFunc
	args []string
}

// 解析语法，将语句拆分成调用和参数
func parseDesc(desc string) (string, []string, error) {
	desc = strings.TrimLeft(desc, " ")
	if desc == "" {
		return "", nil, fmt.Errorf("bad eval syntax")
	}
	words := strings.Split(desc, ",")
	var args []string
	if len(words) > 1 {
		args = words[1:]
	}
	for i, arg := range args {
		args[i] = strings.TrimSpace(arg)
	}
	cmd := strings.TrimSpace(words[0])
	return cmd, args, nil
}

// 注册无参调用
func EmptyArgsExecutor(f func() (string, bool, error)) ExecutorFunc {
	return func(...string) (string, bool, error) {
		return f()
	}
}

// 注册单参数调用
func SingleArgsExecutor(f func(string) (string, bool, error)) ExecutorFunc {
	return func(args ...string) (string, bool, error) {
		if len(args) < 1 {
			return "", false, fmt.Errorf("at least one args")
		}
		return f(args[0])
	}
}

// 解析对象
// 里面所有的字符串都会进行eval处理
// 必须传递Ptr
func Complete(v interface{}) error {
	return complete(v)
}
