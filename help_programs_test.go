package irdmtools

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// helpPrograms are the nine programs and the topics each is expected to have,
// written out here as an independent check on what is cut from the manuals.
var helpPrograms = map[string][]string{
	"rdmutil": {"setup", "get_modified_ids", "get_all_ids", "get_all_stale_ids", "check_doi", "get_record",
		"get_raw_record", "get_record_versions", "get_files", "get_file", "retrieve_file", "get_versions",
		"get_latest_version", "set_version", "set_publication_date", "new_record", "new_version", "publish_version",
		"new_draft", "get_draft", "update_draft", "set_files_enable", "upload_files", "get_draft_files", "delete_files",
		"discard_draft", "review_comment", "send_to_community", "get_review", "review_request", "get_access",
		"set_access", "harvest", "get_endpoint", "post_endpoint", "put_endpoint", "patch_endpoint", "delete_endpoint"},
	"ep3util": {"setup", "get_all_ids", "get_modified_ids", "get_record", "harvest"},
	"doi2rdm": {"options_yaml", "text_normalization", "warnings", "exit_status", "options", "examples"},
	// the single-purpose programs have no topics
	"eprint2rdm":        nil,
	"rdm2eprint":        nil,
	"eprintrest":        nil,
	"ep3ds2citations":   nil,
	"rdmds2citations":   nil,
	"people2vocabulary": nil,
}

var (
	helpBuildMu  sync.Mutex
	helpBinaries = map[string]string{}
	helpBinDir   string
)

// helpBinary builds a program once and returns the path of the binary.
func helpBinary(t *testing.T, name string) string {
	t.Helper()
	helpBuildMu.Lock()
	defer helpBuildMu.Unlock()
	if helpBinDir == "" {
		dir, err := os.MkdirTemp("", "irdmtools-help-")
		if err != nil {
			t.Fatal(err)
		}
		helpBinDir = dir // removed by TestMain
	}
	if bin, ok := helpBinaries[name]; ok {
		return bin
	}
	bin := filepath.Join(helpBinDir, name)
	if out, err := exec.Command("go", "build", "-o", bin, "./cmd/"+name).CombinedOutput(); err != nil {
		t.Fatalf("go build %s: %s\n%s", name, err, out)
	}
	helpBinaries[name] = bin
	return bin
}

// runProgram runs a program and returns its exit code, stdout and stderr.
func runProgram(t *testing.T, name string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(helpBinary(t, name), args...)
	var out, eout bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &eout
	err := cmd.Run()
	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return 0, out.String(), eout.String()
	case errors.As(err, &exitErr):
		return exitErr.ExitCode(), out.String(), eout.String()
	}
	t.Fatal(err)
	return -1, "", ""
}

// indexNames returns the topic names in a program's index: the first field of
// every line that is not a header line.
func indexNames(index string) []string {
	var names []string
	for _, line := range strings.Split(strings.TrimSuffix(index, "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		names = append(names, strings.Fields(line)[0])
	}
	return names
}

// TestHelpFlagStillPrintsTheManual checks -help alone is unchanged: the whole
// manual, exit 0.
func TestHelpFlagStillPrintsTheManual(t *testing.T) {
	for name := range helpPrograms {
		code, out, eout := runProgram(t, name, "-help")
		if code != 0 || eout != "" || !strings.HasPrefix(out, "%"+name+"(1) irdmtools user manual") || !strings.Contains(out, "\n# NAME\n") {
			t.Errorf("%s -help: expected the manual and exit 0, got %d, %q", name, code, out[:min(len(out), 80)])
		}
	}
}

// TestHelpTopicsListing checks -help topics for every program: exit 0, the
// topics the program is expected to have in manual order, and for a program
// with none a header line that says so.
func TestHelpTopicsListing(t *testing.T) {
	for name, want := range helpPrograms {
		code, out, eout := runProgram(t, name, "-help", "topics")
		if code != 0 || eout != "" {
			t.Errorf("%s -help topics: expected exit 0 and nothing on stderr, got %d, %q", name, code, eout)
		}
		if got := indexNames(out); strings.Join(got, " ") != strings.Join(want, " ") {
			t.Errorf("%s: unexpected topics\n got  %v\n want %v", name, got, want)
		}
		if want == nil && !strings.HasPrefix(out, "# "+name+" has no help topics") {
			t.Errorf("%s: expected a header saying there are no topics, got %q", name, out)
		}
		if want != nil && !strings.HasPrefix(out, "# "+name+" help topics.") {
			t.Errorf("%s: expected the index header, got %q", name, out[:min(len(out), 60)])
		}
	}
}

// TestHelpTopicPages checks every topic of the three programs that have them
// prints a page: exit 0, the manual's header lines with the topic in the page
// name and the program's own version and hash, and a synopsis.
func TestHelpTopicPages(t *testing.T) {
	for name, topics := range helpPrograms {
		if topics == nil {
			continue
		}
		_, version, _ := runProgram(t, name, "-version")
		fields := strings.Fields(version) // name version hash
		if len(fields) < 3 {
			t.Fatalf("%s -version: unexpected %q", name, version)
		}
		for _, topic := range topics {
			code, out, eout := runProgram(t, name, "-help", topic)
			if code != 0 || eout != "" {
				t.Errorf("%s -help %s: expected exit 0, got %d, %q", name, topic, code, eout)
				continue
			}
			first := strings.SplitN(out, "\n", 2)[0]
			want := "%" + name + "-" + topic + "(1) irdmtools user manual | version " + fields[1] + " " + fields[2]
			if first != want {
				t.Errorf("%s -help %s: expected the first line %q, got %q", name, topic, want, first)
			}
			for _, section := range []string{"# NAME", "# SYNOPSIS", "# DESCRIPTION", "# SEE ALSO"} {
				if !strings.Contains(out, "\n"+section+"\n") {
					t.Errorf("%s -help %s: expected %s", name, topic, section)
				}
			}
			if name != "doi2rdm" && !strings.Contains(out, "# SYNOPSIS\n\n"+name+" "+topic) {
				t.Errorf("%s -help %s: expected the synopsis to start with the action", name, topic)
			}
		}
	}
}

// TestHelpUsageErrors checks an unknown topic and a surplus word are usage
// errors for every program: exit 2, nothing on standard output, a message that
// for a program with topics lists them.
func TestHelpUsageErrors(t *testing.T) {
	for name, topics := range helpPrograms {
		for _, args := range [][]string{{"-help", "nosuchtopic"}, {"-help", "topics", "extra"}} {
			code, out, eout := runProgram(t, name, args...)
			if code != 2 || out != "" || eout == "" {
				t.Errorf("%s %v: expected exit 2, nothing on stdout and a message, got %d, %q, %q", name, args, code, out, eout)
			}
			if topics != nil && args[1] == "nosuchtopic" && !strings.Contains(eout, topics[0]) {
				t.Errorf("%s %v: expected the message to list the topics, got %q", name, args, eout)
			}
			if topics == nil && args[1] == "nosuchtopic" && !strings.Contains(eout, "no help topics") {
				t.Errorf("%s %v: expected the message to say there are no topics, got %q", name, args, eout)
			}
		}
	}
}

// TestHelpDoi2RdmExitStatusTopic checks the exit_status topic of doi2rdm lists
// every exit code the program returns.
func TestHelpDoi2RdmExitStatusTopic(t *testing.T) {
	code, out, _ := runProgram(t, "doi2rdm", "-help", "exit_status")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	for _, c := range []string{"0", "1", "2", "65", "66", "69", "70", "74", "75", "78"} {
		if !strings.Contains(out, "\n"+c+"\n: ") {
			t.Errorf("expected the exit_status topic to list exit code %s", c)
		}
	}
}

// TestManualsNameHelpTopics checks the manual of each program that has topics
// says how to list them.
func TestManualsNameHelpTopics(t *testing.T) {
	for name, topics := range helpPrograms {
		if topics == nil {
			continue
		}
		_, out, _ := runProgram(t, name, "-help")
		if !strings.Contains(out, "-help topics") || !strings.Contains(out, "-help TOPIC") {
			t.Errorf("%s: expected the manual to name -help topics and -help TOPIC", name)
		}
	}
}
