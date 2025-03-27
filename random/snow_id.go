package random

import (
	"fmt"
	"github.com/Anonymouscn/go-partner/base"
	"time"
)

const (
	SnowIDSign               = 0  // 雪花 ID 符号位
	SnowIDBits               = 64 // 雪花 ID 总比特位
	SnowIDCustomBits         = 22 // 雪花 ID 可客制化比特位
	SnowIDSignBits           = 1  // 雪花 ID 符号比特位
	SnowIDTimestampBits      = 41 // 雪花 ID 时间戳比特位
	SnowIDDefaultMachineBits = 8  // 雪花 ID 机器数默认比特位
	SnowIDDefaultOffsetBits  = 12 // 雪花 ID 时间偏移默认比特位
	SnowIDDefaultSeqBits     = 2  // 雪花 ID 序列默认比特位
)

// SnowIDDefine 雪花 id 客制化定义
type SnowIDDefine struct {
	Machine int // 机器 id
	Offset  int // 时间偏移量
	Seq     int // 序列号
}

// SnowIDGenerator 雪花 id 生成器
type SnowIDGenerator struct {
	Define    *SnowIDDefine // 雪花 id 配置
	MachineID int           // 机器数
	Offset    int           // 偏移量
	Seq       int           // 序列号
}

// CreateSnowIDGenerator 创建雪花 id 生成器
func CreateSnowIDGenerator(config *SnowIDDefine,
	machineID, offset, seq int) *SnowIDGenerator {
	// 雪花 ID 配置校验
	checkSnowIDCustomDefine(generateDefine(config))
	return &SnowIDGenerator{
		Define:    config,
		MachineID: machineID,
		Offset:    offset,
		Seq:       seq,
	}
}

// 校验雪花 ID (自定义部分) 定义
func checkSnowIDCustomDefine(define *SnowIDDefine) {
	if define == nil {
		panic("SnowIDDefine cannot be nil")
	}
	c := define.Machine + define.Offset + define.Seq
	if c != SnowIDCustomBits {
		panic(fmt.Sprintf("SnowIDDefine illegal: custom bit sum %v is illegal (must be %v) !!!", c, SnowIDCustomBits))
	}
}

// 生成雪花 ID 定义
func generateDefine(define *SnowIDDefine) *SnowIDDefine {
	return &SnowIDDefine{
		Machine: base.SetOrDefault[int](define.Machine, SnowIDDefaultMachineBits),
		Offset:  base.SetOrDefault[int](define.Offset, SnowIDDefaultOffsetBits),
		Seq:     base.SetOrDefault[int](define.Seq, SnowIDDefaultSeqBits),
	}
}

// Generate 生成雪花 id
func (generator *SnowIDGenerator) Generate() int64 {
	// 符号位
	sign := int64(SnowIDSign)
	sis, _ := base.NumberToBinaryString(sign, SnowIDSignBits)
	// 时间戳
	timestamp := time.Now().UnixMicro()
	bs, _ := base.NumberToBinaryString(timestamp, SnowIDTimestampBits)
	// 机器 id
	machineID := int64(generator.MachineID)
	ms, _ := base.NumberToBinaryString(machineID, generator.Define.Machine)
	// 偏移量
	offset := int64(generator.Offset)
	os, _ := base.NumberToBinaryString(offset, generator.Define.Offset)
	// 序列号
	seq := int64(generator.Seq)
	ss, _ := base.NumberToBinaryString(seq, generator.Define.Seq)
	// 构建雪花 payload
	payload := sis + bs + ms + os + ss
	result, _ := base.BinaryStringToInt64(payload, SnowIDBits)
	return result
}
