package sort

// Sortable 可排序接口
type Sortable interface {
	Compare(x, y any) int
}
