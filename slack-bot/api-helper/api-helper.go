package apihelper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DefaultBaseURL = "http://127.0.0.1:5000"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// MaxRadiusM caps the search radius a user can request.
const MaxRadiusM = 5000

var nearMePattern = regexp.MustCompile(`^!nearme\s+\(\s*(\S+?)\s*,\s*(\S+?)\s*\)(?:\s+(\S+))?\s*$`)

// normalize maps look-alike Unicode dashes and spaces (added by smart-punctuation keyboards) to ASCII.
var normalizer = strings.NewReplacer(
	"\u2010", "-", "\u2011", "-", "\u2012", "-", "\u2013", "-", "\u2014", "-", "\u2212", "-",
	"\u00a0", " ", "\u2009", " ", "\u202f", " ",
)

func normalize(text string) string {
	return strings.TrimSpace(normalizer.Replace(text))
}

// ParseNearMe extracts the arguments from a "!nearme (<lat>, <long>) [distance]"
// message. distance is in metres and is 0 when omitted.
func ParseNearMe(text string) (lat, lon, radius float64, err error) {
	m := nearMePattern.FindStringSubmatch(normalize(text))
	if m == nil {
		return 0, 0, 0, fmt.Errorf("usage: !nearme (<lat>, <long>) [distance]")
	}
	if lat, lon, err = parseCoords(m[1], m[2]); err != nil {
		return 0, 0, 0, err
	}
	if m[3] != "" {
		if radius, err = strconv.ParseFloat(m[3], 64); err != nil || radius <= 0 || radius > MaxRadiusM {
			return 0, 0, 0, fmt.Errorf("invalid distance %q (must be 1-%d metres)", m[3], MaxRadiusM)
		}
	}
	return lat, lon, radius, nil
}

func parseCoords(latStr, lonStr string) (lat, lon float64, err error) {
	if lat, err = strconv.ParseFloat(latStr, 64); err != nil || lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("invalid latitude %q", latStr)
	}
	if lon, err = strconv.ParseFloat(lonStr, 64); err != nil || lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("invalid longitude %q", lonStr)
	}
	return lat, lon, nil
}

var routePattern = regexp.MustCompile(`^!route\s+\(\s*(\S+?)\s*,\s*(\S+?)\s*\)\s+(\S+)\s+(\S+)\s*$`)
var identPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

// ParseRoute extracts the arguments from a "!route (<lat>, <lon>) <operator> <line>" message.
func ParseRoute(text string) (lat, lon float64, operator, line string, err error) {
	m := routePattern.FindStringSubmatch(normalize(text))
	if m == nil {
		return 0, 0, "", "", fmt.Errorf("usage: !route (<lat>, <lon>) <operator> <line>")
	}
	if lat, lon, err = parseCoords(m[1], m[2]); err != nil {
		return 0, 0, "", "", err
	}
	operator, line = m[3], m[4]
	if !identPattern.MatchString(operator) {
		return 0, 0, "", "", fmt.Errorf("invalid operator %q", operator)
	}
	if !identPattern.MatchString(line) {
		return 0, 0, "", "", fmt.Errorf("invalid line %q", line)
	}
	return lat, lon, operator, line, nil
}

// Route handles a "!route (<lat>, <lon>) <operator> <line>" message and returns the raw
// /RouteInfo JSON response.
func (c *Client) Route(ctx context.Context, text string) ([]byte, error) {
	lat, lon, operator, line, err := ParseRoute(text)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	query.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	query.Set("operator", operator)
	query.Set("line", line)
	return c.get(ctx, "/RouteInfo", query)
}

// NearMe handles a "!nearme (<lat>, <long>) [distance]" message and returns the raw
// /LocationInfo JSON response.
func (c *Client) NearMe(ctx context.Context, text string) ([]byte, error) {
	lat, lon, radius, err := ParseNearMe(text)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	query.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	if radius > 0 {
		query.Set("radius", strconv.FormatFloat(radius, 'f', -1, 64))
	}
	return c.get(ctx, "/LocationInfo", query)
}

func (c *Client) get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading API response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return body, nil
}
