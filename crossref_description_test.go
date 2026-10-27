package irdmtools

import (
	"regexp"
	"strings"
	"testing"
)

// sanitizerTags is RDM's allowed HTML tag list, copied from marshmallow_utils
// 0.15.2 (ALLOWED_HTML_TAGS) as crossrefapi's tests do. A description must
// stay inside it, or RDM's sanitizer will remove what it does not know.
var sanitizerTags = map[string]bool{
	"a": true, "abbr": true, "acronym": true, "b": true, "blockquote": true, "br": true,
	"code": true, "col": true, "colgroup": true, "div": true, "table": true, "tbody": true,
	"tfoot": true, "thead": true, "td": true, "th": true, "tr": true, "em": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "i": true, "li": true,
	"ol": true, "p": true, "pre": true, "s": true, "span": true, "strike": true,
	"strong": true, "sub": true, "sup": true, "u": true, "ul": true,
}

var descriptionTagRE = regexp.MustCompile(`</?([A-Za-z][A-Za-z0-9]*)`)

// TestCrosswalkDescriptions checks the description a crosswalked CrossRef
// record gets from its JATS abstract: HTML RDM's sanitizer keeps, with MathML as
// TeX between \( and \), the leading "Abstract" heading and the publisher's
// copyright line dropped (irdmtools DR-0001; irdmtools#36).
func TestCrosswalkDescriptions(t *testing.T) {
	tests := []struct {
		doi      string
		prefix   string
		contains []string
		absent   []string
	}{
		// MathML in the abstract, a copyright line and a bare heading to drop
		{"10.1103/physrevb.109.035167", "", []string{`\(1T-\mathrm{TaS}_{2}\)`, `\(d_{z^{2}}\)`},
			[]string{"Published by the American Physical Society", "<b></b>", "jats", "mml", "xmlns", "<b>Abstract</b>"}},
		// formulas written as TeX in the source, and a leading Abstract heading
		{"10.1007/s41114-018-0012-9", "", []string{`\(90\%\)`, `\(20~\mathrm{deg}^{2}\)`},
			[]string{"$$", "tex-math", "jats", "<b>Abstract</b>"}},
		// italics stay italics, the layout whitespace around them is tidied
		{"10.1051/0004-6361/202660273", "<p>", []string{"(<i>k</i> &lt; 0.5 <i>h</i>/Mpc)", "cloned <i>Euclid</i> large"},
			[]string{"jats", "\n"}},
		// a structured abstract keeps its headings as bold paragraphs; the
		// leading "Abstract" heading, alone in its section, goes
		{"10.1007/s11695-024-07110-x", "<p><b>Background</b></p><p>Bariatric surgery",
			[]string{"<p><b>Aim</b></p>", "<p><b>Methods</b></p>"}, []string{"<b>Abstract</b></p><p><b>Background"}},
		// the comparison keeps its spaces, escaped for HTML
		{"10.1002/2016GL070081", "<p>Earthquake onsets", []string{"(4 &lt; <i>M</i> &lt; 8)"}, []string{"jats"}},
	}
	for _, tt := range tests {
		got := crosswalkFixture(t, tt.doi).Metadata.Description
		if got == "" {
			t.Errorf("%s: expected a description", tt.doi)
			continue
		}
		if tt.prefix != "" && !strings.HasPrefix(got, tt.prefix) {
			t.Errorf("%s: expected the description to start %q, got %q", tt.doi, tt.prefix, got[:min(len(got), len(tt.prefix)+20)])
		}
		for _, want := range tt.contains {
			if !strings.Contains(got, want) {
				t.Errorf("%s: description lacks %q", tt.doi, want)
			}
		}
		for _, bad := range tt.absent {
			if strings.Contains(got, bad) {
				t.Errorf("%s: description has %q", tt.doi, bad)
			}
		}
	}
}

// TestCrosswalkDescriptionsUseSanitizerTags checks every fixture's
// description is HTML the sanitizer keeps, with nothing of JATS or MathML left.
func TestCrosswalkDescriptionsUseSanitizerTags(t *testing.T) {
	withAbstract := 0
	for _, doi := range crossRefFixtureDOIs {
		if doi == journalIssueDOI {
			continue // its resource type is not mapped, see TestCrosswalkNoTitle
		}
		description := crosswalkFixture(t, doi).Metadata.Description
		if description == "" {
			continue
		}
		withAbstract++
		for _, m := range descriptionTagRE.FindAllStringSubmatch(description, -1) {
			if !sanitizerTags[m[1]] {
				t.Errorf("%s: tag <%s> is not in the sanitizer's list", doi, m[1])
			}
		}
		for _, bad := range []string{"jats", "mml", "xmlns", "tex-math", "<math", "$$"} {
			if strings.Contains(description, bad) {
				t.Errorf("%s: description has %q", doi, bad)
			}
		}
		if strings.Count(description, `\(`) != strings.Count(description, `\)`) {
			t.Errorf("%s: unbalanced formula delimiters", doi)
		}
	}
	if withAbstract < 8 {
		t.Errorf("expected most fixtures to carry an abstract, only %d do", withAbstract)
	}
}

// TestCrosswalkNoAbstract checks that a work with no abstract gets no
// description. The fixture is the journal issue, typed as an article as in
// TestCrosswalkNoTitle.
func TestCrosswalkNoAbstract(t *testing.T) {
	work := loadCrossRefFixture(t, journalIssueDOI)
	if work.Message.Abstract != "" {
		t.Fatalf("expected a fixture with no abstract")
	}
	work.Message.Type = "journal-article"
	if got := crosswalkWork(t, journalIssueDOI, work).Metadata.Description; got != "" {
		t.Errorf("expected no description, got %q", got)
	}
}
