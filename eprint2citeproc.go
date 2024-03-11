package irdmtools

import (
	// Caltech Library packages
	"github.com/caltechlibrary/eprinttools"
	"github.com/caltechlibrary/simplified"
)

// This will talk to a EPrints3 database and retrieve an EPrint record
// and output a CiteProc record in JSON format.

// EPrintToCiteProc takes a single EPrint records and returns a single
// CiteProc struct
func EPrintToCiteProcItem(eprint *eprinttools.EPrint, resourceTypes map[string]string, contributorTypes map[string]string) (*CiteProcItem, error) {
	rec := &simplified.Record{}
    err := CrosswalkEPrintToRecord(eprint, rec, resourceTypes, contributorTypes)
    if err != nil {
    	return nil, err
    }    
    citeProcItem := new(CiteProcItem)
    err = citeProcItem.CrosswalkRecord(rec)
    return citeProcItem, err
}

// EPrintsToCiteProc takes an array of EPrint records and returns an array of
// citation information as CiteProc structs
func EPrintsToCiteProc(eprints *eprinttools.EPrints, resourceTypes map[string]string, contributorTypes map[string]string) ([]*CiteProcItem, error) {
	l := []*CiteProcItem{}
	for _, eprint := range eprints.EPrint {
		citeproc, err := EPrintToCiteProcItem(eprint, resourceTypes, contributorTypes)
		if err != nil {
			return nil, err
		} 
		l = append(l, citeproc)
	}
	return l, nil
}
