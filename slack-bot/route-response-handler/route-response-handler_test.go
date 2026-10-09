package routeresponsehandler

import (
	"strings"
	"testing"
	"time"
)

const routeJSON = `{
  "buses": [
    {"destination": "Bus_Station", "distance_m": 69722.2, "origin": "Central_Railway_Station", "recorded_at": "2026-10-08 18:29:15+00:00", "vehicle_ref": "3868"},
    {"destination": "Central_Railway_Station", "distance_m": 450.4, "origin": "Bus_Station", "recorded_at": "2026-10-08 18:27:57+00:00", "vehicle_ref": "3309"}
  ],
  "line": "4",
  "operator": "ARBB"
}`

func TestFormatRouteInfo(t *testing.T) {
	now := time.Date(2026, 10, 8, 18, 30, 0, 0, time.UTC)

	got, err := formatRouteInfo([]byte(routeJSON), now)
	if err != nil {
		t.Fatal(err)
	}

	want := "*Line 4 (ARBB) buses*\n" +
		"```\n" +
		"Vehicle | Destination             | Distance | Age\n" +
		"--------+-------------------------+----------+----\n" +
		"3868    | Bus Station             | 69.7km   | 45s\n" +
		"3309    | Central Railway Station | 450m     | 2m\n" +
		"```"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatRouteInfoNoBuses(t *testing.T) {
	got, err := FormatRouteInfo([]byte(`{"line":"4","operator":"ARBB","buses":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Line 4") || !strings.Contains(got, "No live buses") {
		t.Errorf("unexpected output: %s", got)
	}
}

func TestFormatRouteInfoInvalidJSON(t *testing.T) {
	if _, err := FormatRouteInfo([]byte("not json")); err == nil {
		t.Error("expected error")
	}
}
