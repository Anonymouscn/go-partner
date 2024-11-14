package restful

import (
	"crypto/tls"
	"net/http"
	"time"
)

// Path Restful path params
type Path []any

// Data Restful data
type Data map[string]any

// Method Restful method
type Method string

const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
)

// Request Restful request
type Request struct {
	req   http.Request // http 请求
	url   string       // url 地址
	path  Path         // 路径参数
	data  Data         // 自动化处理数据
	query Data         // 查询参数
	body  Data         // 请求体数据
}

// Response Restful 响应
type Response struct {
	StatusCode int                  // 响应状态码
	Proto      string               // 响应协议
	Raw        []byte               // 原始响应信息
	Headers    *http.Header         // 响应头
	Request    *http.Request        // 已发送请求
	TLS        *tls.ConnectionState // tls 连接状态
	Err        error                // 执行错误
	Time       time.Duration        // 响应用时
}
