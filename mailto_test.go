package irdmtools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestMailToFromEnv checks the live tests' mailto comes from TEST_MAILTO and
// that an unset, empty or blank value means there is none.
func TestMailToFromEnv(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  string
		ok    bool
	}{
		{"unset", "", false, "", false},
		{"empty", "", true, "", false},
		{"blank", "  \t ", true, "", false},
		{"set", "someone@example.edu", true, "someone@example.edu", true},
		{"trimmed", "  someone@example.edu\n", true, "someone@example.edu", true},
	}
	for _, tt := range tests {
		if tt.set {
			t.Setenv("TEST_MAILTO", tt.value)
		} else {
			// t.Setenv restores the value at the end of the test; unset it for this case
			t.Setenv("TEST_MAILTO", "")
			os.Unsetenv("TEST_MAILTO")
		}
		got, ok := mailToFromEnv()
		if got != tt.want || ok != tt.ok {
			t.Errorf("%s: expected %q, %t, got %q, %t", tt.name, tt.want, tt.ok, got, ok)
		}
	}
}

// fakeSkipper records what liveMailTo does with the test.
type fakeSkipper struct{ skipped string }

// Helper does nothing.
func (f *fakeSkipper) Helper() {}

// Skip records the skip message.
func (f *fakeSkipper) Skip(args ...interface{}) { f.skipped = fmt.Sprint(args...) }

// TestLiveMailTo checks a live test gets the mailto when TEST_MAILTO is set, and
// is skipped with a message saying how to set it when it is not.
func TestLiveMailTo(t *testing.T) {
	t.Setenv("TEST_MAILTO", "someone@example.edu")
	f := new(fakeSkipper)
	if got := liveMailTo(f); got != "someone@example.edu" || f.skipped != "" {
		t.Errorf("expected the mailto and no skip, got %q, skipped %q", got, f.skipped)
	}
	t.Setenv("TEST_MAILTO", "")
	f = new(fakeSkipper)
	liveMailTo(f)
	if !strings.Contains(f.skipped, "TEST_MAILTO") || !strings.Contains(f.skipped, "make test-live") {
		t.Errorf("expected a skip that says how to set TEST_MAILTO, got %q", f.skipped)
	}
}

// addressRE matches something that looks like an email address.
var addressRE = regexp.MustCompile(`[A-Za-z0-9._+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)+`)

// addressProblems returns what is wrong in the text of a test file under
// irdmtools DR-0004: an email address of a person or a group written in it, or
// a mailto assigned a literal. What is allowed is the product's own default (the
// help desk address doi2rdm falls back to), an example.edu address that only a
// local test server sees, and the @author lines of a file header.
func addressProblems(name, src string) []string {
	var problems []string
	for i, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "@author") {
			continue
		}
		if strings.Contains(line, "options.MailTo = \"") {
			problems = append(problems, fmt.Sprintf("%s:%d: a mailto is assigned a literal; a live test uses liveMailTo(t)", name, i+1))
		}
		for _, address := range addressRE.FindAllString(line, -1) {
			if strings.HasSuffix(address, "@example.edu") || address == "helpdesk@library.caltech.edu" {
				continue
			}
			problems = append(problems, fmt.Sprintf("%s:%d: %q is written in a test; live tests read TEST_MAILTO", name, i+1, address))
		}
	}
	return problems
}

// TestAddressProblems checks what addressProblems accepts and refuses.
func TestAddressProblems(t *testing.T) {
	ok := []string{
		"// @author R. S. Doiel, <someone@caltech.edu>",
		`mailto: helpdesk@library.caltech.edu`,
		`options := &Doi2RdmOptions{MailTo: "test@example.edu"}`,
		`options.MailTo = liveMailTo(t)`,
		`x := "no address here"`,
	}
	for _, line := range ok {
		if got := addressProblems("x_test.go", line); len(got) != 0 {
			t.Errorf("expected %q to be accepted, got %v", line, got)
		}
	}
	bad := []string{
		`options.MailTo = "someone@caltech.edu"`,
		`options.MailTo = "test@example.edu"`,
		`m := "dld-test@library.caltech.edu"`,
		`m := "person@example.org"`,
		`m := "someone@caltech.edu"`,
		`m := "a@b.co" // and another: c@d.edu`,
	}
	for _, line := range bad {
		if got := addressProblems("x_test.go", line); len(got) == 0 {
			t.Errorf("expected %q to be refused", line)
		}
	}
	if got := addressProblems("x_test.go", "a\nb\nm := \"z@y.edu\""); len(got) != 1 || !strings.HasPrefix(got[0], "x_test.go:3:") {
		t.Errorf("expected one problem on line 3, got %v", got)
	}
	if got := addressProblems("x_test.go", "m := \"a@b.co\" // c@d.edu"); len(got) != 2 {
		t.Errorf("expected both addresses on a line to be reported, got %v", got)
	}
}

// TestNoAddressInTests checks no test file in the repository has an address
// problem (irdmtools DR-0004).
func TestNoAddressInTests(t *testing.T) {
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	more, _ := filepath.Glob("cmd/*/*_test.go")
	files = append(files, more...)
	if len(files) < 10 {
		t.Fatalf("expected to find the test files, found %d", len(files))
	}
	scanned := strings.Join(files, " ")
	for _, want := range []string{"doi2rdm_test.go", "cmd/doi2rdm/doi2rdm_test.go"} {
		if !strings.Contains(scanned, want) {
			t.Errorf("expected %s to be scanned", want)
		}
	}
	for _, fName := range files {
		if fName == "mailto_test.go" {
			continue
		}
		src, err := os.ReadFile(fName)
		if err != nil {
			t.Fatal(err)
		}
		for _, problem := range addressProblems(fName, string(src)) {
			t.Error(problem)
		}
	}
}
