package irdmtools

import (
	"bytes"
	"os"
	"path"
	"encoding/json"
	"testing"

	// Caltech Library Packages
	"github.com/caltechlibrary/simplified"
)

func TestCiteProcItemCrosswalkRecord(t *testing.T) {
	sampleName := path.Join("testdata", "10.5281-inveniordm.1234.json")
	src, err := os.ReadFile(sampleName)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	rec := new(simplified.Record)
	decoder := json.NewDecoder(bytes.NewBuffer(src))
	decoder.UseNumber()
	if err := decoder.Decode(rec); err != nil {
		t.Error(err)
		t.FailNow()
	}
	
	item := &CiteProcItem{}
	if err := item.CrosswalkRecord(rec); err != nil {
		t.Error(err)
	}
}
