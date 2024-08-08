package restful

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/Anonymouscn/go-partner/base"
	customerror "github.com/Anonymouscn/go-partner/error"
	iotools "github.com/Anonymouscn/go-partner/io"
	"io"
	"net/http"
	urlpkg "net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// RestClient Restful 客户端
type RestClient struct {
	conf      *RestClientConfig // 客户端配置
	client    http.Client       // http 客户端
	request   *Request          // 请求
	responses []*Response       // 响应栈
}

// RestClientConfig RestClient 配置
type RestClientConfig struct {
	EnableRetry    bool           // 启用重试
	MaxRetry       int            // 最大重试次数
	RetryDelay     time.Duration  // 重试间隔时间
	RequestTimeout time.Duration  // 超时时间
	Transport      http.Transport // transport 配置
}

// ParamsConfig 参数配置
type ParamsConfig struct {
	URL     string
	Path    Path
	Headers Data
	Query   Data
	Body    Body
}

// NewRestClient 新建 Restful 客户端
func NewRestClient() *RestClient {
	client := &RestClient{
		conf: &RestClientConfig{
			EnableRetry:    false,           // 默认不启用重试
			MaxRetry:       0,               // 默认最大重试次数 0
			RetryDelay:     0,               // 默认重试间隔时间 0
			RequestTimeout: 2 * time.Minute, // 默认请求超时 2 min
		},
		request: &Request{
			url:   "",
			path:  make(Path, 0),
			query: make(Data),
			body:  struct{}{},
		},
		responses: make([]*Response, 0),
	}
	// 初始化 headers map
	client.request.req.Header = make(http.Header)
	return client
}

// ApplyConfig 应用配置文件
func (rc *RestClient) ApplyConfig(conf *RestClientConfig) *RestClient {
	rc.conf = conf
	return rc
}

// ApplyParams 应用参数
func (rc *RestClient) ApplyParams(paramsConfig *ParamsConfig) *RestClient {
	if paramsConfig != nil {
		rc.request = &Request{
			url:   paramsConfig.URL,
			path:  paramsConfig.Path,
			query: paramsConfig.Query,
			body:  paramsConfig.Body,
		}
		if paramsConfig.Headers != nil {
			rc.SetHeaders(paramsConfig.Headers)
		}
	}
	return rc
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
	rc.responses = make([]*Response, 0)
	return rc
}

// DisableCertAuth 禁用 TLS 证书验证
func (rc *RestClient) DisableCertAuth() *RestClient {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, DisableKeepAlives: true}
	rc.client.Transport = transport
	return rc
}

// ApplyTransPort 应用 transport
func (rc *RestClient) ApplyTransPort(transport *http.Transport) *RestClient {
	rc.client.Transport = transport
	return rc
}

// 生成请求 URL
func (rc *RestClient) generateURL() string {
	url := rc.buildPathParamsToURL(rc.request.url, rc.request.path)
	url = rc.buildQueryParamsToURL(url, rc.request.query)
	return url
}

// 生成请求体
func (rc *RestClient) generateBody() error {
	data, err := json.Marshal(rc.request.body)
	if err != nil {
		return err
	}
	if rc.request.body != nil {
		rc.request.req.Body = io.NopCloser(bytes.NewReader(data))
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

// 处理请求
func (rc *RestClient) handleRequest() *RestClient {
	rc.setRequestURL(rc.generateURL())
	err := rc.generateBody()
	if err != nil {
		rc.handleRequestError(err)
		return rc
	}
	retry := rc.conf.MaxRetry
	if !rc.conf.EnableRetry {
		retry = 0
	}
	for len(rc.responses) <= retry {
		response, err := rc.executeRequest()
		if err != nil {
			rc.handleRequestError(err)
		} else {
			rc.responses = append(rc.responses, response)
			break
		}
		time.Sleep(rc.conf.RetryDelay)
	}
	return rc
}

// 执行一次请求
func (rc *RestClient) executeRequest() (*Response, error) {
	resp, err := rc.client.Do(&rc.request.req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, &customerror.NoContentResponse{}
	}
	defer iotools.CloseReader(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 { // 非正常响应
		return nil, &customerror.RequestFail{
			Details: "raw response: " + string(body),
		}
	} else if !rc.isJsonResponse(resp) { // 非 json 响应
		return nil, &customerror.NoRegularJsonResponse{
			Details: "raw response: " + string(body),
		}
	}
	// 正常响应
	return &Response{
		StatusCode: resp.StatusCode,
		Proto:      resp.Proto,
		Raw:        body,
		Headers:    &resp.Header,
		Request:    &rc.request.req,
		TLS:        resp.TLS,
	}, nil
}

// Get 发送 GET 请求
func (rc *RestClient) Get() *RestClient {
	rc.setRequestMethod(GET)
	return rc.handleRequest()
}

// Post 发送 POST 请求
func (rc *RestClient) Post() *RestClient {
	rc.setRequestMethod(POST)
	return rc.handleRequest()
}

// Put 发送 PUT 请求
func (rc *RestClient) Put() *RestClient {
	rc.setRequestMethod(PUT)
	return rc.handleRequest()
}

// Delete 发送 DELETE 请求
func (rc *RestClient) Delete() *RestClient {
	rc.setRequestMethod(DELETE)
	return rc.handleRequest()
}

// Stringify 获取响应数据字符串
func (rc *RestClient) Stringify() (string, error) {
	length := len(rc.responses)
	lastResp := rc.responses[length-1]
	if lastResp.Err != nil {
		return "", lastResp.Err
	}
	if lastResp.StatusCode >= 300 {
		return string(lastResp.Raw), &customerror.RequestFail{
			Details: strconv.Itoa(lastResp.StatusCode) + ": " + string(lastResp.Raw),
		}
	}
	return string(lastResp.Raw), nil
}

// Bind 获取响应数据绑定到结构
func (rc *RestClient) Bind(v any) error {
	length := len(rc.responses)
	lastResp := rc.responses[length-1]
	if lastResp.Err != nil {
		return lastResp.Err
	}
	if lastResp.StatusCode >= 300 {
		return &customerror.RequestFail{
			Details: strconv.Itoa(lastResp.StatusCode) + ": " + string(lastResp.Raw),
		}
	}
	return json.Unmarshal(lastResp.Raw, v)
}

// ResponseHeaders 获取响应请求 (以最后次请求为准)
func (rc *RestClient) ResponseHeaders() *http.Header {
	return rc.responses[len(rc.responses)-1].Headers
}

// TimesOfRetry 获取重试次数
func (rc *RestClient) TimesOfRetry() int {
	return len(rc.responses) - 1
}

// GetResponseStack 获取响应堆栈
func (rc *RestClient) GetResponseStack() []*Response {
	return rc.responses
}

// RestClient 请求错误处理
func (rc *RestClient) handleRequestError(err error) {
	rc.responses = append(rc.responses, &Response{
		Err: err,
	})
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
