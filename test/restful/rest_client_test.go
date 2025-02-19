package test

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/base"
	"github.com/Anonymouscn/go-partner/restful"
	"testing"
	"time"
)

// ================================================================================ //
//                                                                                  //
//  rest client 测试                                                                 //
//  @author anonymous                                                               //
//  @updated_at 2024.11.16 20:46:49                                                 //
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
var testRestClientDoRequestList = []base.TestFn{
	func() string {
		rc := restful.NewRestClient().
			SetURL("https://api.bilibili.com/x/web-interface/wbi/search/square").
			SetQuery(
				&struct {
					Limit    int
					Platform string
					WebID    string    `json:"wrid"`
					Time     time.Time `json:"wts"`
				}{
					Limit:    10,
					Platform: "web",
					WebID:    "0219d6977966caeb40540387ebfa35e5",
					Time:     time.Now(),
				},
			).
			Get()
		resp := &BiliBiliHotResp[BiliBiliHotData]{}
		if err := rc.Map(resp); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("response: %v %v", resp, *resp.Data.Trending.List[0])
	},
	func() string {
		rc := restful.NewRestClient().SetURL("https://dog.ceo/api/breeds/image/random").Get()
		resp := &RandomDogResp{}
		if err := rc.Map(resp); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("response: %v", resp)
	},
	func() string {
		rc := restful.NewRestClient().SetURL("https://tenapi.cn/v2/toutiaohot").Get()
		resp := &NewsTodayHotResp[[]NewsTodayHotItem]{}
		if err := rc.Map(resp); err != nil {
			return fmt.Sprintf("error %v", err)
		}
		return fmt.Sprintf("response: %v", resp)
	},
	func() string {
		rc := restful.NewRestClient().
			SetURL("https://tenapi.cn/v2/bing").
			SetForm(
				struct {
					Format string
				}{
					Format: "json",
				},
			).
			Post()
		resp := &TenApiResp[BingPhotoInfo]{}
		if err := rc.Map(resp); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("response: %v", resp)
	},
	func() string {
		rc := restful.NewRestClient().SetURL("https://tenapi.cn/v2/history").Get()
		resp := &TenApiResp[HistoryToday]{}
		if err := rc.Map(resp); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("response: %v", resp)
	},
	func() string {
		rc := restful.NewRestClient().
			SetURL("https://tenapi.cn/v2/phone").
			SetForm(
				struct {
					Phone string `json:"tel"`
				}{
					Phone: "$phone",
				},
			).
			Post()
		resp := &TenApiResp[PhoneInfo]{}
		if err := rc.Map(resp); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("response: %v", resp)
	},
	func() string {
		rc := restful.NewRestClient().
			SetURL("https://$device_ip/redfish/v1/SessionService/Sessions").
			DisableCertAuth().
			SetData(
				&struct {
					Username string `json:"UserName"`
					Password string `json:"Password"`
				}{
					Username: "$username",
					Password: "$password",
				},
			).
			Post()
		resp := &ServerResp{}
		if err := rc.Map(resp); err != nil {
			return fmt.Sprintf("error: %v", err)
		}
		return fmt.Sprintf("response: %v", resp)
	},
}

type ServerResp struct {
	Context string `json:"@odata.context"`
	MID     string `json:"@odata.id"`
	ID      string `json:"Id"`
	Name    string `json:"Name"`
	OEM     OEM    `json:"Oem"`
}

type OEM struct {
	Info HuaweiServerInfo `json:"Huawei"`
}

type HuaweiServerInfo struct {
	Account                      string    `json:"UserAccount"`
	LoginTime                    string    `json:"LoginTime"`
	IP                           string    `json:"UserIP"`
	UserTag                      string    `json:"UserTag"`
	MySession                    bool      `json:"MySession"`
	UserID                       string    `json:"UserId"`
	UserValidDays                int       `json:"UserValidDays"`
	AccountInsecurePromptEnabled bool      `json:"AccountInsecurePromptEnabled"`
	UserRole                     []*string `json:"UserRole"`
}

type PhoneInfo struct {
	Local string
	Num   string
	Type  string
	ISP   string `json:"isp"`
	STD   string `json:"std"`
}

type HistoryToday struct {
	Today string
	List  []HistoryTodayItem
}

type HistoryTodayItem struct {
	Title string
	Year  string
	URL   string `json:"url"`
}

type TenApiResp[T any] struct {
	Code    int
	Message string `json:"msg"`
	Data    T
}

type BingPhotoInfo struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	Data          string `json:"data"`
	Width         int
	Height        int
	Copyright     string
	CopyrightLink string `json:"copyrightlink"`
}

type NewsTodayHotResp[T any] struct {
	Code    int
	Message string `json:"msg"`
	Data    T
}

type NewsTodayHotItem struct {
	Name string
	URL  string `json:"url"`
}

type RandomDogResp struct {
	Message string
	Status  string
}

type BiliBiliHotResp[T any] struct {
	Code    int
	Message string
	TTL     int `json:"ttl"`
	Data    T
}

type BiliBiliHotData struct {
	Trending TrendingInfo
}

type TrendingInfo struct {
	Title   string
	TrackID string             `json:"trackid"`
	List    []*BiliBiliHotItem `json:"list"`
	TopList []BiliBiliHotItem
}

type BiliBiliHotItem struct {
	Keyword   string
	ShowName  string
	Icon      string
	URI       string `json:"uri"`
	GoTo      string `json:"goto"`
	HeatScore int64
}

// TestRestClientDoRequest1 RestClient 请求单元测试-1
func TestRestClientDoRequest1(t *testing.T) {
	fmt.Println(testRestClientDoRequestList[0]())
}

// TestRestClientDoRequest2 RestClient 请求单元测试-2
func TestRestClientDoRequest2(t *testing.T) {
	fmt.Println(testRestClientDoRequestList[1]())
}

// TestRestClientDoRequest3 RestClient 请求单元测试-3
func TestRestClientDoRequest3(t *testing.T) {
	fmt.Println(testRestClientDoRequestList[2]())
}

// TestRestClientDoRequest4 RestClient 请求单元测试-4
func TestRestClientDoRequest4(t *testing.T) {
	fmt.Println(testRestClientDoRequestList[3]())
}

// TestRestClientDoRequest5 RestClient 请求单元测试-5
func TestRestClientDoRequest5(t *testing.T) {
	fmt.Println(testRestClientDoRequestList[4]())
}

// TestRestClientDoRequest6 RestClient 请求单元测试-6
func TestRestClientDoRequest6(t *testing.T) {
	fmt.Println(testRestClientDoRequestList[5]())
}

// TestRestClientDoRequest7 RestClient 请求单元测试-7
func TestRestClientDoRequest7(t *testing.T) {
	fmt.Println(testRestClientDoRequestList[6]())
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

func TestGetRequest(t *testing.T) {
	client := restful.NewRestClient().SetBodyRaw(nil).
		SetURL("https://api.openai.com/v1/models").
		SetHeaders(restful.Data{"Authorization": "Bearer " + "$key"}).
		Get()
	resp, err := client.Stringify()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(resp)
}

func TestGetSSEStream(t *testing.T) {
	client := restful.NewRestClient().
		SetURL("https://api.deepseek.com/chat/completions").
		SetHeaders(
			restful.Data{
				"Authorization": "Bearer " + "$Bearer",
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			},
		).SetBody(
		restful.Data{
			"model": "deepseek-chat",
			"messages": []restful.Data{
				{
					"role":    "system",
					"content": "你是智探科技研发的AI助手，请你解答用户提出的问题",
				},
				{
					"role":    "user",
					"content": "你是谁",
				},
			},
			"stream": true,
		},
	).Post()
	if err := client.Stream(func(chunk string, params ...any) {
		fmt.Println("receive: ", chunk)
	}); err != nil {
		fmt.Println(err)
	}
}
