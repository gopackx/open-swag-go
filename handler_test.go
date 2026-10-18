package openswag

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gopackx/open-swag-go/pkg/ui"
)

func newTestServer(t *testing.T, cfg Config) *httptest.Server {
	t.Helper()
	if cfg.Info.Title == "" {
		cfg.Info = Info{Title: "Test API", Version: "1.0.0"}
	}
	docs := New(cfg)
	mux := http.NewServeMux()
	docs.Mount(mux, "/docs")
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func doGet(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", req.URL, err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func get(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return doGet(t, req)
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}

func TestMount_UIServesPinnedScalar(t *testing.T) {
	srv := newTestServer(t, Config{})

	resp := get(t, srv.URL+"/docs/")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", ct)
	}
	body := readBody(t, resp)
	want := "@scalar/api-reference@" + ui.DefaultScalarVersion
	if !strings.Contains(body, want) {
		t.Fatalf("body does not contain %q", want)
	}
}

func TestMount_UIEscapesTitle(t *testing.T) {
	srv := newTestServer(t, Config{Info: Info{Title: "</title><script>alert(1)</script>", Version: "1"}})

	body := readBody(t, get(t, srv.URL+"/docs/"))
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Fatal("title was not HTML-escaped")
	}
}

func TestMount_UIScalarVersionOverride(t *testing.T) {
	srv := newTestServer(t, Config{UI: UIConfig{ScalarVersion: "1.2.3"}})

	body := readBody(t, get(t, srv.URL+"/docs/"))
	if !strings.Contains(body, "@scalar/api-reference@1.2.3") {
		t.Fatal("ScalarVersion override not applied")
	}
}

func TestMount_SpecDefaultCORS(t *testing.T) {
	srv := newTestServer(t, Config{})

	resp := get(t, srv.URL+"/docs/openapi.json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
}

func TestMount_SpecCORSDisabled(t *testing.T) {
	srv := newTestServer(t, Config{UI: UIConfig{SpecCORSOrigin: "-"}})

	resp := get(t, srv.URL+"/docs/openapi.json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if _, ok := resp.Header["Access-Control-Allow-Origin"]; ok {
		t.Fatalf("Access-Control-Allow-Origin present: %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestMount_SpecCORSCustomOrigin(t *testing.T) {
	srv := newTestServer(t, Config{UI: UIConfig{SpecCORSOrigin: "https://app.example.com"}})

	resp := get(t, srv.URL+"/docs/openapi.json")
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestDocsAuth_APIKey(t *testing.T) {
	srv := newTestServer(t, Config{DocsAuth: &DocsAuth{Enabled: true, APIKey: "secret"}})

	if resp := get(t, srv.URL+"/docs/openapi.json"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("without key: status = %d, want 401", resp.StatusCode)
	}
	if resp := get(t, srv.URL+"/docs/openapi.json?key=wrong"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong key: status = %d, want 401", resp.StatusCode)
	}
	if resp := get(t, srv.URL+"/docs/openapi.json?key=secret"); resp.StatusCode != http.StatusOK {
		t.Fatalf("right key: status = %d, want 200", resp.StatusCode)
	}
	if resp := get(t, srv.URL+"/docs/?key=secret"); resp.StatusCode != http.StatusOK {
		t.Fatalf("UI with right key: status = %d, want 200", resp.StatusCode)
	}
}

func TestDocsAuth_BasicAuth(t *testing.T) {
	srv := newTestServer(t, Config{DocsAuth: &DocsAuth{Enabled: true, Username: "admin", Password: "pw"}})

	if resp := get(t, srv.URL+"/docs/"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no credentials: status = %d, want 401", resp.StatusCode)
	} else if resp.Header.Get("WWW-Authenticate") == "" {
		t.Fatal("missing WWW-Authenticate header on 401")
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/docs/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("admin", "wrong")
	if resp := doGet(t, req); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password: status = %d, want 401", resp.StatusCode)
	}

	req, err = http.NewRequest(http.MethodGet, srv.URL+"/docs/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetBasicAuth("admin", "pw")
	if resp := doGet(t, req); resp.StatusCode != http.StatusOK {
		t.Fatalf("right credentials: status = %d, want 200", resp.StatusCode)
	}
}

func TestDocsAuth_EnabledWithoutCredentials(t *testing.T) {
	srv := newTestServer(t, Config{DocsAuth: &DocsAuth{Enabled: true}})

	resp := get(t, srv.URL+"/docs/openapi.json")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	if body := readBody(t, resp); !strings.Contains(body, "DocsAuth enabled but no credentials configured") {
		t.Fatalf("unexpected body %q", body)
	}
}
