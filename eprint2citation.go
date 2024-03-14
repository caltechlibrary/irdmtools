package irdmtools

import (
	// Caltech Library packages
	"github.com/caltechlibrary/eprinttools"
	"github.com/caltechlibrary/simplified"
)

// This will talk to a EPrints3 database and retrieve an EPrint record
// and output a CiteProc record in JSON format.

// EPrintToCitation takes a single EPrint records and returns a single
// Citation struct
func EPrintToCitation(eprint *eprinttools.EPrint, resourceTypes map[string]string, contributorTypes map[string]string) (*Citation, error) {
	rec := &simplified.Record{}
    err := CrosswalkEPrintToRecord(eprint, rec, resourceTypes, contributorTypes)
    if err != nil {
    	return nil, err
    }    
    citeProcItem := new(Citation)
    err = citeProcItem.CrosswalkRecord(rec)
    return citeProcItem, err
}

// EPrintsToCitation takes an array of EPrint records and returns an array of
// citation information as Citation structs
func EPrintsToCitation(eprints *eprinttools.EPrints, resourceTypes map[string]string, contributorTypes map[string]string) ([]*Citation, error) {
	l := []*Citation{}
	for _, eprint := range eprints.EPrint {
		citeproc, err := EPrintToCitation(eprint, resourceTypes, contributorTypes)
		if err != nil {
			return nil, err
		} 
		l = append(l, citeproc)
	}
	return l, nil
}
