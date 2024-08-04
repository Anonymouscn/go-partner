package error

// TypeConvertError 类型转换错误
type TypeConvertError struct {
	name string
}

func (*TypeConvertError) Error() string {
	return "Type convert error"
}
