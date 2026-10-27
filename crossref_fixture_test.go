package irdmtools

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/caltechlibrary/crossrefapi"
)

// crossRefFixtureDOIs are the captured CrossRef responses in testdata/crossref
// that the doi2rdm normalization work is tested against.
var crossRefFixtureDOIs = []string{
	"10.1051/0004-6361/202660273",
	"10.1051/0004-6361/202659007",
	"10.1007/978-3-319-20325-6",
	"10.1080/00253359.2024.2340221",
	"10.1103/physrevb.85.144303",
	"10.1103/physrevb.109.035167",
	"10.1103/physrevlett.134.141902",
	"10.1103/sh13-grfw",
	"10.1007/s00717-026-00622-0",
	"10.24272/j.issn.2095-8137.2025.121",
	"10.67686/didaktika.v10i2.2338",
	"10.1111/jdv.v40.6",
	"10.1002/2016GL070081",
	"10.1007/s11695-024-07110-x",
	"10.3847/1538-4357/ae99c4",
	"10.1007/s41114-018-0012-9",
	"10.1093/mnras/stae1289",
	"10.47456/cad.astro.v7nespecial.53384",
}

// TestCrossRefFixturesLoad checks that every fixture loads into
// *crossrefapi.Works and is the record its file name says it is.
func TestCrossRefFixturesLoad(t *testing.T) {
	for _, doi := range crossRefFixtureDOIs {
		work := loadCrossRefFixture(t, doi)
		if work.Message == nil {
			t.Errorf("%s: no message", doi)
			continue
		}
		if !strings.EqualFold(work.Message.DOI, doi) {
			t.Errorf("%s: fixture holds DOI %q", doi, work.Message.DOI)
		}
	}
}

// TestCrossRefFixtureProperties checks the properties the fixtures were
// chosen for, so a replaced or mistyped fixture is noticed.
func TestCrossRefFixtureProperties(t *testing.T) {
	tests := []struct {
		doi       string
		titles    int
		subtitles int
	}{
		{"10.1051/0004-6361/202660273", 1, 1},
		{"10.1007/978-3-319-20325-6", 1, 1},
		{"10.1111/jdv.v40.6", 0, 0},
		{"10.47456/cad.astro.v7nespecial.53384", 2, 0},
		{"10.1007/s00717-026-00622-0", 2, 0},
	}
	for _, tt := range tests {
		work := loadCrossRefFixture(t, tt.doi)
		if got := len(work.Message.Title); got != tt.titles {
			t.Errorf("%s: expected %d titles, got %d", tt.doi, tt.titles, got)
		}
		if got := len(work.Message.SubTitle); got < tt.subtitles {
			t.Errorf("%s: expected at least %d subtitles, got %d", tt.doi, tt.subtitles, got)
		}
	}
}

// loadCrossRefFixture reads the captured CrossRef response for a DOI from
// testdata/crossref into a *crossrefapi.Works, the way doi2rdm reads a diff
// file. It stops the test if the fixture is missing or does not parse.
func loadCrossRefFixture(t *testing.T, doi string) *crossrefapi.Works {
	t.Helper()
	fName := path.Join("testdata", "crossref", "crossref-"+strings.ReplaceAll(doi, "/", "_")+".json")
	src, err := os.ReadFile(fName)
	if err != nil {
		t.Fatal(err)
	}
	work := new(crossrefapi.Works)
	if err := JSONUnmarshal(src, &work); err != nil {
		t.Fatalf("%s: %s", fName, err)
	}
	return work
}
