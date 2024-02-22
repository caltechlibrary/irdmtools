package irdmtools

import (
	"testing"
)

func TestDoiFuncs(t *testing.T) {
	input := "https://doi.org/10.48550/arXiv.2104.02480"
	expected := "10.48550/arXiv.2104.02480"
	got, err := LinkToDoi(input)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}

	expected = "10.48550"
	got, err = DoiPrefix(input)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}

	input = "10.1029/2022PA004571"
	expected = "10.1029/2022PA004571"
	got, err = LinkToDoi(input)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	expected = "10.1029"
	got, err = DoiPrefix(input)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if expected != got {
		t.Errorf("expected %q, got %q", expected, got)
	}

}
