package lock

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/async/lock"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ================================================================================ //
//                                                                                  //
//  lock 工具集 测试                                                                  //
//  @author anonymous                                                               //
//  @updated_at 2024.11.22 17:51:19                                                 //
//                                                                                  //
//  @cmd_help:                                                                      //
//  1. unit test:                                                                   //
//     $ go test xxx                                                                //
//  2. bench test:                                                                  //
//     $ go test -benchmem -run=^$ -bench ^<$function_name>$ -count=<$count> -v     //
//                                                                                  //
//                                                                                  //
// ================================================================================ //

// TestCustomLock 测试自定义锁
func TestCustomLock(t *testing.T) {
	goroutineCount := 2*runtime.NumCPU() + 1
	c := make(chan string, goroutineCount*2)
	l := lock.CustomLock{}
	var wg sync.WaitGroup
	// 用于检测并发访问的共享变量
	var sharedCounter int32 = 0
	wg.Add(goroutineCount)
	for i := 0; i < goroutineCount; i++ {
		id := i
		// 模拟并发争夺锁
		go func() {
			defer wg.Done()
			// 尝试获取锁
			if ok := l.TryLock(1 * time.Second); ok {
				defer l.Unlock()
				// 增加共享计数器
				current := atomic.AddInt32(&sharedCounter, 1)
				// 临界区断言
				if current != 1 {
					t.Errorf("Critical section violation: %d goroutines in critical section", current)
				}
				// 记录获取锁
				c <- fmt.Sprintf("goroutine[%v] get lock", id)
				// 模拟工作负载
				time.Sleep(50 * time.Millisecond)
				// 减少共享计数器
				atomic.AddInt32(&sharedCounter, -1)
				// 记录释放锁
				c <- fmt.Sprintf("goroutine[%v] return lock", id)
			}
		}()
	}
	// 等待所有 goroutine 完成并关闭 channel
	go func() {
		wg.Wait()
		close(c)
	}()
	// 收集并验证所有消息
	lockEvents := make(map[int]bool) // 记录每个 goroutine 的锁状态
	for msg := range c {
		fmt.Println(msg)
		var id int
		if strings.Contains(msg, "get lock") {
			_, _ = fmt.Sscanf(msg, "goroutine[%d]", &id)
			if lockEvents[id] {
				t.Errorf("Goroutine %d acquired lock while already holding it", id)
			}
			lockEvents[id] = true
		} else if strings.Contains(msg, "return lock") {
			_, _ = fmt.Sscanf(msg, "goroutine[%d]", &id)
			if !lockEvents[id] {
				t.Errorf("Goroutine %d released lock without acquiring it", id)
			}
			delete(lockEvents, id)
		}
	}
	// 验证最终状态
	if len(lockEvents) > 0 {
		t.Errorf("Some goroutines did not release their locks: %v", lockEvents)
	}
	if atomic.LoadInt32(&sharedCounter) != 0 {
		t.Errorf("Final counter value is %d, expected 0", atomic.LoadInt32(&sharedCounter))
	}
}

// 模拟临界区的工作负载
func doSomething() {
	// 模拟一个很小的工作量
	_ = 1 + 1
}

// BenchmarkCustomLock 测试 CustomLock
func BenchmarkCustomLock(b *testing.B) {
	// 测试不同的并发数
	for _, goroutines := range []int{1, 10, 100, 1000} {
		b.Run("Goroutines-"+strconv.Itoa(goroutines), func(b *testing.B) {
			l := &lock.CustomLock{}
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					// 使用推荐的 TryLock 方法，设置合理的超时时间
					if l.TryLock(time.Millisecond * 100) {
						doSomething()
						l.Unlock()
					}
					//doSomething()
					//lock.Unlock()
				}
			})
		})
	}
}

// BenchmarkMutex 测试 sync.Mutex
func BenchmarkMutex(b *testing.B) {
	// 测试相同的并发数
	for _, goroutines := range []int{1, 10, 100, 1000} {
		b.Run("Goroutines-"+strconv.Itoa(goroutines), func(b *testing.B) {
			l := &sync.Mutex{}
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					l.Lock()
					doSomething()
					l.Unlock()
				}
			})
		})
	}
}

// BenchmarkLongCriticalSection 测试临界区场景
func BenchmarkLongCriticalSection(b *testing.B) {
	longWork := func() {
		time.Sleep(time.Millisecond * 200) // 模拟较长的临界区操作
	}

	b.Run("CustomLock", func(b *testing.B) {
		l := &lock.CustomLock{}
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if l.TryLock(time.Millisecond * 100) {
					longWork()
					l.Unlock()
				}
			}
		})
	})

	b.Run("Mutex", func(b *testing.B) {
		l := &sync.Mutex{}
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Lock()
				longWork()
				l.Unlock()
			}
		})
	})
}
