package irdmtools

import (
	"fmt"
	"strings"
	
	// Caltech Library packages
	"github.com/caltechlibrary/eprinttools"
	"github.com/caltechlibrary/simplified"
)

// This will talk to a EPrints3 database and retrieve an EPrint record
// and output a CiteProc record in JSON format.

// EPrintToCitation takes a single EPrint records and returns a single
// Citation struct
func EPrintToCitation(eprint *eprinttools.EPrint, baseURL string, resourceTypes map[string]string, contributorTypes map[string]string) (*Citation, error) {
	rec := &simplified.Record{}
    err := CrosswalkEPrintToRecord(eprint, rec, resourceTypes, contributorTypes)
    if err != nil {
    	return nil, err
    }
    eprintID := eprint.ID
    if eprintID == "" {
    	eprintID = fmt.Sprintf("%d", eprint.EPrintID)
    }
   	// This is the way an EPrint URL is actually formed.
   	eprintURL := fmt.Sprintf("%s/%d", baseURL, eprint.EPrintID)
    // NOTE: We're dealing with a squirly situation of URLs to use during our migration and
    // before the feeds v2.0 implementation.
    if strings.HasPrefix(eprint.ID, "http") || strings.HasPrefix(eprint.ID, "/") {
    	eprintURL = eprint.ID
    } else if eprint.OfficialURL != "" {
    	eprintURL = eprint.OfficialURL
    }
    citation := new(Citation)
    err = citation.CrosswalkRecord(eprint.Collection, eprintID, eprintURL, rec)
    return citation, err
}

// EPrintsToCitation takes an array of EPrint records and returns an array of
// citation information as Citation structs
func EPrintsToCitation(eprints *eprinttools.EPrints, baseURL string, resourceTypes map[string]string, contributorTypes map[string]string) ([]*Citation, error) {
	l := []*Citation{}
	for _, eprint := range eprints.EPrint {
		citation, err := EPrintToCitation(eprint, baseURL, resourceTypes, contributorTypes)
		if err != nil {
			return nil, err
		} 
		l = append(l, citation)
	}
	return l, nil
}
