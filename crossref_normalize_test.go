package irdmtools

import (
	"strings"
	"testing"

	"github.com/caltechlibrary/crossrefapi"
	"github.com/caltechlibrary/simplified"
	"gopkg.in/yaml.v3"
)

// journalIssueDOI is the fixture of a CrossRef journal issue, which has no
// title and a resource type the default options do not map.
const journalIssueDOI = "10.1111/jdv.v40.6"

// crosswalkFixture crosswalks a captured CrossRef response with the default
// options, as doi2rdm does for a record it has fetched.
func crosswalkFixture(t *testing.T, doi string) *simplified.Record {
	t.Helper()
	return crosswalkWork(t, doi, loadCrossRefFixture(t, doi))
}

// crosswalkWork crosswalks a CrossRef work with the default options. The
// funders are dropped first: CrosswalkCrossRefWork looks a funder's ROR up on
// api.ror.org, and these tests are offline and about text, not funding.
func crosswalkWork(t *testing.T, doi string, work *crossrefapi.Works) *simplified.Record {
	t.Helper()
	work.Message.Funder = nil
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(TestDefaultOptions, &options); err != nil {
		t.Fatal(err)
	}
	record, err := CrosswalkCrossRefWork(new(Config), work, options)
	if err != nil {
		t.Fatalf("%s: %s", doi, err)
	}
	if record == nil || record.Metadata == nil {
		t.Fatalf("%s: expected a record with metadata", doi)
	}
	return record
}

// TestCrosswalkTitles checks the title a crosswalked CrossRef record gets:
// plain text, with the subtitle appended (irdmtools DR-0001, crossrefapi
// DR-0003; irdm_harvester#42).
func TestCrosswalkTitles(t *testing.T) {
	tests := []struct {
		doi  string
		want string
	}{
		// italic markup and layout newline in the title, subtitle after it
		{"10.1051/0004-6361/202660273", "Euclid preparation: CII. Non-Gaussianity of two-point statistics likelihood: Parameter inference with a non-Gaussian likelihood in Fourier and configuration space"},
		// subtitle after a bracketed title
		{"10.1051/0004-6361/202659007", "A Virgo Environmental Survey Tracing Ionised Gas Emission (VESTIGE): XX. Star formation in the tidal tail of NGC 4254"},
		{"10.1007/978-3-319-20325-6", "Data Assimilation: A Mathematical Introduction"},
		// MathML in the title becomes text (irdmtools#36)
		{"10.1103/physrevb.85.144303", "Line of continuous phase transitions in a three-dimensional U(1) loop model with 1/r² current-current interactions"},
		// a plain title is unchanged
		{"10.1093/mnras/stae1289", "Searching for magnetar binaries disrupted by core-collapse supernovae"},
		// the first of two titles leads
		{"10.1007/s00717-026-00622-0", "X-chromosomale juvenile Retinoschisis"},
		{"10.24272/j.issn.2095-8137.2025.121", "巴马香猪与野猪肾脏发育的动态单核转录图谱"},
	}
	for _, tt := range tests {
		record := crosswalkFixture(t, tt.doi)
		if record.Metadata.Title != tt.want {
			t.Errorf("%s:\n got  %q\n want %q", tt.doi, record.Metadata.Title, tt.want)
		}
	}
}

// TestCrosswalkTitleHoldingItsSubtitle checks that a title which already
// contains its subtitle is not doubled, and that its markup is removed.
func TestCrosswalkTitleHoldingItsSubtitle(t *testing.T) {
	record := crosswalkFixture(t, "10.1080/00253359.2024.2340221")
	title := record.Metadata.Title
	if n := strings.Count(title, "by T. Michals"); n != 1 {
		t.Errorf("expected the subtitle once, found it %d times in %q", n, title)
	}
	if strings.ContainsAny(title, "<>\n") {
		t.Errorf("expected plain text on one line, got %q", title)
	}
}

// TestCrosswalkAdditionalTitles checks the additional titles: the CrossRef
// titles after the first that add something, each typed "other" (issue #87).
func TestCrosswalkAdditionalTitles(t *testing.T) {
	tests := []struct {
		doi  string
		want []string
	}{
		{"10.1007/s00717-026-00622-0", []string{"X-linked juvenile retinoschisis"}},
		{"10.24272/j.issn.2095-8137.2025.121", []string{"Dynamic single-nucleus transcriptomic profiling of kidney development in Bamaxiang pigs and wild boar"}},
		{"10.47456/cad.astro.v7nespecial.53384", []string{"The Process of Investigating Planetary Transits in Photometric Light Curves from TESS"}},
		// the second title differs only by capitalization
		{"10.67686/didaktika.v10i2.2338", nil},
		{"10.1007/978-3-319-20325-6", nil},
	}
	for _, tt := range tests {
		record := crosswalkFixture(t, tt.doi)
		var got []string
		for i, detail := range record.Metadata.AdditionalTitles {
			got = append(got, detail.Title)
			if detail.Type == nil || detail.Type.ID != "other" {
				t.Errorf("%s: additional_titles[%d] expected type %q, got %v", tt.doi, i, "other", detail.Type)
			}
		}
		if strings.Join(got, "|") != strings.Join(tt.want, "|") {
			t.Errorf("%s: additional titles\n got  %q\n want %q", tt.doi, got, tt.want)
		}
	}
}

// TestCrosswalkTitlesArePlainText checks that no fixture leaves markup or a
// layout newline in a title or additional title.
func TestCrosswalkTitlesArePlainText(t *testing.T) {
	for _, doi := range crossRefFixtureDOIs {
		if doi == journalIssueDOI {
			continue // its resource type is not mapped, see TestCrosswalkNoTitle
		}
		record := crosswalkFixture(t, doi)
		titles := []string{record.Metadata.Title}
		for _, detail := range record.Metadata.AdditionalTitles {
			titles = append(titles, detail.Title)
		}
		for _, title := range titles {
			if strings.ContainsAny(title, "<>\n") || title != strings.TrimSpace(title) {
				t.Errorf("%s: title is not plain text: %q", doi, title)
			}
		}
	}
}

// TestCrosswalkNoTitle checks that a work with no title gives a record with no
// title and no additional titles. The fixture is a journal issue, whose
// resource type is not mapped, so it is typed as an article here: the type is
// not what is under test.
func TestCrosswalkNoTitle(t *testing.T) {
	work := loadCrossRefFixture(t, journalIssueDOI)
	if len(work.Message.Title) != 0 {
		t.Fatalf("expected a fixture with no title, got %q", work.Message.Title)
	}
	work.Message.Type = "journal-article"
	record := crosswalkWork(t, journalIssueDOI, work)
	if record.Metadata.Title != "" || len(record.Metadata.AdditionalTitles) != 0 {
		t.Errorf("expected no titles, got %q and %d additional", record.Metadata.Title, len(record.Metadata.AdditionalTitles))
	}
}
