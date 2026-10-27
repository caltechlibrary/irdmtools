package irdmtools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/caltechlibrary/crossrefapi"
	"github.com/caltechlibrary/simplified"
	"gopkg.in/yaml.v3"
)

// recordTimesRE matches the created and updated times the crosswalk stamps on
// a record with the current time, so two crosswalks of one work never match
// byte for byte.
var recordTimesRE = regexp.MustCompile(`"(created|updated)": "[^"]*"`)

// withoutTimes returns record JSON with the created and updated times removed.
func withoutTimes(src string) string {
	return recordTimesRE.ReplaceAllString(src, `"$1": ""`)
}

// warningOptions returns the default doi2rdm options.
func warningOptions(t *testing.T) *Doi2RdmOptions {
	t.Helper()
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(DefaultDoi2RdmOptionsYAML, &options); err != nil {
		t.Fatal(err)
	}
	return options
}

// warningFixture loads a fixture without its funders, which the crosswalk
// would look up on api.ror.org (see crosswalkWork).
func warningFixture(t *testing.T, doi string) *crossrefapi.Works {
	t.Helper()
	work := loadCrossRefFixture(t, doi)
	work.Message.Funder = nil
	return work
}

// TestCrosswalkWorkWarnings checks that the crosswalk reports what the
// normalizer could not convert cleanly, naming the field it came from.
func TestCrosswalkWorkWarnings(t *testing.T) {
	// a title with a second title that differs: one multiple-titles warning on the title
	record, warnings, err := CrosswalkCrossRefWorkWarnings(new(Config), warningFixture(t, "10.1007/s00717-026-00622-0"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	if record == nil || record.Metadata.Title != "X-chromosomale juvenile Retinoschisis" {
		t.Errorf("expected the record to be crosswalked, got %v", record)
	}
	var multiple []CrossRefWarning
	for _, w := range warnings {
		if w.Warning.Code == crossrefapi.WarnMultipleTitles {
			multiple = append(multiple, w)
		}
	}
	if len(multiple) != 1 || multiple[0].Field != "title" || multiple[0].Warning.Fragment != "X-linked juvenile retinoschisis" {
		t.Errorf("expected one multiple-titles warning on the title, got %+v", warnings)
	}
	// a formula with no TeX form warns on the title
	_, warnings, err = CrosswalkCrossRefWorkWarnings(new(Config), warningFixture(t, "10.1103/physrevlett.134.141902"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range warnings {
		if w.Field == "title" && w.Warning.Code == crossrefapi.WarnMathLinear && w.Warning.Fragment != "" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a math-linear warning on the title, got %+v", warnings)
	}
	// a clean record has none
	_, warnings, err = CrosswalkCrossRefWorkWarnings(new(Config), warningFixture(t, "10.1002/2016GL070081"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for a clean record, got %+v", warnings)
	}
}

// TestCrosswalkWorkWarningFields checks each field a warning can come from.
func TestCrosswalkWorkWarningFields(t *testing.T) {
	work := &crossrefapi.Works{Message: &crossrefapi.Message{
		Type:     "journal-article",
		Title:    []string{"T<sub>y</sub>", "Other"},
		Abstract: "when a < b <jats:p>and</jats:p> c",
	}}
	_, warnings, err := CrosswalkCrossRefWorkWarnings(new(Config), work, warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"title":       crossrefapi.WarnScriptUnmapped,
		"description": crossrefapi.WarnParseFallback,
	}
	got := map[string]string{}
	for _, w := range warnings {
		if w.Warning.Code == crossrefapi.WarnMultipleTitles {
			continue
		}
		got[w.Field] = w.Warning.Code
	}
	for field, code := range want {
		if got[field] != code {
			t.Errorf("expected a %s warning on the %s, got %+v", code, field, warnings)
		}
	}
	// an alternate title that needs conversion warns as an additional title
	work = &crossrefapi.Works{Message: &crossrefapi.Message{
		Type:  "journal-article",
		Title: []string{"First", "Second T<sub>y</sub>"},
	}}
	_, warnings, err = CrosswalkCrossRefWorkWarnings(new(Config), work, warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range warnings {
		if w.Field == "additional title" && w.Warning.Code == crossrefapi.WarnScriptUnmapped {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a script-unmapped warning on an additional title, got %+v", warnings)
	}
}

// TestCrosswalkCrossRefWorkDropsWarnings checks the older function keeps its
// result and reports no warnings.
func TestCrosswalkCrossRefWorkDropsWarnings(t *testing.T) {
	a, err := CrosswalkCrossRefWork(new(Config), warningFixture(t, "10.1007/s00717-026-00622-0"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	b, _, err := CrosswalkCrossRefWorkWarnings(new(Config), warningFixture(t, "10.1007/s00717-026-00622-0"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	srcA, _ := JSONMarshalIndent(a, "", "  ")
	srcB, _ := JSONMarshalIndent(b, "", "  ")
	if withoutTimes(string(srcA)) != withoutTimes(string(srcB)) {
		t.Error("expected the same record from both functions")
	}
}

// TestCrossRefWarningString checks the line a warning becomes: one line, with
// the DOI, the field, the code and the fragment, whatever the fragment holds.
func TestCrossRefWarningString(t *testing.T) {
	w := CrossRefWarning{Field: "title", Warning: crossrefapi.Warning{Code: "math-linear", Fragment: "a\nb\r\nc\rd"}}
	got := w.Line("10.1/x")
	want := "doi2rdm: 10.1/x: title: math-linear: a b c d"
	if got != want || strings.ContainsAny(got, "\r\n") {
		t.Errorf("expected one line like %q, got %q", want, got)
	}
	if plain := (CrossRefWarning{Field: "description", Warning: crossrefapi.Warning{Code: "parse-fallback", Fragment: "x < y"}}).Line("10.2/y"); plain != "doi2rdm: 10.2/y: description: parse-fallback: x < y" {
		t.Errorf("unexpected line %q", plain)
	}
}

// renderFixture runs the part of RunCrossRefToRdm after the live query on a
// fixture and returns what it wrote and its result.
func renderFixture(t *testing.T, app *Doi2Rdm, doi string, old *simplified.Record, diff bool) (string, string, int, error) {
	t.Helper()
	var out, eout bytes.Buffer
	code, err := app.renderCrossRefWork(&out, &eout, doi, warningFixture(t, doi), warningOptions(t), old, diff)
	return out.String(), eout.String(), code, err
}

// TestRenderCrossRefWork checks stdout is the record JSON and nothing else,
// the warnings go to stderr one per line, and the result is EXIT_OK either way.
func TestRenderCrossRefWork(t *testing.T) {
	app := &Doi2Rdm{Cfg: new(Config)}
	tests := []struct {
		doi    string
		title  string
		stderr []string
	}{
		{"10.1007/s00717-026-00622-0", "X-chromosomale juvenile Retinoschisis",
			[]string{"doi2rdm: 10.1007/s00717-026-00622-0: title: multiple-titles: X-linked juvenile retinoschisis"}},
		{"10.1002/2016GL070081", "Evidence for universal earthquake rupture initiation behavior", nil},
	}
	for _, tt := range tests {
		stdout, stderr, code, err := renderFixture(t, app, tt.doi, nil, false)
		if err != nil || code != EXIT_OK {
			t.Errorf("%s: expected EXIT_OK, got %d, %v", tt.doi, code, err)
			continue
		}
		record := new(simplified.Record)
		if err := json.Unmarshal([]byte(stdout), record); err != nil {
			t.Errorf("%s: stdout is not the record JSON: %s", tt.doi, err)
			continue
		}
		if record.Metadata.Title != tt.title {
			t.Errorf("%s: expected title %q, got %q", tt.doi, tt.title, record.Metadata.Title)
		}
		if strings.Contains(stdout, "doi2rdm:") {
			t.Errorf("%s: a warning reached stdout", tt.doi)
		}
		var lines []string
		if stderr != "" {
			lines = strings.Split(strings.TrimSuffix(stderr, "\n"), "\n")
		}
		if strings.Join(lines, "\n") != strings.Join(tt.stderr, "\n") {
			t.Errorf("%s: stderr\n got  %q\n want %q", tt.doi, lines, tt.stderr)
		}
	}
	// the formula warning of the Z_N record is on its own line
	_, stderr, code, err := renderFixture(t, app, "10.1103/physrevlett.134.141902", nil, false)
	if err != nil || code != EXIT_OK {
		t.Fatalf("expected EXIT_OK, got %d, %v", code, err)
	}
	if !strings.Contains(stderr, "doi2rdm: 10.1103/physrevlett.134.141902: title: math-linear: ") {
		t.Errorf("expected a math-linear line for the title, got %q", stderr)
	}
	// a missing stderr writer is tolerated
	if code, err := app.renderCrossRefWork(&bytes.Buffer{}, nil, "10.1007/s00717-026-00622-0", warningFixture(t, "10.1007/s00717-026-00622-0"), warningOptions(t), nil, false); err != nil || code != EXIT_OK {
		t.Errorf("expected EXIT_OK with no stderr writer, got %d, %v", code, err)
	}
}

// TestRenderCrossRefWorkDiff checks the diff output is what DiffAsJSON gives for
// the two records, as before, and that only the new work's warnings are written.
func TestRenderCrossRefWorkDiff(t *testing.T) {
	app := &Doi2Rdm{Cfg: new(Config)}
	// the old work has a warning, the new one has none
	old, err := CrosswalkCrossRefWork(app.Cfg, warningFixture(t, "10.1007/s00717-026-00622-0"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	nRecord, err := CrosswalkCrossRefWork(app.Cfg, warningFixture(t, "10.1002/2016GL070081"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	want, err := old.DiffAsJSON(nRecord)
	if err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code, err := renderFixture(t, app, "10.1002/2016GL070081", old, true)
	if err != nil || code != EXIT_OK {
		t.Fatalf("expected EXIT_OK, got %d, %v", code, err)
	}
	if withoutTimes(stdout) != withoutTimes(string(want))+"\n" {
		t.Errorf("expected the diff output as before\n got  %q\n want %q", stdout, string(want)+"\n")
	}
	if stderr != "" {
		t.Errorf("expected only the new work's warnings, got %q", stderr)
	}
	// the new work's warnings are written in diff mode too
	_, stderr, _, err = renderFixture(t, app, "10.1007/s00717-026-00622-0", nRecord, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr, "multiple-titles") {
		t.Errorf("expected the new work's warning in diff mode, got %q", stderr)
	}
}

// TestRenderCrossRefWorkErrors checks a work that cannot be crosswalked writes
// nothing and returns 65, and that the crosswalk functions refuse a nil work.
func TestRenderCrossRefWorkErrors(t *testing.T) {
	app := &Doi2Rdm{Cfg: new(Config)}
	var out, eout bytes.Buffer
	code, err := app.renderCrossRefWork(&out, &eout, "10.1/x", nil, warningOptions(t), nil, false)
	if err == nil || code != int(ExitData) {
		t.Errorf("expected exit code 65 and an error, got %d, %v", code, err)
	}
	if out.Len() != 0 || eout.Len() != 0 {
		t.Errorf("expected nothing written, got %q and %q", out.String(), eout.String())
	}
	if _, err := CrosswalkCrossRefWork(new(Config), nil, warningOptions(t)); err == nil {
		t.Error("expected an error for a nil work")
	}
	if _, _, err := CrosswalkCrossRefWorkWarnings(new(Config), nil, warningOptions(t)); err == nil {
		t.Error("expected an error for a nil work")
	}
}

// fixtureQuery returns a query function that serves a fixture instead of
// calling CrossRef, for testing RunCrossRefToRdm offline.
func fixtureQuery(t *testing.T) func(*Config, string, *Doi2RdmOptions) (*crossrefapi.Works, error) {
	return func(cfg *Config, doi string, options *Doi2RdmOptions) (*crossrefapi.Works, error) {
		return warningFixture(t, doi), nil
	}
}

// TestRunCrossRefToRdmWritesWarnings checks the whole run, with the live query
// replaced by a fixture: the record JSON on stdout, one warning line per
// warning on stderr, and EXIT_OK; and in diff mode the diff on stdout and only
// the new work's warnings on stderr.
func TestRunCrossRefToRdmWritesWarnings(t *testing.T) {
	app := &Doi2Rdm{Cfg: new(Config), queryCrossRef: fixtureQuery(t)}
	var out, eout bytes.Buffer
	code, err := app.RunCrossRefToRdm(strings.NewReader(""), &out, &eout, "", "10.1007/s00717-026-00622-0", "")
	if err != nil || code != EXIT_OK {
		t.Fatalf("expected EXIT_OK, got %d, %v", code, err)
	}
	record := new(simplified.Record)
	if err := json.Unmarshal(out.Bytes(), record); err != nil {
		t.Fatalf("stdout is not the record JSON: %s", err)
	}
	if record.Metadata.Title != "X-chromosomale juvenile Retinoschisis" {
		t.Errorf("unexpected title %q", record.Metadata.Title)
	}
	if want := "doi2rdm: 10.1007/s00717-026-00622-0: title: multiple-titles: X-linked juvenile retinoschisis\n"; eout.String() != want {
		t.Errorf("expected stderr %q, got %q", want, eout.String())
	}
	// diff mode: the old work is a file of CrossRef JSON, the new one is queried
	old, err := CrosswalkCrossRefWork(app.Cfg, warningFixture(t, "10.1007/s00717-026-00622-0"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	nRecord, err := CrosswalkCrossRefWork(app.Cfg, warningFixture(t, "10.1002/2016GL070081"), warningOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	want, err := old.DiffAsJSON(nRecord)
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	eout.Reset()
	// the old work is read from a file, so write it without its funders too
	oldSrc, err := JSONMarshalIndent(warningFixture(t, "10.1007/s00717-026-00622-0"), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	diffFName := filepath.Join(t.TempDir(), "old.json")
	if err := os.WriteFile(diffFName, oldSrc, 0o644); err != nil {
		t.Fatal(err)
	}
	code, err = app.RunCrossRefToRdm(strings.NewReader(""), &out, &eout, "", "10.1002/2016GL070081", diffFName)
	if err != nil || code != EXIT_OK {
		t.Fatalf("expected EXIT_OK in diff mode, got %d, %v", code, err)
	}
	if withoutTimes(out.String()) != withoutTimes(string(want))+"\n" {
		t.Errorf("expected the diff on stdout\n got  %q\n want %q", out.String(), string(want)+"\n")
	}
	if eout.Len() != 0 {
		t.Errorf("expected no warnings from the new work, got %q", eout.String())
	}
	// a failed query has the exit class of its error ("not found" is 1), and writes nothing
	failing := &Doi2Rdm{Cfg: new(Config), queryCrossRef: func(*Config, string, *Doi2RdmOptions) (*crossrefapi.Works, error) {
		return nil, NegativeError(fmt.Errorf("no such work"))
	}}
	out.Reset()
	eout.Reset()
	if code, err := failing.RunCrossRefToRdm(strings.NewReader(""), &out, &eout, "", "10.1/none", ""); err == nil || code != int(ExitNegative) {
		t.Errorf("expected exit code 1 and an error, got %d, %v", code, err)
	}
	if out.Len() != 0 || eout.Len() != 0 {
		t.Errorf("expected nothing written, got %q and %q", out.String(), eout.String())
	}
}
