package irdmtools

import (
	"testing"

	// Caltech Library Packages
	"github.com/caltechlibrary/eprinttools"
	"github.com/caltechlibrary/simplified"
)

// TestThesisCustomFieldUsesUniversityKey checks that a thesis eprint's
// institution is carried into the "thesis:thesis" custom field under the
// key InvenioRDM defines, "university". The key was once misspelled
// "unversity", which silently dropped the institution on load.
func TestThesisCustomFieldUsesUniversityKey(t *testing.T) {
	eprint := &eprinttools.EPrint{
		Type:        "thesis",
		ThesisType:  "phd",
		Institution: "California Institute of Technology",
		Department:  "Chemistry",
	}
	rec := new(simplified.Record)
	if err := customFieldsMetadataFromEPrint(eprint, rec); err != nil {
		t.Fatalf("customFieldsMetadataFromEPrint returned an error: %s", err)
	}
	raw, ok := rec.CustomFields["thesis:thesis"]
	if !ok {
		t.Fatalf("expected a thesis:thesis custom field, got %+v", rec.CustomFields)
	}
	thesis, ok := raw.(map[string]interface{})
	if !ok {
		t.Fatalf("thesis:thesis is %T, expected map[string]interface{}", raw)
	}
	if got, ok := thesis["university"]; !ok || got != "California Institute of Technology" {
		t.Errorf("expected university %q, got %v (present: %t)",
			"California Institute of Technology", got, ok)
	}
	if _, ok := thesis["unversity"]; ok {
		t.Errorf("the misspelled key %q must not be written", "unversity")
	}
	if got := thesis["department"]; got != "Chemistry" {
		t.Errorf("expected department %q, got %v", "Chemistry", got)
	}
	if got := thesis["type"]; got != "PhD" {
		t.Errorf("expected type %q, got %v", "PhD", got)
	}
}
