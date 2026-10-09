package table

import (
	"strings"
	"unicode/utf8"
)

// MaxDestinationWidth keeps rows narrow because Slack wraps wide code-block lines, which breaks alignment.
const MaxDestinationWidth = 24

func Truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// Render lays out rows as a monospace table with a header and separator line.
func Render(headers []string, rows [][]string) string {
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
