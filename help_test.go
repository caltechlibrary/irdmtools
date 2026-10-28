package irdmtools

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

// helpTestManual is a small manual in the shape of the programs' own: three
// header lines, sections, an ACTION section of definition-list entries (one
// with a second paragraph), and sections that are topics by name.
const helpTestManual = `%app(1) irdmtools user manual | version 1.2.3 abc1234
% A. Author
% 2026-10-27

# NAME

app

# SYNOPSIS

app [OPTIONS] ACTION

# ACTION

__app__ supports the following actions.

setup
: Display an example. More text follows.

get_record RECORD_ID
: Returns a record. The RECORD_ID is
required.

A second paragraph of get_record.

new_draft RECORD_ID [OPTION]
: Creates a draft.

# EXIT STATUS

Standard output is the record.

0
: ok

1
: not found

# OPTIONS

-help
: display help

-version
: display version
`

// topicNames returns the names of topics in order.
func topicNames(topics []HelpTopic) []string {
	var names []string
	for _, t := range topics {
		names = append(names, t.Name)
	}
	return names
}

// TestParseHelpTopicsActions checks each definition-list entry of the ACTION
// section is a topic named by the first word of its term, with the term as its
// synopsis, the first sentence as its summary and everything up to the next
// entry, a second paragraph included, as its text.
func TestParseHelpTopicsActions(t *testing.T) {
	topics, err := ParseHelpTopics(helpTestManual, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(topicNames(topics), " "); got != "setup get_record new_draft" {
		t.Fatalf("expected the three actions in manual order, got %q", got)
	}
	setup, record := topics[0], topics[1]
	if setup.Summary != "Display an example." {
		t.Errorf("expected the first sentence as the summary, got %q", setup.Summary)
	}
	if setup.Synopsis != "setup" || setup.Text != "Display an example. More text follows." {
		t.Errorf("unexpected setup topic %+v", setup)
	}
	if record.Synopsis != "get_record RECORD_ID" {
		t.Errorf("expected the term as the synopsis, got %q", record.Synopsis)
	}
	if want := "Returns a record. The RECORD_ID is\nrequired.\n\nA second paragraph of get_record."; record.Text != want {
		t.Errorf("expected the entry and its second paragraph\n got  %q\n want %q", record.Text, want)
	}
	if record.Summary != "Returns a record." {
		t.Errorf("unexpected summary %q", record.Summary)
	}
}

// TestParseHelpTopicsSections checks a listed section is a topic named by its
// heading in lower case with underscores, in the manual's order whatever order
// the program listed them in, and that its text is the section body.
func TestParseHelpTopicsSections(t *testing.T) {
	topics, err := ParseHelpTopics(helpTestManual, []string{"OPTIONS", "EXIT STATUS"})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(topicNames(topics), " "); got != "setup get_record new_draft exit_status options" {
		t.Fatalf("expected actions then sections in manual order, got %q", got)
	}
	exit := topics[3]
	if exit.Synopsis != "" {
		t.Errorf("expected a section to have no term as its synopsis, got %q", exit.Synopsis)
	}
	if !strings.HasPrefix(exit.Text, "Standard output is the record.") || !strings.HasSuffix(exit.Text, ": not found") {
		t.Errorf("expected the whole section body, got %q", exit.Text)
	}
	if exit.Summary != "Standard output is the record." {
		t.Errorf("unexpected summary %q", exit.Summary)
	}
	options := topics[4]
	if options.Name != "options" || options.Summary != "-help, -version" {
		t.Errorf("expected a section that begins with a list of terms to be summarized by the terms, got %q", options.Summary)
	}
}

// TestParseHelpTopicsErrors checks what makes a manual unfit for an index: an
// action documented twice (a documentation bug), a listed section that is not
// there, a name that cannot be a topic, a name used twice, and the reserved word.
func TestParseHelpTopicsErrors(t *testing.T) {
	withAction := func(extra string) string {
		return strings.Replace(helpTestManual, "new_draft RECORD_ID [OPTION]", extra+"\n: x.\n\nnew_draft RECORD_ID [OPTION]", 1)
	}
	tests := []struct {
		name     string
		manual   string
		sections []string
		want     string
	}{
		{"an action twice", withAction("setup"), nil, "more than once"},
		{"a section that is not there", helpTestManual, []string{"NO SUCH SECTION"}, "NO SUCH SECTION"},
		{"an action with an invalid name", withAction("Get-Record X"), nil, "Get-Record"},
		{"an action with a path in its name", withAction("../x"), nil, "../x"},
		{"the reserved word as an action", withAction("topics"), nil, "topics"},
		{"the reserved word as a section", strings.Replace(helpTestManual, "# OPTIONS", "# TOPICS", 1), []string{"TOPICS"}, "topics"},
		{"a section named like an action", strings.Replace(helpTestManual, "# OPTIONS", "# SETUP", 1), []string{"SETUP"}, "setup"},
	}
	for _, tt := range tests {
		topics, err := ParseHelpTopics(tt.manual, tt.sections)
		if err == nil {
			t.Errorf("%s: expected an error, got topics %v", tt.name, topicNames(topics))
			continue
		}
		if !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: expected the error to mention %q, got %q", tt.name, tt.want, err)
		}
	}
}

// TestParseHelpTopicsNone checks a manual with no ACTION section and no
// listed sections has no topics, and that the line before an entry that is not
// followed by ": " is not taken for a term.
func TestParseHelpTopicsNone(t *testing.T) {
	manual := "%one(1) irdmtools user manual | version 1 abc\n% A\n% d\n\n# NAME\n\none\n\n# OPTIONS\n\n-help\n: display help\n"
	topics, err := ParseHelpTopics(manual, nil)
	if err != nil || len(topics) != 0 {
		t.Errorf("expected no topics and no error, got %v, %v", topicNames(topics), err)
	}
}

// TestHelpSummary checks the summary is the first sentence on one line, cut
// to fit the index.
func TestHelpSummary(t *testing.T) {
	long := "This is a very long first sentence that goes on and on and on without a full stop for far longer than one line of an index should hold"
	manual := strings.Replace(helpTestManual, "Creates a draft.", long, 1)
	topics, err := ParseHelpTopics(manual, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := topics[2].Summary
	if len([]rune(got)) > 100 || !strings.HasSuffix(got, "...") || !strings.HasPrefix(got, "This is a very long first sentence") {
		t.Errorf("expected a summary cut to 100 characters with an ellipsis, got %q (%d)", got, len([]rune(got)))
	}
	multi := strings.Replace(helpTestManual, "Creates a draft.", "Creates a\ndraft. And more.", 1)
	topics, _ = ParseHelpTopics(multi, nil)
	if topics[2].Summary != "Creates a draft." {
		t.Errorf("expected the sentence on one line, got %q", topics[2].Summary)
	}
}

// TestHelpSummarySentences checks where the first sentence ends: not at
// "e.g." or "i.e.", at a full stop before a capital letter or a digit, or at the end.
func TestHelpSummarySentences(t *testing.T) {
	tests := []struct{ text, want string }{
		{"Returns a record, e.g. bq3se-47g50. The ID is required.", "Returns a record, e.g. bq3se-47g50."},
		{"Use it, i.e. now. Then stop.", "Use it, i.e. now."},
		{"Items like a, b, etc. are listed. More.", "Items like a, b, etc. are listed."},
		{"Standard output is the record. 0 : ok", "Standard output is the record."},
		{"No full stop at all", "No full stop at all"},
		{"Ends here.", "Ends here."},
	}
	for _, tt := range tests {
		manual := strings.Replace(helpTestManual, "Creates a draft.", tt.text, 1)
		topics, err := ParseHelpTopics(manual, nil)
		if err != nil {
			t.Fatal(err)
		}
		if got := topics[2].Summary; got != tt.want {
			t.Errorf("%q: expected %q, got %q", tt.text, tt.want, got)
		}
	}
}

// TestHelpIndex checks the index: header lines start with "#", every other
// line is a topic name then its summary, aligned, in manual order, so a script
// can read the names from the first field.
func TestHelpIndex(t *testing.T) {
	topics, _ := ParseHelpTopics(helpTestManual, []string{"EXIT STATUS"})
	got := HelpIndex("app", topics)
	var names []string
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			t.Errorf("expected a name and a summary, got %q", line)
			continue
		}
		names = append(names, fields[0])
		if !strings.HasPrefix(line, fields[0]+"  ") {
			t.Errorf("expected two spaces after the name, got %q", line)
		}
	}
	if strings.Join(names, " ") != "setup get_record new_draft exit_status" {
		t.Errorf("unexpected names %v in\n%s", names, got)
	}
	if !strings.HasPrefix(got, "# app help topics.") || !strings.Contains(got, `"app -help TOPIC"`) || !strings.Contains(got, `"app -help"`) {
		t.Errorf("expected a header saying how to ask for a topic and the whole manual, got\n%s", got)
	}
	column := -1
	for _, line := range strings.Split(strings.TrimSuffix(got, "\n"), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		at := strings.Index(line, fields[1])
		if column == -1 {
			column = at
		}
		if at != column {
			t.Errorf("expected the summaries to line up at column %d, got %d in %q", column, at, line)
		}
	}
	if !regexp.MustCompile(`(?m)^get_record +Returns a record\.$`).MatchString(got) {
		t.Errorf("expected the summary after the name, got\n%s", got)
	}
	none := HelpIndex("app", nil)
	if !strings.HasPrefix(none, "# app has no help topics") || strings.Contains(strings.TrimPrefix(none, "#"), "\n\n") {
		t.Errorf("expected one header line for a program with no topics, got %q", none)
	}
	for _, line := range strings.Split(strings.TrimSuffix(none, "\n"), "\n") {
		if !strings.HasPrefix(line, "#") {
			t.Errorf("expected only header lines when there are no topics, got %q", line)
		}
	}
}

// TestHelpPage checks a topic page: the header lines of the manual with the
// topic in the page name (so release-check finds the version and hash), then
// NAME, SYNOPSIS, DESCRIPTION and SEE ALSO.
func TestHelpPage(t *testing.T) {
	topics, _ := ParseHelpTopics(helpTestManual, []string{"EXIT STATUS"})
	page, err := HelpPage("app", helpTestManual, topics[1])
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(page, "\n")
	if lines[0] != "%app-get_record(1) irdmtools user manual | version 1.2.3 abc1234" || lines[1] != "% A. Author" || lines[2] != "% 2026-10-27" {
		t.Errorf("unexpected header lines %q", lines[:3])
	}
	for _, want := range []string{"# NAME\n\napp-get_record - Returns a record.\n", "# SYNOPSIS\n\napp get_record RECORD_ID\n", "# DESCRIPTION\n\nReturns a record. The RECORD_ID is\nrequired.\n\nA second paragraph of get_record.\n", "# SEE ALSO\n\napp(1)\n"} {
		if !strings.Contains(page, want) {
			t.Errorf("expected the page to contain %q, got\n%s", want, page)
		}
	}
	if strings.Index(page, "# NAME") > strings.Index(page, "# SYNOPSIS") || strings.Index(page, "# SYNOPSIS") > strings.Index(page, "# DESCRIPTION") || strings.Index(page, "# DESCRIPTION") > strings.Index(page, "# SEE ALSO") {
		t.Error("expected the sections in order")
	}
	section, err := HelpPage("app", helpTestManual, topics[3])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(section, "# SYNOPSIS\n\napp -help exit_status\n") || !strings.HasPrefix(section, "%app-exit_status(1) ") {
		t.Errorf("expected a section's synopsis to be how to ask for it, got\n%s", section)
	}
	if !strings.HasSuffix(page, "\n") || strings.HasSuffix(page, "\n\n") {
		t.Error("expected exactly one trailing newline")
	}
	if _, err := HelpPage("app", "no header here\n", topics[0]); err == nil {
		t.Error("expected an error for a manual with no header lines")
	}
}

// runHelp runs RunHelp and returns its code, stdout and stderr.
func runHelp(manual string, sections []string, args ...string) (int, string, string) {
	var out, eout bytes.Buffer
	code := RunHelp("app", manual, sections, args, &out, &eout)
	return code, out.String(), eout.String()
}

// TestRunHelp checks the exit codes and what is written where: the index and a
// topic go to standard output with exit 0, a usage error (an unknown topic, a
// surplus word) goes to standard error with exit 2 and lists the topics, a manual
// that cannot be indexed is 70.
func TestRunHelp(t *testing.T) {
	sections := []string{"EXIT STATUS"}
	code, out, eout := runHelp(helpTestManual, sections, "topics")
	if code != 0 || eout != "" || !regexp.MustCompile(`(?m)^get_record +Returns a record\.$`).MatchString(out) {
		t.Errorf("topics: expected 0 and the index, got %d, %q, %q", code, out, eout)
	}
	code, out, eout = runHelp(helpTestManual, sections, "get_record")
	if code != 0 || eout != "" || !strings.HasPrefix(out, "%app-get_record(1)") {
		t.Errorf("a topic: expected 0 and its page, got %d, %q, %q", code, out, eout)
	}
	code, out, eout = runHelp(helpTestManual, sections)
	if code != 0 || !strings.HasPrefix(out, "# app help topics.") {
		t.Errorf("no word: expected the index, got %d, %q", code, out)
	}
	code, out, eout = runHelp(helpTestManual, sections, "nosuchthing")
	if code != 2 || out != "" || !strings.Contains(eout, `no help topic "nosuchthing"`) || !strings.Contains(eout, "app -help topics") || !strings.Contains(eout, "get_record") {
		t.Errorf("unknown topic: expected 2, nothing on stdout and the topics on stderr, got %d, %q, %q", code, out, eout)
	}
	code, out, eout = runHelp(helpTestManual, sections, "get_record", "setup")
	if code != 2 || out != "" || eout == "" {
		t.Errorf("surplus word: expected 2, got %d, %q, %q", code, out, eout)
	}
	none := "%one(1) irdmtools user manual | version 1 abc\n% A\n% d\n\n# NAME\n\none\n"
	code, out, eout = runHelp(none, nil, "topics")
	if code != 0 || !strings.HasPrefix(out, "# app has no help topics") || eout != "" {
		t.Errorf("no topics, topics: expected 0 and a header, got %d, %q, %q", code, out, eout)
	}
	code, out, eout = runHelp(none, nil, "x")
	if code != 2 || out != "" || !strings.Contains(eout, "no help topics") {
		t.Errorf("no topics, a topic: expected 2 and a message, got %d, %q, %q", code, out, eout)
	}
	broken := strings.Replace(helpTestManual, "new_draft RECORD_ID [OPTION]", "setup\n: x.\n\nnew_draft RECORD_ID [OPTION]", 1)
	code, out, eout = runHelp(broken, sections, "topics")
	if code != 70 || out != "" || !strings.Contains(eout, "more than once") {
		t.Errorf("broken manual: expected 70 and the reason, got %d, %q, %q", code, out, eout)
	}
	code, out, eout = runHelp(helpTestManual, []string{"NO SUCH SECTION"}, "topics")
	if code != 70 || out != "" || eout == "" {
		t.Errorf("a listed section that is missing: expected 70, got %d, %q, %q", code, out, eout)
	}
}

// TestParseHelpTopicsTermRules checks which lines are the term of an entry: not
// one that begins with a colon or a space, and the line after it must begin with
// a colon and a space.
func TestParseHelpTopicsTermRules(t *testing.T) {
	manual := "%app(1) irdmtools user manual | version 1 abc\n% A\n% d\n\n# ACTION\n\n" +
		"real_one X\n: The first. It has\n  an indented line\n: and a line that begins with a colon and a space.\n" +
		":starts_with_colon\n: after a colon line.\n\n" +
		"  indented_term\n: after an indented line.\n\n" +
		"no_space_after_colon\n:text without a space\n\n" +
		"real_two\n: The second.\n"
	topics, err := ParseHelpTopics(manual, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(topicNames(topics), " "); got != "real_one real_two" {
		t.Errorf("expected only the two real entries, got %q", got)
	}
	if want := "The first. It has\n  an indented line\n: and a line that begins with a colon and a space.\n:starts_with_colon\n: after a colon line.\n\n  indented_term\n: after an indented line.\n\nno_space_after_colon\n:text without a space"; topics[0].Text != want {
		t.Errorf("expected the text of the first entry to run to the next real term\n got  %q\n want %q", topics[0].Text, want)
	}
}

// TestParseHelpTopicsActionsHeading checks both spellings of the heading, ACTION
// and ACTIONS, and that another heading is not one.
func TestParseHelpTopicsActionsHeading(t *testing.T) {
	for _, heading := range []string{"ACTION", "ACTIONS"} {
		manual := "%app(1) irdmtools user manual | version 1 abc\n% A\n% d\n\n# " + heading + "\n\nsetup\n: Does it.\n"
		topics, err := ParseHelpTopics(manual, nil)
		if err != nil || len(topics) != 1 || topics[0].Name != "setup" {
			t.Errorf("%s: expected the topic setup, got %v, %v", heading, topicNames(topics), err)
		}
	}
	manual := "%app(1) irdmtools user manual | version 1 abc\n% A\n% d\n\n# ACTION_PARAMETERS\n\nsetup\n: Does it.\n"
	if topics, _ := ParseHelpTopics(manual, nil); len(topics) != 0 {
		t.Errorf("expected another heading not to hold actions, got %v", topicNames(topics))
	}
}

// TestParseHelpTopicsNames checks the characters a topic name may have.
func TestParseHelpTopicsNames(t *testing.T) {
	for _, name := range []string{"GetRecord", "1st_action", "get.record", "get-record", "_x", "get/record"} {
		manual := "%app(1) irdmtools user manual | version 1 abc\n% A\n% d\n\n# ACTION\n\n" + name + "\n: Does it.\n"
		if _, err := ParseHelpTopics(manual, nil); err == nil {
			t.Errorf("expected %q to be refused as a topic name", name)
		}
	}
	for _, name := range []string{"get_record", "a", "x1", "v2_api"} {
		manual := "%app(1) irdmtools user manual | version 1 abc\n% A\n% d\n\n# ACTION\n\n" + name + "\n: Does it.\n"
		if _, err := ParseHelpTopics(manual, nil); err != nil {
			t.Errorf("expected %q to be accepted, got %v", name, err)
		}
	}
}

// TestHelpSummaryCut checks a long summary is cut at a word, not in the middle
// of one, and is nearly as long as the index allows.
func TestHelpSummaryCut(t *testing.T) {
	text := strings.TrimSpace(strings.Repeat("abcdefg ", 20))
	manual := strings.Replace(helpTestManual, "Creates a draft.", text, 1)
	topics, err := ParseHelpTopics(manual, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := topics[2].Summary
	if n := len([]rune(got)); n < 90 || n > 100 {
		t.Errorf("expected a summary of 90 to 100 characters, got %d: %q", n, got)
	}
	if !strings.HasSuffix(got, "abcdefg...") {
		t.Errorf("expected the cut to fall between words, got %q", got)
	}
	short := strings.Repeat("x", 100)
	manual = strings.Replace(helpTestManual, "Creates a draft.", short, 1)
	topics, _ = ParseHelpTopics(manual, nil)
	if topics[2].Summary != short {
		t.Errorf("expected a summary of exactly 100 characters to be kept whole, got %q", topics[2].Summary)
	}
}

// TestHelpSummaryAbbreviationBeforeNumber checks "e.g." followed by a number does
// not end the sentence.
func TestHelpSummaryAbbreviationBeforeNumber(t *testing.T) {
	manual := strings.Replace(helpTestManual, "Creates a draft.", "See the record, e.g. 23808. Then stop.", 1)
	topics, _ := ParseHelpTopics(manual, nil)
	if topics[2].Summary != "See the record, e.g. 23808." {
		t.Errorf("unexpected summary %q", topics[2].Summary)
	}
}

// TestHelpPageHeaderRules checks a manual whose header is not three "%" lines
// naming the program is refused, and that RunHelp reports it as a bug (70).
func TestHelpPageHeaderRules(t *testing.T) {
	topic := HelpTopic{Name: "setup", Summary: "Does it.", Text: "Does it."}
	for name, manual := range map[string]string{
		"a third line that is not a header":  "%app(1) irdmtools user manual | version 1 abc\n% A\nnot a header\n",
		"a second line that is not a header": "%app(1) irdmtools user manual | version 1 abc\nnot a header\n% d\n",
		"a first line that is not a header":  "app(1) irdmtools user manual | version 1 abc\n% A\n% d\n",
		"another program's header":           "%other(1) irdmtools user manual | version 1 abc\n% A\n% d\n",
		"only two lines":                     "%app(1) irdmtools user manual | version 1 abc\n% A\n",
		"one line and no newline":            "%app(1) irdmtools user manual | version 1 abc",
	} {
		if _, err := HelpPage("app", manual, topic); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
	manual := "# NAME\n\napp\n\n# ACTION\n\nsetup\n: Does it.\n"
	code, out, eout := runHelp(manual, nil, "setup")
	if code != 70 || out != "" || !strings.Contains(eout, "header") {
		t.Errorf("expected 70 and the reason for a manual with no header, got %d, %q, %q", code, out, eout)
	}
}
