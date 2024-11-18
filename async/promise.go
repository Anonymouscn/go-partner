package async

import (
	"context"
	"reflect"
)

// Action 异步操作
// res: 上一步执行结果, fn: 本次执行函数, args: 本次执行函数入参
//type Action func(fn any, args ...any)

type Action struct {
	Fn   any   // 本次执行函数
	Args []any // 本次执行函数入参
}

// Valid 验证 action 合法性
func (action *Action) Valid() (bool, reflect.Value) {
	fn := reflect.ValueOf(action.Fn)
	if fn.Kind() != reflect.Func {
		return false, fn
	}
	for _, arg := range action.Args {
		v := reflect.ValueOf(arg)
		if !v.IsValid() {
			return false, fn
		}
	}
	return true, fn
}

// Promise 异步结果 (promise 语法糖)
type Promise struct {
	actions [][]*Action     // 操作表
	ctx     context.Context // 上下文
	cursor  int64           // 调用栈指针
	errors  []*error        // 错误栈
	config  *PromiseConfig  // 配置
}

// PromiseConfig Promise 配置
type PromiseConfig struct {
}

// NewPromise 新建 Promise 异步逻辑
func NewPromise() *Promise {
	return &Promise{
		actions: make([][]*Action, 0),
		ctx:     context.Background(),
		cursor:  0,
		errors:  make([]*error, 0),
	}
}

// Apply 应用 Promise 配置
func (p *Promise) Apply(config *PromiseConfig) *Promise {
	p.config = config
	return p
}

// Then 异步执行下一步方法
func (p *Promise) Then(actions ...Action) *Promise {
	if p.shouldIgnore(actions) {
		return p
	}
	a := make([]*Action, 0)
	for _, action := range actions {
		a = append(a, &action)
	}
	p.actions = append(p.actions, a)
	go p.do()
	return p
}

// shouldIgnore 忽略无效 Action
func (*Promise) shouldIgnore(actions []Action) bool {
	return actions == nil || len(actions) == 0
}

// do 执行方法
func (p *Promise) do() *Promise {
	if len(p.actions[p.cursor]) > 1 {
		return p.async()
	}
	return p.sync()
}

// sync 同步执行方法
func (p *Promise) sync() *Promise {
	current := p.actions[p.cursor]
	if len(current) == 1 {
		a := current[0]
		if ok, _ := a.Valid(); ok {
			fn := reflect.ValueOf(a.Fn)
			resultList := fn.Call(convertParams(a.Args))
			p.ctx = context.WithValue(p.ctx, "", convertResult(resultList))
		}
	}
	p.cursor++
	return p
}

// convertParams 参数转换
func convertParams(args []any) []reflect.Value {
	res := make([]reflect.Value, 0)
	for _, arg := range args {
		res = append(res, reflect.ValueOf(arg))
	}
	return res
}

// convertResult 返回值转换
func convertResult(res []reflect.Value) []any {
	r := make([]any, 0)
	for _, rd := range res {
		r = append(r, rd.Interface())
	}
	return r
}

// async 异步执行方法
func (p *Promise) async() *Promise {
	return p
}

// Get 同步等待获取结果
func (*Promise) Get() {

}
