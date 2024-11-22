package flow

import (
	"github.com/Anonymouscn/go-partner/async/lock"
	"runtime"
	"sync"
	"sync/atomic"
)

// todo 新增数据分发器 - 实现多个消费方分发

// DataDispatcher 数据分发器
type DataDispatcher[T any] struct {
	InOperationState  lock.CustomLock // 输入流操作锁
	OutOperationState lock.CustomLock // 输出流操作锁
	InBoundMap        sync.Map        // 输入流表 map[Flow ID]Flow
	OutBoundMap       sync.Map        // 输出流表 map[Flow ID]Flow
	RouteMap          sync.Map        // 路由表 map[InBound Flow ID][InBound Flow ID List]
	//InBounds          []*DataFlow[T] // 输入流
	//OutBounds         []*DataFlow[T] // 输出流
}

// NewDataDispatcher 新建数据分发器
func NewDataDispatcher[T any]() *DataDispatcher[T] {
	return &DataDispatcher[T]{
		InBoundMap:  sync.Map{},
		OutBoundMap: sync.Map{},
		RouteMap:    sync.Map{},
		//InBounds:  make([]*DataFlow[T], 0),
		//OutBounds: make([]*DataFlow[T], 0),
	}
}

// AddInBound 添加输入流
func (d *DataDispatcher[T]) AddInBound(flow *DataFlow[T]) *DataDispatcher[T] {
	d.InOperationState.Lock()
	defer d.InOperationState.Unlock()

	//d.InBounds = append(d.InBounds, flow)
	return d
}

// RemoveInBound 移除输入流
func (d *DataDispatcher[T]) RemoveInBound(flow *DataFlow[T]) *DataDispatcher[T] {
	d.InOperationState.Lock()
	defer d.InOperationState.Unlock()
	return d
}

// AddOutBound 添加输出流
func (d *DataDispatcher[T]) AddOutBound(flow *DataFlow[T]) *DataDispatcher[T] {
	//d.OutBounds = append(d.OutBounds, flow)
	return d
}

// RemoveOutBound 移除输出流
func (d *DataDispatcher[T]) RemoveOutBound(flow *DataFlow[T]) *DataDispatcher[T] {
	return d
}

// AddRoute 添加路由规则
func (d *DataDispatcher[T]) AddRoute(inboundID, outboundID string) *DataDispatcher[T] {
	return d
}

// RemoveRoute 移除路由规则
func (d *DataDispatcher[T]) RemoveRoute(inboundID, outboundID string) *DataDispatcher[T] {
	return d
}

// ClearAllRouteWithInbound 清除所有指定输入流的路由规则
func (d *DataDispatcher[T]) ClearAllRouteWithInbound(inboundID string) *DataDispatcher[T] {
	return d
}

// ClearAllRouteWithOutbound 清除所有指定输出流的路由规则
func (d *DataDispatcher[T]) ClearAllRouteWithOutbound(outboundID string) *DataDispatcher[T] {
	return d
}

// ResetRoute 重置路由规则
func (d *DataDispatcher[T]) ResetRoute() *DataDispatcher[T] {
	return d
}

// DataFlow 数据流
type DataFlow[T any] struct {
	ID          string     // 数据流 ID
	DataChannel chan T     // 数据管道
	ErrChannel  chan error // 错误管道
	Counter     int64      // 生产者计数器
	State       int64      // 管道状态
}

// NewDataFlow 新建数据流
func NewDataFlow[T any](bufSize uint) *DataFlow[T] {
	flow := &DataFlow[T]{
		DataChannel: make(chan T, bufSize),
		ErrChannel:  make(chan error, bufSize/20+2),
	}
	return flow.Start()
}

// Status 获取数据流状态
func (f *DataFlow[T]) Status() int64 {
	return atomic.LoadInt64(&f.State)
}

// Start 开启数据流
func (f *DataFlow[T]) Start() *DataFlow[T] {
	if atomic.LoadInt64(&f.State) == 0 {
		// 启用管道状态, 允许注册生产方
		for !atomic.CompareAndSwapInt64(&f.State, 0, 1) {
		}
	}
	return f
}

// Stop 关闭数据流
func (f *DataFlow[T]) Stop() *DataFlow[T] {
	defer close(f.DataChannel)
	defer close(f.ErrChannel)
	// 关闭管道状态, 禁止注册生产方
	for !atomic.CompareAndSwapInt64(&f.State, 1, 0) {
	}
	// 自旋等待生产方结束
	for atomic.LoadInt64(&f.Counter) > 0 {
		// 让出 cpu 时间片
		runtime.Gosched()
	}
	return f
}

// ProduceFn 生产方法
type ProduceFn[T any] func(dc chan<- T, ec chan<- error, args ...any)

// Produce 注入生产方
func (f *DataFlow[T]) Produce(fn ProduceFn[T], args ...any) *DataFlow[T] {
	// 注册生产者
	for atomic.LoadInt64(&f.State) > 0 && !atomic.CompareAndSwapInt64(&f.Counter, f.Counter, f.Counter+1) {
	}
	if atomic.LoadInt64(&f.State) == 0 {
		return f
	}
	defer func() {
		for !atomic.CompareAndSwapInt64(&f.Counter, f.Counter, f.Counter-1) {
		}
	}()
	fn(f.DataChannel, f.ErrChannel, args)
	return f
}

// ConsumeFn 消费方法
type ConsumeFn[T any] func(dc <-chan T, ec <-chan error, args ...any)

// Consume 注入消费方
func (f *DataFlow[T]) Consume(fn ConsumeFn[T], args ...any) *DataFlow[T] {
	fn(f.DataChannel, f.ErrChannel, args)
	return f
}
