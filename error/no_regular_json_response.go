package error

import "strings"

// NoRegularJsonResponse 非常规json错误
type NoRegularJsonResponse struct {
	name    string
	Details string
}

func (err *NoRegularJsonResponse) Error() string {
	var builder strings.Builder
	builder.WriteString("No regular json response")
	if err.Details != "" {
		builder.WriteString(" - ")
		builder.WriteString(err.Details)
	}
	return builder.String()
}
