package core

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
