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

// TriggerOption reserves trigger-level configuration without changing constructors.
type TriggerOption func(*triggerConfig)

type triggerConfig struct{}

type onceTrigger struct{ at time.Time }

func Once(at time.Time) Trigger { return onceTrigger{at: at} }

func (t onceTrigger) Next(after time.Time) (time.Time, bool) {
	if t.at.After(after) {
		return t.at, true
	}
	return time.Time{}, false
}

type intervalTrigger struct{ every time.Duration }

func Every(d time.Duration, _ ...TriggerOption) Trigger { return intervalTrigger{every: d} }
func (t intervalTrigger) Next(after time.Time) (time.Time, bool) {
	if t.every <= 0 {
		return time.Time{}, false
	}
	return after.Add(t.every), true
}

type dailyTrigger struct {
	hour, minute, sec int
	loc               *time.Location
}

func DailyAt(hour, minute int, loc *time.Location, _ ...TriggerOption) Trigger {
	if loc == nil {
		loc = time.UTC
	}
	return dailyTrigger{hour: hour, minute: minute, sec: 0, loc: loc}
}
func (t dailyTrigger) Next(after time.Time) (time.Time, bool) {
	if t.hour < 0 || t.hour > 23 || t.minute < 0 || t.minute > 59 {
		return time.Time{}, false
	}
	local := after.In(t.loc)
	n := time.Date(local.Year(), local.Month(), local.Day(), t.hour, t.minute, t.sec, 0, t.loc)
	if !n.After(local) {
		n = n.AddDate(0, 0, 1)
	}
	return n, true
}

type cronTrigger struct {
	minuteStep int
	hourStep   int
	minute     *int
	hour       *int
	loc        *time.Location
}

// Cron supports deterministic L1 cron-like expressions with five fields.
// Supported minute/hour forms: *, */N, and fixed integers. Other fields must be *.
func Cron(expr string, loc *time.Location, _ ...TriggerOption) (Trigger, error) {
	if loc == nil {
		loc = time.UTC
	}
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return nil, fmt.Errorf("schedulex: cron expects 5 fields")
	}
	for _, p := range parts[2:] {
		if p != "*" {
			return nil, fmt.Errorf("schedulex: only wildcard day/month/week fields supported")
		}
	}
	step, minute, err := parseCronField(parts[0], 0, 59)
	if err != nil {
		return nil, err
	}
	hourStep, hour, err := parseCronField(parts[1], 0, 23)
	if err != nil {
		return nil, err
	}
	if step == 0 {
		step = 1
	}
	if hourStep == 0 {
		hourStep = 1
	}
	return cronTrigger{minuteStep: step, hourStep: hourStep, minute: minute, hour: hour, loc: loc}, nil
}

func parseCronField(v string, floor, ceiling int) (int, *int, error) {
	if v == "*" {
		return 1, nil, nil
	}
	if strings.HasPrefix(v, "*/") {
		n, err := strconv.Atoi(strings.TrimPrefix(v, "*/"))
		if err != nil || n <= 0 || n > ceiling+1 {
			return 0, nil, fmt.Errorf("schedulex: invalid cron step %q", v)
		}
		return n, nil, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < floor || n > ceiling {
		return 0, nil, fmt.Errorf("schedulex: invalid cron value %q", v)
	}
	return 0, &n, nil
}

func (t cronTrigger) Next(after time.Time) (time.Time, bool) {
	base := after.In(t.loc).Truncate(time.Minute).Add(time.Minute)
	for i := 0; i < 366*24*60; i++ {
		c := base.Add(time.Duration(i) * time.Minute)
		if t.hour != nil && c.Hour() != *t.hour {
			continue
		}
		if t.hour == nil && c.Hour()%t.hourStep != 0 {
			continue
		}
		if t.minute != nil {
			if c.Minute() != *t.minute {
				continue
			}
		} else if c.Minute()%t.minuteStep != 0 {
			continue
		}
		return c, true
	}
	return time.Time{}, false
}
