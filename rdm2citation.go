package irdmtools

import (
	"fmt"

	// Caltech Library packages
	"github.com/caltechlibrary/simplified"
)

// This will talk to a RDM database and retrieve an RDM record
// and output a Citation record in JSON format.

// Convert an RDM record to a citation in a Citation struct
func RdmToCitation(record *simplified.Record) (*Citation, error) {
	return nil, fmt.Errorf("RdmToCitation() not implemented")
}
