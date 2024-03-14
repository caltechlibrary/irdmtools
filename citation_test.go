package irdmtools

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"testing"

	// Caltech Library Packages
	"github.com/caltechlibrary/simplified"
)

func TestCitationCrosswalkRecord(t *testing.T) {
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

	item := &Citation{}
	if err := item.CrosswalkRecord("rdm_website", "10.5281-inveniordm.1234", "", rec); err != nil {
		t.Error(err)
	}
	expectedS := "rdm_website"
	if item.Repository != expectedS {
		t.Errorf("incorrect repository, expected %q, got %q", expectedS, item.Repository)
	}
	expectedS = "10.5281-inveniordm.1234"
	if item.RepositoryRecordID != expectedS {
		t.Errorf("incorrect repo. rec. id, expected %q, got %q", expectedS, item.RepositoryRecordID)
	}
	expectedS = "InvenioRDM"
	if item.Title != expectedS {
		t.Errorf("incorrect title, expected %q, got %q", expectedS, item.Title)
	}
	expectedS = "InvenioRDM"
	if item.Publisher != expectedS {
		t.Errorf("incorrect publisher, expected %q, got %q", expectedS, item.Publisher)
	}
	expectedS = ""
	if item.Publication != expectedS {
		t.Errorf("incorrect publication, expected %q, got %q", expectedS, item.Publication)
	}
	expectedS = "2018/2020-09"
	if item.PublicationDate != expectedS {
		t.Errorf("incorrect publication date, expected %q, got %q", expectedS, item.PublicationDate)
	}
	expectedS = ""
	if item.Publication != expectedS {
		t.Errorf("incorrect publication, expected %q, got %q", expectedS, item.Publication)
	}
	expectedS = ""
	if item.Series != expectedS {
		t.Errorf("incorrect series, expected %q, got %q", expectedS, item.Series)
	}
	expectedS = ""
	if item.SeriesNumber != expectedS {
		t.Errorf("incorrect series number, expected %q, got %q", expectedS, item.SeriesNumber)
	}
	expectedS = ""
	if item.Volume != expectedS {
		t.Errorf("incorrect volume, expected %q, got %q", expectedS, item.Volume)
	}
	expectedS = ""
	if item.Issue != expectedS {
		t.Errorf("incorrect issue, expected %q, got %q", expectedS, item.Issue)
	}
	expectedS = "https://doi.org/10.5281/inveniordm.1234"
	if item.CiteUsingURL != expectedS {
		t.Errorf("incorrect cite using url, expected %q, got %q", expectedS, item.CiteUsingURL)
	}
	expectedI := len(rec.Metadata.Creators)
	if len(item.Author) != expectedI {
		t.Errorf("author count wrong, expected %d, got %d", expectedI, len(item.Author))
	}
	expectedI = 0
	if len(item.Editor) != expectedI {
		t.Errorf("editor count wrong, expected %d, got %d", expectedI, len(item.Editor))
	}
	expectedI = 0
	if len(item.ThesisAdvisor) != expectedI {
		t.Errorf("thesis advisor count wrong, expected %d, got %d", expectedI, len(item.ThesisAdvisor))
	}
	expectedI = 0
	if len(item.ThesisCommittee) != expectedI {
		t.Errorf("thesis committee count wrong, expected %d, got %d", expectedI, len(item.ThesisCommittee))
	}
	expectedI = 0
	if len(item.Reviewer) != expectedI {
		t.Errorf("reviewers count wrong, expected %d, got %d", expectedI, len(item.Reviewer))
	}
	expectedI = 1
	if len(item.Contributor) != expectedI {
		t.Errorf("contributor count wrong, expected %d, got %d -> %s", expectedI, len(item.Contributor), func(contributor []*CitationAgent) string {
			src, _ := json.Marshal(contributor)
			return string(src)
		}(item.Contributor))
	}
}

func TestCrosswalkCreatorToCitationAgent(t *testing.T) {
	src := []byte(`[
  {
    "person_or_org": {
      "given_name": "Fayth Hui",
      "type": "personal",
      "identifiers": [
        {
          "identifier": "Tan-Fayth-Hui",
          "scheme": "clpid"
        },
        {
          "identifier": "0000-0002-2160-5311",
          "scheme": "orcid"
        }
      ],
      "family_name": "Tan",
      "name": "Tan, Fayth Hui"
    }
  },
  {
    "affiliations": [
      {
        "id": "05dxps055",
        "name": "California Institute of Technology"
      }
    ],
    "person_or_org": {
      "given_name": "Marianne E.",
      "type": "personal",
      "identifiers": [
        {
          "identifier": "Bronner-M-E",
          "scheme": "clpid"
        },
        {
          "identifier": "0000-0003-4274-1862",
          "scheme": "orcid"
        }
      ],
      "family_name": "Bronner",
      "name": "Bronner, Marianne E."
    }
  }
]
`)
	creatorList := []*simplified.Creator{}
	if err := json.Unmarshal(src, &creatorList); err != nil {
		t.Errorf("failed to unmarshal test data, %s", err)
		t.FailNow()
	}
	expectedPeople := []*CitationAgent{}
	src = []byte(`[
	{
	      "given_name": "Fayth Hui",
	      "family_name": "Tan",
	      "clpid": "Tan-Fayth-Hui",
          "orcid": "0000-0002-2160-5311"
    },
	{
	      "given_name": "Marianne E.",
	      "family_name": "Bronner",
          "clpid": "Bronner-M-E",
          "orcid": "0000-0003-4274-1862"
    }
]`)
	if err := json.Unmarshal(src, &expectedPeople); err != nil {
		t.Errorf("failed to unmarshal expected data, %s", err)
		t.FailNow()
	}
	for i := 0; i < len(creatorList); i++ {
		agent, role, err := CrosswalkCreatorToCitationAgent(creatorList[i])
		if err != nil {
			t.Error(err)
		}
		expectedRole := ""
		if role != "" {
			t.Errorf("expected role %q, got %q", expectedRole, role)
		}

		expected := expectedPeople[i]
		if expected.FamilyName != agent.FamilyName {
			t.Errorf("expected agent family %q, got %q", expected.FamilyName, agent.FamilyName)
		}
		if expected.LivedName != agent.LivedName {
			t.Errorf("expected agent lived %q, got %q", expected.LivedName, agent.LivedName)
		}
		if expected.ORCID != agent.ORCID {
			t.Errorf("expected agent orcid %q, got %q", expected.ORCID, agent.ORCID)
		}
		if expected.ISNI != agent.ISNI {
			t.Errorf("expected agent ISNI %q, got %q", expected.ISNI, agent.ISNI)
		}
		if expected.CLpid != agent.CLpid {
			t.Errorf("expected agent CL Pid %q got %q", expected.CLpid, agent.CLpid)
		}
	}
	// Now test contributors
	contributorList := []*simplified.Creator{
		&simplified.Creator{
			PersonOrOrg: &simplified.PersonOrOrg{
				Name: "Nielsen, Lars Holm",
				FamilyName: "Nielsen",
				GivenName: "Lars Holm",
				Type: "person",
				Identifiers: []*simplified.Identifier{
					&simplified.Identifier{
						Scheme: "orcid",
						Identifier: "0000-0001-8135-3489",
					},
					&simplified.Identifier{
						Scheme: "clpid",
						Identifier: "Nielsen-Lars-Holm",
					},
				},
			},
			Role: &simplified.Role{
				ID: "editor",
			},
			Affiliations: []*simplified.Affiliation{
				&simplified.Affiliation{
					ID: "01ggx415",
					Name: "CERN",
				},
			},
		},
	}
	expectedPeople = []*CitationAgent{
		&CitationAgent{
			FamilyName: "Nielsen",
			LivedName:  "Lars Holm",
			ORCID:  "0000-0001-8135-3489",
			CLpid:  "Nielsen-Lars-Holm",
		},
	}
	expectedRole := "editor"
	for i, creator := range contributorList {
		agent, role, err := CrosswalkCreatorToCitationAgent(creator)
		if err != nil {
			t.Error(err)
		}
		if expectedRole != role {
			t.Errorf("expected role %q, got %q", expectedRole, role)
		}
		expected := expectedPeople[i]
		if expected.FamilyName != agent.FamilyName {
			t.Errorf("expected agent family %q, got %q", expected.FamilyName, agent.FamilyName)
		}
		if expected.LivedName != agent.LivedName {
			t.Errorf("expected agent lived %q, got %q", expected.LivedName, agent.LivedName)
		}
		if expected.ORCID != agent.ORCID {
			t.Errorf("expected agent orcid %q, got %q", expected.ORCID, agent.ORCID)
		}
		if expected.ISNI != agent.ISNI {
			t.Errorf("expected agent ISNI %q, got %q", expected.ISNI, agent.ISNI)
		}
		if expected.CLpid != agent.CLpid {
			t.Errorf("expected agent CL Pid %q got %q", expected.CLpid, agent.CLpid)
		}
	}
}

func TestCrosswalkPersonOrOrgToCitationAgent(t *testing.T) {
	src := []byte(`[
	  {
	      "given_name": "Fayth Hui",
	      "identifiers": [
	        {
	          "identifier": "Tan-Fayth-Hui",
	          "scheme": "clpid"
	        },
	        {
	          "identifier": "0000-0002-2160-5311",
	          "scheme": "orcid"
	        }
	      ],
	      "family_name": "Tan",
	      "name": "Tan, Fayth Hui"
	  },
	  {
	      "given_name": "Marianne E.",
	      "identifiers": [
	        {
	          "identifier": "Bronner-M-E",
	          "scheme": "clpid"
	        },
	        {
	          "identifier": "0000-0003-4274-1862",
	          "scheme": "orcid"
	        }
	      ],
	      "family_name": "Bronner"
	  }
	]`)
	personList := []*simplified.PersonOrOrg{}
	if err := json.Unmarshal(src, &personList); err != nil {
		t.Errorf("failed to unmarshal test data, skipping %s", err)
		t.FailNow()
	}
	src = []byte(`[
	{
	      "given_name": "Fayth Hui",
	      "family_name": "Tan",
	      "clpid": "Tan-Fayth-Hui",
          "orcid": "0000-0002-2160-5311"
    },
	{
	      "given_name": "Marianne E.",
	      "family_name": "Bronner"
          "clpid": "Bronner-M-E",
          "orcid": "0000-0003-4274-1862",
    }
]`)
	expectedPeople := []*CitationAgent{}
	if err := json.Unmarshal(src, &expectedPeople); err != nil {
		t.Errorf("failed to unmarshal expected data, skipping %s", err)
		t.FailNow()
	}
	for i, creator := range personList {
		agent, err := CrosswalkPersonOrOrgToCitationAgent(creator)
		if err != nil {
			t.Error(err)
		}
		expected := expectedPeople[i]
		if expected.FamilyName != agent.FamilyName {
			t.Errorf("expected agent family %q, got %q", expected.FamilyName, agent.FamilyName)
		}
		if expected.LivedName != agent.LivedName {
			t.Errorf("expected agent lived %q, got %q", expected.LivedName, agent.LivedName)
		}
		if expected.ORCID != agent.ORCID {
			t.Errorf("expected agent orcid %q, got %q", expected.ORCID, agent.ORCID)
		}
		if expected.ISNI != agent.ISNI {
			t.Errorf("expected agent ISNI %q, got %q", expected.ISNI, agent.ISNI)
		}
		if expected.CLpid != agent.CLpid {
			t.Errorf("expected agent CL Pid %q got %q", expected.CLpid, agent.CLpid)
		}
	}
}
