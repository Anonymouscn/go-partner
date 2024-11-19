package runtime

import "time"

var (
	GlobalTypeFactory = newTypeFactory() // 全局类型工厂
)

// typeFactory 类型工厂
type typeFactory struct {
	config *TypeFactoryConfig // 配置信息
}

// typeInfo 类型信息
type typeInfo struct {
	ID   string // 类型 ID
	Name string // 类型名称
}

// typeCache 类型缓存
type typeCache struct {
	TypeInfo *typeInfo     // 类型信息
	TTL      time.Duration // 缓存生存时间
}

// typeCacheMonitor 类型缓存监视器
type typeCacheMonitor struct {
}

// newTypeFactory 新建类型工厂
func newTypeFactory() *typeFactory {
	return &typeFactory{}
}

// TypeFactoryConfig 类型工厂配置
type TypeFactoryConfig struct {
	Enable bool // 是否启用类型工厂

}

// Init 类型工厂初始化
func (*typeFactory) Init() {

}

// ApplyConfig 应用类型工厂配置
func (f *typeFactory) ApplyConfig() *typeFactory {
	return f
}

// GetConfig 获取类型工厂配置
func (f *typeFactory) GetConfig() TypeFactoryConfig {
	return *f.config
}

// Reload 重载类型工厂
func (f *typeFactory) Reload() {

}

// Registry 类型注册
func (*typeFactory) Registry() {

}

// GetTypeInfo 获取类型信息
func (*typeFactory) GetTypeInfo() {

}

// GetTypeInfoByName 根据类型名称获取类型信息
func (*typeFactory) GetTypeInfoByName() {

}

// NewInstance 实例化类型
func (*typeFactory) NewInstance() {

}

// NewInstanceByName 根据类型名称实例化类型
func (*typeFactory) NewInstanceByName() {

}
