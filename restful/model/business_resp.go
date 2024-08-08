package restful_model

// Result Restful 业务结果
type Result[T any] struct {
	Code    int    `json:"code"`    // 业务状态码
	Message string `json:"message"` // 业务信息
	Data    T      `json:"data"`    // 业务数据
}

// Success 响应成功
func Success() *Result[any] {
	return SuccessWithData(nil)
}

// SuccessWithData 响应成功 (带数据)
func SuccessWithData(data any) *Result[any] {
	return ReplyWithData(200, "Success", data)
}

// Reply 普通响应
func Reply(code int, msg string) *Result[any] {
	return ReplyWithData(code, msg, nil)
}

// ReplyWithData 普通响应 (带数据)
func ReplyWithData(code int, msg string, data any) *Result[any] {
	return &Result[any]{
		Code:    code,
		Message: msg,
		Data:    data,
	}
}
