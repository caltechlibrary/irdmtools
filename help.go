package irdmtools

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

// HelpTopic is one topic of a program's help index. The topics are cut from
// the program's manual, not written a second time, so editing the manual
// changes the topic and the documentation generated from it together.
type HelpTopic struct {
	// Name is the topic's name, as typed after -help: an action's name
	// (get_record) or a section's heading in lower case with underscores
	// (exit_status).
	Name string
	// Summary is the first sentence of the topic's text on one line, cut to
	// fit one line of the index.
	Summary string
	// Synopsis is the term line of an action (get_record RECORD_ID), and ""
	// for a section.
	Synopsis string
	// Text is the topic's text: an action's description up to the next
	// action, or a section's body.
	Text string
}

// helpTopicNameRE is what a topic name may be. Names become file names
// (docs/NAME-TOPIC.1.md) and are typed on a command line.
var helpTopicNameRE = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// helpReservedTopic is the word that asks for the index, so it is not a topic.
const helpReservedTopic = "topics"

// helpSummaryWidth is the longest summary, in characters, the index shows.
const helpSummaryWidth = 100

// helpPeriodRE finds a full stop followed by white space or the end of the text.
var helpPeriodRE = regexp.MustCompile(`\.(?:\s+|$)`)

// helpAbbreviations are the words whose full stop does not end a sentence.
var helpAbbreviations = map[string]bool{"e.g": true, "i.e": true, "etc": true, "vs": true, "cf": true}

// helpCut cuts a line to helpSummaryWidth characters at a space, with an
// ellipsis, when it is longer.
func helpCut(line string) string {
	runes := []rune(line)
	if len(runes) <= helpSummaryWidth {
		return line
	}
	cut := helpSummaryWidth - 3
	for i := cut; i > 0; i-- {
		if unicode.IsSpace(runes[i]) {
			cut = i
			break
		}
	}
	return strings.TrimRightFunc(string(runes[:cut]), unicode.IsSpace) + "..."
}

// helpSummary returns the first sentence of text on one line, cut to
// helpSummaryWidth characters with an ellipsis when it is longer. A sentence
// ends at a full stop that is followed by a capital letter, a digit or the end
// of the text, and is not the full stop of "e.g.", "i.e." and the like.
func helpSummary(text string) string {
	line := strings.Join(strings.Fields(text), " ")
	for _, loc := range helpPeriodRE.FindAllStringIndex(line, -1) {
		before := line[:loc[0]]
		if helpAbbreviations[strings.ToLower(before[strings.LastIndex(before, " ")+1:])] {
			continue
		}
		after := []rune(line[loc[1]:])
		if len(after) == 0 || unicode.IsUpper(after[0]) || unicode.IsDigit(after[0]) {
			line = line[:loc[0]+1]
			break
		}
	}
	return helpCut(line)
}

// helpIsTerm reports whether body[i] is a term line: a line that is followed
// by one beginning ": ".
func helpIsTerm(body []string, i int) bool {
	line := body[i]
	return line != "" && !strings.HasPrefix(line, ":") && !strings.HasPrefix(line, " ") &&
		i+1 < len(body) && strings.HasPrefix(body[i+1], ": ")
}

// helpSectionSummary summarizes a section's text: the first sentence, or, for a
// section that begins with a list of terms (OPTIONS), the terms themselves.
func helpSectionSummary(body []string, text string) string {
	first := 0
	for first < len(body) && body[first] == "" {
		first++
	}
	if first >= len(body) || !helpIsTerm(body, first) {
		return helpSummary(text)
	}
	var terms []string
	for i := range body {
		if helpIsTerm(body, i) {
			terms = append(terms, strings.Fields(body[i])[0])
		}
	}
	return helpCut(strings.Join(terms, ", "))
}

// helpSection is a section of a manual: its heading and the lines of its body.
type helpSection struct {
	title string
	body  []string
}

// helpSections splits a manual into its "# HEADING" sections.
func helpSections(manual string) []helpSection {
	var sections []helpSection
	for _, line := range strings.Split(manual, "\n") {
		if strings.HasPrefix(line, "# ") {
			sections = append(sections, helpSection{title: strings.TrimSpace(line[2:])})
			continue
		}
		if len(sections) > 0 {
			sections[len(sections)-1].body = append(sections[len(sections)-1].body, line)
		}
	}
	return sections
}

// helpActionEntries cuts the body of an ACTION section into topics. An entry is
// a term line (the action's name, then its parameters) that is followed by a
// line beginning ": "; its text runs to the next term line or the end of the
// section, so a second paragraph stays with its entry.
func helpActionEntries(body []string) []HelpTopic {
	isTerm := func(i int) bool { return helpIsTerm(body, i) }
	var topics []HelpTopic
	for i := 0; i < len(body); i++ {
		if !isTerm(i) {
			continue
		}
		end := i + 1
		for end < len(body) && !isTerm(end) {
			end++
		}
		text := append([]string{strings.TrimPrefix(body[i+1], ": ")}, body[i+2:end]...)
		joined := strings.Trim(strings.Join(text, "\n"), "\n")
		topics = append(topics, HelpTopic{
			Name:     strings.Fields(body[i])[0],
			Synopsis: strings.TrimSpace(body[i]),
			Summary:  helpSummary(joined),
			Text:     joined,
		})
		i = end - 1
	}
	return topics
}

// ParseHelpTopics cuts a program's manual into help topics, in the order they
// appear in the manual: each definition-list entry of an ACTION (or ACTIONS)
// section is a topic named by the first word of its term, and each section
// named in sections is a topic named by its heading in lower case with spaces as
// underscores. A manual with neither has no topics. It refuses a manual that
// documents an action twice (a documentation bug), names a section that is not
// there, has a topic name that is not lower case letters, digits and
// underscores, uses the reserved name "topics", or gives two topics one name.
//
// @param manual {string} the manual text, as the program's -help prints it
// @param sections {[]string} headings of sections that are topics, such as "EXIT STATUS"
// @returns {[]HelpTopic} the topics in manual order, empty when there are none
// @returns {error} an error naming the defect in the manual
// @example
//
//	topics, err := ParseHelpTopics(manual, []string{"EXIT STATUS"})
//	for _, t := range topics {
//		fmt.Println(t.Name, t.Summary)
//	}
func ParseHelpTopics(manual string, sections []string) ([]HelpTopic, error) {
	all := helpSections(manual)
	titles := map[string]bool{}
	for _, s := range all {
		titles[s.title] = true
	}
	listed := map[string]bool{}
	for _, name := range sections {
		if !titles[name] {
			return nil, fmt.Errorf("the help section %q is not in the manual", name)
		}
		listed[name] = true
	}
	var topics []HelpTopic
	seen := map[string]bool{}
	add := func(t HelpTopic, what string) error {
		if !helpTopicNameRE.MatchString(t.Name) {
			return fmt.Errorf("%q cannot be a help topic name: use lower case letters, digits and underscores", t.Name)
		}
		if t.Name == helpReservedTopic {
			return fmt.Errorf("%q cannot be a help topic: it is the word that asks for the index", t.Name)
		}
		if seen[t.Name] {
			return fmt.Errorf("the help topic %q is documented more than once (%s); the manual has a duplicate", t.Name, what)
		}
		seen[t.Name] = true
		topics = append(topics, t)
		return nil
	}
	for _, s := range all {
		switch {
		case s.title == "ACTION" || s.title == "ACTIONS":
			for _, t := range helpActionEntries(s.body) {
				if err := add(t, "action "+t.Name); err != nil {
					return nil, err
				}
			}
		case listed[s.title]:
			text := strings.Trim(strings.Join(s.body, "\n"), "\n")
			name := strings.ReplaceAll(strings.ToLower(s.title), " ", "_")
			if err := add(HelpTopic{Name: name, Summary: helpSectionSummary(s.body, text), Text: text}, "section "+s.title); err != nil {
				return nil, err
			}
		}
	}
	return topics, nil
}

// HelpIndex returns the list of topics a program answers "-help topics" with.
// Lines that start with "#" are the header; every other line is a topic name,
// at least two spaces, and its summary, with the summaries lined up, so a script
// reads the names from the first field.
//
// @param appName {string} the program's name
// @param topics {[]HelpTopic} the program's topics, which may be none
// @returns {string} the index, ending with a newline
// @example
//
//	fmt.Print(HelpIndex("rdmutil", topics))
func HelpIndex(appName string, topics []HelpTopic) string {
	if len(topics) == 0 {
		return fmt.Sprintf("# %s has no help topics. Run \"%s -help\" for the manual.\n", appName, appName)
	}
	width := 0
	for _, t := range topics {
		if len(t.Name) > width {
			width = len(t.Name)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s help topics. Run \"%s -help TOPIC\" for one, or \"%s -help\" for the whole manual.\n", appName, appName, appName)
	for _, t := range topics {
		fmt.Fprintf(&b, "%-*s  %s\n", width, t.Name, t.Summary)
	}
	return b.String()
}

// HelpPage returns a topic as a manual page in Markdown. Its first three lines
// are the manual's own header lines with the topic in the page name
// (%app-topic(1)), so the version and hash that release-check looks for are
// there; then NAME, SYNOPSIS, DESCRIPTION and SEE ALSO.
//
// @param appName {string} the program's name
// @param manual {string} the manual text the topic was cut from, for its header lines
// @param topic {HelpTopic} the topic
// @returns {string} the page, ending with one newline
// @returns {error} an error if the manual does not begin with its three "%" header lines
// @example
//
//	page, err := HelpPage("rdmutil", manual, topics[0])
func HelpPage(appName, manual string, topic HelpTopic) (string, error) {
	lines := strings.SplitN(manual, "\n", 4)
	if len(lines) < 3 || !strings.HasPrefix(lines[0], "%") || !strings.HasPrefix(lines[1], "%") || !strings.HasPrefix(lines[2], "%") ||
		!strings.Contains(lines[0], appName+"(1)") {
		return "", fmt.Errorf("the manual of %s does not begin with its three %% header lines", appName)
	}
	synopsis := fmt.Sprintf("%s -help %s", appName, topic.Name)
	if topic.Synopsis != "" {
		synopsis = appName + " " + topic.Synopsis
	}
	var b strings.Builder
	b.WriteString(strings.Replace(lines[0], appName+"(1)", appName+"-"+topic.Name+"(1)", 1) + "\n")
	b.WriteString(lines[1] + "\n" + lines[2] + "\n\n")
	fmt.Fprintf(&b, "# NAME\n\n%s-%s - %s\n\n", appName, topic.Name, topic.Summary)
	fmt.Fprintf(&b, "# SYNOPSIS\n\n%s\n\n", synopsis)
	fmt.Fprintf(&b, "# DESCRIPTION\n\n%s\n\n", topic.Text)
	fmt.Fprintf(&b, "# SEE ALSO\n\n%s(1)\n", appName)
	return b.String(), nil
}

// RunHelp answers the words that follow -help: "topics" (or nothing) with the
// index, a topic's name with its page. The index and a page go to out with exit
// code 0. An unknown topic, a topic for a program that has none, or more than one
// word is a usage error: a message to eout and exit code 2. A manual that cannot
// be indexed is a bug in the program, exit code 70. Workspace DR-0014 gives the
// numbers.
//
// @param appName {string} the program's name
// @param manual {string} the manual text, as the program's -help prints it
// @param sections {[]string} headings of the manual's sections that are topics
// @param args {[]string} the words after -help
// @param out {io.Writer} where the index or the page is written
// @param eout {io.Writer} where a message is written
// @returns {int} the exit code
// @example
//
//	os.Exit(RunHelp("rdmutil", manual, nil, flag.Args(), os.Stdout, os.Stderr))
func RunHelp(appName, manual string, sections []string, args []string, out io.Writer, eout io.Writer) int {
	topics, err := ParseHelpTopics(manual, sections)
	if err != nil {
		fmt.Fprintf(eout, "%s: %s\n", appName, err)
		return int(ExitInternal)
	}
	if len(args) > 1 {
		fmt.Fprintf(eout, "%s: expected one help topic, or %q, got %d words\n", appName, helpReservedTopic, len(args))
		return int(ExitUsage)
	}
	if len(args) == 0 || args[0] == helpReservedTopic {
		io.WriteString(out, HelpIndex(appName, topics))
		return int(ExitOK)
	}
	for _, topic := range topics {
		if topic.Name != args[0] {
			continue
		}
		page, err := HelpPage(appName, manual, topic)
		if err != nil {
			fmt.Fprintf(eout, "%s: %s\n", appName, err)
			return int(ExitInternal)
		}
		if _, err := io.WriteString(out, page); err != nil {
			return int(ExitIO)
		}
		return int(ExitOK)
	}
	if len(topics) == 0 {
		fmt.Fprintf(eout, "%s has no help topics; run \"%s -help\" for the manual\n", appName, appName)
		return int(ExitUsage)
	}
	var names []string
	for _, topic := range topics {
		names = append(names, topic.Name)
	}
	fmt.Fprintf(eout, "no help topic %q; the topics are: %s; try: %s -help topics\n", args[0], strings.Join(names, ", "), appName)
	return int(ExitUsage)
}
