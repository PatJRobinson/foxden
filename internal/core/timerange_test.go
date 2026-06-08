package core

import (
	"testing"
	"time"
)

func TestAllTimeRanges(t *testing.T) {
	got := AllTimeRanges()

	want := []TimeRange{
		TimeRangeDay,
		TimeRangeWeek,
		TimeRangeMonth,
		TimeRangeQuarter,
		TimeRangeYear,
		TimeRangeFiveY,
	}

	if len(got) != len(want) {
		t.Fatalf("AllTimeRanges() returned %d ranges, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("AllTimeRanges()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestTimeRangeNext(t *testing.T) {
	tests := []struct {
		name string
		in   TimeRange
		want TimeRange
	}{
		{
			name: "day advances to week",
			in:   TimeRangeDay,
			want: TimeRangeWeek,
		},
		{
			name: "week advances to month",
			in:   TimeRangeWeek,
			want: TimeRangeMonth,
		},
		{
			name: "month advances to quarter",
			in:   TimeRangeMonth,
			want: TimeRangeQuarter,
		},
		{
			name: "quarter advances to year",
			in:   TimeRangeQuarter,
			want: TimeRangeYear,
		},
		{
			name: "year advances to five years",
			in:   TimeRangeYear,
			want: TimeRangeFiveY,
		},
		{
			name: "five years wraps to day",
			in:   TimeRangeFiveY,
			want: TimeRangeDay,
		},
		{
			name: "unknown value falls back to day",
			in:   TimeRange("nonsense"),
			want: TimeRangeDay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Next()

			if got != tt.want {
				t.Fatalf("%q.Next() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTimeRangeString(t *testing.T) {
	tests := []struct {
		name string
		in   TimeRange
		want string
	}{
		{
			name: "day",
			in:   TimeRangeDay,
			want: "day",
		},
		{
			name: "week",
			in:   TimeRangeWeek,
			want: "week",
		},
		{
			name: "month",
			in:   TimeRangeMonth,
			want: "month",
		},
		{
			name: "quarter",
			in:   TimeRangeQuarter,
			want: "quarter",
		},
		{
			name: "year",
			in:   TimeRangeYear,
			want: "year",
		},
		{
			name: "five years",
			in:   TimeRangeFiveY,
			want: "5y",
		},
		{
			name: "unknown value",
			in:   TimeRange("custom"),
			want: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.String()

			if got != tt.want {
				t.Fatalf("%q.String() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTimeRangeSince(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   TimeRange
		want time.Time
	}{
		{
			name: "day",
			in:   TimeRangeDay,
			want: time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "week",
			in:   TimeRangeWeek,
			want: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "month",
			in:   TimeRangeMonth,
			want: time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "quarter",
			in:   TimeRangeQuarter,
			want: time.Date(2026, 3, 8, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "year",
			in:   TimeRangeYear,
			want: time.Date(2025, 6, 8, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "five years",
			in:   TimeRangeFiveY,
			want: time.Date(2021, 6, 8, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "unknown falls back to week",
			in:   TimeRange("nonsense"),
			want: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Since(now)
			if !got.Equal(tt.want) {
				t.Fatalf("%q.Since(%v) = %v, want %v", tt.in, now, got, tt.want)
			}
		})
	}
}
