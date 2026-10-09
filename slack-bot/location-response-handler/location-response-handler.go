package locationresponsehandler

import (
	"encoding/json"
	"fmt"

	"slack-bot/table"
)

type Bus struct {
	Line            string   `json:"line"`
	Destination     string   `json:"destination"`
	Origin          string   `json:"origin"`
	Operator        string   `json:"operator"`
	VehicleRef      string   `json:"vehicle_ref"`
	DistanceM       float64  `json:"distance_m"`
	Bearing         *float64 `json:"bearing"`
	BearingFromUser *int     `json:"bearing_from_user"`
}

type LocationInfo struct {
	Name    *string `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	RadiusM float64 `json:"radius_m"`
	Buses   []Bus   `json:"buses"`
}

// FormatLocationInfo parses a /LocationInfo JSON response into a Slack-ready
// message: a header line followed by a monospace table.
func FormatLocationInfo(data []byte) (string, error) {
	var info LocationInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return "", fmt.Errorf("parsing location info: %w", err)
	}

	place := fmt.Sprintf("%.4f, %.4f", info.Lat, info.Lon)
	if info.Name != nil && *info.Name != "" {
		place = *info.Name
	}
	header := fmt.Sprintf("*Buses within %.0fm of %s*", info.RadiusM, place)

	if len(info.Buses) == 0 {
		return header + "\nNo live buses found.", nil
	}

	rows := make([][]string, len(info.Buses))
	for i, bus := range info.Buses {
		rows[i] = []string{
			bus.Line,
			table.Truncate(bus.Destination, table.MaxDestinationWidth),
			bus.VehicleRef,
			fmt.Sprintf("%.0fm", bus.DistanceM),
			table.FormatBearing(bus.BearingFromUser),
		}
	}

	tbl := table.Render([]string{"Line", "Destination", "Vehicle", "Distance", "From user"}, rows)
	return header + "\n```\n" + tbl + "```", nil
}
