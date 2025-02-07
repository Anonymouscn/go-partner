package calculate

// SafeDivideFloat 安全除法
// 当触发 NaN 异常时 (如:被除数为0时，使用默认值; 无默认值，0填充)
func SafeDivideFloat[T float64 | float32](a, b T, defaultValue ...T) T {
	if a == 0 {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return a / b
}
