package jsonize

import (
	"encoding/json"
)

// Jsonize 返回对象的json内容，无视错误
func V(object interface{}, indent bool) string {
	var indentUnit string
	if indent {
		indentUnit = "  "
		b, _ := json.MarshalIndent(object, "", indentUnit)
		return string(b)
	}
	b, _ := json.Marshal(object)
	return string(b)
}
