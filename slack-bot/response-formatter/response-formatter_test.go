package responseformatter

import (
	"os"
	"strings"
	"testing"
)

func TestFormatLocationInfo(t *testing.T) {
	data, err := os.ReadFile("../res.json")
	if err != nil {
		t.Fatal(err)
	}

	got, err := FormatLocationInfo(data)
	if err != nil {
		t.Fatal(err)
	}

	want := "*Buses within 50m of 51.5077, -0.1297*\n" +
		"```\n" +
		"Line | Destination             | Vehicle | Distance\n" +
		"-----+-------------------------+---------+---------\n" +
		"139  | Golders Green Station   | LK15CUH | 25m\n" +
		"9    | Hammersmith Bus Station | LTZ1096 | 31m\n" +
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
