package test

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/restful"
	"testing"
)

// ================================================================================ //
//                                                                                  //
//  rest client 测试                                                                 //
//  @author anonymous                                                               //
//  @updated_at 2024.11.15 00:13:09                                                 //
//                                                                                  //
//  @cmd_help:                                                                      //
//  1. unit test:                                                                   //
//     $ go test xxx                                                                //
//  2. bench test:                                                                  //
//     $ go test -benchmem -run=^$ -bench ^<$function_name>$ -count=<$count> -v     //
//                                                                                  //
//  @notice:                                                                        //
//   1. 单元测试 api 均为公共 api, 请合理使用, 勿用于非法用途或恶意攻击, 否则后果自负。        //
//   2. 若有提交 PR 补充测试用例或提供复现 bug 场景, 提供的 json 信息和代码注意脱敏处理。      //
//   3. 欢迎提交 PR, 不论是 bug 反馈还是代码优化, 感谢~                                   //
//                                                                                  //
//                                                                                  //
// ================================================================================ //

// testRestClientDoRequestList 测试 RestClient 请求测试用例列表
var testRestClientDoRequestList = []TestFn{
	func() string {
		return ""
	},
}

// BaiduHotResp 百度热榜响应
type BaiduHotResp[T any] struct {
	Code int
	Msg  string
	Data T
	Time float64 `json:"exec_time"`
	IP   string  `json:"ip"`
}

type BaiduHotData struct {
	Content    []*BaiduHotContent
	TopContent []BaiduHotContent
	UpdateTime string
	TypeName   string
}

type BaiduHotContent struct {
	Word      string
	Desc      string
	HotChange string
	HotScore  int64
	Index     int
	HotTag    string
	HotTagImg string
	Img       string
	URL       string `json:"url"`
}

// TestRestClientDoRequest1 RestClient 请求单元测试-1
func TestRestClientDoRequest1(t *testing.T) {
	fmt.Println(testRestClientDoRequestList[0]())
	rc := restful.NewRestClient().
		SetURL("https://cn.apihz.cn/api/xinwen/baidu.php").
		SetData(
			&struct {
				ID  int `json:"id"`
				Key int
			}{
				ID:  8,
				Key: 8,
			}).
		Get()
	resp := &BaiduHotResp[BaiduHotData]{}
	err := rc.Map(resp)
	if err != nil {
		fmt.Println("error: ", err)
		return
	}
	fmt.Println("response: ", resp)
}

// BenchmarkRestClientDoRequest RestClient 请求基准测试
func BenchmarkRestClientDoRequest(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, fn := range testRestClientDoRequestList {
			fn()
		}
	}
}
