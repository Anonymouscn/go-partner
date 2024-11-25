package env

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/Anonymouscn/go-partner/flow"
	"log"
	"os"
	"regexp"
	"sort"
	"testing"
)

const batchSize = 2000 // 每批处理的行数

const max = 2000000000000

func TestReadSQLReport(t *testing.T) {
	// 录入 csv
	csvFile, err := os.Open("/Users/anonymous/Downloads/1252524126_cdb-4k5ej1u2_1732505152_15392844.csv")
	if err != nil {
		log.Fatalln("Unable to read input file:", err)
	}
	defer csvFile.Close()
	// 创建一个CSV读取器
	reader := csv.NewReader(bufio.NewReader(csvFile))
	// 读取CSV文件的标题行
	headers, err := reader.Read()
	if err != nil {
		log.Fatalln("Unable to read headers:", err)
	}
	fmt.Println("Headers:", headers)

	summary := make(map[string]int64)

	// 新建数据流
	f := flow.NewDataFlow[string](batchSize * 2)

	// 注入消费者
	f.Consume(func(dc <-chan string, ec chan<- error, args ...any) {
		for v := range dc {
			summary[v]++
		}
	})

	seq := 1                   // 批次序号
	lines := make([]string, 0) // SQL 缓存块

	// 逐行读取 CSV 文件
	for i := 0; i < max; i++ {
		record, err := reader.Read()
		if err != nil {
			break // 当读取到文件末尾时，跳出循环
		}
		lines = append(lines, record[6])
		if len(lines) == batchSize {
			// 批量处理 SQL 块
			handleBatchSQLBlock(lines, f)
			lines = make([]string, 0) // 清空当前批次的行
			fmt.Printf("process in seq[%v]\n", seq)
			seq++
		}
	}
	// 处理剩余缓存块
	if len(lines) != 0 {
		// 批量处理剩余 SQL 块
		handleBatchSQLBlock(lines, f)
		lines = make([]string, 0) // 清空当前批次的行
		fmt.Printf("process in seq[%v]\n", seq)
		seq++
	}

	// 等待处理结果
	f.Stop()

	// 统计计数
	sum := int64(0)

	// ========== 排序 ========== //
	type kv struct {
		Key   string
		Value int64
	}
	var sortedSummary []kv
	for k, v := range summary {
		sortedSummary = append(sortedSummary, kv{k, v})
	}
	// 按照值进行排序
	sort.Slice(sortedSummary, func(i, j int) bool {
		return sortedSummary[i].Value > sortedSummary[j].Value
	})

	for _, kv := range sortedSummary {
		fmt.Println(kv.Value, kv.Key)
		sum += kv.Value
	}
	fmt.Println("sum: ", sum)
}

// handleBatchSQLBlock 批量处理 sql 块
func handleBatchSQLBlock(lines []string, f *flow.DataFlow[string]) {
	// 拷贝切片
	l := make([]string, 0)
	for _, item := range lines {
		l = append(l, item)
	}
	// 新建生产者
	f.Produce(func(dc chan<- string, ec chan<- error, args ...any) {
		for _, line := range l {
			p := obfuscateSQL(line)
			dc <- p
		}
	})
}

// SQL 混淆
func obfuscateSQL(sql string) string {
	// 正则表达式匹配 SQL 变量值
	// 1. 单引号括起来的字符串: '.*?'
	// 2. 数字（包括浮点数和整数）: \b\d+(\.\d+)?\b
	// 3. NULL 值: \bNULL\b
	// 注意：正则表达式需要处理 SQL 特殊字符和转义字符
	re := regexp.MustCompile(`'[^']*'|\b\d+(\.\d+)?\b|\bNULL\b`)
	// 将匹配的部分替换为 "?"
	return replacePatterns(re.ReplaceAllString(sql, "?"))
}

// replacePatterns 替换字符串中的模式 ?* 或 *? 为 ?
func replacePatterns(input string) string {
	re := regexp.MustCompile(`[^()\s]*\?[^()\s]*`)
	return re.ReplaceAllString(input, "?")
}
