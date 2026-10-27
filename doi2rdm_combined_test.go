package irdmtools

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/caltechlibrary/crossrefapi"
)

// combinedApp returns a Doi2Rdm whose CrossRef query serves a fixture and whose
// DataCite query counts its calls and finds nothing, so the default path (which
// tries CrossRef, then DataCite) runs offline.
func combinedApp(t *testing.T, dataCiteCalls *int) *Doi2Rdm {
	return &Doi2Rdm{
		Cfg:           new(Config),
		queryCrossRef: fixtureQuery(t),
		queryDataCite: func(*Config, string, *Doi2RdmOptions) (map[string]interface{}, error) {
			*dataCiteCalls++
			return nil, NegativeError(fmt.Errorf("datacite: no such DOI"))
		},
	}
}

// TestCombinedBadSymbolsStopsBeforeDataCite checks that the default path, which
// is what irdm_harvester runs, reports a bad symbols file with its own exit
// code and does not go on to DataCite, whose "not found" would hide it.
func TestCombinedBadSymbolsStopsBeforeDataCite(t *testing.T) {
	dir := t.TempDir()
	writeSymbolsFile(t, dir, "invalid.json", `{"superscript": {}}`)
	tests := []struct {
		name    string
		symbols string
		code    int
	}{
		{"missing", "absent.json", int(ExitNoInput)},
		{"invalid", "invalid.json", int(ExitConfig)},
	}
	for _, tt := range tests {
		calls := 0
		app := combinedApp(t, &calls)
		var out, eout bytes.Buffer
		code, err := app.RunDoiToRdmCombined(strings.NewReader(""), &out, &eout, writeOptionsFile(t, dir, tt.symbols), "10.1007/s00717-026-00622-0", "")
		if err == nil || code != tt.code {
			t.Errorf("%s: expected exit code %d and an error, got %d, %v", tt.name, tt.code, code, err)
		}
		if err != nil && !strings.Contains(err.Error(), tt.symbols) {
			t.Errorf("%s: expected the error to name %q, got %q", tt.name, tt.symbols, err)
		}
		if calls != 0 {
			t.Errorf("%s: expected DataCite not to be tried, it was queried %d times", tt.name, calls)
		}
		if out.Len() != 0 {
			t.Errorf("%s: expected nothing on stdout, got %q", tt.name, out.String())
		}
	}
}

// TestCombinedCrossRefSuccess checks that a DOI CrossRef has is written, with
// its warnings, and DataCite is not asked.
func TestCombinedCrossRefSuccess(t *testing.T) {
	calls := 0
	app := combinedApp(t, &calls)
	var out, eout bytes.Buffer
	code, err := app.RunDoiToRdmCombined(strings.NewReader(""), &out, &eout, "", "10.1007/s00717-026-00622-0", "")
	if err != nil || code != EXIT_OK {
		t.Fatalf("expected EXIT_OK, got %d, %v", code, err)
	}
	if calls != 0 {
		t.Errorf("expected DataCite not to be asked, got %d queries", calls)
	}
	if !strings.Contains(out.String(), `"X-chromosomale juvenile Retinoschisis"`) {
		t.Errorf("expected the CrossRef record on stdout, got %q", out.String())
	}
	if !strings.Contains(eout.String(), "multiple-titles") {
		t.Errorf("expected the warning on stderr, got %q", eout.String())
	}
}

// TestCombinedFallsBackToDataCite checks that a DOI CrossRef does not have is
// tried at DataCite, and when neither has it the exit code is the not-found one
// and the message names both.
func TestCombinedFallsBackToDataCite(t *testing.T) {
	calls := 0
	app := combinedApp(t, &calls)
	app.queryCrossRef = func(*Config, string, *Doi2RdmOptions) (*crossrefapi.Works, error) {
		return nil, NegativeError(fmt.Errorf("crossref: no such DOI"))
	}
	var out, eout bytes.Buffer
	code, err := app.RunDoiToRdmCombined(strings.NewReader(""), &out, &eout, "", "10.1/none", "")
	if err == nil || code != int(ExitNegative) {
		t.Errorf("expected exit code 1 and an error, got %d, %v", code, err)
	}
	if calls != 1 {
		t.Errorf("expected DataCite to be tried once, got %d", calls)
	}
	if err != nil && (!strings.Contains(err.Error(), "crossref:") || !strings.Contains(err.Error(), "datacite:")) {
		t.Errorf("expected the error to name both services, got %q", err)
	}
	if out.Len() != 0 {
		t.Errorf("expected nothing on stdout, got %q", out.String())
	}
}
