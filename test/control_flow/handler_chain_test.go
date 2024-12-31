package control_flow

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/control_flow"
	"testing"
)

type HandlerA[T, R int] struct {
	control_flow.DefaultControlFlowHandler[T, R]
}

func (*HandlerA[T, R]) ShouldHandle(x T, _ ...any) bool {
	if x == 1 {
		return true
	}
	return false
}

func (*HandlerA[T, R]) Handle(_ T, _ ...any) (*R, error) {
	var res *R
	return res, nil
}

func TestControlFlowChain(t *testing.T) {
	chain := control_flow.ControlFlowChain[int, int]{
		&control_flow.DefaultControlFlowHandler[int, int]{},
	}
	fmt.Println(chain.Handle(1, true))
}
