package base

import (
	"strings"
	"unicode"
)

// CamelToSnake 将大驼峰命名转换为下划线分隔的字符串
func CamelToSnake(str string) string {
	var result strings.Builder
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i != 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
