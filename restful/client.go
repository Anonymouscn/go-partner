package restful

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/Anonymouscn/go-partner/base"
	customerror "github.com/Anonymouscn/go-partner/error"
	iotools "github.com/Anonymouscn/go-partner/io"
	"github.com/Anonymouscn/go-partner/net"
	"github.com/bytedance/sonic"
	"io"
	"net/http"
	urlpkg "net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
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
	URL    string        // 请求 url
	Path   Path          // 请求路径
	Header Data          // 请求头数据
	Query  any           // 请求行数据
	Body   any           // 请求体数据
	Data   any           // 自动化处理数据
	Raw    []byte        // 原生请求体数据 (不需要 json 转换处理)
	Form   urlpkg.Values // 表单数据
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
			path:  make(Path, 0),
			query: make(Data),
			body:  make(Data),
			data:  make(Data),
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
			query: make(map[string]any),
			body:  make(map[string]any),
			data:  make(map[string]any),
			raw:   paramsConfig.Raw,
			form:  paramsConfig.Form,
		}
		// 请求头处理
		if paramsConfig.Header != nil {
			rc.ApplyHeaders(paramsConfig.Header)
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

// ResetPath 重置路径参数
func (rc *RestClient) ResetPath() *RestClient {
	rc.request.path = make(Path, 0)
	return rc
}

// SetHeaders 设置请求头 (等同于 ApplyHeaders, 兼容旧版本)
func (rc *RestClient) SetHeaders(headers Data) *RestClient {
	return rc.ApplyHeaders(headers)
}

// ApplyHeaders 应用请求头
func (rc *RestClient) ApplyHeaders(headers Data) *RestClient {
	for k, v := range headers {
		rc.request.req.Header.Set(k, rc.buildParams(v))
	}
	return rc
}

// AddHeaders 添加请求头
func (rc *RestClient) AddHeaders(headers Data) *RestClient {
	for k, v := range headers {
		rc.request.req.Header.Add(k, rc.buildParams(v))
	}
	return rc
}

// ResetHeaders 重置请求头
func (rc *RestClient) ResetHeaders() *RestClient {
	rc.request.req.Header = make(http.Header)
	return rc
}

// ====================================== Cookie 相关配置 ====================================== //

// ContainsCookie 是否存在 Cookie
func (rc *RestClient) ContainsCookie() bool {
	c := rc.request.req.Header.Values("Cookie")
	return c != nil && len(c) > 0
}

// GetCookie 获取 Cookie
func (rc *RestClient) GetCookie() map[string]string {
	if !rc.ContainsCookie() {
		return map[string]string{}
	}
	c := strings.Split(rc.request.req.Header.Values("Cookie")[0], ";")
	cookie := make(map[string]string)
	for _, item := range c {
		entry := strings.Split(item, "=")
		if len(entry) == 2 {
			cookie[entry[0]] = entry[1]
		}
	}
	return cookie
}

// GetCookieValue 获取 Cookie 值
func (rc *RestClient) GetCookieValue(key string) string {
	return rc.GetCookie()[key]
}

// SetCookies 设置 Cookies
func (rc *RestClient) SetCookies(cookies Data) *RestClient {
	m := make(map[string]string)
	for k, v := range cookies {
		val, err := base.AnyToString(v)
		if err != nil {
			fmt.Printf("set cookies error: %v\n", err)
			return rc
		}
		m[k] = val
	}
	rc.ApplyHeaders(Data{"Cookie": rc.generateCookie(m)})
	return rc
}

// RemoveCookies 移除 Cookies
func (rc *RestClient) RemoveCookies(cookies ...string) *RestClient {
	cookie := rc.GetCookie()
	for _, c := range cookies {
		cookie[c] = ""
	}
	rc.ApplyHeaders(Data{"Cookie": rc.generateCookie(cookie)})
	return rc
}

// ResetCookies 重置 Cookies
func (rc *RestClient) ResetCookies() *RestClient {
	rc.ApplyHeaders(Data{"Cookie": ""})
	return rc
}

// generateCookie 生成 Cookie 字符串
func (rc *RestClient) generateCookie(cm map[string]string) string {
	var builder strings.Builder
	for k, v := range cm {
		builder.WriteString(k + "=" + v)
		builder.WriteString(";")
	}
	return builder.String()
}

// ==================================== 参数直接装配方法定义 ==================================== //

// paramsToMap 参数转 map
func (*RestClient) paramsToMap(params any) (map[string]any, error) {
	pm, err := base.AnyToMap(params)
	if err != nil {
		fmt.Printf("convert params[%v] fail: %v\n", params, err)
		return nil, err
	}
	return pm, nil
}

// applyParams 应用参数
func (rc *RestClient) applyParams(params []any) map[string]any {
	l := len(params)
	m := map[string]any{}
	if l == 0 {
		return m
	} else if l > 1 {
		tm := make(map[string]any)
		for _, p := range params {
			if pm, err := rc.paramsToMap(p); err == nil {
				base.MapCopyOnNotExist(tm, pm)
			}
		}
		return tm
	} else {
		if pm, err := rc.paramsToMap(params[0]); err == nil {
			return pm
		}
	}
	return m
}

// addParams 添加参数
func (rc *RestClient) addParams(m map[string]any, params any, overwrite bool) {
	if pm, err := rc.paramsToMap(params); err == nil {
		if overwrite {
			base.MapCopy(m, pm)
		} else {
			base.MapCopyOnNotExist(m, pm)
		}
	}
}

// applyFormParams 应用表单参数
func (rc *RestClient) applyFormParams(params []any) (map[string][]string, error) {
	form := make(map[string][]string)
	for _, item := range params {
		m, err := rc.paramsToMap(item)
		if err != nil {
			return nil, err
		}
		for k, v := range m {
			if err := rc.setFormParams(form, k, v); err != nil {
				return nil, err
			}
		}
	}
	return form, nil
}

// addFormParams 添加表单参数
func (rc *RestClient) addFormParams(m map[string][]string, params any, overwrite bool) error {
	if pm, err := rc.paramsToMap(params); err == nil {
		if overwrite {
			for k, v := range pm {
				if err := rc.setFormParams(m, k, v); err != nil {
					return err
				}
			}
		} else {
			for k, v := range pm {
				if m[k] == nil {
					if err := rc.setFormParams(m, k, v); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// setFormParams 设置表单参数
func (*RestClient) setFormParams(m map[string][]string, key string, params any) error {
	p, err := base.AnyToString(params)
	if err != nil {
		return err
	}
	m[key] = []string{p}
	return nil
}

// SetData 设置自动化参数 (GET/DELETE: SetQuery, POST/PUT: SetBody)
func (rc *RestClient) SetData(params ...any) *RestClient {
	rc.request.data = rc.applyParams(params)
	return rc
}

// AddData 添加自动化参数 (GET/DELETE: SetQuery, POST/PUT: SetBody)
func (rc *RestClient) AddData(params any) *RestClient {
	rc.addParams(rc.request.data, params, true)
	return rc
}

// ResetData 重置自动化参数
func (rc *RestClient) ResetData() *RestClient {
	rc.request.data = make(Data)
	return rc
}

// SetQuery 设置请求行参数
func (rc *RestClient) SetQuery(params ...any) *RestClient {
	rc.request.query = rc.applyParams(params)
	return rc
}

// AddQuery 添加请求行参数
func (rc *RestClient) AddQuery(params any) *RestClient {
	rc.addParams(rc.request.query, params, true)
	return rc
}

// ResetQuery 重置请求行参数
func (rc *RestClient) ResetQuery() *RestClient {
	rc.request.query = make(Data)
	return rc
}

// SetBody 设置请求体参数
func (rc *RestClient) SetBody(params ...any) *RestClient {
	rc.request.body = rc.applyParams(params)
	return rc
}

// AddBody 添加请求体参数
func (rc *RestClient) AddBody(params any) *RestClient {
	rc.addParams(rc.request.body, params, true)
	return rc
}

// ResetBody 重置请求体参数
func (rc *RestClient) ResetBody() *RestClient {
	rc.request.body = make(Data)
	return rc
}

// SetBodyRaw 设置原生请求体参数 (参数直接装配, 无需 json 化)
func (rc *RestClient) SetBodyRaw(raw []byte) *RestClient {
	rc.request.raw = raw
	return rc
}

// SetBodyRawString 设置原生请求体参数 (参数直接装配, 无需 json 化)
func (rc *RestClient) SetBodyRawString(raw string) *RestClient {
	rc.request.raw = []byte(raw)
	return rc
}

// ResetBodyRaw 重置原生请求体参数
func (rc *RestClient) ResetBodyRaw() *RestClient {
	rc.request.raw = nil
	return rc
}

// SetForm 设置表单参数
func (rc *RestClient) SetForm(params ...any) *RestClient {
	rc.request.form, _ = rc.applyFormParams(params)
	return rc
}

// AddForm 添加表单参数
func (rc *RestClient) AddForm(params any) *RestClient {
	_ = rc.addFormParams(rc.request.form, params, true)
	return rc
}

// ResetForm 重置表单参数
func (rc *RestClient) ResetForm() *RestClient {
	rc.request.form = nil
	return rc
}

// SetJsonString 直接装配 json 字符串请求体 (待废除方法, 与新方法 SetBodyRaw, SetBodyRawString 作用一致)
// Deprecated: SetJsonString 方法将于 2024.12.01 废除, 请使用新方法: SetBodyRawString
func (rc *RestClient) SetJsonString(json string) *RestClient {
	return rc.SetBodyRawString(json)
}

// ClearResponses 清除所有响应信息
func (rc *RestClient) ClearResponses() *RestClient {
	rc.responses = make([]*Response, 0)
	return rc
}

// Reset 参数重置
func (rc *RestClient) Reset() *RestClient {
	return rc.ResetHeaders().
		ResetPath().
		ResetQuery().
		ResetBody().
		ResetBodyRaw().
		ResetData().
		ResetForm().
		ClearResponses()
}

// ==================================================================================== //

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
	if !rc.isEmptyRaw() {
		rc.request.req.Body = io.NopCloser(bytes.NewReader(rc.request.raw))
	} else if rc.request.body != nil {
		data, err := sonic.Marshal(rc.request.body)
		if err != nil {
			return err
		}
		if rc.request.body != nil {
			rc.request.req.Body = io.NopCloser(bytes.NewReader(data))
		}
	}
	return nil
}

func (rc *RestClient) isEmptyRaw() bool {
	r := rc.request.raw
	return r == nil || len(r) <= 0
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
func (rc *RestClient) isJsonResponse(body []byte) bool {
	return sonic.Valid(body)
}

// 处理请求
func (rc *RestClient) handleRequest() *RestClient {
	// 处理自动参数
	rc.handleData()
	// 生成请求行
	rc.setRequestURL(rc.generateURL())
	// 生成请求体
	err := rc.generateBody()
	if err != nil {
		rc.handleRequestError(err)
		return rc
	}
	// 应用重试配置
	retry := rc.conf.MaxRetry
	if !rc.conf.EnableRetry {
		retry = 0
	}
	// 自动计算 Content-Length
	if rc.request.body != nil {
		rc.AddHeaders(Data{"Content-Length": strconv.Itoa(int(unsafe.Sizeof(rc.request.body)))})
	}
	// 应用重试方案并请求
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
	var (
		resp *http.Response
		err  error
	)
	if rc.request.form != nil {
		resp, err = rc.client.PostForm(rc.request.url, rc.request.form)
	} else {
		resp, err = rc.client.Do(&rc.request.req)
	}
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
	} else if !rc.isJsonResponse(body) { // 非 json 响应
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

// handleData 处理动态参数数据
func (rc *RestClient) handleData() *RestClient {
	if len(rc.request.data) > 0 {
		switch rc.request.req.Method {
		case GET, DELETE:
			rc.AddQuery(rc.request.data)
		case POST, PUT:
			rc.AddBody(rc.request.data)
		}
	}
	return rc
}

// Do 发起指定类型请求
func (rc *RestClient) Do(method Method) *RestClient {
	rc.setRequestMethod(method)
	return rc.handleRequest()
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

// action 实际发送请求方法
func (rc *RestClient) action() (*Response, error) {
	lastResp := rc.responses[len(rc.responses)-1]
	if lastResp.Err != nil {
		return nil, lastResp.Err
	}
	if lastResp.StatusCode >= 300 {
		return nil, &customerror.RequestFail{
			Details: strconv.Itoa(lastResp.StatusCode) + ": " + string(lastResp.Raw),
		}
	}
	return lastResp, nil
}

// Stringify 获取响应数据字符串
func (rc *RestClient) Stringify() (string, error) {
	resp, err := rc.action()
	if err != nil {
		return "", err
	}
	return string(resp.Raw), nil
}

// Bind 获取响应数据绑定到结构 (字段强校验)
func (rc *RestClient) Bind(v any) error {
	resp, err := rc.action()
	if err != nil {
		return err
	}
	return sonic.Unmarshal(resp.Raw, v)
}

// Map 获取响应数据映射到结构 (字段弱校验)
func (rc *RestClient) Map(v any) error {
	resp, err := rc.action()
	if err != nil {
		return err
	}
	m := make(map[string]any)
	if err := sonic.Unmarshal(resp.Raw, &m); err != nil {
		return err
	}
	base.MapToStruct(m, v)
	return nil
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
	Build(value any) (string, bool)
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

func (*IntegerValueBuilder) Build(value any) (string, bool) {
	result, err := base.IntegerToString(value)
	if err != nil {
		return "", false
	}
	return result, true
}

// FloatValueBuilder 浮点型参数构造器
type FloatValueBuilder struct{}

func (*FloatValueBuilder) Build(value any) (string, bool) {
	result, err := base.FloatToString(value)
	if err != nil {
		return "", false
	}
	return result, true
}

// BoolValueBuilder 布尔型参数值构造器
type BoolValueBuilder struct{}

func (*BoolValueBuilder) Build(value any) (string, bool) {
	result, err := base.BoolToString(value)
	if err != nil {
		return "", false
	}
	return result, true
}

func (*RestClient) buildParams(param any) string {
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
				url += net.EncodeURL(val)
				continue
			}
		}
	}
	return url
}

// 拼接 Path 参数到 URL
func (*RestClient) buildPathParamsToURL(url string, params []any) string {
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

// 是否是合法数据
func (rc *RestClient) isLegalData(data any) (reflect.Kind, bool, string) {
	if data == nil {
		return reflect.Invalid, false, "<nil>"
	}
	t := reflect.TypeOf(data)
	k := t.Kind()
	sign := k.String()
	switch k {
	case reflect.Struct:
		return k, true, sign
	case reflect.Map:
		keyTypeDesc := t.Key().String()
		valueTypeDesc := t.Elem().String()
		sign = fmt.Sprintf("%v[%v]%v", k.String(), keyTypeDesc, valueTypeDesc)
		if keyTypeDesc == "string" {
			return k, true, sign
		}
		return k, false, sign
	case reflect.Ptr:
		_, ok, _ := rc.isLegalData(reflect.Indirect(reflect.ValueOf(data)))
		return reflect.Pointer, ok, fmt.Sprintf("<ptr>")
	}
	return k, false, sign
}

// ConvertData 转换数据 any(struct/map[string]T) => map[string]any
func (rc *RestClient) ConvertData(data any) (map[string]any, error) {
	d, ok, sign := rc.isLegalData(data)
	if !ok {
		return nil, errors.New(fmt.Sprintf("params type %v is illegal !", sign))
	}
	if d == reflect.Pointer {
		return rc.ConvertData(reflect.ValueOf(data).Elem().Interface())
	}
	if d == reflect.Struct {
		return base.StructToMap(data), nil
	}
	if strings.Contains(sign, "interface") {
		if m, ok := data.(map[string]any); ok {
			return m, nil
		}
	} else {
		m := make(map[string]any)
		v := reflect.ValueOf(data)
		for _, key := range v.MapKeys() {
			m[key.String()] = v.MapIndex(key).Interface()
		}
		return m, nil
	}
	return nil, errors.New(fmt.Sprintf("params %v convert error !", sign))
}
