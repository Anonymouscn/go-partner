package test

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/base"
	"strconv"
	"testing"
)

type Object struct {
	A int
	B string
	C float64
	D bool
}

func TestSliceToMap(t *testing.T) {
	l := make([]Object, 0)
	for i := 0; i < 20; i++ {
		l = append(l, Object{
			A: i,
			B: strconv.Itoa(i),
			C: float64(i) + 0.01,
			D: i%2 == 0,
		})
	}
	m := base.SliceToMap[Object, int](l, func(object Object) int {
		return object.A
	})
	fmt.Println(m)
}
