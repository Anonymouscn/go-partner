package base

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// todo 评估'结构信息缓存'方案, 动态缓存结构信息, 减少反射调用优化性能

const (
	SubMapStructSign  = "map[string]interface {}" // 子 map 结构体签名
	TimeStampSign     = "int64"                   // 时间戳签名
	TimeStructSign    = "time.Time"               // 时间结构体签名
	TimeStructPtrSign = "*time.Time"              // 时间结构体指针签名
)

var (
	TimeType = reflect.TypeOf(time.Time{}) // 时间类型 time.Time
	// BasicTypeSignMap 基础类型签名表
	BasicTypeSignMap = map[string]any{
		"int":     1,
		"int8":    2,
		"int16":   3,
		"int32":   4,
		"int64":   5,
		"uint":    6,
		"uint8":   7,
		"uint16":  8,
		"uint32":  9,
		"uint64":  10,
		"byte":    11,
		"float32": 12,
		"float64": 13,
		"string":  14,
		"bool":    15,
	}
)

// StructToMap struct 结构体转 map
func StructToMap(obj any) map[string]any {
	// 判空
	if obj == nil {
		return nil
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	// 指针类型处理
	if t.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		t = t.Elem()
		v = v.Elem()
	}
	data := make(map[string]any)
	// 结构体字段处理
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		val := v.Field(i)
		// 忽略非公开字段
		if !val.CanInterface() {
			continue
		}
		// 读取 json 标签
		tag := GetJsonKeyOrDefaultFromStructField(field, CamelToSnake(field.Name))
		// === 处理匿名字段 === //
		if field.Anonymous {
			if val.Kind() == reflect.Pointer && val.IsNil() {
				continue
			}
			val = reflect.Indirect(val)
			if val.Kind() == reflect.Struct {
				embedded := StructToMap(val.Interface())
				for k, v := range embedded {
					if _, exists := data[k]; !exists {
						data[k] = v
					}
				}
			}
			continue
		}
		// === 处理具名字段 === //
		// 指针类型处理
		if val.Kind() == reflect.Pointer {
			if val.IsNil() {
				continue
			}
			val = val.Elem()
		}
		// 时间类型自动转换处理
		if val.Type() == TimeType {
			if tm, ok := val.Interface().(time.Time); ok {
				data[tag] = tm.Unix()
			}
			continue
		}
		// 嵌套结构处理
		if val.Kind() == reflect.Struct {
			data[GetJsonKeyOrDefaultFromStructField(field, CamelToSnake(field.Name))] = StructToMap(val.Interface())
			continue
		}
		// 数组/切片类型处理
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			l := val.Len()
			s := make([]any, l)
			for i := 0; i < l; i++ {
				elem := val.Index(i)
				if elem.Kind() == reflect.Pointer && !elem.IsNil() {
					elem = elem.Elem()
				}
				if elem.Kind() == reflect.Struct {
					s[i] = StructToMap(elem.Interface())
				} else {
					s[i] = elem.Interface()
				}
			}
			data[tag] = s
			continue
		}
		data[tag] = val.Interface()
	}
	return data
}

// MapToStruct map 转 struct 结构体 (公开形式调用方法)
func MapToStruct(m map[string]any, obj any) {
	sT := reflect.TypeOf(obj)
	// 接收对象如果不是指针, 禁止执行
	if sT.Kind() != reflect.Pointer {
		return
	}
	// map 转 struct 结构体
	MapToStructDetails(m, sT, reflect.ValueOf(obj))
}

// MapToStructDetails map 转 struct 结构体 (实际调用方法)
// m: map, st: 结构体类型, sv: 结构体值
func MapToStructDetails(m map[string]any, st reflect.Type, sv reflect.Value) {
	// 空指针处理
	if sv.Kind() == reflect.Invalid {
		return
	}
	// 结构体指针处理
	if sv.Kind() == reflect.Pointer {
		sv = sv.Elem()
		// 空指针处理
		if sv.Kind() == reflect.Invalid {
			return
		}
		st = st.Elem()
	}
	// 结构体字段处理
	for i := 0; i < sv.NumField(); i++ {
		sFT := st.Field(i) // 结构体字段类型
		sFV := sv.Field(i) // 结构体字段值
		// 排除不可导出字段
		if !sFV.CanInterface() {
			continue
		}
		// 处理匿名结构体
		if sFT.Anonymous {
			mapToAnonymousStruct(m, sFT.Type, sFV)
			continue
		}
		tag := GetJsonKeyOrDefaultFromStructField(sFT, CamelToSnake(sFT.Name))
		mV, ok := m[tag]
		if !ok || mV == nil {
			continue
		}
		v := reflect.ValueOf(mV)
		checkAndInstantiatePointerField(sFT.Type, sFV, v)
		// 处理结构体字段
		setStructField(sFV, v)
	}
}

// 指针字段处理，将未实例化的指针实例化
// ft: 结构体字段, fv: 结构体值, av: 期望初始化的值 (只对基本类型有效)
func checkAndInstantiatePointerField(ft reflect.Type, fv reflect.Value, av ...reflect.Value) bool {
	if strings.HasPrefix(ft.String(), "*") {
		// 空指针处理
		if fv.Kind() == reflect.Pointer && fv.Elem().Kind() == reflect.Invalid {
			// 初始化内存空间
			fv.Set(reflect.New(fv.Type().Elem()))
			// 基本类型给默认值 (实际地址指向常量池)
			if len(av) > 0 && BasicTypeSignMap[strings.Trim(ft.String(), "*")] != nil {
				fv = fv.Elem()
				if fv.Kind() == av[0].Kind() {
					fv.Set(av[0])
				}
			}
			return true
		}
	}
	return false
}

// 处理匿名结构体
// m: map, st: 结构体类型, sv: 结构体值
func mapToAnonymousStruct(m map[string]any, st reflect.Type, sv reflect.Value) {
	// 匿名结构指针处理
	if sv.Kind() == reflect.Pointer {
		checkAndInstantiatePointerField(st, sv)
		st = st.Elem()
		sv = sv.Elem()
	}
	// 默认映射字段到同层 map
	MapToStructDetails(m, st, sv)
	// 匿名结构默认名: 匿名结构/指针名称 (下划线格式)
	defaultName := CamelToSnake(st.Name())
	if m[defaultName] != nil {
		if mV, ok := (m[defaultName]).(map[string]any); ok {
			// 下级 map 存在则映射到下级 map
			MapToStructDetails(mV, st, sv)
		}
	}
}

// 结构体字段设置目标值
// fv: 结构体字段值, tv: 目标值
func setStructField(fv, tv reflect.Value) {
	sign := tv.Type().String()
	k := tv.Kind()
	// 处理子结构
	if sign == SubMapStructSign {
		setSubStruct(fv, tv)
		return
	}
	// 处理时间戳
	if sign == TimeStampSign && tv.CanInt() &&
		(fv.Type().String() == TimeStructSign || fv.Type().String() == TimeStructPtrSign) {
		if setTime(fv, tv.Int()) {
			return
		}
	}
	// 处理时间结构
	if sign == TimeStructSign {
		fv.Set(tv)
		return
	}
	// 处理时间结构指针
	if sign == TimeStructPtrSign {
		fv.Set(tv.Elem())
		return
	}
	if k == reflect.Array || k == reflect.Slice {
		_ = setSliceOrArray(fv, tv)
	}
	// 处理基本类型
	_ = setBasicValue(fv, tv)
}

// setSliceOrArray 设置切片/数组
func setSliceOrArray(fv, tv reflect.Value) error {
	l := tv.Len()
	// 切片/数组自动扩容
	if fv.Len() < l {
		fv.Set(reflect.MakeSlice(fv.Type(), l, l))
	}
	if l > 0 {
		for i := 0; i < l; i++ {
			src := reflect.ValueOf(tv.Index(i).Interface())
			dst := fv.Index(i)
			// 处理数组单元
			setStructField(dst, src)
		}
	}
	return nil
}

// 设置子结构体
// fv: 结构体字段值, tv: 目标值
func setSubStruct(fv, tv reflect.Value) {
	ov := tv.Interface()
	sm, ok := ov.(map[string]any)
	if !ok {
		return
	}
	switch fv.Kind() {
	case reflect.Struct:
		MapToStructDetails(sm, fv.Type(), fv)
	case reflect.Pointer:
		eT := fv.Type()
		checkAndInstantiatePointerField(eT, fv)
		MapToStructDetails(sm, eT.Elem(), fv.Elem())
	}
}

// setTime 填充结构体时间字段 (时间戳自动转换 time.Time)
// 注意: 10位时间戳 (如需高精度时间请自定义纳秒级字段)
func setTime(point reflect.Value, timestamp int64) bool {
	tm := time.Unix(timestamp, 0)
	switch point.Kind() {
	case reflect.Pointer:
		checkAndInstantiatePointerField(point.Type(), point)
		point.Elem().Set(reflect.ValueOf(tm))
		return true
	case reflect.Struct:
		point.Set(reflect.ValueOf(tm))
		return true
	}
	return false
}

// setBasicValue 设置基本类型
func setBasicValue(point, value reflect.Value) error {
	// 类型检查点
	if err := checkPoint(point); err != nil {
		return err
	}
	// 尝试自动类型转换
	if value.CanConvert(point.Type()) {
		point.Set(value.Convert(point.Type()))
		return nil
	}
	// 类型指针转换
	if point.Kind() == reflect.Pointer && !checkAndInstantiatePointerField(point.Type(), point, value) {
		return setBasicValue(point.Elem(), value)
	}
	switch value.Kind() {
	case reflect.Struct: // 非法结构
		return fmt.Errorf("type of value[%v] is not a basic type or a pointer of basic type", value.Kind())
	// 基础类型转换
	case reflect.Interface: // 类型接口
		if !value.IsNil() {
			return setBasicValue(point, value.Elem())
		}
	}
	return nil
}

// checkPoint 检查点校验
func checkPoint(point reflect.Value) error {
	// 有效值校验
	if !point.IsValid() {
		return fmt.Errorf("not a valid point")
	}
	// 能否修改检查
	if !point.CanSet() {
		return fmt.Errorf("point can not be set")
	}
	return nil
}

// GetTagFromStructField 从结构体字段获取标签
func GetTagFromStructField(field reflect.StructField, tag string) string {
	return strings.Trim(field.Tag.Get(tag), " ")
}

// GetJsonTagFromStructField 从结构体字段获取 json 标签
func GetJsonTagFromStructField(field reflect.StructField) []string {
	return strings.Split(GetTagFromStructField(field, "json"), ",")
}

// GetFormTagFromStructField 从结构体获取 form 标签
func GetFormTagFromStructField(field reflect.StructField) []string {
	return strings.Split(GetTagFromStructField(field, "form"), ",")
}

// GetFormKeyFromStructField 从结构体字段获取 form key
func GetFormKeyFromStructField(field reflect.StructField) string {
	tag := GetFormTagFromStructField(field)
	if len(tag) == 0 {
		return ""
	}
	return tag[0]
}

// GetGormTagFromStructField 从结构体获取 gorm 标签
func GetGormTagFromStructField(field reflect.StructField) []string {
	return strings.Split(GetTagFromStructField(field, "gorm"), ",")
}

// GetJsonKeyFromStructField 从结构体字段获取 json key
func GetJsonKeyFromStructField(field reflect.StructField) string {
	tag := GetJsonTagFromStructField(field)
	if len(tag) == 0 {
		return ""
	}
	return tag[0]
}

// GetJsonKeyOrDefaultFromStructField 从结构体字段获取 json key, 获取不到则给默认值
func GetJsonKeyOrDefaultFromStructField(field reflect.StructField, defaultValue string) string {
	name := GetJsonKeyFromStructField(field)
	if name == "" {
		return defaultValue
	}
	return name
}
