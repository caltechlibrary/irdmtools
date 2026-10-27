package irdmtools

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caltechlibrary/crossrefapi"
)

// failingWriter is a writer whose writes fail.
type failingWriter struct{}

// Write always fails.
func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

// fixtureServerBody returns a CrossRef response for a DOI without its funders,
// which the crosswalk would look up on api.ror.org, as JSON text.
func fixtureServerBody(t *testing.T, doi string) string {
	t.Helper()
	src, err := JSONMarshalIndent(warningFixture(t, doi), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

// runner is a way into doi2rdm that returns an exit code and an error.
type runner struct {
	name string
	run  func(app *Doi2Rdm, out, eout io.Writer, optionFName, doi, diffFName string) (int, error)
}

var runners = []runner{
	{"crossref", func(app *Doi2Rdm, out, eout io.Writer, o, doi, d string) (int, error) {
		return app.RunCrossRefToRdm(strings.NewReader(""), out, eout, o, doi, d)
	}},
	{"datacite", func(app *Doi2Rdm, out, eout io.Writer, o, doi, d string) (int, error) {
		return app.RunDataCiteToRdm(strings.NewReader(""), out, eout, o, doi, d)
	}},
	{"default", func(app *Doi2Rdm, out, eout io.Writer, o, doi, d string) (int, error) {
		return app.RunDoiToRdmCombined(strings.NewReader(""), out, eout, o, doi, d)
	}},
}

// TestRunExitCodesForFiles checks the exit code for each file problem, through
// every way into doi2rdm: a missing file is 66, a file that is present but
// wrong is 78 (options YAML) or 65 (a -diff file), and nothing is written.
func TestRunExitCodesForFiles(t *testing.T) {
	dir := t.TempDir()
	writeSymbolsFile(t, dir, "bad-options.yaml", "mailto: [unclosed\n")
	writeSymbolsFile(t, dir, "bad-diff.json", "not json")
	writeSymbolsFile(t, dir, "invalid.json", `{"superscript": {}}`)
	tests := []struct {
		name    string
		options string
		diff    string
		code    int
	}{
		{"missing options file", filepath.Join(dir, "absent.yaml"), "", 66},
		{"invalid options YAML", filepath.Join(dir, "bad-options.yaml"), "", 78},
		// writeOptionsFile always writes options.yaml, so each case gets its own directory
		{"missing symbols file", writeOptionsFile(t, t.TempDir(), "absent.json"), "", 66},
		{"invalid symbols file", writeOptionsFile(t, dir, "invalid.json"), "", 78},
		{"missing -diff file", "", filepath.Join(dir, "absent-diff.json"), 66},
		{"invalid -diff file", "", filepath.Join(dir, "bad-diff.json"), 65},
	}
	for _, r := range runners {
		for _, tt := range tests {
			// the DataCite way into doi2rdm does not read a symbols file
			if r.name == "datacite" && strings.Contains(tt.name, "symbols") {
				continue
			}
			calls := 0
			app := combinedApp(t, &calls)
			var out, eout bytes.Buffer
			code, err := r.run(app, &out, &eout, tt.options, "10.1007/s00717-026-00622-0", tt.diff)
			if err == nil || code != tt.code {
				t.Errorf("%s, %s: expected exit code %d and an error, got %d, %v", r.name, tt.name, tt.code, code, err)
			}
			if code != ExitCode(err) {
				t.Errorf("%s, %s: expected the returned code %d to be ExitCode(err) %d", r.name, tt.name, code, ExitCode(err))
			}
			if out.Len() != 0 {
				t.Errorf("%s, %s: expected nothing on stdout, got %q", r.name, tt.name, out.String())
			}
			if calls != 0 {
				t.Errorf("%s, %s: expected no DataCite query, got %d", r.name, tt.name, calls)
			}
		}
	}
}

// TestRunExitCodeForUnconvertibleRecord checks a fetched record that cannot be
// converted is 65, and that the default path reports it at once instead of
// trying DataCite.
func TestRunExitCodeForUnconvertibleRecord(t *testing.T) {
	for _, r := range []runner{runners[0], runners[2]} {
		calls := 0
		app := combinedApp(t, &calls)
		// a journal issue, whose resource type the default options do not map
		var out, eout bytes.Buffer
		code, err := r.run(app, &out, &eout, "", journalIssueDOI, "")
		if err == nil || code != 65 {
			t.Errorf("%s: expected exit code 65 and an error, got %d, %v", r.name, code, err)
		}
		if err != nil && !strings.Contains(err.Error(), "journal-issue") {
			t.Errorf("%s: expected the error to name the resource type, got %q", r.name, err)
		}
		if calls != 0 {
			t.Errorf("%s: expected DataCite not to be tried, got %d queries", r.name, calls)
		}
		if out.Len() != 0 {
			t.Errorf("%s: expected nothing on stdout, got %q", r.name, out.String())
		}
	}
}

// TestRunExitCodeForFailedWrite checks a failed write of the record is 74.
func TestRunExitCodeForFailedWrite(t *testing.T) {
	calls := 0
	app := combinedApp(t, &calls)
	code, err := app.RunCrossRefToRdm(strings.NewReader(""), failingWriter{}, &bytes.Buffer{}, "", "10.1007/s00717-026-00622-0", "")
	if err == nil || code != 74 {
		t.Errorf("expected exit code 74 and an error, got %d, %v", code, err)
	}
}

// TestRunExitCodeFollowsTheQueryError checks the exit code of a run is the
// class of the error the query gave.
func TestRunExitCodeFollowsTheQueryError(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{NegativeError(errors.New("no such DOI")), 1},
		{TempFailError(errors.New("rate limited")), 75},
		{UnavailableError(errors.New("down")), 69},
		{errors.New("unclassified"), 70},
	}
	for _, tt := range tests {
		app := &Doi2Rdm{Cfg: new(Config), queryCrossRef: func(*Config, string, *Doi2RdmOptions) (*crossrefapi.Works, error) {
			return nil, tt.err
		}}
		var out bytes.Buffer
		code, err := app.RunCrossRefToRdm(strings.NewReader(""), &out, &bytes.Buffer{}, "", "10.1/x", "")
		if err == nil || code != tt.code {
			t.Errorf("%v: expected exit code %d, got %d, %v", tt.err, tt.code, code, err)
		}
	}
}

// TestCombinedExitCodes checks the default path's rule with both services
// answering over HTTP: a record from either is 0; "not found" (1) needs both to
// say so; otherwise the class of the first failure that was not "not found",
// CrossRef's before DataCite's.
func TestCombinedExitCodes(t *testing.T) {
	good := fixtureServerBody(t, "10.1007/s00717-026-00622-0")
	tests := []struct {
		name        string
		crossStatus int
		crossBody   string
		dataStatus  int
		code        int
		dataCalls   int32
	}{
		{"found at CrossRef", 200, good, 404, 0, 0},
		{"not found at either", 404, "", 404, 1, 1},
		{"CrossRef down, DataCite not found", 503, "", 404, 69, 1},
		{"CrossRef not found, DataCite down", 404, "", 503, 69, 1},
		{"CrossRef rate limited, DataCite not found", 429, "", 404, 75, 1},
		{"CrossRef not found, DataCite rate limited", 404, "", 429, 75, 1},
		{"both down", 503, "", 500, 69, 1},
		{"CrossRef rate limited and DataCite down", 429, "", 503, 75, 1},
		{"CrossRef down and DataCite rate limited", 503, "", 429, 69, 1},
	}
	for _, tt := range tests {
		statusServer(t, &crossRefAPIURL, tt.crossStatus, tt.crossBody)
		dataCalls := statusServer(t, &dataCiteAPIURL, tt.dataStatus, "")
		app := &Doi2Rdm{Cfg: new(Config)}
		var out, eout bytes.Buffer
		code, err := app.RunDoiToRdmCombined(strings.NewReader(""), &out, &eout, "", "10.1007/s00717-026-00622-0", "")
		if code != tt.code || (tt.code == 0) != (err == nil) {
			t.Errorf("%s: expected exit code %d, got %d (%v)", tt.name, tt.code, code, err)
		}
		if *dataCalls != tt.dataCalls {
			t.Errorf("%s: expected %d DataCite queries, got %d", tt.name, tt.dataCalls, *dataCalls)
		}
		if tt.code != 0 && out.Len() != 0 {
			t.Errorf("%s: expected nothing on stdout, got %q", tt.name, out.String())
		}
		if tt.code != 0 && err != nil && (!strings.Contains(err.Error(), "CrossRef") || !strings.Contains(err.Error(), "DataCite")) {
			t.Errorf("%s: expected the error to name both services, got %q", tt.name, err)
		}
	}
}

// dataCiteObject is the smallest DataCite object the crosswalk accepts: a
// title and a type.
var dataCiteObject = map[string]interface{}{
	"data": map[string]interface{}{
		"id": "10.1/x",
		"attributes": map[string]interface{}{
			"doi":    "10.1/x",
			"titles": []interface{}{map[string]interface{}{"title": "A DataCite title"}},
			"types":  map[string]interface{}{"resourceTypeGeneral": "Text"},
		},
	},
}

// TestCombinedFindsTheDOIAtDataCite checks that a DOI CrossRef has no record
// for, but DataCite has, is written: exit 0, the DataCite record on stdout.
func TestCombinedFindsTheDOIAtDataCite(t *testing.T) {
	app := &Doi2Rdm{
		Cfg: new(Config),
		queryCrossRef: func(*Config, string, *Doi2RdmOptions) (*crossrefapi.Works, error) {
			return nil, NegativeError(errors.New("no such DOI at CrossRef"))
		},
		queryDataCite: func(*Config, string, *Doi2RdmOptions) (map[string]interface{}, error) {
			return dataCiteObject, nil
		},
	}
	var out, eout bytes.Buffer
	code, err := app.RunDoiToRdmCombined(strings.NewReader(""), &out, &eout, "", "10.1/x", "")
	if err != nil || code != 0 {
		t.Fatalf("expected exit code 0, got %d, %v", code, err)
	}
	if !strings.Contains(out.String(), "A DataCite title") {
		t.Errorf("expected the DataCite record on stdout, got %q", out.String())
	}
}

// TestRunDataCiteExitCodes checks the DataCite way into doi2rdm: an empty
// answer is 65, and a record that is found is written with exit 0.
func TestRunDataCiteExitCodes(t *testing.T) {
	empty := &Doi2Rdm{Cfg: new(Config), queryDataCite: func(*Config, string, *Doi2RdmOptions) (map[string]interface{}, error) {
		return map[string]interface{}{}, nil
	}}
	var out bytes.Buffer
	if code, err := empty.RunDataCiteToRdm(strings.NewReader(""), &out, &bytes.Buffer{}, "", "10.1/x", ""); err == nil || code != 65 {
		t.Errorf("expected exit code 65 for an empty answer, got %d, %v", code, err)
	}
	found := &Doi2Rdm{Cfg: new(Config), queryDataCite: func(*Config, string, *Doi2RdmOptions) (map[string]interface{}, error) {
		return dataCiteObject, nil
	}}
	out.Reset()
	if code, err := found.RunDataCiteToRdm(strings.NewReader(""), &out, &bytes.Buffer{}, "", "10.1/x", ""); err != nil || code != 0 {
		t.Errorf("expected exit code 0, got %d, %v", code, err)
	}
	if !strings.Contains(out.String(), "A DataCite title") {
		t.Errorf("expected the record on stdout, got %q", out.String())
	}
	// a failed write is 74 here too
	if code, err := found.RunDataCiteToRdm(strings.NewReader(""), failingWriter{}, &bytes.Buffer{}, "", "10.1/x", ""); err == nil || code != 74 {
		t.Errorf("expected exit code 74 for a failed write, got %d, %v", code, err)
	}
	// a record that cannot be converted is 65
	broken := &Doi2Rdm{Cfg: new(Config), queryDataCite: func(*Config, string, *Doi2RdmOptions) (map[string]interface{}, error) {
		return map[string]interface{}{"data": map[string]interface{}{"attributes": map[string]interface{}{
			"types": map[string]interface{}{"resourceTypeGeneral": "NoSuchType"},
		}}}, nil
	}}
	if code, err := broken.RunDataCiteToRdm(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, "", "10.1/x", ""); err == nil || code != 65 {
		t.Errorf("expected exit code 65 for a record that cannot be converted, got %d, %v", code, err)
	}
}
