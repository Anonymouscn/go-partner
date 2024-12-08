package random

import "fmt"

const (
	SnowIDCustomBitSum  = 22 // 雪花 id 可客制化比特位
	SnowIDSignBits      = 1  // 雪花 id 符号比特位
	SnowIDTimestampBits = 41 // 雪花 id 时间戳比特位
	SnowIDMachineBits   = 8  // 雪花 id 机器数比特位
	SnowIDLoopBits      = 12 // 雪花 id 时间偏移比特位
	SnowIDSeqBits       = 2  // 雪花 id 序列比特位
)

// SnowIDDefine 雪花 id 定义
type SnowIDDefine struct {
	sign      int // 符号位填充
	timestamp int // 毫秒时间戳
	machine   int // 机器 id
	loop      int // 时间偏移量
	seq       int // 序列号
}

// CustomSnowIDTemplate 客制化雪花 id 模版
type CustomSnowIDTemplate struct {
	Machine int // 机器 id
	Loop    int // 时间偏移量
	Seq     int // 序列号
}

// DefaultSnowIDDefine 默认雪花 id 定义
func DefaultSnowIDDefine() *SnowIDDefine {
	return &SnowIDDefine{
		sign:      SnowIDSignBits,
		timestamp: SnowIDTimestampBits,
		machine:   SnowIDMachineBits,
		loop:      SnowIDLoopBits,
		seq:       SnowIDSeqBits,
	}
}

// CustomSnowIDDefine 客制化雪花 id 定义
func CustomSnowIDDefine(template *CustomSnowIDTemplate) *SnowIDDefine {
	checkCustomTemplate(template)
	return &SnowIDDefine{
		sign:      SnowIDSignBits,
		timestamp: SnowIDTimestampBits,
		machine:   template.Machine,
		loop:      template.Loop,
		seq:       template.Seq,
	}
}

// checkCustomTemplate 校验客制化模版是否合规
func checkCustomTemplate(template *CustomSnowIDTemplate) {
	if template == nil {
		panic("CustomSnowIDTemplate cannot be nil !!!")
	}
	c := template.Seq + template.Loop + template.Machine
	if c != SnowIDCustomBitSum {
		panic(fmt.Sprintf("CustomSnowIDTemplate illegal: custom bit sum %v is illegal (must be %v) !!!", c, SnowIDCustomBitSum))
	}
}

// SnowIDGenerator 雪花 id 生成器
type SnowIDGenerator struct {
	config *SnowIDDefine // 雪花 id 配置
}

// CreateSnowIDGenerator 创建雪花 id 生成器
func CreateSnowIDGenerator(config *SnowIDDefine, machineID, loop *int32, seq int) *SnowIDGenerator {
	return &SnowIDGenerator{
		config: config,
	}
}

// Generate 生成雪花 id
func (generator *SnowIDGenerator) Generate() {

}
