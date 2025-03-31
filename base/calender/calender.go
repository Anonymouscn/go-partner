package calender

import (
	"time"
)

// Period 时间段
type Period struct {
	StartTime time.Time
	EndTime   time.Time
}

// Calender 日历
type Calender struct {
	// ======== 初始生成字段 ======== //
	year  int
	month time.Month
	day   int
	now   time.Time
	// ======== 惰性计算字段 ======== //
	todayPeriod     *Period
	thisWeekPeriod  *Period
	thisMonthPeriod *Period
	thisYearPeriod  *Period
}

// NewCalender 新建日历
func NewCalender() *Calender {
	c := &Calender{}
	return c.Flush()
}

// Flush 刷新当前时间
func (c *Calender) Flush() *Calender {
	c.now = time.Now()
	c.year, c.month, c.day = c.now.Date()
	return c
}

// Now 获取当前时间
func (c *Calender) Now() time.Time {
	return c.now
}

// 内部计算今天日期
func (c *Calender) calculateToday() {
	todayStart := time.Date(c.year, c.month, c.day, 0, 0, 0, 0, time.Local)
	todayEnd := time.Date(c.year, c.month, c.day, 23, 59, 59, 999999999, time.Local)
	c.todayPeriod = &Period{
		StartTime: todayStart,
		EndTime:   todayEnd,
	}
}

// Today 获取今天时间段 [结算: 天]
func (c *Calender) Today() *Period {
	c.calculateToday()
	return c.todayPeriod
}

// Yesterday 获取昨天时间段 [结算: 天]
func (c *Calender) Yesterday() *Period {
	return c.DaysAgo(1)
}

// Tomorrow 获取明天时间段 [结算: 天]
func (c *Calender) Tomorrow() *Period {
	return c.DaysAfter(1)
}

// DaysAgo 几天前时间段 [结算: 天]
func (c *Calender) DaysAgo(n int) *Period {
	c.calculateToday()
	todayPeriod := c.todayPeriod
	dayStart := todayPeriod.StartTime.AddDate(0, 0, -n)
	dayEnd := todayPeriod.EndTime.AddDate(0, 0, -n)
	return &Period{
		StartTime: dayStart,
		EndTime:   dayEnd,
	}
}

// DaysAfter 几天后时间段 [结算: 天]
func (c *Calender) DaysAfter(n int) *Period {
	c.calculateToday()
	todayPeriod := c.todayPeriod
	dayStart := todayPeriod.StartTime.AddDate(0, 0, n)
	dayEnd := todayPeriod.EndTime.AddDate(0, 0, n)
	return &Period{
		StartTime: dayStart,
		EndTime:   dayEnd,
	}
}

// 内部计算本周日期
func (c *Calender) calculateThisWeek() {
	c.calculateToday()
	todayPeriod := c.todayPeriod
	todayStart := todayPeriod.StartTime
	// Get the current weekday (0 = Sunday, 1 = Monday, ..., 6 = Saturday)
	weekday := todayStart.Weekday()
	// Calculate the start of the week (assuming weeks start on Monday)
	daysFromMonday := int(weekday)
	if daysFromMonday == 0 { // If today is Sunday, go back 6 days to get to Monday
		daysFromMonday = 6
	} else {
		daysFromMonday--
	}
	thisWeekStart := todayStart.AddDate(0, 0, -daysFromMonday)
	thisWeekEnd := thisWeekStart.AddDate(0, 0, 6).
		Add(23*time.Hour + 59*time.Minute + 59*time.Second + 999999999*time.Nanosecond)
	c.thisWeekPeriod = &Period{
		StartTime: thisWeekStart,
		EndTime:   thisWeekEnd,
	}
}

// ThisWeek 本周时间段 [结算: 周]
func (c *Calender) ThisWeek() *Period {
	c.calculateThisWeek()
	return c.thisWeekPeriod
}

// LastWeek 上周时间段 [结算: 周]
func (c *Calender) LastWeek() *Period {
	return c.WeeksAgo(1)
}

// NextWeek 下周时间段 [结算: 周]
func (c *Calender) NextWeek() *Period {
	return c.WeeksAfter(1)
}

// WeeksAgo 几周前时间段 [结算: 周]
func (c *Calender) WeeksAgo(n int) *Period {
	c.calculateThisWeek()
	thisWeekPeriod := c.thisWeekPeriod
	lastWeekStart := thisWeekPeriod.StartTime.AddDate(0, 0, -7*n)
	lastWeekEnd := thisWeekPeriod.StartTime.AddDate(0, 0, -7*(n-1)).Add(-1)
	return &Period{
		StartTime: lastWeekStart,
		EndTime:   lastWeekEnd,
	}
}

// WeeksAfter 几周后时间段 [结算: 周]
func (c *Calender) WeeksAfter(n int) *Period {
	c.calculateThisWeek()
	thisWeekPeriod := c.thisWeekPeriod
	nextWeekStart := thisWeekPeriod.StartTime.AddDate(0, 0, 7*n)
	nextWeekEnd := thisWeekPeriod.StartTime.AddDate(0, 0, 7*(n+1)).Add(-1)
	return &Period{
		StartTime: nextWeekStart,
		EndTime:   nextWeekEnd,
	}
}

// 内部计算本月日期
func (c *Calender) calculateThisMonth() {
	thisMonthStart := time.Date(c.year, c.month, 1, 0, 0, 0, 0, time.Local)
	// Calculate the last day of this month
	nextMonth := c.month + 1
	nextMonthYear := c.year
	if nextMonth > 12 {
		nextMonth = 1
		nextMonthYear++
	}
	thisMonthEnd := time.Date(nextMonthYear, nextMonth, 1, 0, 0, 0, 0, time.Local).
		Add(-1)
	c.thisMonthPeriod = &Period{
		StartTime: thisMonthStart,
		EndTime:   thisMonthEnd,
	}
}

// ThisMonth 这个月时间段 [结算: 月]
func (c *Calender) ThisMonth() *Period {
	c.calculateThisMonth()
	return c.thisMonthPeriod
}

// LastMonth 上个月时间段 [结算: 月]
func (c *Calender) LastMonth() *Period {
	return c.MonthsAgo(1)
}

// NextMonth 下个月时间段 [结算: 月]
func (c *Calender) NextMonth() *Period {
	return c.MonthsAfter(1)
}

// MonthsAgo 几个月前时间段 [结算: 月]
func (c *Calender) MonthsAgo(n int) *Period {
	c.calculateThisMonth()
	thisMonthPeriod := c.thisMonthPeriod
	monthStart := thisMonthPeriod.StartTime.AddDate(0, -n, 0)
	monthEnd := thisMonthPeriod.StartTime.AddDate(0, -(n - 1), 0).Add(-1)
	return &Period{
		StartTime: monthStart,
		EndTime:   monthEnd,
	}
}

// MonthsAfter 几个月后时间段 [结算: 月]
func (c *Calender) MonthsAfter(n int) *Period {
	c.calculateThisMonth()
	thisMonthPeriod := c.thisMonthPeriod
	monthStart := thisMonthPeriod.StartTime.AddDate(0, n, 0)
	monthEnd := thisMonthPeriod.StartTime.AddDate(0, n+1, 0).Add(-1)
	return &Period{
		StartTime: monthStart,
		EndTime:   monthEnd,
	}
}

// 内部计算本月日期
func (c *Calender) calculateThisYear() {
	thisYearStart := time.Date(c.year, 1, 1, 0, 0, 0, 0, time.Local)
	thisYearEnd := time.Date(c.year, 12, 31, 23, 59, 59, 999999999, time.Local)
	c.thisYearPeriod = &Period{
		StartTime: thisYearStart,
		EndTime:   thisYearEnd,
	}
}

// ThisYear 今年时间段 [结算: 年]
func (c *Calender) ThisYear() *Period {
	c.calculateThisYear()
	return c.thisYearPeriod
}

// LastYear 去年时间段 [结算: 年]
func (c *Calender) LastYear() *Period {
	return c.YearsAgo(1)
}

// NextYear 明年时间段 [结算: 年]
func (c *Calender) NextYear() *Period {
	return c.YearsAfter(1)
}

// YearsAgo 几年前时间段 [结算: 年]
func (c *Calender) YearsAgo(n int) *Period {
	c.calculateThisYear()
	thisYearPeriod := c.thisYearPeriod
	yearStart := thisYearPeriod.StartTime.AddDate(-n, 0, 0)
	yearEnd := thisYearPeriod.StartTime.AddDate(-(n - 1), 0, 0).Add(-1)
	return &Period{
		StartTime: yearStart,
		EndTime:   yearEnd,
	}
}

// YearsAfter 几年后时间段 [结算: 年]
func (c *Calender) YearsAfter(n int) *Period {
	c.calculateThisYear()
	thisYearPeriod := c.thisYearPeriod
	yearStart := thisYearPeriod.StartTime.AddDate(n, 0, 0)
	yearEnd := thisYearPeriod.StartTime.AddDate(n+1, 0, 0).Add(-1)
	return &Period{
		StartTime: yearStart,
		EndTime:   yearEnd,
	}
}
