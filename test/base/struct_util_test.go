package test

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/Anonymouscn/go-partner/base"
	jsonutil "github.com/Anonymouscn/go-partner/json"
	"io"
	"strconv"
	"testing"
	"time"
)

// ================================================================================ //
//                                                                                  //
//  struct util 测试                                                                //
//  @author anonymous                                                               //
//  @updated_at 2024.11.16 20:45:19                                                 //
//                                                                                  //
//  @cmd_help:                                                                      //
//  1. unit test:                                                                   //
//     $ go test xxx                                                                //
//  2. bench test:                                                                  //
//     $ go test -benchmem -run=^$ -bench ^<$function_name>$ -count=<$count> -v     //
//                                                                                  //
//                                                                                  //
// ================================================================================ //

// reqData1 请求数据-1
type reqData1 struct {
	V1  int
	V2  int8
	V3  int16
	V4  int32
	V5  int64
	V6  uint
	V7  uint8
	V8  uint16
	V9  uint32
	V10 uint64
	V11 float32
	V12 float64
	V13 bool
	V14 byte
	V15 string
}

// reqDataMap1 请求数据map-1
var reqDataMap1 = map[string]any{
	"v1":  1,
	"v2":  int8(86),
	"v3":  int16(7567),
	"v4":  int32(-856),
	"v5":  int64(7567),
	"v6":  uint(65464364),
	"v7":  uint8(36),
	"v8":  uint16(53454),
	"v9":  uint32(56757),
	"v10": uint64(99999999999),
	"v11": float32(1.21),
	"v12": 1024.13,
	"v13": true,
	"v14": byte(2),
	"v15": "success",
}

// reqData2 请求数据-2
type reqData2 struct {
	V1  *int
	V2  *int8
	V3  *int16
	V4  *int32
	V5  *int64
	V6  *uint
	V7  *uint8
	V8  *uint16
	V9  *uint32
	V10 *uint64
	V11 *float32
	V12 *float64
	V13 *bool
	V14 *byte
	V15 *string
}

// reqData3 请求数据-3
type reqData3 struct {
	V1 time.Time
	V2 *time.Time
	V3 int64
	V4 time.Time
}

// reqDataMap3 请求数据map-3
var reqDataMap3 = map[string]any{
	"v1": time.Now().Unix(),
	"v2": time.Now().Unix(),
	"v3": time.Now().Unix(),
	"v4": time.Now(),
}

// reqData4 请求数据-4
type reqData4 struct {
	V *req4SubData1
}

type req4SubData1 struct {
	V req4SubData2
}

type req4SubData2 struct {
	V *req4SubData3
}

type req4SubData3 struct {
	V req4SubData4
}

type req4SubData4 struct {
	V *req4SubData5
}

type req4SubData5 struct {
	C int
	S string
	T *time.Time
}

var reqDataMap4 = map[string]any{
	"v": map[string]any{
		"v": map[string]any{
			"v": map[string]any{
				"v": map[string]any{
					"v": map[string]any{
						"c": 200,
						"s": "success",
						"t": time.Now().Unix(),
					},
				},
			},
		},
	},
}

// JsonResp json 响应
type JsonResp[T any] struct {
	Code int    `json:"code"` // 业务响应码
	Msg  string // 业务消息
	Data *T     // 业务数据
}

// RespData1 响应数据-1
type RespData1 struct {
	User       Resp1SubData1
	Role       *Resp1SubData2
	Permission *Resp1SubData3
	*Resp1SubData4
}

// Resp1SubData1 响应数据1子数据-1
type Resp1SubData1 struct {
	Name    string
	Age     int
	Phone   string
	IsAdmin bool
	Tag     []string
}

// Resp1SubData2 响应数据1子数据-2
type Resp1SubData2 struct {
	Name      string
	CreatedAt *time.Time
	CreatedBy string
}

// Resp1SubData3 响应数据1子数据-3
type Resp1SubData3 struct {
}

// Resp1SubData4 响应数据1子数据-4
type Resp1SubData4 struct {
	Timestamp time.Time
}

// RespData2 响应数据-2
type RespData2 struct {
	V1 *int
	V2 *int16
	V3 *float64
	V4 *string
}

// RespData3 响应数据-3
type RespData3 struct {
	Resp3Anon1
	Resp3Anon2
	Grey bool
}

type Resp3Anon1 struct {
	Signature string
	Timestamp *time.Time
}

type Resp3Anon2 struct {
	Resp3Anon3
	SN string `json:"sn"`
}

type Resp3Anon3 struct {
	CorpID string `json:"corp_id"`
	AppID  string `json:"app_id"`
}

// testStructToMapExampleList 结构体转 map 测试用例表
var testStructToMapExampleList = []base.TestFn{
	func() string {
		t := time.Now()
		resp := &JsonResp[RespData1]{
			Code: 200,
			Msg:  "success",
			Data: &RespData1{
				User: Resp1SubData1{
					Name:    "anonymous",
					Age:     36,
					Phone:   "7746273846",
					IsAdmin: true,
					Tag:     []string{"programmer", "cool-guys", "crazy-man"},
				},
				Role: &Resp1SubData2{
					Name:      "admin",
					CreatedAt: &t,
					CreatedBy: "system",
				},
				Permission: nil,
			},
		}
		resp.Data.Resp1SubData4 = &Resp1SubData4{}
		resp.Data.Timestamp = time.Now()
		res := base.StructToMap(resp)
		j, _ := jsonutil.MapToJsonString(res)
		return fmt.Sprintf("%v %v\n %v\n", res, res["data"], j)
	},
	func() string {
		resp := &JsonResp[any]{
			Code: 500,
			Msg:  "Error: record not found",
		}
		res := base.StructToMap(resp)
		j, _ := jsonutil.MapToJsonString(res)
		return fmt.Sprintf("%v %v\n %v\n", res, res["data"], j)
	},
	func() string {
		v1, v2, v3, v4 := 1, int16(2), 3.14, "success"
		resp := &JsonResp[RespData2]{
			Code: 200,
			Msg:  "success",
			Data: &RespData2{
				V1: &v1,
				V2: &v2,
				V3: &v3,
				V4: &v4,
			},
		}
		res := base.StructToMap(resp)
		return fmt.Sprintf("%v\n", res)
	},
	func() string {
		resp := &JsonResp[RespData3]{
			Code: 200,
			Msg:  "success",
			Data: &RespData3{
				Grey: true,
			},
		}
		t := time.Now()
		resp.Data.CorpID = "Gdwuygdu3ge-e2df2f3revr"
		resp.Data.AppID = "app424729472424"
		resp.Data.SN = "SN-dejwfhewfwkefbwjekfwkefwefewf"
		resp.Data.Timestamp = &t
		h := md5.New()
		_, _ = io.WriteString(h, resp.Data.SN+"$"+strconv.FormatInt(resp.Data.Timestamp.Unix(), 10))
		resp.Data.Signature = base64.StdEncoding.EncodeToString(h.Sum(nil))
		res := base.StructToMap(resp)
		j, _ := jsonutil.MapToJsonString(res)
		return fmt.Sprintf("%v %v\n %v\n", res, res["data"], j)
	},
}

// TestStructToMap1 结构体转 map 单元测试-1
func TestStructToMap1(t *testing.T) {
	fmt.Println(testStructToMapExampleList[0]())
}

// TestStructToMap2 结构体转 map 单元测试-2
func TestStructToMap2(t *testing.T) {
	fmt.Println(testStructToMapExampleList[1]())
}

// TestStructToMap3 结构体转 map 单元测试-3
func TestStructToMap3(t *testing.T) {
	fmt.Println(testStructToMapExampleList[2]())
}

// TestStructToMap4 结构体转 map 单元测试-4
func TestStructToMap4(t *testing.T) {
	fmt.Println(testStructToMapExampleList[3]())
}

// BenchmarkStructToMap 结构体转 map 基准测试
func BenchmarkStructToMap(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, fn := range testStructToMapExampleList {
			fn()
		}
	}
}

// testMapToStructExampleList map 转结构体测试用例表
var testMapToStructExampleList = []func() string{
	func() string {
		d1 := &reqData1{}
		base.MapToStruct(reqDataMap1, d1)
		return fmt.Sprintf("%v\n", d1)
	},
	func() string {
		reqDataMap2 := make(map[string]any)
		for k, v := range reqDataMap1 {
			reqDataMap2[k] = &v
		}
		d2 := &reqData2{}
		base.MapToStruct(reqDataMap1, d2)
		return fmt.Sprintf("%v\n", d2)
	},
	func() string {
		d3 := &reqData3{}
		base.MapToStruct(reqDataMap3, d3)
		return fmt.Sprintf("%v\n", d3)
	},
	func() string {
		d4 := &reqData4{}
		base.MapToStruct(reqDataMap4, d4)
		return fmt.Sprintf("%v %v %v %v %v %v\n", d4, d4.V, d4.V.V, d4.V.V.V, d4.V.V.V.V, d4.V.V.V.V.V)
	},
}

// TestMapToStruct1 map 转结构体单元测试-1
func TestMapToStruct1(t *testing.T) {
	fmt.Println(testMapToStructExampleList[0]())
}

// TestMapToStruct2 map 转结构体单元测试-2
func TestMapToStruct2(t *testing.T) {
	fmt.Println(testMapToStructExampleList[1]())
}

// TestMapToStruct3 map 转结构体单元测试-3
func TestMapToStruct3(t *testing.T) {
	fmt.Println(testMapToStructExampleList[2]())
}

// TestMapToStruct4 map 转结构体单元测试-4
func TestMapToStruct4(t *testing.T) {
	fmt.Println(testMapToStructExampleList[3]())
}

// BenchmarkMapToStruct map 转结构体基准测试-1
func BenchmarkMapToStruct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, fn := range testMapToStructExampleList {
			fn()
		}
	}
}
