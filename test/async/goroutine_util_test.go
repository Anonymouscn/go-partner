package async

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/async"
	"github.com/Anonymouscn/go-partner/base"
	"strings"
	"sync"
	"testing"
)

// testGetGoRoutineIDExampleList 获取 go routine id 测试用例列表
var testGetGoRoutineIDExampleList = []base.TestFn{
	func() string {
		wg := sync.WaitGroup{}
		c := make(chan string)
		defer close(c)
		for i := 0; i < 24; i++ {
			wg.Add(1)
			go func() {
				c <- fmt.Sprintf("goroutine id [%v]", async.GetGoRoutineID())
				wg.Done()
			}()
		}
		var builder strings.Builder
		go func() {
			for v := range c {
				builder.WriteString(v + "\n")
			}
		}()
		wg.Wait()
		return builder.String()
	},
}

// TestGetGoRoutineID 获取 go routine id 单元测试
func TestGetGoRoutineID(t *testing.T) {
	fmt.Println(testGetGoRoutineIDExampleList[0]())
}

// BenchmarkGetGoRoutineID
func BenchmarkGetGoRoutineID(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, e := range testGetGoRoutineIDExampleList {
			e()
		}
	}
}
