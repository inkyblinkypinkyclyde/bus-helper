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

// ParseNearMe extracts the arguments from a "!nearme (<lat>, <long>) [distance]"
// message. distance is in metres and is 0 when omitted.
func ParseNearMe(text string) (lat, lon, radius float64, err error) {
	m := nearMePattern.FindStringSubmatch(strings.TrimSpace(text))
	if m == nil {
		return 0, 0, 0, fmt.Errorf("usage: !nearme (<lat>, <long>) [distance]")
	}
	if lat, err = strconv.ParseFloat(m[1], 64); err != nil || lat < -90 || lat > 90 {
		return 0, 0, 0, fmt.Errorf("invalid latitude %q", m[1])
	}
	if lon, err = strconv.ParseFloat(m[2], 64); err != nil || lon < -180 || lon > 180 {
		return 0, 0, 0, fmt.Errorf("invalid longitude %q", m[2])
	}
	if m[3] != "" {
		if radius, err = strconv.ParseFloat(m[3], 64); err != nil || radius <= 0 || radius > MaxRadiusM {
			return 0, 0, 0, fmt.Errorf("invalid distance %q (must be 1-%d metres)", m[3], MaxRadiusM)
		}
	}
	return lat, lon, radius, nil
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
