package random

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/random"
	"testing"
)

// TestSnowIDGenerate 雪花 ID 生成单元测试
func TestSnowIDGenerate(t *testing.T) {
	IDGenerator := random.CreateSnowIDGenerator(
		&random.SnowIDDefine{
			Machine: 8,
			Offset:  12,
			Seq:     2,
		},
		0,
		0,
		0,
	)
	fmt.Println(IDGenerator.Generate())
}
