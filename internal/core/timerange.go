package core

import "time"

type TimeRange string

const (
	TimeRangeDay     TimeRange = "day"
	TimeRangeWeek    TimeRange = "week"
	TimeRangeMonth   TimeRange = "month"
	TimeRangeQuarter TimeRange = "quarter"
	TimeRangeYear    TimeRange = "year"
	TimeRangeFiveY   TimeRange = "5y"
)

func AllTimeRanges() []TimeRange {
	return []TimeRange{
		TimeRangeDay,
		TimeRangeWeek,
		TimeRangeMonth,
		TimeRangeQuarter,
		TimeRangeYear,
		TimeRangeFiveY,
	}
}

func (r TimeRange) Next() TimeRange {
	ranges := AllTimeRanges()

	for i, candidate := range ranges {
		if candidate == r {
			return ranges[(i+1)%len(ranges)]
		}
	}

	return TimeRangeDay
}

func (r TimeRange) String() string {
	return string(r)
}

func (r TimeRange) Since(now time.Time) time.Time {
	switch r {
	case TimeRangeDay:
		return now.AddDate(0, 0, -1)
	case TimeRangeWeek:
		return now.AddDate(0, 0, -7)
	case TimeRangeMonth:
		return now.AddDate(0, -1, 0)
	case TimeRangeQuarter:
		return now.AddDate(0, -3, 0)
	case TimeRangeYear:
		return now.AddDate(-1, 0, 0)
	case TimeRangeFiveY:
		return now.AddDate(-5, 0, 0)
	default:
		return TimeRangeWeek.Since(now)
	}
}
