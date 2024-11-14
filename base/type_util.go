package base

import (
	customerror "github.com/Anonymouscn/go-partner/error"
	"reflect"
	"strconv"
	"strings"
)

// GetTypeName 获取类型名称
func GetTypeName(value interface{}) string {
	valueType := reflect.TypeOf(value)
	if valueType == nil {
		return "unknown"
	}
	return valueType.Name()
}

// IsInteger 判断类型是否是整型
func IsInteger(value interface{}) (bool, string) {
	typeName := GetTypeName(value)
	return strings.Contains(typeName, "int"), typeName
}

// IsFloat 判断类型是否是浮点型
func IsFloat(value interface{}) (bool, string) {
	typeName := GetTypeName(value)
	return strings.Contains(typeName, "float"), typeName
}

// IsBool 判断类型是否是布尔型
func IsBool(value interface{}) (bool, string) {
	typeName := GetTypeName(value)
	return typeName == "bool", typeName
}

// BoolToString 布尔型转字符串
func BoolToString(value interface{}) (string, error) {
	isBool, _ := IsBool(value)
	if !isBool {
		return "", &customerror.TypeConvertError{}
	}
	v, _ := value.(bool)
	if v {
		return "true", nil
	} else {
		return "false", nil
	}
}

// IntegerToString 整型转换字符串 [int*, uint* => string]
func IntegerToString(value interface{}) (string, error) {
	isInteger, valueName := IsInteger(value)
	if !isInteger {
		return "", &customerror.TypeConvertError{}
	}
	switch valueName {
	case "int64":
		v, _ := value.(int64)
		return strconv.FormatInt(v, 10), nil
	case "int32":
		v, _ := value.(int32)
		return strconv.FormatInt(int64(v), 10), nil
	case "int16":
		v, _ := value.(int16)
		return strconv.Itoa(int(v)), nil
	case "int8":
		v, _ := value.(int8)
		return strconv.Itoa(int(v)), nil
	case "uint64":
		v, _ := value.(uint64)
		return strconv.FormatInt(int64(v), 10), nil
	case "uint32":
		v, _ := value.(uint32)
		return strconv.FormatInt(int64(v), 10), nil
	case "uint16":
		v, _ := value.(uint16)
		return strconv.Itoa(int(v)), nil
	case "uint8":
		v, _ := value.(uint8)
		return strconv.Itoa(int(v)), nil
	case "uint":
		v, _ := value.(uint)
		return strconv.Itoa(int(v)), nil
	default:
		v, _ := value.(int)
		return strconv.Itoa(v), nil
	}
}

// FloatToString 浮点型转字符串
func FloatToString(value interface{}) (string, error) {
	isFloat, valueName := IsFloat(value)
	if !isFloat {
		return "", &customerror.TypeConvertError{}
	}
	switch valueName {
	case "float32":
		v, _ := value.(float32)
		return strconv.FormatFloat(float64(v), 'g', -1, 64), nil
	default:
		v, _ := value.(float64)
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	}
}
