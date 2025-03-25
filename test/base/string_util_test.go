package test

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/base"
	"testing"
)

// TestRandStr 测试生成随机字符串
func TestRandStr(t *testing.T) {
	for i := 0; i < 32; i++ {
		hash, err := base.RandString(32)
		if err != nil {
			panic(err)
		}
		serviceToken := "sk_" + hash
		fmt.Println(serviceToken)
	}
}
