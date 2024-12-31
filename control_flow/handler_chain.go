package control_flow

// ControlFlowChain 控制流处理链
type ControlFlowChain[T, R any] []ControlFlowHandler[T, R]

func (flow ControlFlowChain[T, R]) Handle(params T, once bool, opts ...any) (product *R, err error) {
	for _, handler := range flow {
		if handler.ShouldHandle(params, opts) {
			product, err = handler.Handle(params, opts)
			if once || err != nil {
				return product, err
			}
		}
	}
	return
}

// ControlFlowHandler 控制流处理器
type ControlFlowHandler[T any, R any] interface {
	ShouldHandle(T, ...any) bool
	Handle(T, ...any) (*R, error)
}

// DefaultControlFlowHandler 默认控制流处理器
type DefaultControlFlowHandler[T any, R any] struct{}

func (*DefaultControlFlowHandler[T, R]) ShouldHandle(_ T, _ ...any) bool {
	return true
}

func (*DefaultControlFlowHandler[T, R]) Handle(_ T, _ ...any) (*R, error) {
	return nil, nil
}
