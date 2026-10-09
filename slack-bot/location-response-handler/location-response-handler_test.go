package locationresponsehandler

import (
	"strings"
	"testing"
)

const locationJSON = `{
  "buses": [
	{"bearing": 270.0, "bearing_from_user": 90, "destination": "Golders Green Station", "distance_m": 25.36, "line": "139", "operator": "TFLO", "origin": "Waterloo Station / Tenison Way", "vehicle_ref": "LK15CUH"},
	{"bearing": null, "bearing_from_user": null, "destination": "Hammersmith Bus Station", "distance_m": 30.72, "line": "9", "operator": "TFLO", "origin": "Aldwych / Bush House", "vehicle_ref": "LTZ1096"}
  ],
  "lat": 51.5077,
  "lon": -0.1297,
  "name": null,
  "radius_m": 50.0
}`

func TestFormatLocationInfo(t *testing.T) {
	got, err := FormatLocationInfo([]byte(locationJSON))
	if err != nil {
		t.Fatal(err)
	}

	want := "*Buses within 50m of 51.5077, -0.1297*\n" +
		"```\n" +
		"Line | Destination             | Vehicle | Distance | From user\n" +
		"-----+-------------------------+---------+----------+----------\n" +
		"139  | Golders Green Station   | LK15CUH | 25m      | 090°\n" +
		"9    | Hammersmith Bus Station | LTZ1096 | 31m      | ?\n" +
		"```"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFormatLocationInfoNoBuses(t *testing.T) {
	got, err := FormatLocationInfo([]byte(`{"name":"Stop","radius_m":100,"buses":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Stop") || !strings.Contains(got, "No live buses") {
		t.Errorf("unexpected output: %s", got)
	}
}

func TestFormatLocationInfoInvalidJSON(t *testing.T) {
	if _, err := FormatLocationInfo([]byte("not json")); err == nil {
		t.Error("expected error")
	}
}
