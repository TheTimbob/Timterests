package server_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"timterests/internal/server"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	s := &server.Server{}

	svr := httptest.NewServer(http.HandlerFunc(s.HelloWorldHandler))
	defer svr.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, svr.URL, nil)
	if err != nil {
		t.Fatalf("error creating request. Err: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("error making request to svr. Err: %v", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Errorf("error closing response body: %v", err)
		}
	}()
	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}

	expected := "{\"message\":\"Hello World\"}"

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}

	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}
