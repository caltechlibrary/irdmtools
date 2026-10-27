package main

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// manualActions returns the action names the manual documents: each entry of
// the ACTION section is a term line (the action name, then its parameters)
// followed by a line that begins with ": ".
func manualActions(manual string) []string {
	var names []string
	in := false
	lines := strings.Split(manual, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			in = line == "# ACTION" || line == "# ACTIONS"
			continue
		}
		if in && line != "" && !strings.HasPrefix(line, ":") && !strings.HasPrefix(line, " ") &&
			i+1 < len(lines) && strings.HasPrefix(lines[i+1], ": ") {
			names = append(names, strings.Fields(line)[0])
		}
	}
	return names
}

// dispatchedActions returns the action names the program's Run method accepts,
// read from the case labels of its source.
func dispatchedActions(t *testing.T) []string {
	t.Helper()
	src, err := os.ReadFile("../../rdmutil.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	start := strings.Index(text, "func (app *RdmUtil) Run(")
	if start < 0 {
		t.Fatal("Run not found")
	}
	body := text[start:]
	if end := strings.Index(body[1:], "\nfunc "); end >= 0 {
		body = body[:end+1]
	}
	quoted := regexp.MustCompile("\"([a-z_0-9]+)\"")
	var names []string
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "case \"") {
			for _, m := range quoted.FindAllStringSubmatch(line, -1) {
				names = append(names, m[1])
			}
		}
	}
	return names
}

// TestManualActionsAreUnique checks no action is documented twice: the help
// topics are cut from the manual, one per action name.
func TestManualActionsAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range manualActions(helpText) {
		if seen[name] {
			t.Errorf("the manual documents the action %q more than once", name)
		}
		seen[name] = true
	}
	if len(seen) == 0 {
		t.Error("expected the manual to document actions")
	}
}

// TestManualMatchesDispatcher checks the manual documents exactly the actions
// the program runs: a documented action that does not exist, or an existing one
// the manual does not mention, is a documentation bug.
func TestManualMatchesDispatcher(t *testing.T) {
	documented := map[string]bool{}
	for _, name := range manualActions(helpText) {
		documented[name] = true
	}
	dispatched := map[string]bool{}
	for _, name := range dispatchedActions(t) {
		dispatched[name] = true
	}
	var missing, stale []string
	for name := range dispatched {
		if !documented[name] {
			missing = append(missing, name)
		}
	}
	for name := range documented {
		if !dispatched[name] {
			stale = append(stale, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(stale)
	if len(missing) > 0 {
		t.Errorf("actions the program runs but the manual does not document: %v", missing)
	}
	if len(stale) > 0 {
		t.Errorf("actions the manual documents but the program does not run: %v", stale)
	}
}
