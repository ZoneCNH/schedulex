package schedulex

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Trigger calculates the next scheduled instant after a reference time.
type Trigger interface {
	Next(after time.Time) (time.Time, bool)
}

type onceTrigger struct{ at time.Time }

// Once fires exactly once at at.
func Once(at time.Time) Trigger { return onceTrigger{at: at} }

func (t onceTrigger) Next(after time.Time) (time.Time, bool) {
	if after.Before(t.at) {
		return t.at, true
	}
	return time.Time{}, false
}

type everyTrigger struct{ interval time.Duration }

// Every fires repeatedly at interval boundaries after the reference time.
func Every(interval time.Duration) Trigger { return everyTrigger{interval: interval} }

func (t everyTrigger) Next(after time.Time) (time.Time, bool) {
	if t.interval <= 0 {
		return time.Time{}, false
	}
	return after.Add(t.interval), true
}

type dailyAtTrigger struct {
	hour int
	min  int
	loc  *time.Location
}

// DailyAt fires once per day at hour:minute in loc. loc defaults to UTC.
func DailyAt(hour, minute int, loc *time.Location) Trigger {
	if loc == nil {
		loc = time.UTC
	}
	return dailyAtTrigger{hour: hour, min: minute, loc: loc}
}

func (t dailyAtTrigger) Next(after time.Time) (time.Time, bool) {
	if t.hour < 0 || t.hour > 23 || t.min < 0 || t.min > 59 {
		return time.Time{}, false
	}
	local := after.In(t.loc)
	candidate := time.Date(local.Year(), local.Month(), local.Day(), t.hour, t.min, 0, 0, t.loc)
	if !candidate.After(local) {
		candidate = candidate.AddDate(0, 0, 1)
	}
	return candidate, true
}

type cronTrigger struct{ interval time.Duration }

// Cron supports the deterministic v0.1 subset: @hourly, @daily, and */N * * * *.
func Cron(expr string) (Trigger, error) {
	switch strings.TrimSpace(expr) {
	case "@hourly":
		return cronTrigger{interval: time.Hour}, nil
	case "@daily":
		return DailyAt(0, 0, time.UTC), nil
	}
	fields := strings.Fields(expr)
	if len(fields) == 5 && strings.HasPrefix(fields[0], "*/") && fields[1] == "*" && fields[2] == "*" && fields[3] == "*" && fields[4] == "*" {
		mins, err := strconv.Atoi(strings.TrimPrefix(fields[0], "*/"))
		if err != nil || mins <= 0 {
			return nil, fmt.Errorf("schedulex: invalid cron minute interval %q", fields[0])
		}
		return cronTrigger{interval: time.Duration(mins) * time.Minute}, nil
	}
	return nil, fmt.Errorf("schedulex: unsupported cron expression %q", expr)
}

func (t cronTrigger) Next(after time.Time) (time.Time, bool) {
	if t.interval <= 0 {
		return time.Time{}, false
	}
	return after.Truncate(t.interval).Add(t.interval), true
}
