package lock

import (
	"github.com/Anonymouscn/go-partner/async"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	locked     = uint64(1) // 锁定值
	ownerShift = 32        // 锁拥有者 id 偏移量
	maxSpins   = 4         // 最大自旋次数
	yieldSpins = 2         // 进入 yield 阶段临界值
	maxBackoff = 32        // 最大退避时间
)

// CustomLock 自定义锁 (试验中)
type CustomLock struct {
	state  uint64   // 高32位存储 owner，低32位存储 state
	waiter uint32   // 等待者计数
	_      [48]byte // 字节填充缓存行, 避免伪共享
}

// 快速获取锁（无竞争路径）
//
//go:nosplit
func (l *CustomLock) fastLock(gid uint64) bool {
	return atomic.CompareAndSwapUint64(&l.state, 0, locked|gid)
}

// 快速检查重入
//
//go:nosplit
func (l *CustomLock) isReentrant(gid uint64) bool {
	return atomic.LoadUint64(&l.state)>>ownerShift == gid>>ownerShift
}

// Lock 上锁
func (l *CustomLock) Lock() {
	// 获取当前goroutine ID（只获取一次）
	gid := uint64(async.GetGoRoutineID()) << ownerShift
	// 快速路径1：无竞争直接获取
	if l.fastLock(gid) {
		return
	}
	// 快速路径2：检查重入
	if l.isReentrant(gid) {
		return
	}
	// 增加等待者计数
	atomic.AddUint32(&l.waiter, 1)
	defer atomic.AddUint32(&l.waiter, ^uint32(0))
	// 自旋获取锁
	spins := 0
	backoff := uint32(1)
	var lastState uint64
	for {
		state := atomic.LoadUint64(&l.state)
		// 避免在状态未改变时进行 CAS
		if state == lastState {
			// 优化的自旋策略
			if spins < maxSpins {
				spins++
				if spins < yieldSpins {
					// 短自旋
					for i := 0; i < int(backoff); i++ {
						runtime.Gosched()
					}
					// 指数退避，但有上限
					if backoff < maxBackoff {
						backoff <<= 1
					}
				} else {
					runtime.Gosched()
				}
				continue
			}
			// 超过自旋限制，进入睡眠
			runtime.Gosched()
			spins = 0
			backoff = 1
			continue
		}
		// 状态发生改变，尝试获取锁
		if state == 0 {
			if atomic.CompareAndSwapUint64(&l.state, 0, locked|gid) {
				return
			}
		}
		// 更新上次观察到的状态
		lastState = state
	}
}

// TryLock 尝试上锁
func (l *CustomLock) TryLock(ttl time.Duration) bool {
	gid := uint64(async.GetGoRoutineID()) << ownerShift
	deadline := time.Now().Add(ttl)
	// 快速路径
	if l.fastLock(gid) {
		return true
	}
	// 检查重入
	if l.isReentrant(gid) {
		return true
	}
	// 增加等待者计数
	atomic.AddUint32(&l.waiter, 1)
	defer atomic.AddUint32(&l.waiter, ^uint32(0))
	spins := 0
	backoff := uint32(1)
	var lastState uint64
	for time.Now().Before(deadline) {
		state := atomic.LoadUint64(&l.state)
		if state == lastState {
			if spins < maxSpins {
				spins++
				if spins < yieldSpins {
					for i := 0; i < int(backoff); i++ {
						runtime.Gosched()
					}
					if backoff < maxBackoff {
						backoff <<= 1
					}
				} else {
					runtime.Gosched()
				}
				continue
			}
			runtime.Gosched()
			spins = 0
			backoff = 1
			continue
		}
		if state == 0 && atomic.CompareAndSwapUint64(&l.state, 0, locked|gid) {
			return true
		}
		lastState = state
	}
	return false
}

// Unlock 解锁
func (l *CustomLock) Unlock() {
	gid := uint64(async.GetGoRoutineID()) << ownerShift
	// 快速检查所有者
	state := atomic.LoadUint64(&l.state)
	if state>>ownerShift != gid>>ownerShift {
		panic("unlock of mutex not owned by current goroutine")
	}
	// 检查是否有等待者
	if atomic.LoadUint32(&l.waiter) == 0 {
		// 无等待者，直接释放
		atomic.StoreUint64(&l.state, 0)
		return
	}
	// 有等待者，使用 自旋 + CAS 确保正确释放
	for !atomic.CompareAndSwapUint64(&l.state, state, 0) {
		state = atomic.LoadUint64(&l.state)
		if state>>ownerShift != gid>>ownerShift {
			panic("unlock of mutex not owned by current goroutine")
		}
	}
}

// IsLocked 是否上锁
func (l *CustomLock) IsLocked() bool {
	return atomic.LoadUint64(&l.state)&locked == locked
}
