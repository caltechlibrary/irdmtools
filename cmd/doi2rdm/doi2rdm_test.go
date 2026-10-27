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
	tests := []struct {
		name string
		args []string
		code int
	}{
		{"help", []string{"-help"}, 0},
		{"version", []string{"-version"}, 0},
		{"show-yaml", []string{"-show-yaml"}, 0},
		{"unknown flag", []string{"-no-such-flag"}, 2},
		{"no DOI", nil, 1},
		{"missing options file", []string{filepath.Join(dir, "absent.yaml"), "10.1/none"}, 2},
		{"missing symbols file", []string{options("missing.yaml", "absent.json"), "10.1/none"}, 66},
		{"invalid symbols file", []string{options("invalid.yaml", "invalid.json"), "10.1/none"}, 78},
		{"missing symbols file, crossref", []string{options("missing2.yaml", "absent.json"), "crossref", "10.1/none"}, 66},
		{"invalid options YAML", []string{write("bad.yaml", "mailto: [unclosed\n"), "10.1/none"}, 8},
	}
	for _, tt := range tests {
		code, out, eout := run(t, bin, tt.args...)
		if code != tt.code {
			t.Errorf("%s: expected exit code %d, got %d (stderr %q)", tt.name, tt.code, code, eout)
		}
		if tt.code != 0 && tt.code != 2 && strings.TrimSpace(out) != "" {
			t.Errorf("%s: expected nothing on stdout, got %q", tt.name, out)
		}
	}
}
