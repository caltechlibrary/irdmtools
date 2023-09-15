// irdmtools is a package for working with institutional repositories and
// data management systems. Current implementation targets Invenio-RDM.
//
// @author R. S. Doiel, <rsdoiel@caltech.edu>
// @author Tom Morrell, <tmorrell@caltech.edu>
//
// Copyright (c) 2023, Caltech
// All rights not granted herein are expressly reserved by Caltech.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
package irdmtools

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	// Caltech Library Packages
	"github.com/caltechlibrary/dataset/v2"
	"github.com/caltechlibrary/eprinttools"
	"github.com/caltechlibrary/simplified"
)

const (
	timestamp = `2006-01-02 15:04:05`
	datestamp = `2006-01-02`
)

// EPrint2Rdm holds the configuration for rdmutil cli.
type EPrint2Rdm struct {
	Cfg *Config
}

// EPrintKeysPage holds the structure of the HTML page with the
// EPrint IDs embedded from the EPrint REST API.
type EPrintKeysPage struct {
	XMLName xml.Name `xml:"html"`
	Anchors []string `xml:"body>ul>li>a"`
}

var (
	defaultEPrintContributorTypeMap = map[string]string{
		// Caltech Library Roles

		// Author of section
		"http://coda.library.caltech.edu/ARA": "author_section",
		// Astronaut
		"http://coda.library.caltech.edu/AST": "astronaut",

		// LOC Terms and URI

		// Author of afterword, colophon, etc.
		"http://www.loc.gov/loc.terms/relators/AFT": "other",
		// Bibliographic antecedent
		"http://www.loc.gov/loc.terms/relators/ANT": "other",
		// Author in quotations or text abstracts
		"http://www.loc.gov/loc.terms/relators/AQT": "other",
		// Screenwriter or co-Screenwriter
		"http://www.loc.gov/loc.terms/relators/AUS": "screenwriter",
		// Author, co-Author
		"http://www.loc.gov/loc.terms/relators/AUT": "author",
		// Collaborator
		"http://www.loc.gov/loc.terms/relators/CLB": "collaborator",
		// Compiler
		"http://www.loc.gov/loc.terms/relators/COM": "compiler",
		// Contributor
		"http://www.loc.gov/loc.terms/relators/CTB": "contributor",
		// Director (e.g. Movie, theatre)
		"http://www.loc.gov/loc.terms/relators/DRT": "director",
		// Editor
		"http://www.loc.gov/loc.terms/relators/EDT": "editor",
		// Narrator
		"http://www.loc.gov/loc.terms/relators/NRT": "narrator",
		// Other
		"http://www.loc.gov/loc.terms/relators/OTH": "other",
		// Publishing Directory (publications)
		"http://www.loc.gov/loc.terms/relators/PBD": "publishing_directory",
		// Programming
		"http://www.loc.gov/loc.terms/relators/PRG": "programmer",
		// Reviewer
		"http://www.loc.gov/loc.terms/relators/REV": "reviewer",
		// Research Team Member
		"http://www.loc.gov/loc.terms/relators/RTM": "research_team",
		// Speaker
		"http://www.loc.gov/loc.terms/relators/SPK": "speaker",
		// Teacher
		"http://www.loc.gov/loc.terms/relators/TCH": "teacher",
		// Translator
		"http://www.loc.gov/loc.terms/relators/TRL": "translator",

		// LOC Term URI to DataCite terms

		// Process Contact -> ContactPerson
		"https://www.loc.gov/loc.terms/relators/prc": "contact_person",
		// DataCollector <- Data Contributor
		"https://www.loc.gov/loc.terms/relators/dtc": "data_collector",
		// DataCurator <- ???
		//"https://www.loc.gov/loc.terms/relators/???": "data_curator",
		// DataManager -> DataManager
		"https://www.loc.gov/loc.terms/relators/dtm": "data_manager",
		// Distributor -> Distrubtor
		"https://www.loc.gov/loc.terms/relators/dst": "distributor",
		// Host Institution -> HostingInstitution
		"https://www.loc.gov/loc.terms/relators/his": "hosting_institution",
		// Producer -> Producer
		"https://www.loc.gov/loc.terms/relators/pro": "producer",
		// ??? -> Project Leader
		// "https://www.loc.gov/loc.terms/relators/pdr": "project_leader",
		// Project Directory -> Project Manager
		"https://www.loc.gov/loc.terms/relators/pdr": "project_manager",
		// ??? -> ProjectMember
		// ??? -> RegistrationAgency
		// ??? -> RegistrationAuthority
		// ??? -> RelatedPerson
		// Researcher -> Researcher
		"https://www.loc.gov/loc.terms/relators/res": "researcher",
		// ??? -> ResearchGroup
		// Copyright Holder -> RightsHolder
		"https://www.loc.gov/loc.terms/relators/cph": "rights_holder",
		// Sponsor -> Sponsor
		"https://www.loc.gov/loc.terms/relators/spn": "sponsor",
		// ??? -> Supervisor
		// ??? -> WorkPackageLeader

		// Additional LOC -> Caltech Library RDM person or org types

		// Interviewee -> interviewee
		"https://www.loc.gov/loc.terms/relators/ive": "interviewee",
		// Interviewer -> interviewer
		"https://www.loc.gov/loc.terms/relators/ivr": "interviewer",
	}

	defaultEPrintResourceTypeMap = map[string]string{
		"article":           "publication-article",
		"journal-article":   "publication-article",
		"book":              "publication-book",
		"book_section":      "publication-section",
		"conference_item":   "conference-paper",
		"dataset":           "dataset",
		"experiment":        "publication-deliverable",
		"journal_issue":     "publication-issue",
		"lab_notes":         "labnotebook",
		"monograph":         "publication-report",
		"oral_history":      "publication-oralhistory",
		"patent":            "publication-patent",
		"software":          "software",
		"teaching_resource": "teachingresource",
		"thesis":            "publication-thesis",
		"video":             "video",
		// This following will need to be updated and added to RDM
		"geospatial_resource": "other",
		"website":             "other",
		"other":               "other",
		"image":               "other",
	}
)

/**
 * Implements a crosswalk from EPrints to an Invenio RDM JSON
 * representation.
 *
 * See documentation and example on Invenio's structured data:
 *
 * - https://inveniordm.docs.cern.ch/reference/metadata/
 * - https://github.com/caltechlibrary/caltechdata_api/blob/ce16c6856eb7f6424db65c1b06de741bbcaee2c8/tests/conftest.py#L147
 *
 */

// Convert eprint.PubDate() from ends with -00 or -00-00 to
// -01-01 or -01 respectively so they validate in RDM.
func normalizeEPrintDate(s string) string {
	// Normalize eprint.Date to something sensible
	if len(s) > 10 {
		s = s[0:10]
	}
	if strings.HasSuffix(s, "-00") {
		if strings.HasSuffix(s, "-00-00") {
			s = strings.Replace(s, "-00-00", "-01-01", 1)
		} else {
			s = strings.Replace(s, "-00", "-01", 1)
		}
	}
	return s
}

func listMapHasID(l []map[string]string, target string) bool {
	for _, item := range l {
		if id, ok := item["id"]; ok && id == target {
			return true
		}
	}
	return false
}


func normalizeThesisType(s string) string {
	switch s {
	case	"phd":
		return "PhD"
	default:
		if len(s) > 1 {
			parts := strings.SplitN(s, "", 2)
			return fmt.Sprintf("%s%s", strings.ToUpper(parts[0]), strings.ToLower(parts[1]))
		}
	}
	return s
}
func customFieldsMetadataFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	/*NOTE: Custom fields for Journal data
		"custom_fields": {
			"journal:journal": {
				"issue": "7",
				"pages": "15-23",
				"title": "Nature",
				"volume": "645"
			},
			"imprint:imprint": {
				"isbn": "978-3-16-148410-0",
	            "pages": "12-15",
	            "title": "Book title",
	            "place": "Location, place"
			}
		},
	*/
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.CustomFields == nil {
		rec.CustomFields = map[string]interface{}{}
	}
	// NOTE: handle thesis type including capitalization of types.
	if eprint.Type == "article" && eprint.Publication != "" {
		if err := SetJournalField(rec, "title", eprint.Publication); err != nil {
			return err
		}
		if err := SetJournalField(rec, "issue", eprint.Number); err != nil {
			return err
		}
		if err := SetJournalField(rec, "pages", eprint.PageRange); err != nil {
			return err
		}
		if err := SetJournalField(rec, "volume", eprint.Volume); err != nil {
			return err
		}
		if eprint.Publisher != "" {
			rec.Metadata.Publisher = eprint.Publisher
		}
		if eprint.Series != "" {
			if err := SetJournalField(rec, "series", eprint.Series); err != nil {
				return err
			}
		}
		if eprint.ISSN != "" {
			if err := SetJournalField(rec, "issn", eprint.ISSN); err != nil {
				return err
			}
		}
	}
	if eprint.OtherNumberingSystem != nil && eprint.OtherNumberingSystem.Length() > 0 {
		for i := 0; i < eprint.OtherNumberingSystem.Length(); i++ {
			item := eprint.OtherNumberingSystem.IndexOf(i)
			if item.Name != nil && item.Name.Value != "" {
				SetCustomField(rec, "caltech:other_num_name", "", item.Name.Value)
			}
			if item.ID != "" {
				SetCustomField(rec, "caltech:other_num_id", "", item.ID)
			}
		}
	}
	if eprint.Series != "" && eprint.ISBN == "" {
		SetCustomField(rec, "caltech:series", "series", eprint.Series)
	}
	if eprint.PlaceOfPub != "" && eprint.ISBN == "" {
		SetCustomField(rec, "caltech:place_of_publication", "", eprint.PlaceOfPub)
	}
	// NOTE: handle thesis type including capitalization of types.
	if eprint.Type == "thesis" {
		val := map[string]interface{}{
			"type": normalizeThesisType(eprint.ThesisType),
			"unversity": eprint.Institution,
			"department": eprint.Department,
		}
		SetCustomField(rec, "thesis:thesis", "", val)
	}
	// NOTE: handle "local_group" mapped from eprint_local_group table.
	if eprint.LocalGroup != nil && eprint.LocalGroup.Length() > 0 {
		groups := []map[string]string{}
		for i := 0; i < eprint.LocalGroup.Length(); i++ {
			localGroup := eprint.LocalGroup.IndexOf(i)
			if strings.TrimSpace(localGroup.Value) != "" {
				m := map[string]string{}
				m["id"] = strings.ReplaceAll(localGroup.Value, " ", "-")
				if !listMapHasID(groups, localGroup.Value) {
					groups = append(groups, m)
				}
			}
		}
		SetCustomField(rec, "caltech:groups", "", groups)
	}

	// NOTE: handle "event" case, issue #13
	if eprint.EventType != "" || eprint.EventTitle != "" ||
		eprint.EventLocation != "" || eprint.EventDates != "" {
		m := map[string]string{}
		if eprint.EventType != "" {
			m["type"] = eprint.EventType
		}
		if eprint.EventTitle != "" {
			m["title"] = eprint.EventTitle
		}
		if eprint.EventLocation != "" {
			m["place"] = eprint.EventLocation
		}
		if eprint.EventDates != "" {
			m["dates"] = eprint.EventDates
		}
		SetCustomField(rec, "meeting:meeting", "", m)
	}

	// NOTE: Suggests needs to be mapped to a custom field "caltech:internal-note"
	if eprint.Suggestions != "" {
		SetCustomField(rec, "caltech:internal_note", "", eprint.Suggestions)
	}

	// NOTE: Mapping subjects and keyswords to .metadata.subjects. We're handling it hear
	// instread of simplified.go because I need to gaurantee duplicate subjects don't get added
	// as part of the merging of keywords and subjects from EPrints.
	if eprint.Keywords != "" || (eprint.Subjects != nil && eprint.Subjects.Length() > 0) {
		subjectStrings := []string{}
		if eprint.Keywords != "" {
			keywords := strings.Split(eprint.Keywords, ";")
			for _, keyword := range keywords {
				val := strings.TrimSpace(keyword)
				if val != "" && val != "cls" {
					subjectStrings = append(subjectStrings, val)
				}
			}
		}
		for i := 0; i < eprint.Subjects.Length(); i++ {
			subject := eprint.Subjects.IndexOf(i)
			//NOTE: irdmtools issue #51, ignore cls as a subject, this was an EPrints-ism
			// needed for Caltech Library only.
			val := strings.TrimSpace(subject.Value)
			if val != "" && val != "cls" {
				subjectStrings = append(subjectStrings, val)
			}
		}
		if len(subjectStrings) > 0 {
			duplicates := map[string]bool{}
			for _, subject := range subjectStrings {
				if _, duplicate := duplicates[subject]; ! duplicate {
					AddSubject(rec, subject)
				}
				duplicates[subject] = true
			}
		}
	}
	return nil
}

// CrosswalkEPrintToRecord implements a crosswalk between
// an EPrint 3.x EPrint XML record as struct to a Invenio RDM
// record as struct.
func CrosswalkEPrintToRecord(eprint *eprinttools.EPrint, rec *simplified.Record, resourceTypes map[string]string, contributorTypes map[string]string) error {
	rec.Schema = `local://records/record-v2.0.0.json`
	rec.ID = fmt.Sprintf("%s:%d", eprint.Collection, eprint.EPrintID)

	// NOTE: If an eprint is "deleted" we need to render a tombsone record then return
	if eprint.EPrintStatus == "deletion" {
		if err := tombstoneFromEPrint(eprint, rec); err != nil {
			return err
		}
		return nil
	}

	if err := parentFromEPrint(eprint, rec); err != nil {
		return err
	}
	// NOTE: externalPIDFromEPrint needs to happen before called metdataFromEPrint
	if err := externalPIDFromEPrint(eprint, rec); err != nil {
		return err
	}
	if err := recordAccessFromEPrint(eprint, rec); err != nil {
		return err
	}

	if err := metadataFromEPrint(eprint, rec, contributorTypes); err != nil {
		return err
	}
	// NOTE: journal:journal is in CustomFields map as a submap
	if err := customFieldsMetadataFromEPrint(eprint, rec); err != nil {
		return err
	}
	if err := filesFromEPrint(eprint, rec); err != nil {
		return err
	}

	if err := createdUpdatedFromEPrint(eprint, rec); err != nil {
		return err
	}
	// Now finish simple record normalization ...
	if err := mapResourceType(eprint, rec, resourceTypes); err != nil {
		return err
	}
	if err := simplifyCreators(eprint, rec); err != nil {
		return err
	}
	if err := simplifyContributors(eprint, rec, contributorTypes); err != nil {
		return err
	}
	if err := simplifyFunding(eprint, rec); err != nil {
		return err
	}
	return nil
}

func itemToPersonOrOrg(item *eprinttools.Item) *simplified.PersonOrOrg {
	var (
		clPeopleID string
		orcid      string
		ror        string
	)
	person := new(simplified.PersonOrOrg)
	person.Type = "personal"
	if item.Name != nil {
		person.FamilyName = item.Name.Family
		person.GivenName = item.Name.Given
		// FIXME: How do we map honorific and lineage? E.g. William Goddard versus William A. Goddard III
		//honorific := item.Name.Honourific
		//lineage := item.Name.Lineage
		if person.FamilyName != "" || person.GivenName != "" {
			person.Name = fmt.Sprintf("%s, %s", person.FamilyName, person.GivenName)
			clPeopleID = item.ID
			if item.ORCID != "" {
				orcid = item.ORCID
			} else {
				orcid = item.Name.ORCID
			}
			
		} else {
			person.Type = "organizational"
			person.Name = item.Name.Value
		}
	}
	ror = item.ROR
	if clPeopleID != "" {
		person.Identifiers = append(person.Identifiers, mkSimpleIdentifier("clpid", clPeopleID))
	}
	if orcid != "" {
		person.Identifiers = append(person.Identifiers, mkSimpleIdentifier("orcid", orcid))
	}
	if ror != "" {
		person.Identifiers = append(person.Identifiers, mkSimpleIdentifier("ror", ror))
	}
	return person
}

// simplifyCreators make sure the identifiers are mapped to Invenio-RDM
// identifiers.
func simplifyCreators(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	// First map the creators (personal) to RDM .metadata.creators
	// FIXME: Then map the corpCreators (organizational) to RDM .metadata.creators
	creators := []*simplified.Creator{}
	if eprint.Creators != nil && eprint.Creators.Length() > 0 {
		for i := 0; i < eprint.Creators.Length(); i++ {
			if item := eprint.Creators.IndexOf(i); item != nil && item.Name != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					creators = append(creators, &simplified.Creator{
						PersonOrOrg: person,
					})
				}
			}
		}
	}
	if eprint.CorpCreators != nil && eprint.CorpCreators.Length() > 0 {
		for i := 0; i < eprint.CorpCreators.Length(); i++ {
			if item := eprint.CorpCreators.IndexOf(i); item != nil {
				if org := itemToPersonOrOrg(item); org != nil {
					creators = append(creators, &simplified.Creator{
						PersonOrOrg: org,
					})
				}
			}
		}
	}
	// NOTE: If there are no creators AND their are editors then I need
	// to create the editor as a creator and add them here instead of
	// in contributors. This is related to issue #9.
	if len(creators) == 0 && eprint.Editors.Length() > 0 {
		for i := 0; i < eprint.Editors.Length(); i++ {
			if item := eprint.Editors.IndexOf(i); item != nil && item.Name != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					creators = append(creators, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							ID: "editor",
						},
					})
				}
			}
		}
	}
	if len(creators) > 0 {
		rec.Metadata.Creators = creators
	}
	return nil
}

// simplifyContributors make sure the identifiers are mapped to Invenio-RDM
// identifiers.
func simplifyContributors(eprint *eprinttools.EPrint, rec *simplified.Record, contributorTypes map[string]string) error {
	contributors := []*simplified.Creator{}
	// First add contributors, then editors, etc.
	if eprint.Contributors != nil && eprint.Contributors.Length() > 0 {
		for i := 0; i < eprint.Contributors.Length(); i++ {
			if item := eprint.Contributors.IndexOf(i); item != nil && item.Name != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							Title: map[string]string{
								"en": uriToContributorType(item.Role, contributorTypes),
							},
						},
					})
				}
			}
		}
	}
	if eprint.CorpContributors != nil && eprint.CorpContributors.Length() > 0 {
		for i := 0; i < eprint.Contributors.Length(); i++ {
			if item := eprint.Contributors.IndexOf(i); item != nil {
				if org := itemToPersonOrOrg(item); org != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: org,
						Role: &simplified.Role{
							Title: map[string]string{
								"en": "contributor",
							},
						},
					})
				}
			}
		}
	}
	// Add Editors, Adivisors, Committee Members, Thesis Chair, Reviewers, Translators, etc.
	if eprint.Editors != nil && eprint.Editors.Length() > 0 {
		for i := 0; i < eprint.Editors.Length(); i++ {
			if item := eprint.Editors.IndexOf(i); item != nil && item.Name != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							Title: map[string]string{
								"en": "editor",
							},
						},
					})
				}
			}
		}
	}
	if eprint.ThesisAdvisor != nil && eprint.ThesisAdvisor.Length() > 0 {
		for i := 0; i < eprint.ThesisAdvisor.Length(); i++ {
			if item := eprint.ThesisAdvisor.IndexOf(i); item != nil  && item.Name != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							Title: map[string]string{
								"en": "thesis_advisor",
							},
						},
					})
				}
			}
		}
	}
	if eprint.ThesisCommittee != nil && eprint.ThesisCommittee.Length() > 0 {
		for i := 0; i < eprint.ThesisCommittee.Length(); i++ {
			if item := eprint.ThesisCommittee.IndexOf(i); item != nil && item.Name != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							Title: map[string]string{
								"en": "thesis_committee",
							},
						},
					})
				}
			}
		}
	}
	if len(contributors) > 0 {
		rec.Metadata.Contributors = contributors
	}
	return nil
}

func simplifyFunding(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	if rec.Metadata.Funding != nil && len(rec.Metadata.Funding) > 0 {
		for _, funder := range rec.Metadata.Funding {
			if funder.Funder != nil {
				funder.Funder.Scheme = ""
			}
			if funder.Award != nil {
				if funder.Award.Number == "" {
					funder.Award = nil
				} else {
					//NOTE: funder.Award.Title is a struct in
					// Invenio-RDM like
					// ```
					//   title : { "lang": "en", "unavailable" }
					// ```
					// This needs to be normalized in the final
					// Python processing for importing into Invenio-RDM.
					funder.Award.Scheme = ""
				}
			}
		}
	}
	return nil
}

// LoadTypesMap map parses a CSV file where first column is the EPrint value
// and second colume is the Ivenio RDM type string. If the file cannot be
// read or parsed it will return an error otherwise the map[string]string
// value will be updated with the contents of the mapping.
//
// ```
// resourceTypes := map[string]string{}
// if err := LoadTypesMap("resource-types.csv", resourceTypes); err != nil {
//	// ... handle error or continie processing...
// }
//
// contribTypes := map[string]string{}
// if err := LoadTypesMap("contrib-types.csv", contribTypes); err != nil {
//	// ... handle error or continie processing...
// }

// ```
func LoadTypesMap(fName string, mapTypes map[string]string) error {
	if mapTypes == nil {
		mapTypes = map[string]string{}
	}
	src, err := os.ReadFile(fName)
	if err != nil {
		return err
	}
	reader := csv.NewReader(bytes.NewBuffer(src))
	table, err := reader.ReadAll()
	if err != nil {
		return err
	}
	for _, row := range table {
		if len(row) == 2 {
			mapTypes[row[0]] = row[1]
		}
	}
	return err
}

// mapResourceType maps the EPrints record types to a predetermined
// Invenio-RDM record type.
func mapResourceType(eprint *eprinttools.EPrint, rec *simplified.Record, resourceTypesMap map[string]string) error {
	if rec.Metadata.ResourceType == nil {
		rec.Metadata.ResourceType = make(map[string]interface{})
	}
	if len(resourceTypesMap) == 0 {
		// NOTE: This is a default map used of no resource type map is provided.
		for k, v := range defaultEPrintResourceTypeMap {
			resourceTypesMap[k] = v
		}
	}
	val, ok := resourceTypesMap[eprint.Type]
	if !ok {
		return fmt.Errorf("unable to map eprint.type to %q RDM resource type", eprint.Type)
	}
	rec.Metadata.ResourceType["id"] = val
	return nil
}

// parentFromEPrint crosswalks the Perent unique ID from EPrint record.
func parentFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	if eprint.Reviewer != "" {
		parent := new(simplified.RecordIdentifier)
		parent.ID = fmt.Sprintf("%s:%d", eprint.Collection, eprint.EPrintID)
		parent.Access = new(simplified.Access)
		ownedBy := new(simplified.User)
		ownedBy.User = eprint.UserID
		ownedBy.DisplayName = eprint.Reviewer
		parent.Access.OwnedBy = append(parent.Access.OwnedBy, ownedBy)
		rec.Parent = parent
	} else {
		rec.Parent = nil
	}
	return nil
}

// externalPIDFromEPrint aggregates all the external identifiers
// from the EPrint record into Record
func externalPIDFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	rec.ExternalPIDs = map[string]*simplified.PersistentIdentifier{}
	// Pickup DOI
	if eprint.DOI != "" {
		pid := new(simplified.PersistentIdentifier)
		pid.Identifier = eprint.DOI
		pid.Provider = "external"
		pid.Client = ""
		rec.ExternalPIDs["doi"] = pid
	}
	return nil
}

// recordAccessFromEPrint extracts access permissions from the EPrint
func recordAccessFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	isPublic := true
	if (eprint.ReviewStatus == "review") ||
		(eprint.ReviewStatus == "withheld") ||
		(eprint.ReviewStatus == "gradoffice") ||
		(eprint.ReviewStatus == "notapproved") {
		isPublic = false
	}
	if eprint.EPrintStatus != "archive" || eprint.MetadataVisibility != "show" {
		isPublic = false
	}
	rec.RecordAccess = new(simplified.RecordAccess)
	// By default lets assume the files are public.
	rec.RecordAccess.Files = "public"
	if isPublic {
		rec.RecordAccess.Record = "public"
	} else {
		rec.RecordAccess.Record = "restricted"
	}
	// Need to make sure record is not embargoed
	if eprint.Documents != nil {
		for i := 0; i < eprint.Documents.Length(); i++ {
			doc := eprint.Documents.IndexOf(i)
			if doc.Security == "internal" || doc.Security == "validuser" {
				rec.RecordAccess.Files = "restricted"
			}
			if doc.DateEmbargo != "" {
				embargo := new(simplified.Embargo)
				embargo.Until = doc.DateEmbargo
				if eprint.Suggestions != "" {
					embargo.Reason = eprint.Suggestions
				}
				if doc.Security == "internal" {
					embargo.Active = true
				} else {
					embargo.Active = false
				}
				rec.RecordAccess.Embargo = embargo
				break
			}
		}
	}
	return nil
}

func uriToContributorType(role_uri string, contributorTypes map[string]string) string {
	if len(contributorTypes) == 0 {
		for k, v := range defaultEPrintContributorTypeMap {
			contributorTypes[k] = v
		}
	}
	if val, ok := contributorTypes[role_uri]; ok {
		return val
	}
	// FIXME: The default mapping is "other" since we don't know what it should be.
	// Per slack conversation but I'm using "unknown" to confirm that the resource mapping
	// is happening.
	return "other"
}

func dateTypeFromTimestamp(dtType string, timestamp string, description string) *simplified.DateType {
	dt := new(simplified.DateType)
	dt.Type = new(simplified.Type)
	dt.Type.ID = dtType
	dt.Type.Title = map[string]string{
		"en": dtType,
	}
	dt.Description = description
	if len(timestamp) > 9 {
		dt.Date = timestamp[0:10]
	} else {
		dt.Date = timestamp
	}
	return dt
}

func mkSimpleIdentifier(scheme string, value string) *simplified.Identifier {
	identifier := new(simplified.Identifier)
	identifier.Scheme = scheme
	identifier.Identifier = value
	return identifier
}

func funderFromItem(item *eprinttools.Item) *simplified.Funder {
	funder := new(simplified.Funder)
	if item.GrantNumber != "" {
		funder.Award = new(simplified.AwardIdentifier)
		funder.Award.Number = item.GrantNumber
		funder.Award.Scheme = "eprints_grant_number"
	}
	if item.Agency != "" {
		org := new(simplified.FunderIdentifier)
		org.Name = item.Agency
		org.Scheme = "eprints_agency"
		funder.Funder = org
	}
	return funder
}

// metadataFromEPrint extracts metadata from the EPrint record
func metadataFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record, contributorTypes map[string]string) error {
	// NOTE: Creators get listed in the citation, Contributors do not.
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	rec.Metadata.Title = eprint.Title
	if (eprint.AltTitle != nil) && (eprint.AltTitle.Items != nil) {
		for _, item := range eprint.AltTitle.Items {
			if strings.TrimSpace(item.Value) != "" {
				title := new(simplified.TitleDetail)
				title.Title = item.Value
				rec.Metadata.AdditionalTitles = append(rec.Metadata.AdditionalTitles, title)
			}
		}
	}
	if err := simplifyCreators(eprint, rec); err != nil {
		return err
	}
	if err := simplifyContributors(eprint, rec, contributorTypes); err != nil {
		return err
	}
	if eprint.Abstract != "" {
		rec.Metadata.Description = eprint.Abstract
	}

	// Rights are scattered in several EPrints fields, they need to
	// be evaluated to create a "Rights" object used in DataCite/Invenio
	if eprint.Rights != "" {
		rights := new(simplified.Right)
		rights.Title = map[string]string{
			"en": "Other",
		}
		rights.Description = map[string]string{
			"en": eprint.Rights,
		}
		rec.Metadata.Rights = append(rec.Metadata.Rights, rights)
	}
	// Figure out if our copyright information is in the Note field.
	/*
	// NOTE: Removed this mapping based on issue 70 in 
	// caltechlibrary/caltechauthors repo.
	if (eprint.Note != "") && (strings.Contains(eprint.Note, "©") || strings.Contains(eprint.Note, "copyright") || strings.Contains(eprint.Note, "(c)")) {
		addRights = true
		m := map[string]string{
			"en": fmt.Sprintf("%s", eprint.Note),
		}
		rights.Description = m
	}
	*/
	// FIXME: work with Tom to sort out how "Rights" and document level
	// copyright info should work.
	if eprint.CopyrightStatement != "" {
		rights := new(simplified.Right)
		rights.Title = map[string]string{
			"en": "Other",
		}
		rights.Description = map[string]string{
			"en": eprint.CopyrightStatement,
		}
		rec.Metadata.Rights = append(rec.Metadata.Rights, rights)
	}


	// FIXME: Work with Tom to figure out correct mapping of rights from EPrints XML
	// FIXME: Language appears to be at the "document" level, not record level

	// NOTE: RDM Requires a publication date
	// Default to the eprint.Datestamp and correct if DateType is "published"
	rec.Metadata.PublicationDate = normalizeEPrintDate(eprint.Datestamp)

	// NOTE: We have a few records that have NULL or empty string 
	// eprint.date_type fields. These all appear like they should have
	// the value in eprint.date treated as the publication date if it is
	// publicated. 
	// See https://github.com/caltechlibrary/caltechauthors/issues/75
	if (eprint.DateType == "published" || eprint.DateType == "") && (eprint.Date != "") {
		rec.Metadata.Dates = append(rec.Metadata.Dates, dateTypeFromTimestamp("pub_date", eprint.Date, "EPrint's Publication Date"))
		rec.Metadata.PublicationDate = normalizeEPrintDate(eprint.Date)
	}
	// Handle case of overloaded date time from EPrints
	if (eprint.DateType != "") && (eprint.Date != "") {
		rec.Metadata.Dates = append(rec.Metadata.Dates, dateTypeFromTimestamp(eprint.DateType, eprint.Date, "Created from EPrint's date_type and date field"))
	}
	if eprint.Datestamp != "" {
		rec.Metadata.Dates = append(rec.Metadata.Dates, dateTypeFromTimestamp("created", eprint.Datestamp, "Created from EPrint's datestamp field"))
	}
	if eprint.LastModified != "" {
		rec.Metadata.Dates = append(rec.Metadata.Dates, dateTypeFromTimestamp("updated", eprint.LastModified, "Created from EPrint's last_modified field"))
	}
	/*
		// status_changed is not a date type in Invenio-RDM, might be mapped
		// into available object.
		// FIXME: is this date reflect when it changes status or when it was made available?
		if eprint.StatusChanged != "" {
			rec.Metadata.Dates = append(rec.Metadata.Dates, dateTypeFromTimestamp("status_changed", eprint.StatusChanged, "Created from EPrint's status_changed field"))
		}
	*/
	if eprint.Publisher != "" {
		rec.Metadata.Publisher = eprint.Publisher
	} else if eprint.Publication != "" {
		rec.Metadata.Publisher = eprint.Publication
	} else if eprint.DOI == "" {
		rec.Metadata.Publisher = "Caltech Library"
	}

	// Pickup EPrint ID as "external identifier" in .metadata.identifier
	if eprint.EPrintID > 0 {
		AddIdentifier(rec, "eprintid", fmt.Sprintf("%d", eprint.EPrintID))
	}

	// NOTE: I'm adding this in metadataFromEPrint due to cases where multiple
	// records use the same DOI (something that publishers do at times). 
	if eprint.DOI != "" {
		AddIdentifier(rec, "doi", eprint.DOI)
	}
	if eprint.IDNumber != "" {
		AddIdentifier(rec, "resolverid", eprint.IDNumber)
	}
	if eprint.ISBN != "" {
		SetImprintField(rec, "isbn", eprint.ISBN)
		if eprint.BookTitle != "" {
			SetImprintField(rec, "title", eprint.BookTitle)
		}
		if eprint.PageRange != "" {
			SetImprintField(rec, "pages", eprint.PageRange)
		}
		if eprint.PlaceOfPub != "" {
			SetImprintField(rec, "place", eprint.PlaceOfPub)
		}
		if eprint.Series != "" {
			SetImprintField(rec, "series", eprint.Series)
		}
	}
	if eprint.PMCID != "" {
		if strings.Contains(eprint.PMCID, ",") {
			pmcids := strings.Split(eprint.PMCID, ",")
			for _, pmcid := range pmcids {
				pmcid = strings.TrimSpace(strings.ToUpper(pmcid))
				AddIdentifier(rec, "pmcid", pmcid)
			}
		} else if strings.Contains(eprint.PMCID, ";") {
			pmcids := strings.Split(eprint.PMCID, ";")
			for _, pmcid := range pmcids {
				pmcid = strings.TrimSpace(strings.ToUpper(pmcid))
				AddIdentifier(rec, "pmcid", pmcid)
			}
		} else {
			AddIdentifier(rec, "pmcid", strings.TrimSpace(strings.ToUpper(eprint.PMCID)))
		}
	}
	if (eprint.Funders != nil) && (eprint.Funders.Items != nil) {
		for _, item := range eprint.Funders.Items {
			if item.Agency != "" {
				rec.Metadata.Funding = append(rec.Metadata.Funding, funderFromItem(item))
			}
		}
	}
	// NOTE: Handle related URLs and place them into identifiers.
	if eprint.RelatedURL != nil && eprint.RelatedURL.Length() > 0 {
		for _, item := range eprint.RelatedURL.Items {
			urlType, urlValue := strings.TrimSpace(item.Type), strings.TrimSpace(item.Value)
			if urlValue == "" {
				urlValue = strings.TrimSpace(item.URL)
			}
			switch urlType {
			case "doi":
				AddRelatedIdentifier(rec, "doi", "describes", urlValue)
			case "pmcid":
				AddIdentifier(rec, urlType, urlValue)
			case "eprintid":
				AddIdentifier(rec, urlType, urlValue)
			case "resolverid":
				AddIdentifier(rec, urlType, urlValue)
			case "pmc":
				//NOTE: Per Tom we're not migrating these, their sorta useless ...
				//AddIdentifier(rec, urlType, urlValue)
			case "pub":
				AddRelatedIdentifier(rec, "url", "ispublishedin", urlValue)
			default:
				AddRelatedIdentifier(rec, "url", "describes", urlValue)
			}
		}
	}
	// NOTE: Issue #47, Add notes as additional description
	if eprint.Note != "" || eprint.ErrataText != "" {
		if rec.Metadata.AdditionalDescriptions == nil {
			rec.Metadata.AdditionalDescriptions = []*simplified.Description{}
		}
		if eprint.Note != "" {
			rec.Metadata.AdditionalDescriptions = append(rec.Metadata.AdditionalDescriptions , &simplified.Description{
				Type: &simplified.Type{
					ID: "additional",
				},
				Description: eprint.Note,
			})
		}
		if eprint.ErrataText != "" {
			rec.Metadata.AdditionalDescriptions = append(rec.Metadata.AdditionalDescriptions , &simplified.Description{
				Type: &simplified.Type{
					ID: "errata",
				},
				Description: eprint.ErrataText,
			})
		}
	}
	return nil
}

// Decide if to migrate filename based on name and format description
func migrateFile(fName string, doc *eprinttools.Document) bool {
	// Always explude indexcodes.txt and thumbnails. These are 
	// EPrints internal files not user submitted files.
	if (fName == "indexcodes.txt") || (fName == "preview.png") ||
		strings.HasPrefix(doc.FormatDesc, "Generate") ||
		strings.HasPrefix(doc.FormatDesc, "Thumbnail") {
		return false
	}
	return true
}

// filesFromEPrint extracts all the files specific metadata from the
// EPrint record with a specific document.security string (e.g.
// 'internal', 'public', 'staffonly', 'validuser'). NOTE: "staffonly"
// security setting is normalized to "internal" in this func.
func filesFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	// crosswalk Files from EPrints DocumentList
	if (eprint != nil) && (eprint.Documents != nil) && (eprint.Documents.Length() > 0) {
		addFiles := false
		files := new(simplified.Files)
		files.Order = []string{}
		files.Enabled = true
		files.Entries = map[string]*simplified.Entry{}
		for i := 0; i < eprint.Documents.Length(); i++ {
			doc := eprint.Documents.IndexOf(i)
			// NOTE: We normalize staffonly to internal, per Kathy
			// at migration project meeting 2023-08-10. RSD
			if doc.Security == "staffonly" {
				doc.Security = "internal"
			}
			if len(doc.Files) > 0 {
				for _, docFile := range doc.Files {
					// Check to make sure we want to retain file 
					// information.
					if migrateFile(docFile.Filename, doc) {
    					addFiles = true
    					entry := new(simplified.Entry)
    					entry.FileID = docFile.URL
    					entry.Size = docFile.FileSize
    					entry.MimeType = docFile.MimeType
						entry.Metadata = map[string]interface{}{
							"security": doc.Security,
							"format": doc.Format,
							"format_desc": doc.FormatDesc,
							"rev_number": doc.RevNumber,
							"pos": doc.Pos,
							"main": doc.Main,
							"content": doc.Content,
							"file_id": docFile.FileID,
							"object_id": docFile.ObjectID,
							"filename": docFile.Filename,
						}
    					if doc.Content == "submitted"  || 
							doc.Content == "preprint" || 
							doc.Content == "published" {
    						entry.VersionID = doc.Content
    					}
    					if docFile.Hash != "" {
    						entry.CheckSum = fmt.Sprintf("%s:%s", strings.ToLower(docFile.HashType), docFile.Hash)
    					}
    					files.Entries[docFile.Filename] = entry
    					if strings.HasPrefix(docFile.Filename, "preview") {
    						files.DefaultPreview = docFile.Filename
    					}
					}
				}
			}
		}
		if addFiles {
			rec.Files = files
		} else {
			rec.Files = nil
		}
	}
	return nil
}

// tombstoneFromEPrint builds a tombstone is the EPrint record
// eprint_status is deletion.
func tombstoneFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	// FIXME: crosswalk Tombstone
	if eprint.EPrintStatus == "deletion" {
		tombstone := new(simplified.Tombstone)
		tombstone.RemovedBy = new(simplified.User)
		tombstone.RemovedBy.DisplayName = eprint.Reviewer
		tombstone.RemovedBy.User = eprint.UserID
		if eprint.Suggestions != "" {
			tombstone.Reason = eprint.Suggestions
		}
		rec.Tombstone = tombstone
	}
	return nil
}

// createdUpdatedFromEPrint extracts
func createdUpdatedFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	var (
		created, updated time.Time
		err              error
		tmFmt            string
	)
	// crosswalk Created date
	if len(eprint.Datestamp) > 0 {
		tmFmt = timestamp
		if len(eprint.Datestamp) < 11 {
			tmFmt = datestamp
		}
		created, err = time.Parse(tmFmt, eprint.Datestamp)
		if err != nil {
			return fmt.Errorf("Error parsing datestamp, %s", err)
		}
		rec.Created = created
	}
	if len(eprint.LastModified) > 0 {
		tmFmt = timestamp
		if len(eprint.LastModified) == 10 {
			tmFmt = datestamp
		}
		updated, err = time.Parse(tmFmt, eprint.LastModified)
		if err != nil {
			return fmt.Errorf("Error parsing last modified date, %s", err)
		}
		rec.Updated = updated
	}
	return nil
}

// Run implements the eprint2rdm cli behaviors.
//
// ```
//
//		app := new(irdmtools.EPrint2Rdm)
//		eprintUsername := os.Getenv("EPRINT_USERNAME")
//		eprintPassword := os.Getenv("EPRINT_PASSWORD")
//		eprintHost := "eprints.example.edu"
//		eprintId := "11822"
//		resourceTypes := map[string]string{}
//	 if err := LoadTypesMap("resource-types.csv", resourceTypes);
//			err != nil {
//			// ... handle error ...
//		}
//	 contributorTypes := map[string]string{}
//	 if err := LoadTypesMap("contributor-types.csv", contributorTypes);
//			err != nil {
//			// ... handle error ...
//		}
//		src, err := app.Run(os.Stdin, os.Stdout, os.Stderr,
//							eprintUser, eprintPassword,
//							eprintHost, eprintId,
//	                     resourceTypes, contributorsTypes,
//							debug)
//		if err != nil {
//			// ... handle error ...
//		}
//		fmt.Printf("%s\n", src)
//
// ```
func (app *EPrint2Rdm) Run(in io.Reader, out io.Writer, eout io.Writer, username string, password string, host string, eprintId string, resourceTypesFName string, contributorTypesFName string, allIds bool, idList string, cName string, debug bool) error {
	if username == "" || password == "" {
		return fmt.Errorf("username or password missing")
	}
	timeout := time.Duration(timeoutSeconds)
	baseURL := fmt.Sprintf("https://%s:%s@%s", username, password, host)
	if allIds {
		eprintids, err := GetKeys(baseURL, timeout, 3)
		if err != nil {
			return err
		}
		for _, eprintid := range eprintids {
			fmt.Fprintf(out, "%d\n", eprintid)
		}
	} else {
		var (
			c   *dataset.Collection
			err error
		)
		if cName != "" {
			c, err = dataset.Open(cName)
			if err != nil {
				return err
			}
			defer c.Close()
		}
		resourceTypes := map[string]string{}
		if resourceTypesFName != "" {
			if err := LoadTypesMap(resourceTypesFName, resourceTypes); err != nil {
				return fmt.Errorf("loading resource type map, %q, %s", resourceTypesFName, err)
			}
		} else {
			for k, v := range defaultEPrintResourceTypeMap {
				resourceTypes[k] = v
			}
		}
		contributorTypes := map[string]string{}
		if contributorTypesFName != "" {
			if err := LoadTypesMap(contributorTypesFName, contributorTypes); err != nil {
				return fmt.Errorf("loading contributor type map, %q, %s", contributorTypesFName, err)
			}
		} else {
			for k, v := range defaultEPrintContributorTypeMap {
				contributorTypes[k] = v
			}
		}

		// Handle the case when you're havesting an id list.
		if idList != "" {
			if cName == "" {
				return fmt.Errorf("id list must be used with -harvest option")
			}
			fp, err := os.Open(idList)
			if err != nil {
				return fmt.Errorf("read id list failed, %q, %s", idList, err)
			}
			defer fp.Close()
			scanner := bufio.NewScanner(fp)
			eprintIds := []string{}
			for scanner.Scan() {
				eprintId = scanner.Text()
				if eprintId != "" {
					eprintIds = append(eprintIds, eprintId)
				}
			}
			tot := len(eprintIds)
			// Figure out when to report out records processed, e.g. each 1% of records
			t0 := time.Now()
			rptTime := time.Now()
			reportProgress := false
			log.Printf("Start processing %d records", tot)
			for i, eprintId := range eprintIds {
				if rptTime, reportProgress = CheckWaitInterval(rptTime, (1 * time.Minute)); reportProgress || i == 0 {
					log.Printf("processing id %s (%d/%d): %s", eprintId, i, tot, ProgressETR(t0, i, tot))
				}
				// Handle the case when you're harvesting a single record id
				id, err := strconv.Atoi(eprintId)
				if err != nil {
					log.Printf("line %d, concverting %q, %s", i+1, eprintId, err)
					continue
				}
				eprints, err := GetEPrint(baseURL, id, timeout, 3)
				if err != nil {
					log.Printf("line %d, retrieving %q, %s", i+1, eprintId, err)
					continue
				}
				record := new(simplified.Record)
				if err := CrosswalkEPrintToRecord(eprints.EPrint[0], record, resourceTypes, contributorTypes); err != nil {
					log.Printf("line %d, crosswalking %q, %s", i+1, eprintId, err)
					continue
				}
				if c.HasKey(eprintId) {
					if err := c.UpdateObject(eprintId, record); err != nil {
						return fmt.Errorf("error saving %q, line %d, %s", eprintId, i+1, err)
					}
				} else {
					if err := c.CreateObject(eprintId, record); err != nil {
						return fmt.Errorf("error saving %q, line %d, %s", eprintId, i+1, err)
					}
				}
			}
			log.Printf("Finished, processed %d records in %s", tot, time.Since(t0).Round(time.Second))
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to scaner %q, %s", idList, err)
			}
			return nil
		}
		// Handle the case when you're harvesting a single record id
		id, err := strconv.Atoi(eprintId)
		if err != nil {
			return err
		}
		eprints, err := GetEPrint(baseURL, id, timeout, 3)
		if err != nil {
			return err
		}
		record := new(simplified.Record)
		if err := CrosswalkEPrintToRecord(eprints.EPrint[0], record, resourceTypes, contributorTypes); err != nil {
			return err
		}
		if cName != "" {
			if c.HasKey(eprintId) {
				if err := c.UpdateObject(eprintId, record); err != nil {
					return err
				}
			} else {
				if err := c.CreateObject(eprintId, record); err != nil {
					return err
				}
			}
			return nil
		}
		src, err := json.MarshalIndent(record, "", "   ")
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", src)
	}
	return nil
}
