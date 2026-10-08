package responseformatter

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

type Bus struct {
	Line        string   `json:"line"`
	Destination string   `json:"destination"`
	Origin      string   `json:"origin"`
	Operator    string   `json:"operator"`
	VehicleRef  string   `json:"vehicle_ref"`
	DistanceM   float64  `json:"distance_m"`
	Bearing     *float64 `json:"bearing"`
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
			truncate(bus.Destination, maxDestinationWidth),
			bus.VehicleRef,
			fmt.Sprintf("%.0fm", bus.DistanceM),
		}
	}

	table := renderTable([]string{"Line", "Destination", "Vehicle", "Distance"}, rows)
	return header + "\n```\n" + table + "```", nil
}

// Slack wraps code-block lines that are too wide, which breaks column alignment.
const maxDestinationWidth = 24

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

func renderTable(headers []string, rows [][]string) string {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if n := utf8.RuneCountInString(cell); n > widths[i] {
				widths[i] = n
			}
		}
	}

	var sb strings.Builder
	writeRow := func(cells []string) {
		for i, cell := range cells {
			if i > 0 {
				sb.WriteString(" | ")
			}
			sb.WriteString(cell)
			if i < len(cells)-1 {
				sb.WriteString(strings.Repeat(" ", widths[i]-utf8.RuneCountInString(cell)))
			}
		}
		sb.WriteString("\n")
	}

	writeRow(headers)
	separators := make([]string, len(headers))
	for i, w := range widths {
		separators[i] = strings.Repeat("-", w)
	}
	sb.WriteString(strings.Join(separators, "-+-") + "\n")
	for _, row := range rows {
		writeRow(row)
	}
	return sb.String()
}
