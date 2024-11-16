package jsonutil

import (
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
)

// JsonStringToMap json 字符串转 map
func JsonStringToMap(str string) (map[string]any, error) {
	return JsonByteToMap([]byte(str))
}

// MapToJsonString map 转 json 字符串
func MapToJsonString(m map[string]any) (string, error) {
	b, err := MapToJsonByte(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// JsonByteToMap json 字节数据转 map
func JsonByteToMap(b []byte) (map[string]any, error) {
	data := make(map[string]any)
	err := sonic.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("decode json error: %v", err))
	}
	return data, nil
}

// MapToJsonByte map 转 json 字节数据
func MapToJsonByte(m map[string]any) ([]byte, error) {
	res, err := sonic.Marshal(m)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("decode json from map error: %v", err))
	}
	return res, nil
}
