package irdmtools

import (
	//"fmt"
	//"encoding/json"
	"testing"
)

func TestQueryDataCiteObject(t *testing.T) {
	ids := []string{
		"10.22002/D1.868",
		"arXiv:2312.07215",
		"arXiv:2305.06519",
		"arXiv:2312.03791",
		"arXiv:2305.19279",
		"arXiv:2305.05315",
		"arXiv:2305.07673",
		"arXiv:2111.03606",
		"arXiv:2112.06016",
	}
	cfg := new(Config)
	options := new(Doi2RdmOptions)
	options.MailTo = "test@example.edu"
	for _, id := range ids {
		work, err := QueryDataCiteObject(cfg, id, options)
		if err != nil {
			t.Error(err)
		}
		title := getObjectTitle(work)
		description := getObjectDescription(work)
		doi := getObjectDOI(work)
		publicationDate := getObjectPublicationDate(work)
		if title == "" {
			t.Errorf("expected title for %q, got empty string", id)
		}
		if description == "" {
			t.Errorf("expected description for %q, got empty string", id)
		}
		if doi == "" {
			t.Errorf("expected doi for %q, got empty string", id)
		}
		if publicationDate == "" {
			t.Errorf("expected publicationDate for %q, got empty string", id)
		}
	}
}

func TestQueryDatasetDOI(t *testing.T) {
	data := map[string]map[string]interface{}{
		"10.22002/d1.868": map[string]interface{}{
			"title": "caltechlibrary/dataset: issues #13, #14, #15 fixes",
			"author": []map[string]string{
				map[string]string{
					"literal": "Robert Doiel",
				},
			},
			"description": "This release is primarily implementing fixes.",
			"doi":         "10.22002/d1.868",
			"identifier":  "https://doi.org/10.22002/d1.868",
			"published":   "2018",
		},
	}

	cfg := new(Config)
	options := new(Doi2RdmOptions)
	options.MailTo = "test@example.edu"
	for doi, expected := range data {
		m, err := QueryDataCiteObject(cfg, doi, options)
		if err != nil {
			t.Error(err)
		}
		if m == nil || len(m) == 0 {
			t.Errorf("no data returned for doi %q", doi)
		}
		expectedS := expected["title"]
		gotS := getObjectTitle(m)
		if expectedS != gotS {
			t.Errorf("expected (%q) to get %q, got %q", doi, expectedS, gotS)
		}
		creators := getObjectCreators(m)
		if creators != nil {
			for i, creator := range creators {
				if i > 0 {
					t.Errorf("got too many authors, %+v", creators)
					break
				}
				if creator.PersonOrOrg == nil {
					t.Errorf("creator.PersonOrOrg is nil")
				} else if creator.PersonOrOrg.Name != "Robert Doiel" {
					t.Errorf("expected \"Robert Doiel\", got %q", creator.PersonOrOrg.Name)
				}
			}
		} else {
			t.Errorf("expected authors, got nil")
		}
		doiVal := getObjectDOI(m)
		if doi != doiVal {
			t.Errorf("expected DOI (%T) %q, got (%T) %q", doi, doi, doiVal, doiVal)
		}
		expectedS = "This release is primarily implementing fixes."
		gotS = getObjectDescription(m)
		if expectedS != gotS {
			t.Errorf("expected description %q, got %q", expectedS, gotS)
		}
		expectedS = "2018"
		gotS = getObjectPublicationDate(m)
		if expectedS != gotS {
			t.Errorf("expected published %q, got %q", expectedS, gotS)
		}
		expectedS = "https://doi.org/10.22002/d1.868"
		gotS = getObjectIdentifier(m)
		if expectedS != gotS {
			t.Errorf("expected identifier %q, got %q", expectedS, gotS)
		}
		/*
			src, _ := json.MarshalIndent(m, "", "    ")
			fmt.Printf("DEBUG m ->\n%s\n", src)
		*/
	}
}
