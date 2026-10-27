package irdmtools

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

// statusServer starts a test server that answers every request with a status
// and a body, counts the requests, and points an API base URL at it until the
// test ends.
func statusServer(t *testing.T, apiURL *string, status int, body string) *int32 {
	t.Helper()
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	old := *apiURL
	*apiURL = server.URL
	t.Cleanup(func() {
		*apiURL = old
		server.Close()
	})
	return &calls
}

// closedServer points an API base URL at an address nothing listens on.
func closedServer(t *testing.T, apiURL *string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()
	old := *apiURL
	*apiURL = url
	t.Cleanup(func() { *apiURL = old })
}

// statusOptions returns options with a mailto that only the test server sees.
func statusOptions() *Doi2RdmOptions {
	return &Doi2RdmOptions{MailTo: "test@example.edu"}
}

// fixtureBody returns the captured CrossRef response for a DOI as text.
func fixtureBody(t *testing.T, doi string) string {
	t.Helper()
	src, err := os.ReadFile("testdata/crossref/crossref-" + strings.ReplaceAll(doi, "/", "_") + ".json")
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

const dataCiteBody = `{"data":{"id":"10.1/x","type":"dois","attributes":{"doi":"10.1/x"}}}`

// TestQueryCrossRefWorkStatus checks the HTTP status CrossRef answers with
// becomes the exit class a caller can act on, and that a good answer is a record.
func TestQueryCrossRefWorkStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		code   int
	}{
		{"found", 200, fixtureBody(t, "10.1007/s00717-026-00622-0"), 0},
		{"not found", 404, "Resource not found.", 1},
		{"rate limited", 429, "slow down", 75},
		{"server error", 500, "oops", 69},
		{"unavailable", 503, "later", 69},
		{"forbidden", 403, "no", 69},
		{"empty answer", 200, "", 65},
		{"not JSON", 200, "<html>maintenance</html>", 65},
	}
	for _, tt := range tests {
		statusServer(t, &crossRefAPIURL, tt.status, tt.body)
		work, err := QueryCrossRefWork(new(Config), "10.1007/s00717-026-00622-0", statusOptions())
		if got := ExitCode(err); got != tt.code {
			t.Errorf("%s: expected exit code %d, got %d (%v)", tt.name, tt.code, got, err)
		}
		if tt.code == 0 && (work == nil || work.Message == nil) {
			t.Errorf("%s: expected a record", tt.name)
		}
		if tt.code != 0 && work != nil {
			t.Errorf("%s: expected no record with an error", tt.name)
		}
		if err != nil && !strings.Contains(err.Error(), "10.1007/s00717-026-00622-0") {
			t.Errorf("%s: expected the error to name the DOI, got %q", tt.name, err)
		}
		if err != nil && !strings.Contains(err.Error(), "CrossRef") {
			t.Errorf("%s: expected the error to name the service, got %q", tt.name, err)
		}
		if err != nil && tt.status != 200 && !strings.Contains(err.Error(), fmt.Sprint(tt.status)) {
			t.Errorf("%s: expected the error to give the HTTP status %d, got %q", tt.name, tt.status, err)
		}
		if tt.name == "unavailable" && (err == nil || !strings.Contains(err.Error(), "Service Unavailable")) {
			t.Errorf("%s: expected the status text in the error, got %v", tt.name, err)
		}
	}
	closedServer(t, &crossRefAPIURL)
	if _, err := QueryCrossRefWork(new(Config), "10.1/x", statusOptions()); ExitCode(err) != 69 {
		t.Errorf("expected exit code 69 when CrossRef cannot be reached, got %d (%v)", ExitCode(err), err)
	}
}

// TestQueryDataCiteObjectStatus is the same for DataCite.
func TestQueryDataCiteObjectStatus(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		code   int
	}{
		{"found", 200, dataCiteBody, 0},
		{"not found", 404, `{"errors":[{"status":"404"}]}`, 1},
		{"rate limited", 429, "slow down", 75},
		{"server error", 500, "oops", 69},
		{"unavailable", 503, "later", 69},
		{"empty answer", 200, "", 65},
		{"not JSON", 200, "<html>maintenance</html>", 65},
	}
	for _, tt := range tests {
		statusServer(t, &dataCiteAPIURL, tt.status, tt.body)
		object, err := QueryDataCiteObject(new(Config), "10.1/x", statusOptions())
		if got := ExitCode(err); got != tt.code {
			t.Errorf("%s: expected exit code %d, got %d (%v)", tt.name, tt.code, got, err)
		}
		if tt.code == 0 && len(object) == 0 {
			t.Errorf("%s: expected an object", tt.name)
		}
		if err != nil && (!strings.Contains(err.Error(), "10.1/x") || !strings.Contains(err.Error(), "DataCite")) {
			t.Errorf("%s: expected the error to name the DOI and the service, got %q", tt.name, err)
		}
	}
	closedServer(t, &dataCiteAPIURL)
	if _, err := QueryDataCiteObject(new(Config), "10.1/x", statusOptions()); ExitCode(err) != 69 {
		t.Errorf("expected exit code 69 when DataCite cannot be reached, got %d (%v)", ExitCode(err), err)
	}
}

// TestQueryWithoutMailto checks a missing mailto is a configuration error,
// reported before any request is made.
func TestQueryWithoutMailto(t *testing.T) {
	calls := statusServer(t, &crossRefAPIURL, 200, "{}")
	dataCalls := statusServer(t, &dataCiteAPIURL, 200, "{}")
	options := &Doi2RdmOptions{}
	if _, err := QueryCrossRefWork(new(Config), "10.1/x", options); ExitCode(err) != 78 {
		t.Errorf("expected exit code 78 without a mailto for CrossRef, got %d (%v)", ExitCode(err), err)
	}
	if _, err := QueryDataCiteObject(new(Config), "10.1/x", options); ExitCode(err) != 78 {
		t.Errorf("expected exit code 78 without a mailto for DataCite, got %d (%v)", ExitCode(err), err)
	}
	if *calls != 0 || *dataCalls != 0 {
		t.Errorf("expected no request to be made, got %d and %d", *calls, *dataCalls)
	}
}
