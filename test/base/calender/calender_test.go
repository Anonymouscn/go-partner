package calender

import (
	"fmt"
	calender2 "github.com/Anonymouscn/go-partner/base/calender"
	"testing"
)

// TestCalender 日历单元测试
func TestCalender(t *testing.T) {
	calender := calender2.NewCalender()
	// ============================================= 按天计算测试 ============================================= //
	today := calender.Today()
	fmt.Printf("today: [%v - %v]\n", today.StartTime, today.EndTime)
	yesterday := calender.Yesterday()
	fmt.Printf("yesterday: [%v - %v]\n", yesterday.StartTime, yesterday.EndTime)
	tomorrow := calender.Tomorrow()
	fmt.Printf("tomorrow: [%v - %v]\n", tomorrow.StartTime, tomorrow.EndTime)
	twoDaysAgo := calender.DaysAgo(2)
	fmt.Printf("2 days ago: [%v - %v]\n", twoDaysAgo.StartTime, twoDaysAgo.EndTime)
	threeDaysAgo := calender.DaysAgo(3)
	fmt.Printf("3 days ago: [%v - %v]\n", threeDaysAgo.StartTime, threeDaysAgo.EndTime)
	twoDaysAfter := calender.DaysAfter(2)
	fmt.Printf("2 days after: [%v - %v]\n", twoDaysAfter.StartTime, twoDaysAfter.EndTime)
	threeDaysAfter := calender.DaysAfter(3)
	fmt.Printf("3 days after: [%v - %v]\n", threeDaysAfter.StartTime, threeDaysAfter.EndTime)
	// ============================================= 按周计算测试 ============================================= //
	thisWeek := calender.ThisWeek()
	fmt.Printf("this week: [%v - %v]\n", thisWeek.StartTime, thisWeek.EndTime)
	lastWeek := calender.LastWeek()
	fmt.Printf("last week: [%v - %v]\n", lastWeek.StartTime, lastWeek.EndTime)
	nextWeek := calender.NextWeek()
	fmt.Printf("next week: [%v - %v]\n", nextWeek.StartTime, nextWeek.EndTime)
	twoWeeksAgo := calender.WeeksAgo(2)
	fmt.Printf("2 weeks ago: [%v - %v]\n", twoWeeksAgo.StartTime, twoWeeksAgo.EndTime)
	threeWeeksAgo := calender.WeeksAgo(3)
	fmt.Printf("3 weeks ago: [%v - %v]\n", threeWeeksAgo.StartTime, threeWeeksAgo.EndTime)
	twoWeeksAfter := calender.WeeksAfter(2)
	fmt.Printf("2 weeks after: [%v - %v]\n", twoWeeksAfter.StartTime, twoWeeksAfter.EndTime)
	threeWeeksAfter := calender.WeeksAfter(3)
	fmt.Printf("3 weeks after: [%v - %v]\n", threeWeeksAfter.StartTime, threeWeeksAfter.EndTime)
	// ============================================= 按月计算测试 ============================================= //
	thisMonth := calender.ThisMonth()
	fmt.Printf("this month: [%v - %v]\n", thisMonth.StartTime, thisMonth.EndTime)
	lastMonth := calender.LastMonth()
	fmt.Printf("last month: [%v - %v]\n", lastMonth.StartTime, lastMonth.EndTime)
	nextMonth := calender.NextMonth()
	fmt.Printf("next month: [%v - %v]\n", nextMonth.StartTime, nextMonth.EndTime)
	twoMonthsAgo := calender.MonthsAgo(2)
	fmt.Printf("2 months ago: [%v - %v]\n", twoMonthsAgo.StartTime, twoMonthsAgo.EndTime)
	threeMonthsAgo := calender.MonthsAgo(3)
	fmt.Printf("3 months ago: [%v - %v]\n", threeMonthsAgo.StartTime, threeMonthsAgo.EndTime)
	twoMonthsAfter := calender.MonthsAfter(2)
	fmt.Printf("2 months after: [%v - %v]\n", twoMonthsAfter.StartTime, twoMonthsAfter.EndTime)
	threeMonthsAfter := calender.MonthsAfter(3)
	fmt.Printf("3 months after: [%v - %v]\n", threeMonthsAfter.StartTime, threeMonthsAfter.EndTime)
	// ============================================= 按年计算测试 ============================================= //
	thisYear := calender.ThisYear()
	fmt.Printf("this year: [%v - %v]\n", thisYear.StartTime, thisYear.EndTime)
	lastYear := calender.LastYear()
	fmt.Printf("last year: [%v - %v]\n", lastYear.StartTime, lastYear.EndTime)
	nextYear := calender.NextYear()
	fmt.Printf("next year: [%v - %v]\n", nextYear.StartTime, nextYear.EndTime)
	twoYearsAgo := calender.YearsAgo(2)
	fmt.Printf("2 years ago: [%v - %v]\n", twoYearsAgo.StartTime, twoYearsAgo.EndTime)
	threeYearsAgo := calender.YearsAgo(3)
	fmt.Printf("3 years ago: [%v - %v]\n", threeYearsAgo.StartTime, threeYearsAgo.EndTime)
	twoYearsAfter := calender.YearsAfter(2)
	fmt.Printf("2 years after: [%v - %v]\n", twoYearsAfter.StartTime, twoYearsAfter.EndTime)
	threeYearsAfter := calender.YearsAfter(3)
	fmt.Printf("3 years after: [%v - %v]\n", threeYearsAfter.StartTime, threeYearsAfter.EndTime)
}
