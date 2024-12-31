package base

// SliceToMap 切片转 map
func SliceToMap[T any, K comparable](slice []T, keyFn func(T) K) map[K]T {
	result := make(map[K]T)
	for _, item := range slice {
		key := keyFn(item)
		result[key] = item
	}
	return result
}
