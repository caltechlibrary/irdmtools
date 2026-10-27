package irdmtools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/caltechlibrary/crossrefapi"
	"gopkg.in/yaml.v3"
)

// symbolsTitleWork returns a CrossRef work whose title holds the entity
// &eacute;, which the default symbol table converts to "é" and an override
// can redefine.
func symbolsTitleWork() *crossrefapi.Works {
	return &crossrefapi.Works{Message: &crossrefapi.Message{
		Type:  "journal-article",
		Title: []string{"caf&eacute;"},
	}}
}

// writeSymbolsFile writes an override file for the normalizer and returns its path.
func writeSymbolsFile(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// writeOptionsFile writes the default options plus a symbols line.
func writeOptionsFile(t *testing.T, dir, symbols string) string {
	t.Helper()
	src := string(DefaultDoi2RdmOptionsYAML) + "\nsymbols: " + symbols + "\n"
	return writeSymbolsFile(t, dir, "options.yaml", src)
}

const eacuteOverride = `{"entities": {"eacute": "E"}}`

// TestDefaultOptionsHaveNoSymbols checks the default options YAML still parses
// and names no override file.
func TestDefaultOptionsHaveNoSymbols(t *testing.T) {
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(DefaultDoi2RdmOptionsYAML, &options); err != nil {
		t.Fatal(err)
	}
	if options.Symbols != "" {
		t.Errorf("expected no symbols file by default, got %q", options.Symbols)
	}
}

// TestCrosswalkUsesDefaultSymbols checks a record is converted with the default
// table when no override file is named.
func TestCrosswalkUsesDefaultSymbols(t *testing.T) {
	app := &Doi2Rdm{Cfg: new(Config)}
	options, code, err := app.loadCrossRefOptions("")
	if err != nil {
		t.Fatalf("exit code %d: %s", code, err)
	}
	record, err := CrosswalkCrossRefWork(app.Cfg, symbolsTitleWork(), options)
	if err != nil {
		t.Fatal(err)
	}
	if record.Metadata.Title != "café" {
		t.Errorf("expected %q, got %q", "café", record.Metadata.Title)
	}
}

// TestCrosswalkUsesOverrideSymbols checks that an override file named by the
// options changes what the crosswalk produces, with a relative path read
// relative to the options file and an absolute path read as it is.
func TestCrosswalkUsesOverrideSymbols(t *testing.T) {
	dir := t.TempDir()
	abs := writeSymbolsFile(t, dir, "absolute.json", eacuteOverride)
	writeSymbolsFile(t, dir, "relative.json", `{"entities": {"eacute": "R"}}`)
	otherDir := t.TempDir()
	tests := []struct {
		name    string
		options string
		want    string
	}{
		{"relative", writeOptionsFile(t, dir, "relative.json"), "cafR"},
		{"absolute", writeOptionsFile(t, otherDir, abs), "cafE"},
	}
	for _, tt := range tests {
		app := &Doi2Rdm{Cfg: new(Config)}
		options, code, err := app.loadCrossRefOptions(tt.options)
		if err != nil {
			t.Errorf("%s: exit code %d: %s", tt.name, code, err)
			continue
		}
		record, err := CrosswalkCrossRefWork(app.Cfg, symbolsTitleWork(), options)
		if err != nil {
			t.Errorf("%s: %s", tt.name, err)
			continue
		}
		if record.Metadata.Title != tt.want {
			t.Errorf("%s: expected %q, got %q", tt.name, tt.want, record.Metadata.Title)
		}
	}
}

// TestCrosswalkOptionsSymbolsDirect checks the crosswalk reads Symbols from an
// options value that did not come through loadCrossRefOptions, and reports a
// bad file as an error rather than a panic or the default table.
func TestCrosswalkOptionsSymbolsDirect(t *testing.T) {
	dir := t.TempDir()
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(DefaultDoi2RdmOptionsYAML, &options); err != nil {
		t.Fatal(err)
	}
	options.Symbols = writeSymbolsFile(t, dir, "over.json", eacuteOverride)
	record, err := CrosswalkCrossRefWork(new(Config), symbolsTitleWork(), options)
	if err != nil {
		t.Fatal(err)
	}
	if record.Metadata.Title != "cafE" {
		t.Errorf("expected %q, got %q", "cafE", record.Metadata.Title)
	}
	bad := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(DefaultDoi2RdmOptionsYAML, &bad); err != nil {
		t.Fatal(err)
	}
	bad.Symbols = filepath.Join(dir, "absent.json")
	if _, err := CrosswalkCrossRefWork(new(Config), symbolsTitleWork(), bad); err == nil {
		t.Error("expected an error for a missing symbols file")
	}
}

// TestLoadOptionsSymbolsErrors checks that a missing or invalid override file
// stops the run, naming the file, with the workspace exit codes: 66 (no input)
// and 78 (config), which cannot be mistaken for 2, "DOI not found".
func TestLoadOptionsSymbolsErrors(t *testing.T) {
	dir := t.TempDir()
	writeSymbolsFile(t, dir, "invalid.json", `{"superscript": {}}`)
	writeSymbolsFile(t, dir, "notjson.json", `not json`)
	tests := []struct {
		name    string
		symbols string
		code    int
	}{
		{"missing", "absent.json", EX_NOINPUT},
		{"unknown section", "invalid.json", EX_CONFIG},
		{"not JSON", "notjson.json", EX_CONFIG},
	}
	for _, tt := range tests {
		app := &Doi2Rdm{Cfg: new(Config)}
		options, code, err := app.loadCrossRefOptions(writeOptionsFile(t, dir, tt.symbols))
		if err == nil {
			t.Errorf("%s: expected an error", tt.name)
			continue
		}
		if options != nil {
			t.Errorf("%s: expected no options with an error", tt.name)
		}
		if code != tt.code {
			t.Errorf("%s: expected exit code %d, got %d", tt.name, tt.code, code)
		}
		if !strings.Contains(err.Error(), tt.symbols) {
			t.Errorf("%s: expected the error to name %q, got %q", tt.name, tt.symbols, err)
		}
	}
}

// TestNormalizerBuiltOnce checks the options build their normalizer once and
// hand the same one to every record of a run.
func TestNormalizerBuiltOnce(t *testing.T) {
	options := new(Doi2RdmOptions)
	first, err := options.crossRefNormalizer()
	if err != nil {
		t.Fatal(err)
	}
	second, err := options.crossRefNormalizer()
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Error("expected the same normalizer for every call")
	}
}

// TestLoadCrossRefOptionsDefaults checks what loading the options still does
// besides the symbols: the debug flag of the configuration reaches the
// options, and a blank mailto becomes the Library's help desk address.
func TestLoadCrossRefOptionsDefaults(t *testing.T) {
	dir := t.TempDir()
	fName := writeSymbolsFile(t, dir, "options.yaml", "dot_initials: true\n")
	app := &Doi2Rdm{Cfg: &Config{Debug: true}}
	options, code, err := app.loadCrossRefOptions(fName)
	if err != nil {
		t.Fatalf("exit code %d: %s", code, err)
	}
	if !options.Debug {
		t.Error("expected the configuration's debug flag in the options")
	}
	if options.MailTo != "helpdesk@library.caltech.edu" {
		t.Errorf("expected the default mailto, got %q", options.MailTo)
	}
	if !options.DotInitials {
		t.Error("expected dot_initials from the options file")
	}
	if _, code, err := (&Doi2Rdm{Cfg: new(Config)}).loadCrossRefOptions(filepath.Join(dir, "absent.yaml")); err == nil || code != ENOENT {
		t.Errorf("expected ENOENT for a missing options file, got %d, %v", code, err)
	}
	bad := writeSymbolsFile(t, dir, "bad.yaml", "mailto: [unclosed\n")
	if _, code, err := (&Doi2Rdm{Cfg: new(Config)}).loadCrossRefOptions(bad); err == nil || code != ENOEXEC {
		t.Errorf("expected ENOEXEC for an invalid options file, got %d, %v", code, err)
	}
}

// TestRunCrossRefToRdmBadSymbols checks a bad symbols file stops a run before
// anything is fetched or written: no network is needed to see the error.
func TestRunCrossRefToRdmBadSymbols(t *testing.T) {
	dir := t.TempDir()
	var out, eout strings.Builder
	tests := []struct {
		symbols string
		code    int
	}{
		{"absent.json", EX_NOINPUT},
		{writeSymbolsFile(t, dir, "invalid.json", `{"superscript": {}}`), EX_CONFIG},
	}
	for _, tt := range tests {
		app := &Doi2Rdm{Cfg: new(Config)}
		code, err := app.RunCrossRefToRdm(strings.NewReader(""), &out, &eout, writeOptionsFile(t, dir, tt.symbols), "10.1000/none", "")
		if err == nil || code != tt.code {
			t.Errorf("%s: expected exit code %d and an error, got %d, %v", tt.symbols, tt.code, code, err)
		}
	}
	if out.Len() != 0 {
		t.Errorf("expected nothing on stdout, got %q", out.String())
	}
}

// TestWorkspaceExitCodes pins the numbers of the workspace exit codes
// (DR-0014) the symbols option uses: scripts, such as the irdm_harvester,
// read the numbers, not the constants.
func TestWorkspaceExitCodes(t *testing.T) {
	if EX_NOINPUT != 66 {
		t.Errorf("expected EX_NOINPUT to be 66, got %d", EX_NOINPUT)
	}
	if EX_CONFIG != 78 {
		t.Errorf("expected EX_CONFIG to be 78, got %d", EX_CONFIG)
	}
	if ENOENT != 2 {
		t.Errorf("expected ENOENT, the harvester's \"DOI not found\", to stay 2, got %d", ENOENT)
	}
}

// TestCrosswalkDescriptionUsesOverrideSymbols checks the override table
// reaches the description as well as the title.
func TestCrosswalkDescriptionUsesOverrideSymbols(t *testing.T) {
	dir := t.TempDir()
	work := symbolsTitleWork()
	work.Message.Abstract = "<jats:p>caf&eacute;</jats:p>"
	app := &Doi2Rdm{Cfg: new(Config)}
	options, code, err := app.loadCrossRefOptions("")
	if err != nil {
		t.Fatalf("exit code %d: %s", code, err)
	}
	record, err := CrosswalkCrossRefWork(app.Cfg, work, options)
	if err != nil {
		t.Fatal(err)
	}
	if record.Metadata.Description != "<p>café</p>" {
		t.Errorf("expected the default table, got %q", record.Metadata.Description)
	}
	over := new(Doi2RdmOptions)
	*over = *options
	over.normalizer = nil
	over.Symbols = writeSymbolsFile(t, dir, "over.json", eacuteOverride)
	work = symbolsTitleWork()
	work.Message.Abstract = "<jats:p>caf&eacute;</jats:p>"
	record, err = CrosswalkCrossRefWork(app.Cfg, work, over)
	if err != nil {
		t.Fatal(err)
	}
	if record.Metadata.Description != "<p>cafE</p>" {
		t.Errorf("expected the override table, got %q", record.Metadata.Description)
	}
}

// TestGetWorksDescriptionBadSymbols checks the description reports a symbols
// file that cannot be read instead of falling back to the default table. The
// crosswalk reaches the title first, so this is tested directly.
func TestGetWorksDescriptionBadSymbols(t *testing.T) {
	options := new(Doi2RdmOptions)
	options.Symbols = filepath.Join(t.TempDir(), "absent.json")
	work := symbolsTitleWork()
	work.Message.Abstract = "<jats:p>x</jats:p>"
	if got, _, err := getWorksDescription(work, options); err == nil {
		t.Errorf("expected an error for a missing symbols file, got %q", got)
	}
	if title, _, _, err := getWorksTitles(work, options); err == nil {
		t.Errorf("expected an error for a missing symbols file, got %q", title)
	}
}
