package irdmtools

import (
	"testing"
)

func TestLookupROR(t *testing.T) {
	doiSuffix := "100000025"
	expectedROR := "https://ror.org/04xeg9z08"
	ror, ok := lookupROR(doiSuffix, false)
	if ! ok {
		t.Errorf("expected lookupROR to return OK, failed")
	}
	if ror != expectedROR {
		t.Errorf("expected ror %q, got %q", expectedROR, ror)
	}
}
