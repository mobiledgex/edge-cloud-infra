package orm

import (
	"time"
)

type TimeDefinitionObj interface {
	GetStartTime() time.Time
	GetEndTime() time.Time
	GetLast() int
	SetStartTime(t time.Time)
	SetEndTime(t time.Time)
	SetLast(l int)
}

const (
	DefaultTimeWindow = 15 * time.Second
	// Max 100 data points on the graph
	MaxTimeDefinition = 100
)

func getTimeDefinition(obj TimeDefinitionObj, minTimeWindow time.Duration) string {
	duration := getTimeDefinitionDuration(obj, minTimeWindow)
	if duration <= 0 {
		return ""
	}
	return duration.String()
}

func getTimeDefinitionDuration(obj TimeDefinitionObj, minTimeWindow time.Duration) time.Duration {
	start := obj.GetStartTime()
	end := obj.GetEndTime()
	last := obj.GetLast()
	// In case we are requesting last n number of entries and don't provide time window
	// we should skip the function and time-based grouping
	if start.IsZero() && end.IsZero() && last != 0 {
		return 0
	}
	// set the max number of data points per grouping
	if last == 0 {
		obj.SetLast(MaxTimeDefinition)
	}
	if end.IsZero() {
		end = time.Now().UTC()
		obj.SetEndTime(end)
	}
	// Default time to last 12hrs
	if start.IsZero() {
		obj.SetStartTime(end.Add(-12 * time.Hour).UTC())
	}
	// If start time is past end time, cannot group by time
	timeDiff := obj.GetEndTime().Sub(obj.GetStartTime())
	if timeDiff < 0 {
		return 0
	}
	// Make sure we don't have any fractional seconds in here
	timeWindow := time.Duration(timeDiff / time.Duration(obj.GetLast())).Truncate(time.Second)
	if timeWindow < minTimeWindow {
		return minTimeWindow
	}
	return timeWindow
}
