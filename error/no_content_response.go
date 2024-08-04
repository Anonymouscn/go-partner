package error

// NoContentResponse 无内容响应错误
type NoContentResponse struct {
	name string
}

func (*NoContentResponse) Error() string {
	return "No content response"
}
