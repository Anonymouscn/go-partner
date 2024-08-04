package error

import "strings"

// RequestFail 请求失败错误
type RequestFail struct {
	name    string
	Details string
}

func (err *RequestFail) Error() string {
	var builder strings.Builder
	builder.WriteString("Request fail")
	if err.Details != "" {
		builder.WriteString(" - ")
		builder.WriteString(err.Details)
	}
	return builder.String()
}
