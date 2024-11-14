package test

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/net"
	"testing"
)

// ================================================================================ //
//                                                                                  //
//  url util 测试                                                                    //
//  @author anonymous                                                               //
//  @updated_at 2024.11.11 01:44:49                                                 //
//                                                                                  //
//  @cmd_help:                                                                      //
//  1. unit test:                                                                   //
//     $ go test xxx                                                                //
//  2. bench test:                                                                  //
//     $ go test -benchmem -run=^$ -bench ^<$function_name>$ -count=<$count> -v     //
//                                                                                  //
//                                                                                  //
// ================================================================================ //

// testEncodeURLExampleList 编码 url 测试用例列表
var testEncodeURLExampleList = []TestFn{
	func() string {
		return net.EncodeURL("你好, 世界!")
	},
}

// TestEncodeURL1 编码 url 单元测试-1
func TestEncodeURL1(t *testing.T) {
	fmt.Println(testEncodeURLExampleList[0]())
}

// BenchmarkEncodeURL 编码 url 基准测试
func BenchmarkEncodeURL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, t := range testEncodeURLExampleList {
			t()
		}
	}
}
