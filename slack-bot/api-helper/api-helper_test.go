package apihelper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseNearMe(t *testing.T) {
	for _, text := range []string{
		"!nearme (51.5077, -0.1297)",
		"!nearme (51.5077,-0.1297)",
		"!nearme  ( 51.5077 , -0.1297 ) ",
	} {
		lat, lon, radius, err := ParseNearMe(text)
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		if lat != 51.5077 || lon != -0.1297 || radius != 0 {
			t.Errorf("%q: got %v, %v, %v", text, lat, lon, radius)
		}
	}
}

func TestParseNearMeWithRadius(t *testing.T) {
	_, _, radius, err := ParseNearMe("!nearme (51.5077, -0.1297) 100")
	if err != nil {
		t.Fatal(err)
	}
	if radius != 100 {
		t.Errorf("got radius %v", radius)
	}
}

func TestParseNearMeInvalid(t *testing.T) {
	for _, text := range []string{
		"",
		"!nearme",
		"!nearme 51.5 -0.1",
		"!nearme (51.5)",
		"!nearme (51.5, -0.1",
		"!nearme (a, b)",
		"!nearme (91, 0)",
		"!nearme (0, 181)",
		"!status (51.5, -0.1)",
		"!nearme (51.5, -0.1) 0",
		"!nearme (51.5, -0.1) -5",
		"!nearme (51.5, -0.1) abc",
		"!nearme (51.5, -0.1) 99999",
		"!nearme (51.5, -0.1) 100 extra",
	} {
		if _, _, _, err := ParseNearMe(text); err == nil {
			t.Errorf("expected error for %q", text)
		}
	}
}

func TestParseRoute(t *testing.T) {
	for _, text := range []string{
		"!route (51.5074, -0.1278) ARBB 4",
		"!route (51.5074,-0.1278)  ARBB   4 ",
	} {
		lat, lon, operator, line, err := ParseRoute(text)
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		if lat != 51.5074 || lon != -0.1278 || operator != "ARBB" || line != "4" {
			t.Errorf("%q: got %v, %v, %q, %q", text, lat, lon, operator, line)
		}
	}
}

func TestParseRouteInvalid(t *testing.T) {
	for _, text := range []string{
		"",
		"!route",
		"!route ARBB 4",
		"!route (51.5, -0.1)",
		"!route (51.5, -0.1) ARBB",
		"!route (51.5, -0.1) ARBB 4 extra",
		"!route (91, 0) ARBB 4",
		"!route (0, 181) ARBB 4",
		"!route (a, b) ARBB 4",
		"!route (51.5, -0.1) AR&BB 4",
		"!route (51.5, -0.1) ARBB 4;x",
		"!nearme (51.5, -0.1) ARBB 4",
	} {
		if _, _, _, _, err := ParseRoute(text); err == nil {
			t.Errorf("expected error for %q", text)
		}
	}
}

func TestParseRouteUnicodeLookalikes(t *testing.T) {
	for _, text := range []string{
		"!route (52.0414115, \u22120.7563261) ARBB 4",
		"!route (52.0414115, \u20130.7563261) ARBB 4",
		"!route (52.0414115,\u00a0-0.7563261) ARBB 4",
	} {
		lat, lon, _, _, err := ParseRoute(text)
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		if lat != 52.0414115 || lon != -0.7563261 {
			t.Errorf("%q: got %v, %v", text, lat, lon)
		}
	}
}

func TestParseNearMeUnicodeLookalikes(t *testing.T) {
	_, lon, _, err := ParseNearMe("!nearme (52.0414115,\u00a0\u22120.7563261)")
	if err != nil {
		t.Fatal(err)
	}
	if lon != -0.7563261 {
		t.Errorf("got %v", lon)
	}
}

func TestRoute(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/RouteInfo" || q.Get("lat") != "51.5074" || q.Get("lon") != "-0.1278" ||
			q.Get("operator") != "ARBB" || q.Get("line") != "4" {
			t.Errorf("unexpected request: %s", r.URL)
		}
		w.Write([]byte(`{"buses":[]}`))
	}))
	defer srv.Close()

	got, err := NewClient(srv.URL).Route(context.Background(), "!route (51.5074, -0.1278) ARBB 4")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"buses":[]}` {
		t.Errorf("got %s", got)
	}
}

func TestNearMe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/LocationInfo" || q.Get("lat") != "51.5077" || q.Get("lon") != "-0.1297" || q.Has("radius") {
			t.Errorf("unexpected request: %s", r.URL)
		}
		w.Write([]byte(`{"buses":[]}`))
	}))
	defer srv.Close()

	got, err := NewClient(srv.URL).NearMe(context.Background(), "!nearme (51.5077, -0.1297)")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"buses":[]}` {
		t.Errorf("got %s", got)
	}
}

func TestNearMeWithRadius(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("radius") != "100" {
			t.Errorf("unexpected request: %s", r.URL)
		}
		w.Write([]byte(`{"buses":[]}`))
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL).NearMe(context.Background(), "!nearme (51.5077, -0.1297) 100"); err != nil {
		t.Fatal(err)
	}
}

func TestNearMeAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"boom"}`, http.StatusBadGateway)
	}))
	defer srv.Close()

	if _, err := NewClient(srv.URL).NearMe(context.Background(), "!nearme (51.5, -0.1)"); err == nil {
		t.Error("expected error")
	}
}
