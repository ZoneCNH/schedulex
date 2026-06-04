package schedulex

import (
	"testing"
	"time"
)

func TestTriggerDeterministicGolden(t *testing.T) {
	loc := time.UTC
	start := time.Date(2026, 6, 4, 10, 0, 0, 0, loc)
	cron, err := Cron("*/15 * * * *", loc)
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]Trigger{
		"once":      Once(start.Add(time.Hour)),
		"interval":  Every(30 * time.Minute),
		"cron_like": cron,
		"daily":     DailyAt(11, 30, 0, loc),
	}
	got := map[string]string{}
	for name, tr := range cases {
		n, ok := tr.Next(start)
		if !ok {
			t.Fatalf("%s no next", name)
		}
		got[name] = n.Format(time.RFC3339)
	}
	assertGoldenJSON(t, "../../contracts/trigger_cases/l1_golden.json", got)
}

func TestDailyAtDSTGolden(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip(err)
	}
	after := time.Date(2026, 3, 8, 1, 30, 0, 0, loc)
	next, ok := DailyAt(3, 30, 0, loc).Next(after)
	if !ok {
		t.Fatal("no next")
	}
	got := map[string]string{"spring_forward": next.Format(time.RFC3339)}
	assertGoldenJSON(t, "../../contracts/trigger_cases/dst_golden.json", got)
}

func TestCronRejectsUnsupportedExpressions(t *testing.T) {
	if _, err := Cron("0 0 1 * *", time.UTC); err == nil {
		t.Fatal("expected unsupported day field")
	}
}

func assertGoldenJSON(t *testing.T, path string, got any) {
	t.Helper()
	b, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, '\n')
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(want) != string(b) {
		t.Fatalf("golden mismatch for %s\nwant=%s\ngot=%s", path, want, b)
	}
}
