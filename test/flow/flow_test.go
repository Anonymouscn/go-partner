package flow

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/flow"
	"testing"
)

// ================================================================================ //
//                                                                                  //
//  flow 流式处理工具集 测试                                                           //
//  @author anonymous                                                               //
//  @updated_at 2024.11.21 14:32:19                                                 //
//                                                                                  //
//  @cmd_help:                                                                      //
//  1. unit test:                                                                   //
//     $ go test xxx                                                                //
//  2. bench test:                                                                  //
//     $ go test -benchmem -run=^$ -bench ^<$function_name>$ -count=<$count> -v     //
//                                                                                  //
//                                                                                  //
// ================================================================================ //

func TestDataFlow(t *testing.T) {
	f := flow.NewDataFlow[int](20)
	m := make(map[int]int)
	// 注入消费者 1
	f.Consume(func(dc <-chan int, ec chan<- error, args ...any) {
		for v := range dc {
			fmt.Println("receive message: ", v)
			m[v] = v
		}
	})
	// 注入生产者 1
	f.Produce(func(dc chan<- int, ec chan<- error, args ...any) {
		for i := 0; i < 1000; i++ {
			dc <- i
		}
	})
	// 注入生产者 2
	f.Produce(func(dc chan<- int, ec chan<- error, args ...any) {
		for i := 0; i < 1000; i++ {
			dc <- i
		}
	})
	// 注入生产者 3
	f.Produce(func(dc chan<- int, ec chan<- error, args ...any) {
		for i := 0; i < 1000; i++ {
			dc <- i
		}
	})
	// 外部终止数据流
	f.Stop()
	fmt.Println("success !")
	fmt.Println(m)
}

func BenchmarkTestDataFlow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f := flow.NewDataFlow[int](20)
		m := make(map[int]int)
		// 注入消费者 1
		f.Consume(func(dc <-chan int, ec chan<- error, args ...any) {
			for v := range dc {
				fmt.Println("receive message: ", v)
				m[v] = v
			}
		})
		// 注入生产者 1
		f.Produce(func(dc chan<- int, ec chan<- error, args ...any) {
			for i := 0; i < 1000; i++ {
				dc <- i
			}
		})
		// 注入生产者 2
		f.Produce(func(dc chan<- int, ec chan<- error, args ...any) {
			for i := 0; i < 1000; i++ {
				dc <- i
			}
		})
		// 注入生产者 3
		f.Produce(func(dc chan<- int, ec chan<- error, args ...any) {
			for i := 0; i < 1000; i++ {
				dc <- i
			}
		})
		// 外部终止数据流
		f.Stop()
		fmt.Println("success !")
		fmt.Println(m)
	}
}
