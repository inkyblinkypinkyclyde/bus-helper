package routeresponsehandler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"slack-bot/table"
)

type RouteBus struct {
	Destination     string  `json:"destination"`
	Origin          string  `json:"origin"`
	VehicleRef      string  `json:"vehicle_ref"`
	DistanceM       float64 `json:"distance_m"`
	RecordedAt      string  `json:"recorded_at"`
	BearingFromUser *int    `json:"bearing_from_user"`
}

type RouteInfo struct {
	Operator string     `json:"operator"`
	Line     string     `json:"line"`
	Buses    []RouteBus `json:"buses"`
}

// FormatRouteInfo parses a /RouteInfo JSON response into a Slack-ready
// message: a header line followed by a monospace table.
func FormatRouteInfo(data []byte) (string, error) {
	return formatRouteInfo(data, time.Now())
}

func formatRouteInfo(data []byte, now time.Time) (string, error) {
	var info RouteInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return "", fmt.Errorf("parsing route info: %w", err)
	}

	header := fmt.Sprintf("*Line %s (%s) buses*", info.Line, info.Operator)

	if len(info.Buses) == 0 {
		return header + "\nNo live buses found.", nil
	}

	rows := make([][]string, len(info.Buses))
	for i, bus := range info.Buses {
		rows[i] = []string{
			bus.VehicleRef,
			table.Truncate(strings.ReplaceAll(bus.Destination, "_", " "), table.MaxDestinationWidth),
			formatDistance(bus.DistanceM),
			formatAge(bus.RecordedAt, now),
			table.FormatBearing(bus.BearingFromUser),
		}
	}

	tbl := table.Render([]string{"Vehicle", "Destination", "Distance", "Age", "From user"}, rows)
	return header + "\n```\n" + tbl + "```", nil
}

func formatDistance(m float64) string {
	if m >= 1000 {
		return fmt.Sprintf("%.1fkm", m/1000)
	}
	return fmt.Sprintf("%.0fm", m)
}

// formatAge renders how long ago a bus last reported, or "?" if the timestamp is unreadable.
func formatAge(recordedAt string, now time.Time) string {
	t, err := time.Parse("2006-01-02 15:04:05-07:00", recordedAt)
	if err != nil {
		return "?"
	}
	age := now.Sub(t)
	if age < 0 {
		age = 0
	}
	if age < time.Minute {
		return fmt.Sprintf("%ds", int(age.Seconds()))
	}
	return fmt.Sprintf("%dm", int(age.Minutes()))
}
