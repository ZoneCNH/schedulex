package schedulex

import (
	"testing"
	"time"
)

func TestTriggerGoldenCases(t *testing.T) {
	base := time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC)
	t.Run("once", func(t *testing.T) {
		next, ok := Once(base.Add(time.Hour)).Next(base)
		if !ok || !next.Equal(base.Add(time.Hour)) {
			t.Fatalf("Once next = %v,%v", next, ok)
		}
	})
	t.Run("every", func(t *testing.T) {
		next, ok := Every(15 * time.Minute).Next(base)
		if !ok || !next.Equal(base.Add(15*time.Minute)) {
			t.Fatalf("Every next = %v,%v", next, ok)
		}
	})
	t.Run("cron", func(t *testing.T) {
		tr, err := Cron("*/20 * * * *")
		if err != nil {
			t.Fatal(err)
		}
		next, ok := tr.Next(base)
		want := time.Date(2026, 6, 4, 10, 40, 0, 0, time.UTC)
		if !ok || !next.Equal(want) {
			t.Fatalf("Cron next = %v,%v want %v,true", next, ok, want)
		}
	})
}

func TestDailyAtDSTGoldenCase(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	after := time.Date(2026, 3, 8, 1, 30, 0, 0, loc)
	next, ok := DailyAt(2, 30, loc).Next(after)
	if !ok || next.In(loc).Hour() != 1 && next.In(loc).Hour() != 3 {
		t.Fatalf("DailyAt DST next = %v,%v", next, ok)
	}
}

func TestReconcileMisfirePolicies(t *testing.T) {
	missed := []time.Time{{}, time.Unix(10, 0)}
	if got := ReconcileMisfire(MisfireSkip, missed); len(got) != 0 {
		t.Fatalf("skip = %v", got)
	}
	if got := ReconcileMisfire(MisfireRunOnce, missed); len(got) != 1 || !got[0].Equal(missed[1]) {
		t.Fatalf("run_once = %v", got)
	}
	if got := ReconcileMisfire(MisfireCatchUp, missed); len(got) != 2 {
		t.Fatalf("catch_up = %v", got)
	}
}
