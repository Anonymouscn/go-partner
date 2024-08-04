package restful

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Anonymouscn/go-tools/base"
	customerror "github.com/Anonymouscn/go-tools/error"
	iotools "github.com/Anonymouscn/go-tools/io"
	"io"
	"net/http"
	urlpkg "net/url"
	"reflect"
	"strconv"
	"strings"
)

// RestClient Restful 客户端
type RestClient struct {
	client    http.Client // http 客户端
	request   *Request    // 请求
	responses []Response  // 响应栈
}

// RestClientConfig RestClient 配置
type RestClientConfig struct {
	retry *RetryConfig // 请求重试配置
}

// RetryConfig 请求重试配置
type RetryConfig struct {
	enable bool // 是否启用
	max    int  // 最大重试次数
}

// NewRestClient 新建 Restful 客户端
func NewRestClient() *RestClient {
	return &RestClient{
		request: &Request{
			url:   "",
			path:  make(Path, 0),
			query: make(Data),
			body:  struct{}{},
		},
		responses: make([]Response, 0),
	}
}

// SetURL 设置 URL 路径参数
func (rc *RestClient) SetURL(url string) *RestClient {
	rc.request.url = url
	return rc
}

// SetPath 设置路径参数
func (rc *RestClient) SetPath(path Path) *RestClient {
	rc.request.path = path
	return rc
}

// SetHeaders 设置请求头
func (rc *RestClient) SetHeaders(headers Data) *RestClient {
	for k, v := range headers {
		rc.request.req.Header.Set(k, rc.buildParams(v))
	}
	return rc
}

// SetQuery 设置查询参数
func (rc *RestClient) SetQuery(query Data) *RestClient {
	rc.request.query = query
	return rc
}

// SetBody 设置请求体参数
func (rc *RestClient) SetBody(body Body) *RestClient {
	rc.request.body = body
	return rc
}

// Reset 参数重置
func (rc *RestClient) Reset() *RestClient {
	rc.request.path = make([]interface{}, 0)
	rc.request.query = make(Data)
	rc.request.body = struct{}{}
	rc.responses = make([]Response, 0)
	return rc
}

// DisableCertAuth 禁用 TLS 证书验证
func (rc *RestClient) DisableCertAuth() *RestClient {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	rc.client.Transport = transport
	return rc
}

// 生成请求 URL
func (rc *RestClient) generateURL() string {
	url := rc.buildPathParamsToURL(rc.request.url, rc.request.path)
	url = rc.buildQueryParamsToURL(url, rc.request.query)
	fmt.Println(url)
	return url
}

// 生成请求体
func (rc *RestClient) generateBody() error {
	data, err := json.Marshal(rc.request.body)
	if err != nil {
		return err
	}
	if rc.request.body != nil {
		var requestReader io.Reader
		requestReader = bytes.NewReader(data)
		body, ok := requestReader.(io.ReadCloser)
		if !ok {
			body = io.NopCloser(requestReader)
		} else {
			defer iotools.CloseReader(rc.request.req.Body)
		}
		rc.request.req.Body = body
	}
	return nil
}

// 设置请求方法
func (rc *RestClient) setRequestMethod(method Method) *RestClient {
	rc.request.req.Method = string(method)
	return rc
}

// 设置请求 URL
func (rc *RestClient) setRequestURL(url string) *RestClient {
	u, _ := urlpkg.Parse(url)
	rc.request.req.URL = u
	return rc
}

// 是否是 json 响应
func (rc *RestClient) isJsonResponse(resp *http.Response) bool {
	contentType := resp.Header["Content-Type"]
	return contentType != nil && strings.Contains(contentType[0], "application/json")
}

// 执行请求
func (rc *RestClient) executeRequest() *RestClient {
	rc.setRequestURL(rc.generateURL())
	err := rc.generateBody()
	if err != nil {
		rc.handleRequestError(err)
		return rc
	}
	resp, err := rc.client.Do(&rc.request.req)
	if err != nil {
		rc.handleRequestError(err)
		return rc
	}
	if resp == nil {
		rc.handleRequestError(&customerror.NoContentResponse{})
		return rc
	}
	defer iotools.CloseReader(resp.Body)
	return rc.handleHTTPResponse(resp, nil)
}

// Get 发送 GET 请求
func (rc *RestClient) Get() *RestClient {
	rc.setRequestMethod(GET)
	return rc.executeRequest()
}

// Post 发送 POST 请求
func (rc *RestClient) Post() *RestClient {
	rc.setRequestMethod(POST)
	return rc.executeRequest()
}

// Put 发送 PUT 请求
func (rc *RestClient) Put() *RestClient {
	rc.setRequestMethod(PUT)
	return rc.executeRequest()
}

// Delete 发送 DELETE 请求
func (rc *RestClient) Delete() *RestClient {
	rc.setRequestMethod(DELETE)
	return rc.executeRequest()
}

// Stringify 获取响应数据字符串
func (rc *RestClient) Stringify() (string, error) {
	length := len(rc.responses)
	lastResp := rc.responses[length-1]
	if lastResp.err != nil {
		return "", lastResp.err
	}
	if lastResp.statusCode >= 400 {
		return string(lastResp.raw), &customerror.RequestFail{
			Details: strconv.Itoa(lastResp.statusCode) + ": " + string(lastResp.raw),
		}
	}
	return string(lastResp.raw), nil
}

// Bind 获取响应数据绑定到结构
func (rc *RestClient) Bind(v any) error {
	length := len(rc.responses)
	lastResp := rc.responses[length-1]
	if lastResp.err != nil {
		return lastResp.err
	}
	if lastResp.statusCode >= 400 {
		return &customerror.RequestFail{
			Details: strconv.Itoa(lastResp.statusCode) + ": " + string(lastResp.raw),
		}
	}
	return json.Unmarshal(lastResp.raw, v)
}

// TimesOfRetry 获取重试次数
func (rc *RestClient) TimesOfRetry() int {
	return len(rc.responses) - 1
}

// GetResponseStack 获取响应堆栈
func (rc *RestClient) GetResponseStack() []Response {
	return rc.responses
}

// RestClient 请求错误处理
func (rc *RestClient) handleRequestError(err error) {
	rc.responses = append(rc.responses, Response{
		err: err,
	})
}

// 处理 HTTP 响应
func (rc *RestClient) handleHTTPResponse(resp *http.Response, err error) *RestClient {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		rc.handleRequestError(err)
		return rc
	}
	if !rc.isJsonResponse(resp) {
		rc.handleRequestError(&customerror.NoRegularJsonResponse{
			Details: "raw response: " + string(body),
		})
		return rc
	}
	rc.responses = append(rc.responses, Response{
		statusCode: resp.StatusCode,
		Proto:      resp.Proto,
		raw:        body,
		headers:    &resp.Header,
		request:    &rc.request.req,
		TLS:        resp.TLS,
	})
	return rc
}

// ============================ 参数构建处理 =============================== //

// ParamValueBuilder 参数值构造器
type ParamValueBuilder interface {
	Build(value interface{}) (string, bool)
}

// StringValueBuilder 字符串参数值构造器
type StringValueBuilder struct{}

func (*StringValueBuilder) Build(value any) (string, bool) {
	valueType := reflect.TypeOf(value)
	if valueType != nil && valueType.Name() == "string" {
		v, ok := value.(string)
		if !ok {
			return "", false
		}
		return v, true
	}
	return "", false
}

// IntegerValueBuilder 整型参数值构造器
type IntegerValueBuilder struct{}

func (*IntegerValueBuilder) Build(value interface{}) (string, bool) {
	result, err := base.IntegerToString(value)
	if err != nil {
		return "", false
	}
	return result, true
}

// FloatValueBuilder 浮点型参数构造器
type FloatValueBuilder struct{}

func (*FloatValueBuilder) Build(value interface{}) (string, bool) {
	result, err := base.FloatToString(value)
	if err != nil {
		return "", false
	}
	return result, true
}

// BoolValueBuilder 布尔型参数值构造器
type BoolValueBuilder struct{}

func (*BoolValueBuilder) Build(value interface{}) (string, bool) {
	result, err := base.BoolToString(value)
	if err != nil {
		return "", false
	}
	return result, true
}

func (*RestClient) buildParams(param interface{}) string {
	paramBuilderChain := []ParamValueBuilder{
		&StringValueBuilder{},
		&IntegerValueBuilder{},
		&BoolValueBuilder{},
	}
	for _, builder := range paramBuilderChain {
		if val, ok := builder.Build(param); ok {
			return val
		}
	}
	return ""
}

// BuildQueryParamsToURL 拼接 Query 参数到 URL
func (*RestClient) buildQueryParamsToURL(url string, params Data) string {
	paramBuilderChain := []ParamValueBuilder{
		&StringValueBuilder{},
		&IntegerValueBuilder{},
		&BoolValueBuilder{},
	}
	existPrefix := strings.Contains(url, "?")
	for k, v := range params {
		if !existPrefix {
			url += "?"
			existPrefix = true
		} else {
			url += "&"
		}
		url += k
		url += "="
		for _, builder := range paramBuilderChain {
			val, ok := builder.Build(v)
			if ok {
				url += val
				continue
			}
		}
	}
	return url
}

// 拼接 Path 参数到 URL
func (*RestClient) buildPathParamsToURL(url string, params []interface{}) string {
	if len(params) == 0 {
		return url
	}
	var strBuilder strings.Builder
	paramBuilderChain := []ParamValueBuilder{
		&StringValueBuilder{},
		&IntegerValueBuilder{},
		&BoolValueBuilder{},
	}
	queryIndex := strings.IndexByte(url, '?')
	pathIndex := strings.LastIndexByte(url, '/')
	if queryIndex != -1 {
		strBuilder.WriteString(url[:queryIndex])
	} else {
		strBuilder.WriteString(url)
	}
	if strBuilder.Len() > 0 &&
		strBuilder.String()[strBuilder.Len()-1] != '/' &&
		pathIndex != len(url)-1 {
		strBuilder.WriteString("/")
	}
	count := 0
	for _, v := range params {
		for _, builder := range paramBuilderChain {
			val, ok := builder.Build(v)
			if ok {
				if count > 0 {
					strBuilder.WriteString("/")
				}
				strBuilder.WriteString(val)
				count++
				break
			}
		}
	}
	if queryIndex != -1 {
		strBuilder.WriteString(url[queryIndex:])
	}
	return strBuilder.String()
}
