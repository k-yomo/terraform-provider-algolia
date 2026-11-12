package algoliautil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewDebugRequester(t *testing.T) {
	t.Parallel()

	got := NewDebugRequester()
	if got == nil || got.Client == nil {
		t.Fatalf("NewDebugRequester() returned nil requester or client")
	}
	if _, ok := got.Client.Transport.(*debugTransport); !ok {
		t.Errorf("NewDebugRequester() transport = %T, want *debugTransport", got.Client.Transport)
	}
}

func TestDebugRequester_Request(t *testing.T) {
	t.Parallel()

	const respBody = `{"status":"ok"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Algolia-API-Key"); got != "secret-key" {
			t.Errorf("server received api key header %q, want %q", got, "secret-key")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, respBody)
	}))
	defer server.Close()

	req, err := http.NewRequest(http.MethodPost, server.URL+"/1/indexes/test/query", strings.NewReader(`{"query":"phone"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("X-Algolia-API-Key", "secret-key")

	resp, err := NewDebugRequester().Request(req, time.Second, time.Second)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	defer resp.Body.Close()

	// Dumping the response for logging must not consume the body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != respBody {
		t.Errorf("response body = %q, want %q", body, respBody)
	}
}

func TestDebugTransport_RoundTripError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close() // nothing listens anymore, so the round trip fails

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newDebugTransport(http.DefaultTransport).RoundTrip(req); err == nil {
		t.Error("RoundTrip() error = nil, want connection error")
	}
}

func TestPrettyPrintJsonLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		in          string
		want        []string
		mustNotHave []string
	}{
		{
			name: "masks algolia credential headers",
			in:   "POST /1/indexes/test HTTP/1.1\nX-Algolia-Api-Key: secret-key\nX-Algolia-Application-Id: APPID123\nAccept: application/json",
			want: []string{
				"X-Algolia-Api-Key: **********",
				"X-Algolia-Application-Id: ********",
				"Accept: application/json",
			},
			mustNotHave: []string{"secret-key", "APPID123"},
		},
		{
			name:        "masks headers case-insensitively",
			in:          "x-algolia-api-key: secret-key",
			want:        []string{"x-algolia-api-key: **********"},
			mustNotHave: []string{"secret-key"},
		},
		{
			name: "pretty prints json lines",
			in:   "HTTP/1.1 200 OK\n\n{\"status\":200,\"message\":\"ok\"}",
			want: []string{"HTTP/1.1 200 OK", "{\n \"status\": 200,\n \"message\": \"ok\"\n}"},
		},
		{
			name: "leaves non-json lines untouched",
			in:   "GET /1/keys HTTP/1.1\nnot json {",
			want: []string{"GET /1/keys HTTP/1.1\nnot json {"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := prettyPrintJsonLines([]byte(tt.in))
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("prettyPrintJsonLines() = %q, want it to contain %q", got, w)
				}
			}
			for _, s := range tt.mustNotHave {
				if strings.Contains(got, s) {
					t.Errorf("prettyPrintJsonLines() = %q, must not contain %q", got, s)
				}
			}
		})
	}
}
