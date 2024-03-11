package irdmtools

import (
	"fmt"

	// Caltech Library packages
	"github.com/caltechlibrary/simplified"
)

// This will talk to a RDM database and retrieve an RDM record
// and output a CiteProc record in JSON format.

// Convert an RDM record to a citation in a CiteProc struct
func RdmToCiteProc(record *simplified.Record) (*CiteProcItem, error) {
	return nil, fmt.Errorf("RdmToCiteProc() not implemented")
}
