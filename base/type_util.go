package base

import (
	"fmt"
	customerror "github.com/Anonymouscn/go-partner/error"
	"maps"
	"reflect"
	"strconv"
	"strings"
)

// GetTypeName 获取类型名称
func GetTypeName(value any) string {
	name := reflect.ValueOf(value).Kind().String()
	switch name {
	case "ptr":
		name = fmt.Sprintf("ptr[%v]", GetTypeName(reflect.Indirect(reflect.ValueOf(value).Elem())))
	case "struct":
		details := reflect.TypeOf(value).Name()
		if details != "" && details != "Value" {
			name = fmt.Sprintf("struct[%v]", details)
		}
	}
	return name
}

// IsInteger 判断类型是否是整型
func IsInteger(value any) (bool, string) {
	typeName := GetTypeName(value)
	return strings.Contains(typeName, "int"), typeName
}

// IsFloat 判断类型是否是浮点型
func IsFloat(value any) (bool, string) {
	typeName := GetTypeName(value)
	return strings.Contains(typeName, "float"), typeName
}

// IsBool 判断类型是否是布尔型
func IsBool(value any) (bool, string) {
	typeName := GetTypeName(value)
	return typeName == "bool", typeName
}

// IsString 判断类型是否是字符串
func IsString(value any) (bool, string) {
	typeName := GetTypeName(value)
	return typeName == "string", typeName
}

// IsStruct 判断类型是否是结构体
func IsStruct(value any) (bool, string) {
	typeName := GetTypeName(value)
	return typeName == "struct", typeName
}

// IsPointer 判断是否是指针
func IsPointer(value any) (bool, string) {
	typeName := GetTypeName(value)
	return strings.Contains(typeName, "ptr"), typeName
}

// IsToMapAvailable 值是否可转 map
func IsToMapAvailable(value any) (bool, reflect.Kind) {
	k := reflect.ValueOf(value).Kind()
	switch k {
	case reflect.Struct, reflect.Pointer, reflect.Map:
		return true, k
	}
	return false, k
}

// AnyToMap 任意类型尝试转 map
func AnyToMap(value any) (map[string]any, error) {
	return anyToMap(value, reflect.ValueOf(value).Kind())
}

// anyToMap 任意类型尝试转 map (内部包使用, 无类型检查)
func anyToMap(value any, kind reflect.Kind) (map[string]any, error) {
	var m map[string]any
	var err error
	switch kind {
	case reflect.Struct, reflect.Pointer:
		m = StructToMap(value)
	case reflect.Map:
		m, err = StandardMap(value)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%v can not be converted to map type", value)
	}
	return m, nil
}

// MergeAnyToMap 合并任意类型到 map
func MergeAnyToMap(dst map[string]any, src any, overwrite bool) error {
	ok, k := IsToMapAvailable(src)
	if !ok {
		return fmt.Errorf("%v can not be converted to map type", src)
	}
	m, err := anyToMap(src, k)
	if err != nil {
		return err
	}
	if overwrite {
		maps.Copy(dst, m)
	} else {
		MapCopyOnNotExist(dst, m)
	}
	return nil
}

// AnyToString 任意类型尝试转 string
func AnyToString(value any) (string, error) {
	v := reflect.ValueOf(value)
	k := v.Kind()
	t := v.Type()
	tn := t.Name()
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return integerToString(value, tn)
	case reflect.Float32, reflect.Float64:
		return floatToString(value, tn)
	case reflect.Bool:
		return boolToString(value)
	case reflect.String:
		return v.String(), nil
	}
	return "", fmt.Errorf("%v can not be converted to string", value)
}

// BoolToString 布尔型转字符串
func BoolToString(value any) (string, error) {
	isBool, _ := IsBool(value)
	if !isBool {
		return "", &customerror.TypeConvertError{}
	}
	return boolToString(value)
}

// boolToString 布尔型转字符串 (内部包使用, 无类型检查)
func boolToString(value any) (string, error) {
	v, ok := value.(bool)
	if ok {
		if v {
			return "true", nil
		} else {
			return "false", nil
		}
	}
	return "", fmt.Errorf("value %v can not be converted to bool", value)
}

// IntegerToString 整型转换字符串 [int*, uint* => string]
func IntegerToString(value any) (string, error) {
	isInteger, typeName := IsInteger(value)
	if !isInteger {
		return "", &customerror.TypeConvertError{}
	}
	return integerToString(value, typeName)
}

// integerToString 整型转换字符串 (内部包使用, 无类型检查)
func integerToString(value any, typeName string) (string, error) {
	switch typeName {
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
func FloatToString(value any) (string, error) {
	isFloat, typeName := IsFloat(value)
	if !isFloat {
		return "", &customerror.TypeConvertError{}
	}
	return floatToString(value, typeName)
}

// floatToString 浮点型转字符串 (内部包使用, 无类型检查)
func floatToString(value any, typeName string) (string, error) {
	switch typeName {
	case "float32":
		v, _ := value.(float32)
		return strconv.FormatFloat(float64(v), 'g', -1, 64), nil
	default:
		v, _ := value.(float64)
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	}
}
