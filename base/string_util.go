package base

import (
	"crypto/rand"
	"encoding/base64"
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

// RandString 生成指定长度的随机字符串
func RandString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}
