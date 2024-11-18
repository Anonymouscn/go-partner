package base

import (
	"fmt"
	"reflect"
)

// MapCopyOnNotExist 当目标 map 不存在键值时进行复制
func MapCopyOnNotExist[M1 ~map[K]V, M2 ~map[K]V, K comparable, V any](dst M1, src M2) {
	for k, v := range src {
		if _, exist := any(dst[k]).(interface{ IsNil() bool }); !exist {
			dst[k] = v
		}
	}
}

// StandardMap 转换标准 map (map[string]any)
func StandardMap(m any) (map[string]any, error) {
	v := reflect.ValueOf(m)
	k := v.Kind()
	if k == reflect.Map {
		res := make(map[string]any)
		keys := v.MapKeys()
		for _, key := range keys {
			ks, err := AnyToString(key.Interface())
			if err != nil {
				return nil, err
			}
			res[ks] = v.MapIndex(key).Interface()
		}
		return res, nil
	}
	return nil, fmt.Errorf("%v is not a normal map type", m)
}
