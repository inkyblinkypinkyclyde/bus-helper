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
