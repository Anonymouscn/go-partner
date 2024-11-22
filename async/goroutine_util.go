package async

import (
	"github.com/petermattis/goid"
)

// GetGoRoutineID 获取 go routine id
func GetGoRoutineID() int32 {
	return int32(goid.Get())
}
