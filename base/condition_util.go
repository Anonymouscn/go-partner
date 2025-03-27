package base

// SetOrDefault 设置非空值或默认值
func SetOrDefault[T comparable](value, defaultValue T) T {
	var zero T
	if value == zero {
		return defaultValue
	}
	return value
}
