package flow

import (
	"runtime"
	"sync/atomic"
)

// todo 需要新增数据分发器 - 实现多个消费方分发

//// DataDispatcher 数据分发器
//type DataDispatcher[T any] struct {
//	InOperationState  lock.CustomLock // 输入流操作锁
//	OutOperationState lock.CustomLock // 输出流操作锁
//	InBoundMap        sync.Map        // 输入流表 map[Flow ID]Flow
//	OutBoundMap       sync.Map        // 输出流表 map[Flow ID]Flow
//	RouteMap          sync.Map        // 路由表 map[InBound Flow ID][InBound Flow ID List]
//	//InBounds          []*DataFlow[T] // 输入流
//	//OutBounds         []*DataFlow[T] // 输出流
//}
//
//// NewDataDispatcher 新建数据分发器
//func NewDataDispatcher[T any]() *DataDispatcher[T] {
//	return &DataDispatcher[T]{
//		InBoundMap:  sync.Map{},
//		OutBoundMap: sync.Map{},
//		RouteMap:    sync.Map{},
//		//InBounds:  make([]*DataFlow[T], 0),
//		//OutBounds: make([]*DataFlow[T], 0),
//	}
//}
//
//// AddInBound 添加输入流
//func (d *DataDispatcher[T]) AddInBound(flow *DataFlow[T]) *DataDispatcher[T] {
//	d.InOperationState.Lock()
//	defer d.InOperationState.Unlock()
//
//	//d.InBounds = append(d.InBounds, flow)
//	return d
//}
//
//// RemoveInBound 移除输入流
//func (d *DataDispatcher[T]) RemoveInBound(flow *DataFlow[T]) *DataDispatcher[T] {
//	d.InOperationState.Lock()
//	defer d.InOperationState.Unlock()
//	return d
//}
//
//// AddOutBound 添加输出流
//func (d *DataDispatcher[T]) AddOutBound(flow *DataFlow[T]) *DataDispatcher[T] {
//	//d.OutBounds = append(d.OutBounds, flow)
//	return d
//}
//
//// RemoveOutBound 移除输出流
//func (d *DataDispatcher[T]) RemoveOutBound(flow *DataFlow[T]) *DataDispatcher[T] {
//	return d
//}
//
//// AddRoute 添加路由规则
//func (d *DataDispatcher[T]) AddRoute(inboundID, outboundID string) *DataDispatcher[T] {
//	return d
//}
//
//// RemoveRoute 移除路由规则
//func (d *DataDispatcher[T]) RemoveRoute(inboundID, outboundID string) *DataDispatcher[T] {
//	return d
//}
//
//// ClearAllRouteWithInbound 清除所有指定输入流的路由规则
//func (d *DataDispatcher[T]) ClearAllRouteWithInbound(inboundID string) *DataDispatcher[T] {
//	return d
//}
//
//// ClearAllRouteWithOutbound 清除所有指定输出流的路由规则
//func (d *DataDispatcher[T]) ClearAllRouteWithOutbound(outboundID string) *DataDispatcher[T] {
//	return d
//}
//
//// ResetRoute 重置路由规则
//func (d *DataDispatcher[T]) ResetRoute() *DataDispatcher[T] {
//	return d
//}

// DataDispatcher 数据分发器
type DataDispatcher struct {
}

func NewDataDispatcher() *DataDispatcher {
	return &DataDispatcher{}
}

// 数据流模式常量枚举
const (
	DataConsume  = iota // 数据消费
	DataDispatch        // 数据分发
)

// 数据流状态常量枚举
const (
	ClosedState  = iota // 数据流关闭状态
	ClosingState        // 数据流关闭中状态
	StartedState        // 数据流开启状态
)

// DataFlow 数据流
type DataFlow[T any] struct {
	ID             string          // 数据流 ID
	DataChannel    chan T          // 数据管道
	ErrChannel     chan error      // 错误管道
	DataDispatcher *DataDispatcher // 数据分发器
	Counter        int64           // 生产者计数器
	State          int64           // 管道状态 [0:无生产者无消费者, 1:无生产者有消费者, 2:有生产者有消费者]
	FlowMode       int32           // 数据流模式 [DataConsume:数据消费, DataDispatch:数据分发; 默认数据消费模式]
}

// NewDataFlow 新建数据流
func NewDataFlow[T any](bufSize uint) *DataFlow[T] {
	flow := &DataFlow[T]{
		DataChannel:    make(chan T, bufSize),
		ErrChannel:     make(chan error, bufSize/20+2),
		DataDispatcher: NewDataDispatcher(),
	}
	return flow.Start()
}

// UseConsumeMode 使用消费模式
func (f *DataFlow[T]) UseConsumeMode() *DataFlow[T] {
	for !atomic.CompareAndSwapInt32(&f.FlowMode, f.FlowMode, DataConsume) {
	}
	return f
}

// UseDispatchMode 使用分发模式
func (f *DataFlow[T]) UseDispatchMode() *DataFlow[T] {
	for !atomic.CompareAndSwapInt32(&f.FlowMode, f.FlowMode, DataDispatch) {
	}
	return f
}

// CustomDispatcher 自定义分发器
func (f *DataFlow[T]) CustomDispatcher(dispatcher *DataDispatcher) *DataFlow[T] {
	f.DataDispatcher = dispatcher
	return f
}

// Status 获取数据流状态
func (f *DataFlow[T]) Status() int64 {
	return atomic.LoadInt64(&f.State)
}

// Start 开启数据流
func (f *DataFlow[T]) Start() *DataFlow[T] {
	if atomic.LoadInt64(&f.State) == ClosedState {
		// 启用管道, 允许注册生产方
		for !atomic.CompareAndSwapInt64(&f.State, ClosedState, StartedState) {
		}
	}
	return f
}

// Stop 关闭数据流
func (f *DataFlow[T]) Stop() *DataFlow[T] {
	// 关闭管道, 禁止注册生产方
	for !atomic.CompareAndSwapInt64(&f.State, StartedState, ClosingState) {
	}
	// 等待生产方结束生产
	for atomic.LoadInt64(&f.Counter) > 0 {
		runtime.Gosched()
	}
	// 关闭数据和错误处理队列
	close(f.DataChannel)
	close(f.ErrChannel)
	// 等待消费方停止消费
	for atomic.LoadInt64(&f.State) != ClosedState {
		runtime.Gosched()
	}
	return f
}

// ProduceFn 生产方法
type ProduceFn[T any] func(dc chan<- T, ec chan<- error, args ...any)

// Produce 注入生产方
func (f *DataFlow[T]) Produce(fn ProduceFn[T], args ...any) *DataFlow[T] {
	// 注册生产方
	for atomic.LoadInt64(&f.State) == StartedState && !atomic.CompareAndSwapInt64(&f.Counter, f.Counter, f.Counter+1) {
	}
	// 数据流已关闭, 注册失败
	if atomic.LoadInt64(&f.State) != StartedState {
		return f
	}
	// 异步生产
	go func() {
		defer func() {
			for !atomic.CompareAndSwapInt64(&f.Counter, f.Counter, f.Counter-1) {
			}
		}()
		fn(f.DataChannel, f.ErrChannel, args)
	}()
	return f
}

// ConsumeFn 消费方法
type ConsumeFn[T any] func(dc <-chan T, ec chan<- error, args ...any)

// Consume 注入消费方
func (f *DataFlow[T]) Consume(fn ConsumeFn[T], args ...any) *DataFlow[T] {
	// 异步消费
	go func() {
		fn(f.DataChannel, f.ErrChannel, args)
		defer func() {
			for !atomic.CompareAndSwapInt64(&f.State, ClosingState, ClosedState) {
			}
		}()
	}()
	return f
}

// ErrorHandleFn 错误处理方法
type ErrorHandleFn[T any] func(ec <-chan error, args ...any)

// OnError 错误处理
func (f *DataFlow[T]) OnError(fn ErrorHandleFn[T], args ...any) *DataFlow[T] {
	// 异步错误处理
	go func() {
		fn(f.ErrChannel, args)
	}()
	return f
}
