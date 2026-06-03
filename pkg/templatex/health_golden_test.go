package templatex_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ZoneCNH/xlib-standard/pkg/templatex"
	"github.com/ZoneCNH/xlib-standard/testkit"
)

func TestHealthStatusJSONGolden(t *testing.T) {
	payload, err := json.Marshal(templatex.HealthStatus{
		Name:      "templatex",
		Status:    templatex.HealthHealthy,
		Message:   "ok",
		CheckedAt: time.Unix(0, 0).UTC(),
		LatencyMs: 7,
		Metadata: map[string]string{
			"kind": "template",
		},
	})
	if err != nil {
		t.Fatalf("marshal health status: %v", err)
	}

	payload = append(payload, '\n')
	testkit.RequireGolden(t, "testdata/golden/health_status.json", payload)
}
