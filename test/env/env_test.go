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
	"strings"
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
	sum, kind, sc, ic, uc, dc := int64(0), int64(0), int64(0), int64(0), int64(0), int64(0)
	// 统计模版

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
		//fmt.Println(kv.Value, kv.Key)

		// todo 校验正确性
		if strings.Contains(kv.Key, "http") {
			fmt.Println(kv.Value, kv.Key)
		}

		sum += kv.Value
		switch strings.ToLower(kv.Key[:6]) {
		case "select":
			sc++
		case "insert":
			ic++
		case "update":
			uc++
		case "delete":
			dc++
		}
		kind++
	}
	fmt.Printf("sum: %v kind: %v operations: [select %v, insert %v, update %v, delete %v]\n", sum, kind, sc, ic, uc, dc)
}

// handleBatchSQLBlock 批量处理 sql 块
func handleBatchSQLBlock(lines []string, f *flow.DataFlow[string]) {
	// 拷贝切片
	l := make([]string, 0)
	for _, item := range lines {
		l = append(l, item)
	}
	// SQL 替换处理链
	SQLReplacerChain := []SQLTokenReplacer{
		&SQLCharacterReplacer{},
		&SQLVariableTokenReplacer{},
		//&SQLValuesSuffixReplacer{},
		//&SQLInSuffixReplacer{},
		//&SQLLimitSuffixReplacer{},
	}
	// 新建生产者
	f.Produce(func(dc chan<- string, ec chan<- error, args ...any) {
		//i := 0
		for _, line := range l {
			//if i >= 10 {
			//	break
			//}
			//i++
			// todo 处理 sql
			operation := line[:6]
			p := line
			//for _, replacer := range SQLReplacerChain {
			//	if replacer.ShouldReplace(operation) {
			//		p = replacer.Replace(p)
			//	}
			//}
			chainLen := len(SQLReplacerChain)
			for i := 0; i < chainLen; i++ {
				if SQLReplacerChain[i].ShouldReplace(operation) {
					p = SQLReplacerChain[i].Replace(p)
				}
			}
			//p := obfuscateSQL(line)
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
	//re := regexp.MustCompile(`'[^']*'|\b\d+(\.\d+)?\b|\bNULL\b`)
	// 将匹配的部分替换为 "?"
	return sql
	//return replaceValuesClause(replaceInClause(replaceToken(replacePatterns(re.ReplaceAllString(replaceNonASCII(strings.ReplaceAll(sql, "?", "$")), "?")))))
}

// replacePatterns 替换字符串中的模式 ?* 或 *? 为 ?
func replacePatterns(input string) string {
	re := regexp.MustCompile(`[^()\s]*\?[^()\s]*`)
	return re.ReplaceAllString(input, "?")
}

func replaceInClause(input string) string {
	// 定义正则表达式，匹配 in (*) 模式
	re := regexp.MustCompile(`in \([^)]*\)`)
	// 替换匹配到的部分为 in (?)
	result := re.ReplaceAllString(input, "in (?...)")
	return result
}

func replaceNonASCII(input string) string {
	// 定义正则表达式，匹配所有非 ASCII 字符
	re := regexp.MustCompile(`[^\x00-\x7F]`)
	// 替换匹配到的部分为 $non-ascii
	result := re.ReplaceAllString(input, "$nonAscii")
	return result
}

func replaceValuesClause(input string) string {
	//// 定义正则表达式，匹配 in (*) 模式
	//re := regexp.MustCompile(`values \([^)]*\)`)
	//// 替换匹配到的部分为 in (?)
	//result := re.ReplaceAllString(input, "values (?...)")
	//return result

	// 定义正则表达式，匹配 values 后面任意字符
	re := regexp.MustCompile(`(?i)values\s+\([^)]*\)`)
	// 替换匹配到的部分为 values (?...)
	result := re.ReplaceAllString(input, "values (?...)")
	return result
}

func replaceToken(input string) string {
	//// 定义正则表达式，匹配 ? 任意字符 ?
	//re := regexp.MustCompile(`\?.*?\?`)
	//// 将匹配到的部分替换为单个问号 ?
	//result := re.ReplaceAllString(input, "?")
	//return result

	//// 定义正则表达式，匹配 ? 与 ? 之间的任意字符，包括空格
	//re := regexp.MustCompile(`\?[^\?]*\?`)
	//// 将匹配到的部分替换为单个问号 ?
	//result := re.ReplaceAllString(input, "?")
	//return result

	//// 定义正则表达式，匹配问号对之间的任意内容（包括特殊字符、空格等）
	//re := regexp.MustCompile(`\?\s*[^?]*?\?`)
	//// 将匹配到的问号内容替换为单个问号
	//result := re.ReplaceAllString(input, "?")
	//return result

	//// 使用正则表达式匹配空格之间存在一个以上问号的模式
	//re := regexp.MustCompile(`\s\?+\s`)
	//// 替换为 " ? "
	//return re.ReplaceAllString(input, " ? ")

	//// 使用正则表达式匹配空格之间存在一个以上问号的模式
	//re := regexp.MustCompile(`\s\?+\s+`)
	//// 替换为 " ? "
	//return re.ReplaceAllString(input, " ? ")

	// 定义正则表达式，匹配空格之间存在一个或多个?的任意字符
	//re := regexp.MustCompile(`\s+[^?\s]*\?\S*\s+`)
	//// 使用替换函数，将匹配到的内容替换为" ? "
	//result := re.ReplaceAllString(input, " ? ")
	//return result

	//re := regexp.MustCompile(`\s\?+(\s|\w)*\?+\s`)
	//result := re.ReplaceAllString(input, " ? ")
	//return result

	return input
}

// SQL token 替换器
type SQLTokenReplacer interface {
	ShouldReplace(t string) bool
	Replace(token string) string
}

// SQLCharacterReplacer SQL 字符替换器
type SQLCharacterReplacer struct{}

func (r *SQLCharacterReplacer) ShouldReplace(t string) bool {
	return true
}

func (r *SQLCharacterReplacer) Replace(token string) string {
	// 剔除中文字符 & emoji
	re1 := regexp.MustCompile("[^\\p{L}\\p{N}a-zA-Z0-9_=\\s-':,`\\[\\](){}?/.*]+") // [^\p{L}\p{N}a-zA-Z0-9_=\s-':,`\[\]\(\)\{\}]+
	res := re1.ReplaceAllString(token, "$")
	// 剔除 emoji
	//re := regexp.MustCompile("[\u4e00-\u9fa5]+")
	//res = re.ReplaceAllString(res, "$")
	// 剔除艺术字体
	//re2 := regexp.MustCompile("[\u1B00-\u1B7F\u2E00-\u2E7F\uA620-\uA62F]+")
	//res = re2.ReplaceAllString(res, "$")
	// 剔除 ascii 不兼容的 unicode 字符
	re2 := regexp.MustCompile("[\u0080-\uffff]?")
	res = re2.ReplaceAllString(res, "")
	// 剔除错误 json 数组存储格式
	re3 := regexp.MustCompile(`/'([A-Za-z0-9-_]*)/'`)
	res = re3.ReplaceAllString(res, "\"$1\"")
	// 剔除非法转义 (/')
	re4 := regexp.MustCompile(`'([A-Za-z0-9$/\s_-]*)('+)([A-Za-z0-9$/\s_-]*)'`)
	res = re4.ReplaceAllString(res, "'$1$3'")
	return res
}

// SQLVariableTokenReplacer SQL 变量替换器
type SQLVariableTokenReplacer struct{}

func (r *SQLVariableTokenReplacer) ShouldReplace(_ string) bool {
	return true
}

func (r *SQLVariableTokenReplacer) Replace(token string) string {
	// 去除 emoji
	//re0 := regexp.MustCompile(`#*0-9]\x{FE0F}?\x{20E3}|©\x{FE0F}?|[®\x{203C}\x{2049}\x{2122}\x{2139}\x{2194}-\x{2199}\x{21A9}\x{21AA}]\x{FE0F}?|[\x{231A}\x{231B}]|[\x{2328}\x{23CF}]\x{FE0F}?|[\x{23E9}-\x{23EC}]|[\x{23ED}-\x{23EF}]\x{FE0F}?|\x{23F0}|[\x{23F1}\x{23F2}]\x{FE0F}?|\x{23F3}|[\x{23F8}-\x{23FA}\x{24C2}\x{25AA}\x{25AB}\x{25B6}\x{25C0}\x{25FB}\x{25FC}]\x{FE0F}?|[\x{25FD}\x{25FE}]|[\x{2600}-\x{2604}\x{260E}\x{2611}]\x{FE0F}?|[\x{2614}\x{2615}]|\x{2618}\x{FE0F}?|\x{261D}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|[\x{2620}\x{2622}\x{2623}\x{2626}\x{262A}\x{262E}\x{262F}\x{2638}-\x{263A}\x{2640}\x{2642}]\x{FE0F}?|[\x{2648}-\x{2653}]|[\x{265F}\x{2660}\x{2663}\x{2665}\x{2666}\x{2668}\x{267B}\x{267E}]\x{FE0F}?|\x{267F}|\x{2692}\x{FE0F}?|\x{2693}|[\x{2694}-\x{2697}\x{2699}\x{269B}\x{269C}\x{26A0}]\x{FE0F}?|\x{26A1}|\x{26A7}\x{FE0F}?|[\x{26AA}\x{26AB}]|[\x{26B0}\x{26B1}]\x{FE0F}?|[\x{26BD}\x{26BE}\x{26C4}\x{26C5}]|\x{26C8}\x{FE0F}?|\x{26CE}|[\x{26CF}\x{26D1}\x{26D3}]\x{FE0F}?|\x{26D4}|\x{26E9}\x{FE0F}?|\x{26EA}|[\x{26F0}\x{26F1}]\x{FE0F}?|[\x{26F2}\x{26F3}]|\x{26F4}\x{FE0F}?|\x{26F5}|[\x{26F7}\x{26F8}]\x{FE0F}?|\x{26F9}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{26FA}\x{26FD}]|\x{2702}\x{FE0F}?|\x{2705}|[\x{2708}\x{2709}]\x{FE0F}?|[\x{270A}\x{270B}][\x{1F3FB}-\x{1F3FF}]?|[\x{270C}\x{270D}][\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|\x{270F}\x{FE0F}?|[\x{2712}\x{2714}\x{2716}\x{271D}\x{2721}]\x{FE0F}?|\x{2728}|[\x{2733}\x{2734}\x{2744}\x{2747}]\x{FE0F}?|[\x{274C}\x{274E}\x{2753}-\x{2755}\x{2757}]|\x{2763}\x{FE0F}?|\x{2764}(?:\x{200D}[\x{1F525}\x{1FA79}]|\x{FE0F}(?:\x{200D}[\x{1F525}\x{1FA79}])?)?|[\x{2795}-\x{2797}]|\x{27A1}\x{FE0F}?|[\x{27B0}\x{27BF}]|[\x{2934}\x{2935}\x{2B05}-\x{2B07}]\x{FE0F}?|[\x{2B1B}\x{2B1C}\x{2B50}\x{2B55}]|[\x{3030}\x{303D}\x{3297}\x{3299}]\x{FE0F}?|[\x{1F004}\x{1F0CF}]|[\x{1F170}\x{1F171}\x{1F17E}\x{1F17F}]\x{FE0F}?|[\x{1F18E}\x{1F191}-\x{1F19A}]|\x{1F1E6}[\x{1F1E8}-\x{1F1EC}\x{1F1EE}\x{1F1F1}\x{1F1F2}\x{1F1F4}\x{1F1F6}-\x{1F1FA}\x{1F1FC}\x{1F1FD}\x{1F1FF}]|\x{1F1E7}[\x{1F1E6}\x{1F1E7}\x{1F1E9}-\x{1F1EF}\x{1F1F1}-\x{1F1F4}\x{1F1F6}-\x{1F1F9}\x{1F1FB}\x{1F1FC}\x{1F1FE}\x{1F1FF}]|\x{1F1E8}[\x{1F1E6}\x{1F1E8}\x{1F1E9}\x{1F1EB}-\x{1F1EE}\x{1F1F0}-\x{1F1F5}\x{1F1F7}\x{1F1FA}-\x{1F1FF}]|\x{1F1E9}[\x{1F1EA}\x{1F1EC}\x{1F1EF}\x{1F1F0}\x{1F1F2}\x{1F1F4}\x{1F1FF}]|\x{1F1EA}[\x{1F1E6}\x{1F1E8}\x{1F1EA}\x{1F1EC}\x{1F1ED}\x{1F1F7}-\x{1F1FA}]|\x{1F1EB}[\x{1F1EE}-\x{1F1F0}\x{1F1F2}\x{1F1F4}\x{1F1F7}]|\x{1F1EC}[\x{1F1E6}\x{1F1E7}\x{1F1E9}-\x{1F1EE}\x{1F1F1}-\x{1F1F3}\x{1F1F5}-\x{1F1FA}\x{1F1FC}\x{1F1FE}]|\x{1F1ED}[\x{1F1F0}\x{1F1F2}\x{1F1F3}\x{1F1F7}\x{1F1F9}\x{1F1FA}]|\x{1F1EE}[\x{1F1E8}-\x{1F1EA}\x{1F1F1}-\x{1F1F4}\x{1F1F6}-\x{1F1F9}]|\x{1F1EF}[\x{1F1EA}\x{1F1F2}\x{1F1F4}\x{1F1F5}]|\x{1F1F0}[\x{1F1EA}\x{1F1EC}-\x{1F1EE}\x{1F1F2}\x{1F1F3}\x{1F1F5}\x{1F1F7}\x{1F1FC}\x{1F1FE}\x{1F1FF}]|\x{1F1F1}[\x{1F1E6}-\x{1F1E8}\x{1F1EE}\x{1F1F0}\x{1F1F7}-\x{1F1FB}\x{1F1FE}]|\x{1F1F2}[\x{1F1E6}\x{1F1E8}-\x{1F1ED}\x{1F1F0}-\x{1F1FF}]|\x{1F1F3}[\x{1F1E6}\x{1F1E8}\x{1F1EA}-\x{1F1EC}\x{1F1EE}\x{1F1F1}\x{1F1F4}\x{1F1F5}\x{1F1F7}\x{1F1FA}\x{1F1FF}]|\x{1F1F4}\x{1F1F2}|\x{1F1F5}[\x{1F1E6}\x{1F1EA}-\x{1F1ED}\x{1F1F0}-\x{1F1F3}\x{1F1F7}-\x{1F1F9}\x{1F1FC}\x{1F1FE}]|\x{1F1F6}\x{1F1E6}|\x{1F1F7}[\x{1F1EA}\x{1F1F4}\x{1F1F8}\x{1F1FA}\x{1F1FC}]|\x{1F1F8}[\x{1F1E6}-\x{1F1EA}\x{1F1EC}-\x{1F1F4}\x{1F1F7}-\x{1F1F9}\x{1F1FB}\x{1F1FD}-\x{1F1FF}]|\x{1F1F9}[\x{1F1E6}\x{1F1E8}\x{1F1E9}\x{1F1EB}-\x{1F1ED}\x{1F1EF}-\x{1F1F4}\x{1F1F7}\x{1F1F9}\x{1F1FB}\x{1F1FC}\x{1F1FF}]|\x{1F1FA}[\x{1F1E6}\x{1F1EC}\x{1F1F2}\x{1F1F3}\x{1F1F8}\x{1F1FE}\x{1F1FF}]|\x{1F1FB}[\x{1F1E6}\x{1F1E8}\x{1F1EA}\x{1F1EC}\x{1F1EE}\x{1F1F3}\x{1F1FA}]|\x{1F1FC}[\x{1F1EB}\x{1F1F8}]|\x{1F1FD}\x{1F1F0}|\x{1F1FE}[\x{1F1EA}\x{1F1F9}]|\x{1F1FF}[\x{1F1E6}\x{1F1F2}\x{1F1FC}]|\x{1F201}|\x{1F202}\x{FE0F}?|[\x{1F21A}\x{1F22F}\x{1F232}-\x{1F236}]|\x{1F237}\x{FE0F}?|[\x{1F238}-\x{1F23A}\x{1F250}\x{1F251}\x{1F300}-\x{1F320}]|[\x{1F321}\x{1F324}-\x{1F32C}]\x{FE0F}?|[\x{1F32D}-\x{1F335}]|\x{1F336}\x{FE0F}?|[\x{1F337}-\x{1F37C}]|\x{1F37D}\x{FE0F}?|[\x{1F37E}-\x{1F384}]|\x{1F385}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F386}-\x{1F393}]|[\x{1F396}\x{1F397}\x{1F399}-\x{1F39B}\x{1F39E}\x{1F39F}]\x{FE0F}?|[\x{1F3A0}-\x{1F3C1}]|\x{1F3C2}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F3C3}\x{1F3C4}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3C5}\x{1F3C6}]|\x{1F3C7}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F3C8}\x{1F3C9}]|\x{1F3CA}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3CB}\x{1F3CC}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F3CD}\x{1F3CE}]\x{FE0F}?|[\x{1F3CF}-\x{1F3D3}]|[\x{1F3D4}-\x{1F3DF}]\x{FE0F}?|[\x{1F3E0}-\x{1F3F0}]|\x{1F3F3}(?:\x{200D}(?:\x{26A7}\x{FE0F}?|\x{1F308})|\x{FE0F}(?:\x{200D}(?:\x{26A7}\x{FE0F}?|\x{1F308}))?)?|\x{1F3F4}(?:\x{200D}\x{2620}\x{FE0F}?|\x{E0067}\x{E0062}(?:\x{E0065}\x{E006E}\x{E0067}|\x{E0073}\x{E0063}\x{E0074}|\x{E0077}\x{E006C}\x{E0073})\x{E007F})?|[\x{1F3F5}\x{1F3F7}]\x{FE0F}?|[\x{1F3F8}-\x{1F407}]|\x{1F408}(?:\x{200D}\x{2B1B})?|[\x{1F409}-\x{1F414}]|\x{1F415}(?:\x{200D}\x{1F9BA})?|[\x{1F416}-\x{1F43A}]|\x{1F43B}(?:\x{200D}\x{2744}\x{FE0F}?)?|[\x{1F43C}-\x{1F43E}]|\x{1F43F}\x{FE0F}?|\x{1F440}|\x{1F441}(?:\x{200D}\x{1F5E8}\x{FE0F}?|\x{FE0F}(?:\x{200D}\x{1F5E8}\x{FE0F}?)?)?|[\x{1F442}\x{1F443}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F444}\x{1F445}]|[\x{1F446}-\x{1F450}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F451}-\x{1F465}]|[\x{1F466}\x{1F467}][\x{1F3FB}-\x{1F3FF}]?|\x{1F468}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}]|\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?|[\x{1F468}\x{1F469}]\x{200D}(?:\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?)|[\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FC}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?\x{1F468}[\x{1F3FB}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F468}[\x{1F3FB}-\x{1F3FE}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|\x{1F469}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D})?[\x{1F468}\x{1F469}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}]|\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?|\x{1F469}\x{200D}(?:\x{1F466}(?:\x{200D}\x{1F466})?|\x{1F467}(?:\x{200D}[\x{1F466}\x{1F467}])?)|[\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FC}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}]|\x{1F48B}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FF}])|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}[\x{1F468}\x{1F469}][\x{1F3FB}-\x{1F3FE}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|\x{1F46A}|[\x{1F46B}-\x{1F46D}][\x{1F3FB}-\x{1F3FF}]?|\x{1F46E}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F46F}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F470}\x{1F471}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F472}[\x{1F3FB}-\x{1F3FF}]?|\x{1F473}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F474}-\x{1F476}][\x{1F3FB}-\x{1F3FF}]?|\x{1F477}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F478}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F479}-\x{1F47B}]|\x{1F47C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F47D}-\x{1F480}]|[\x{1F481}\x{1F482}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F483}[\x{1F3FB}-\x{1F3FF}]?|\x{1F484}|\x{1F485}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F486}\x{1F487}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F488}-\x{1F48E}]|\x{1F48F}[\x{1F3FB}-\x{1F3FF}]?|\x{1F490}|\x{1F491}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F492}-\x{1F4A9}]|\x{1F4AA}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F4AB}-\x{1F4FC}]|\x{1F4FD}\x{FE0F}?|[\x{1F4FF}-\x{1F53D}]|[\x{1F549}\x{1F54A}]\x{FE0F}?|[\x{1F54B}-\x{1F54E}\x{1F550}-\x{1F567}]|[\x{1F56F}\x{1F570}\x{1F573}]\x{FE0F}?|\x{1F574}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|\x{1F575}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{FE0F}\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F576}-\x{1F579}]\x{FE0F}?|\x{1F57A}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F587}\x{1F58A}-\x{1F58D}]\x{FE0F}?|\x{1F590}[\x{FE0F}\x{1F3FB}-\x{1F3FF}]?|[\x{1F595}\x{1F596}][\x{1F3FB}-\x{1F3FF}]?|\x{1F5A4}|[\x{1F5A5}\x{1F5A8}\x{1F5B1}\x{1F5B2}\x{1F5BC}\x{1F5C2}-\x{1F5C4}\x{1F5D1}-\x{1F5D3}\x{1F5DC}-\x{1F5DE}\x{1F5E1}\x{1F5E3}\x{1F5E8}\x{1F5EF}\x{1F5F3}\x{1F5FA}]\x{FE0F}?|[\x{1F5FB}-\x{1F62D}]|\x{1F62E}(?:\x{200D}\x{1F4A8})?|[\x{1F62F}-\x{1F634}]|\x{1F635}(?:\x{200D}\x{1F4AB})?|\x{1F636}(?:\x{200D}\x{1F32B}\x{FE0F}?)?|[\x{1F637}-\x{1F644}]|[\x{1F645}-\x{1F647}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F648}-\x{1F64A}]|\x{1F64B}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F64C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F64D}\x{1F64E}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F64F}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F680}-\x{1F6A2}]|\x{1F6A3}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F6A4}-\x{1F6B3}]|[\x{1F6B4}-\x{1F6B6}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F6B7}-\x{1F6BF}]|\x{1F6C0}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F6C1}-\x{1F6C5}]|\x{1F6CB}\x{FE0F}?|\x{1F6CC}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F6CD}-\x{1F6CF}]\x{FE0F}?|[\x{1F6D0}-\x{1F6D2}\x{1F6D5}-\x{1F6D7}\x{1F6DD}-\x{1F6DF}]|[\x{1F6E0}-\x{1F6E5}\x{1F6E9}]\x{FE0F}?|[\x{1F6EB}\x{1F6EC}]|[\x{1F6F0}\x{1F6F3}]\x{FE0F}?|[\x{1F6F4}-\x{1F6FC}\x{1F7E0}-\x{1F7EB}\x{1F7F0}]|\x{1F90C}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F90D}\x{1F90E}]|\x{1F90F}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F910}-\x{1F917}]|[\x{1F918}-\x{1F91F}][\x{1F3FB}-\x{1F3FF}]?|[\x{1F920}-\x{1F925}]|\x{1F926}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F927}-\x{1F92F}]|[\x{1F930}-\x{1F934}][\x{1F3FB}-\x{1F3FF}]?|\x{1F935}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F936}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F937}-\x{1F939}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F93A}|\x{1F93C}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F93D}\x{1F93E}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F93F}-\x{1F945}\x{1F947}-\x{1F976}]|\x{1F977}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F978}-\x{1F9B4}]|[\x{1F9B5}\x{1F9B6}][\x{1F3FB}-\x{1F3FF}]?|\x{1F9B7}|[\x{1F9B8}\x{1F9B9}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9BA}|\x{1F9BB}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F9BC}-\x{1F9CC}]|[\x{1F9CD}-\x{1F9CF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9D0}|\x{1F9D1}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}])|\x{1F3FB}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FC}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FC}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}\x{1F3FD}-\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FD}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FE}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}-\x{1F3FD}\x{1F3FF}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?|\x{1F3FF}(?:\x{200D}(?:[\x{2695}\x{2696}\x{2708}]\x{FE0F}?|\x{2764}\x{FE0F}?\x{200D}(?:\x{1F48B}\x{200D}|)\x{1F9D1}[\x{1F3FB}-\x{1F3FE}]|[\x{1F33E}\x{1F373}\x{1F37C}\x{1F384}\x{1F393}\x{1F3A4}\x{1F3A8}\x{1F3EB}\x{1F3ED}\x{1F4BB}\x{1F4BC}\x{1F527}\x{1F52C}\x{1F680}\x{1F692}]|\x{1F91D}\x{200D}\x{1F9D1}[\x{1F3FB}-\x{1F3FF}]|[\x{1F9AF}-\x{1F9B3}\x{1F9BC}\x{1F9BD}]))?)?|[\x{1F9D2}\x{1F9D3}][\x{1F3FB}-\x{1F3FF}]?|\x{1F9D4}(?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|\x{1F9D5}[\x{1F3FB}-\x{1F3FF}]?|[\x{1F9D6}-\x{1F9DD}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?|[\x{1F3FB}-\x{1F3FF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?)?|[\x{1F9DE}\x{1F9DF}](?:\x{200D}[\x{2640}\x{2642}]\x{FE0F}?)?|[\x{1F9E0}-\x{1F9FF}\x{1FA70}-\x{1FA74}\x{1FA78}-\x{1FA7C}\x{1FA80}-\x{1FA86}\x{1FA90}-\x{1FAAC}\x{1FAB0}-\x{1FABA}\x{1FAC0}-\x{1FAC2}]|[\x{1FAC3}-\x{1FAC5}][\x{1F3FB}-\x{1F3FF}]?|[\x{1FAD0}-\x{1FAD9}\x{1FAE0}-\x{1FAE7}]|\x{1FAF0}[\x{1F3FB}-\x{1F3FF}]?|\x{1FAF1}(?:\x{1F3FB}(?:\x{200D}\x{1FAF2}[\x{1F3FC}-\x{1F3FF}])?|\x{1F3FC}(?:\x{200D}\x{1FAF2}[\x{1F3FB}\x{1F3FD}-\x{1F3FF}])?|\x{1F3FD}(?:\x{200D}\x{1FAF2}[\x{1F3FB}\x{1F3FC}\x{1F3FE}\x{1F3FF}])?|\x{1F3FE}(?:\x{200D}\x{1FAF2}[\x{1F3FB}-\x{1F3FD}\x{1F3FF}])?|\x{1F3FF}(?:\x{200D}\x{1FAF2}[\x{1F3FB}-\x{1F3FE}])?)?|[\x{1FAF2}-\x{1FAF6}][\x{1F3FB}-\x{1F3FF}]?`)
	//res := re0.ReplaceAllString(token, "$")

	// 去除转义
	//re1 := regexp.MustCompile(`/'(^[\s,])`)
	//res := re1.ReplaceAllString(token, "")
	// 字符串变量匹配
	//re2 := regexp.MustCompile(`'(\S+)(\s*)(\S*)'`)
	//res = re2.ReplaceAllString(token, " ? ")
	// 字符串变量匹配 +
	re1 := regexp.MustCompile(`'([^']*)'`)
	res := re1.ReplaceAllString(token, "?")
	// 非字符串变量值匹配 +
	re2 := regexp.MustCompile(`=((\s)*)([0-9]+)`)
	res = re2.ReplaceAllString(res, "= ?")

	//re3 := regexp.MustCompile(`.=\s[^?][-:\s'0-9]*`)
	//res = re3.ReplaceAllString(res, "= ? ")
	return res
}

// SQLValuesSuffixReplacer SQL values 子句替换器
type SQLValuesSuffixReplacer struct{}

func (r *SQLValuesSuffixReplacer) ShouldReplace(t string) bool {
	return t == "insert"
}

func (r *SQLValuesSuffixReplacer) Replace(token string) string {
	re1 := regexp.MustCompile(`values[^\n]*`)
	res := re1.ReplaceAllString(token, "values (?...)")
	return res
}

// SQLInSuffixReplacer SQL in 子句替换器
type SQLInSuffixReplacer struct{}

func (r *SQLInSuffixReplacer) ShouldReplace(t string) bool {
	return t != "insert"
}

func (r *SQLInSuffixReplacer) Replace(token string) string {
	re1 := regexp.MustCompile(`in[^\n]*`)
	res := re1.ReplaceAllString(token, "in (?...)")
	return res
}

// SQLLimitSuffixReplacer SQL limit 子句替换器
type SQLLimitSuffixReplacer struct {
}

func (r *SQLLimitSuffixReplacer) ShouldReplace(t string) bool {
	return t != "insert"
}

func (r *SQLLimitSuffixReplacer) Replace(token string) string {
	re1 := regexp.MustCompile(`limit\s[0-9]*`)
	res := re1.ReplaceAllString(token, "limit ? ")
	re2 := regexp.MustCompile(`offset\s[0-9]*`)
	res = re2.ReplaceAllString(res, "offset ? ")
	return res
}

// 0. 字符合规: {[^\p{L}\p{N}a-zA-Z0-9_=\s-':,`\[\]\(\)\{\}]+} => []
// 1. 变量替换: + {.'[^\s]*.'} => [ ? ]
//             - {.=\s[^?][-\:\s\'0-9a-zA-Z]*} => [= ?]
//             + {.=\s[^?][-\:\s\'0-9]*} => [= ? ]
// 2. 匹配 values 子句: (insert)
//             {values[^\n]*} => [values (?...)]
// 3. 匹配 in 子句: (not insert)
//             {in[^\n]*} => [in (?...)]
// 4. 匹配 limit 子句: (not insert)
//             {limit\s[0-9]*} => [limit ? ]
//             {offset\s[0-9]*} => [offset ? ]

// RemoveChineseAndEmoji 移除字符串中的所有中文字符和 Emoji
func RemoveChineseAndEmoji(input string) string {
	// 匹配中文字符的范围和 Emoji 的范围
	re := regexp.MustCompile("[\u4e00-\u9fa5]|[\U0001F600-\U0001F64F]|[\U0001F300-\U0001F5FF]|[\U0001F680-\U0001F6FF]|[\U0001F700-\U0001F77F]|[\U0001F800-\U0001F8FF]|[\U0001F900-\U0001F9FF]|[\U0001FA00-\U0001FAFF]|[\u2600-\u26FF]|[\u2700-\u27BF]")
	return re.ReplaceAllString(input, "")
}
