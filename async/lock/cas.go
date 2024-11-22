package lock

import (
	"sync/atomic"
)

// CASSignal CAS 信号量
type CASSignal struct {
	state int64 // 信号量标记
}

// Add 增加信号量
func (s *CASSignal) Add() {
	for !atomic.CompareAndSwapInt64(&s.state, s.state, s.state+1) {
	}
}

// Done 减少信号量
func (s *CASSignal) Done() {
	for !atomic.CompareAndSwapInt64(&s.state, s.state, s.state-1) {
	}
}

// Status 读取信号量状态
func (s *CASSignal) Status() int64 {
	return atomic.LoadInt64(&s.state)
}

// CASSwitch CAS 原子开关
type CASSwitch struct {
	state int64 // 信号量标记
}

// Off 关闭原子开关
func (s *CASSwitch) Off() {
	for !atomic.CompareAndSwapInt64(&s.state, s.state, 0) {
	}
}

// On 开启原子开关
func (s *CASSwitch) On() {
	for !atomic.CompareAndSwapInt64(&s.state, s.state, 1) {
	}
}

// Status 获取原子开关状态
func (s *CASSwitch) Status() int64 {
	return atomic.LoadInt64(&s.state)
}
