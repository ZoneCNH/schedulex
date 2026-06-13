package schedulex

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestTriggerDeterminism_SameClockSameResult 验证相同 clock 输入产生相同的 next time。
// 使用 StaticClock 固定当前时间，多次调用 Next() 验证结果一致。
func TestTriggerDeterminism_SameClockSameResult(t *testing.T) {
	fixedTime := time.Date(2026, 6, 9, 10, 0, 0, 0, time.UTC)

	triggers := map[string]Trigger{
		"once":     Once(fixedTime.Add(time.Hour)),
		"interval": Every(30 * time.Minute),
		"daily":    DailyAt(11, 30, time.UTC),
	}

	cron, err := Cron("*/15 * * * *", time.UTC)
	if err != nil {
		t.Fatal(err)
	}
	triggers["cron"] = cron

	for name, tr := range triggers {
		t.Run(name, func(t *testing.T) {
			// 调用 100 次，结果应完全一致
			first, ok := tr.Next(fixedTime)
			if !ok {
				t.Fatalf("trigger %s returned no next", name)
			}
			for i := 0; i < 100; i++ {
				got, ok := tr.Next(fixedTime)
				if !ok {
					t.Fatalf("trigger %s returned no next on iteration %d", name, i)
				}
				if !got.Equal(first) {
					t.Fatalf("trigger %s: iteration %d returned %v, want %v", name, i, got, first)
				}
			}
		})
	}
}

// TestTriggerDeterminism_MultipleCallsAdvance 验证多次 advance 后 Next 结果的可重复性。
func TestTriggerDeterminism_MultipleCallsAdvance(t *testing.T) {
	loc := time.UTC
	cron, err := Cron("*/15 * * * *", loc)
	if err != nil {
		t.Fatal(err)
	}

	// 第一轮：从固定时间开始，连续调用 Next 10 次
	base := time.Date(2026, 6, 9, 10, 0, 0, 0, loc)
	sequence1 := make([]time.Time, 10)
	current := base
	for i := 0; i < 10; i++ {
		next, ok := cron.Next(current)
		if !ok {
			t.Fatalf("round 1: no next at step %d", i)
		}
		sequence1[i] = next
		current = next
	}

	// 第二轮：同样的起始点，同样的步骤
	current = base
	for i := 0; i < 10; i++ {
		next, ok := cron.Next(current)
		if !ok {
			t.Fatalf("round 2: no next at step %d", i)
		}
		if !next.Equal(sequence1[i]) {
			t.Fatalf("step %d: round 2 got %v, want %v", i, next, sequence1[i])
		}
		current = next
	}
}

// TestTriggerDeterminism_DifferentInputsDifferentResults 验证不同输入产生不同结果。
func TestTriggerDeterminism_DifferentInputsDifferentResults(t *testing.T) {
	loc := time.UTC
	daily := DailyAt(9, 0, loc)

	t1 := time.Date(2026, 6, 9, 8, 0, 0, 0, loc)
	t2 := time.Date(2026, 6, 10, 8, 0, 0, 0, loc)

	next1, ok := daily.Next(t1)
	if !ok {
		t.Fatal("no next for t1")
	}
	next2, ok := daily.Next(t2)
	if !ok {
		t.Fatal("no next for t2")
	}

	if next1.Equal(next2) {
		t.Fatalf("expected different results for different inputs: %v vs %v", next1, next2)
	}
}

// TestTriggerDeterminism_DSTGoldenValidates 验证 DST golden 文件中的所有场景。
func TestTriggerDeterminism_DSTGoldenValidates(t *testing.T) {
	data, err := os.ReadFile("../../testdata/golden/dst_transitions.json")
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	var golden struct {
		Cases []struct {
			Name     string `json:"name"`
			Timezone string `json:"timezone"`
			Trigger  struct {
				Type   string `json:"type"`
				Hour   int    `json:"hour"`
				Minute int    `json:"minute"`
			} `json:"trigger"`
			After        string `json:"after"`
			ExpectedNext string `json:"expected_next"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("parse golden: %v", err)
	}

	for _, tc := range golden.Cases {
		t.Run(tc.Name, func(t *testing.T) {
			loc, err := time.LoadLocation(tc.Timezone)
			if err != nil {
				t.Fatalf("load location %s: %v", tc.Timezone, err)
			}

			after, err := time.Parse(time.RFC3339, tc.After)
			if err != nil {
				t.Fatalf("parse after: %v", err)
			}
			expected, err := time.Parse(time.RFC3339, tc.ExpectedNext)
			if err != nil {
				t.Fatalf("parse expected_next: %v", err)
			}

			var trigger Trigger
			switch tc.Trigger.Type {
			case "daily_at":
				trigger = DailyAt(tc.Trigger.Hour, tc.Trigger.Minute, loc)
			default:
				t.Fatalf("unsupported trigger type: %s", tc.Trigger.Type)
			}

			// 执行两次验证确定性
			for i := 0; i < 2; i++ {
				got, ok := trigger.Next(after)
				if !ok {
					t.Fatalf("iteration %d: trigger returned no next", i)
				}
				if !got.Equal(expected) {
					t.Fatalf("iteration %d: got %v, want %v (UTC: got %v, want %v)",
						i, got, expected, got.UTC(), expected.UTC())
				}
			}
		})
	}
}
