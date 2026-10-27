package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelpTextDocumentsNormalization checks the manual says what doi2rdm does
// to a CrossRef title and abstract, how to change the symbol table, where the
// warnings go, and which exit codes it returns.
func TestHelpTextDocumentsNormalization(t *testing.T) {
	for _, want := range []string{
		"# TEXT NORMALIZATION",
		"plain text",
		"subtitle",
		"additional title",
		"description",
		"symbols:",
		"# WARNINGS",
		"stderr",
		"never change",
		"# EXIT STATUS",
		"66",
		"78",
		"multiple-titles",
		"math-linear",
	} {
		if !strings.Contains(helpText, want) {
			t.Errorf("expected the manual to contain %q", want)
		}
	}
}

// buildDoi2Rdm builds the program into a temporary directory and returns its path.
func buildDoi2Rdm(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "doi2rdm")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %s\n%s", err, out)
	}
	return bin
}

// run runs the program and returns its exit code and what it wrote to stdout
// and stderr.
func run(t *testing.T, bin string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var out, eout bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &eout
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0, out.String(), eout.String()
	case errors.As(err, &exitErr):
		return exitErr.ExitCode(), out.String(), eout.String()
	default:
		t.Fatal(err)
	}
	return -1, "", ""
}

// TestExitStatus checks the exit codes the manual documents, for every case
// that needs no network: the answer is in the number, so a script can tell a
// configuration problem from an unknown DOI.
func TestExitStatus(t *testing.T) {
	bin := buildDoi2Rdm(t)
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	options := func(name, symbols string) string {
		return write(name, "mailto: helpdesk@library.caltech.edu\nsymbols: "+symbols+"\n")
	}
	write("invalid.json", `{"superscript": {}}`)
	ok := write("ok.yaml", "mailto: helpdesk@library.caltech.edu\n")
	tests := []struct {
		name string
		args []string
		code int
	}{
		{"help", []string{"-help"}, 0},
		{"version", []string{"-version"}, 0},
		{"show-yaml", []string{"-show-yaml"}, 0},
		{"unknown flag", []string{"-no-such-flag"}, 2},
		{"no DOI", nil, 2},
		{"missing options file", []string{filepath.Join(dir, "absent.yaml"), "10.1/none"}, 66},
		{"missing options file, crossref", []string{filepath.Join(dir, "absent.yaml"), "crossref", "10.1/none"}, 66},
		{"missing options file, datacite", []string{filepath.Join(dir, "absent.yaml"), "datacite", "10.1/none"}, 66},
		{"invalid options YAML", []string{write("bad.yaml", "mailto: [unclosed\n"), "10.1/none"}, 78},
		{"missing symbols file", []string{options("missing.yaml", "absent.json"), "10.1/none"}, 66},
		{"invalid symbols file", []string{options("invalid.yaml", "invalid.json"), "10.1/none"}, 78},
		{"missing symbols file, crossref", []string{options("missing2.yaml", "absent.json"), "crossref", "10.1/none"}, 66},
		{"missing -diff file", []string{"-diff", filepath.Join(dir, "absent-diff.json"), "10.1/none"}, 66},
		{"missing -diff file, datacite", []string{"-diff", filepath.Join(dir, "absent-diff.json"), ok, "datacite", "10.1/none"}, 66},
		{"invalid -diff file", []string{"-diff", write("bad-diff.json", "not json"), "10.1/none"}, 65},
		{"invalid -diff file, crossref", []string{"-diff", write("bad-diff2.json", "not json"), ok, "crossref", "10.1/none"}, 65},
		{"invalid -diff file, datacite", []string{"-diff", write("bad-diff3.json", "not json"), ok, "datacite", "10.1/none"}, 65},
	}
	for _, tt := range tests {
		code, out, eout := run(t, bin, tt.args...)
		if code != tt.code {
			t.Errorf("%s: expected exit code %d, got %d (stderr %q)", tt.name, tt.code, code, eout)
		}
		if tt.code != 0 && strings.TrimSpace(out) != "" {
			t.Errorf("%s: expected nothing on stdout, got %q", tt.name, out)
		}
	}
}

// TestHelpTextListsEveryExitCode checks the EXIT STATUS section lists each code
// the program returns as its own entry, and the WARNINGS section each warning
// code the normalizer can give.
func TestHelpTextListsEveryExitCode(t *testing.T) {
	section := func(from, to string) string {
		i := strings.Index(helpText, from)
		j := strings.Index(helpText, to)
		if i < 0 || j < i {
			t.Fatalf("expected the manual to have %q before %q", from, to)
		}
		return helpText[i:j]
	}
	status := section("# EXIT STATUS", "\n# OPTIONS\n")
	for _, code := range []string{"0", "1", "2", "65", "66", "69", "70", "74", "75", "78"} {
		if !strings.Contains(status, "\n"+code+"\n: ") {
			t.Errorf("expected EXIT STATUS to list exit code %s as an entry", code)
		}
	}
	warnings := section("# WARNINGS", "# EXIT STATUS")
	for _, code := range []string{"parse-fallback", "script-unmapped", "math-linear", "math-unconverted", "multiple-titles"} {
		if !strings.Contains(warnings, `"`+code+`"`) {
			t.Errorf("expected WARNINGS to describe the %q code", code)
		}
	}
	for _, field := range []string{`"title"`, `"additional title"`, `"description"`} {
		if !strings.Contains(warnings, field) {
			t.Errorf("expected WARNINGS to name the field %s", field)
		}
	}
}

// TestHelpTextDocumentsDebug checks the manual explains -debug as a way to see
// what CrossRef or DataCite returned, not only what doi2rdm does with it.
func TestHelpTextDocumentsDebug(t *testing.T) {
	i := strings.Index(helpText, "\n# OPTIONS\n")
	if i < 0 {
		t.Fatal("expected the manual to have an OPTIONS section")
	}
	options := helpText[i:]
	for _, want := range []string{"-debug\n: ", "works JSON", "objects JSON", "CrossRef", "DataCite", "standard error", "before"} {
		if !strings.Contains(options, want) {
			t.Errorf("expected the -debug entry and its example to contain %q", want)
		}
	}
	if !strings.Contains(helpText, "-debug ") {
		t.Error("expected an example that uses -debug")
	}
}
