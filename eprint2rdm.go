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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	// Caltech Library Packages
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

// CrosswalkEPrintToRecord implements a crosswalk between
// an EPrint 3.x EPrint XML record as struct to a Invenio RDM
// record as struct.
func CrosswalkEPrintToRecord(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	rec.Schema = `local://records/record-v2.0.0.json`
	rec.ID = fmt.Sprintf("%s:%d", eprint.Collection, eprint.EPrintID)

	if err := parentFromEPrint(eprint, rec); err != nil {
		return err
	}
	if err := externalPIDFromEPrint(eprint, rec); err != nil {
		return err
	}
	if err := recordAccessFromEPrint(eprint, rec); err != nil {
		return err
	}

	if err := metadataFromEPrint(eprint, rec); err != nil {
		return err
	}
	if err := filesFromEPrint(eprint, rec); err != nil {
		return err
	}

	if eprint.EPrintStatus == "deletion" {
		if err := tombstoneFromEPrint(eprint, rec); err != nil {
			return err
		}
	}

	if err := createdUpdatedFromEPrint(eprint, rec); err != nil {
		return err
	}
	/*
		if err := pidFromEPrint(eprint, rec); err != nil {
			return err
		}
	*/
	// Now finish simple record normalization ...
	if err := mapResourceType(eprint, rec); err != nil {
		return err
	}
	if err := simplifyCreators(eprint, rec); err != nil {
		return err
	}
	if err := simplifyContributors(eprint, rec); err != nil {
		return err
	}
	// FIXME: Map eprint record types to invenio RDM record types we've
	// decided on.
	// FIXME: Funders must have a title, could just copy in the funder
	// name for now.
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
	person.Type = "person"
	if item.Name != nil {
		person.FamilyName = item.Name.Family
		person.GivenName = item.Name.Given
		// FIXME: How do we map honorific and lineage? E.g. William Goddard versus William A. Goddard III
		//honorific := item.Name.Honourific
		//lineage := item.Name.Lineage
		if person.FamilyName != "" || person.GivenName != "" {
			person.Name = fmt.Sprintf("%s, %s", person.FamilyName, person.GivenName)
			clPeopleID = item.Name.ID
			orcid = item.Name.ORCID
		} else {
			person.Type = "organizaton"
			person.Name = item.Name.Value
		}
	}
	ror = item.ROR
	if clPeopleID != "" {
		person.Identifiers = append(person.Identifiers, mkSimpleIdentifier("clpid", clPeopleID))
	}
	if orcid != "" {
		person.Identifiers = append(person.Identifiers, mkSimpleIdentifier("ORCID", orcid))
	}
	if ror != "" {
		person.Identifiers = append(person.Identifiers, mkSimpleIdentifier("ROR", ror))
	}
	return person
}

// simplifyCreators make sure the identifiers are mapped to Invenio-RDM
// identifiers.
func simplifyCreators(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	// First map the creators (person) to RDM .metadata.creators
	// FIXME: Then map the corpCreators (org) to RDM .metadata.creators
	creators := []*simplified.Creator{}
	if eprint.Creators != nil && eprint.Creators.Length() > 0 {
		for i := 0; i < eprint.Creators.Length(); i++ {
			if item := eprint.Creators.IndexOf(i); item != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					creators = append(creators, &simplified.Creator{
						PersonOrOrg: person,
						//Role:        &simplified.Role{},
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
						//Role:        &simplified.Role{},
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
func simplifyContributors(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	contributors := []*simplified.Creator{}
	// First add contributors, then editors, etc.
	if eprint.Contributors != nil && eprint.Contributors.Length() > 0 {
		for i := 0; i < eprint.Contributors.Length(); i++ {
			if item := eprint.Contributors.IndexOf(i); item != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							Title: uriToContributorType(item.Role),
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
							Title: "contributor",
						},
					})
				}
			}
		}
	}
	// Add Editors, Adivisors, Committee Members, Thesis Chair, Reviewers, Translators, etc.
	if eprint.Editors != nil && eprint.Editors.Length() > 0 {
		for i := 0; i < eprint.Editors.Length(); i++ {
			if item := eprint.Editors.IndexOf(i); item != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							Title: "editor",
						},
					})
				}
			}
		}
	}
	if eprint.ThesisAdvisor != nil && eprint.ThesisAdvisor.Length() > 0 {
		for i := 0; i < eprint.ThesisAdvisor.Length(); i++ {
			if item := eprint.ThesisAdvisor.IndexOf(i); item != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							Title: "thesis_advisor",
						},
					})
				}
			}
		}
	}
	if eprint.ThesisCommittee != nil && eprint.ThesisCommittee.Length() > 0 {
		for i := 0; i < eprint.ThesisCommittee.Length(); i++ {
			if item := eprint.ThesisCommittee.IndexOf(i); item != nil {
				if person := itemToPersonOrOrg(item); person != nil {
					contributors = append(contributors, &simplified.Creator{
						PersonOrOrg: person,
						Role: &simplified.Role{
							Title: "thesis_committee",
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

// mapResourceType maps the EPrints record types to a predetermined
// Invenio-RDM record type.
func mapResourceType(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	if rec.Metadata.ResourceType == nil {
		rec.Metadata.ResourceType = map[string]string{}
	}
	// FIXME: I need to implement a full map of default resource types.
	// FIXME: need to load this from an optional configuration file
	crosswalkResourceTypes := map[string]string{
		"article": "publication-article",
	}
	val, ok := crosswalkResourceTypes[eprint.Type]
	if !ok {
		return fmt.Errorf("unable to map %q to simple record type", eprint.Type)
	}
	rec.Metadata.ResourceType["id"] = val
	return nil
}

/*
// PIDFromEPrint crosswalks the PID from an EPrint record.
func pidFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
	data := map[string]interface{}{}
	src := fmt.Sprintf(`{
"id": %d,
"pid": { "eprint": "eprintid"}
}`, eprint.EPrintID)
	err := jsonDecode([]byte(src), &data)
	if err != nil {
		return fmt.Errorf("Cannot generate PID from EPrint %d", eprint.EPrintID)
	}
	rec.PID = data
	return nil
}
*/

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
		pid.Provider = "datacite"
		pid.Client = ""
		rec.ExternalPIDs["doi"] = pid
	}
	// Pickup ISSN
	if eprint.ISBN != "" {
		pid := new(simplified.PersistentIdentifier)
		pid.Identifier = eprint.ISSN
		pid.Provider = ""
		pid.Client = ""
		rec.ExternalPIDs["issn"] = pid
	}
	// Pickup ISBN
	if eprint.ISBN != "" {
		pid := new(simplified.PersistentIdentifier)
		pid.Identifier = eprint.ISBN
		pid.Provider = ""
		pid.Client = ""
		rec.ExternalPIDs["isbn"] = pid
	}
	//FIXME: figure out if we have other persistent identifiers
	//scattered in the EPrints XML and map them.
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
	// By default lets assume the files are restricted.
	rec.RecordAccess.Files = "resticted"
	if isPublic {
		rec.RecordAccess.Record = "public"
	} else {
		rec.RecordAccess.Record = "restricted"
	}
	// Need to make sure record is not embargoed
	if eprint.Documents != nil {
		for i := 0; i < eprint.Documents.Length(); i++ {
			doc := eprint.Documents.IndexOf(i)
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

func uriToContributorType(role_uri string) string {
	roles := map[string]string{
		// Article Author
		"http://coda.library.caltech.edu/ARA": "author_section",
		// Astronaut
		"http://coda.library.caltech.edu/AST": "astronaut",
		// Author of afterword, colophon, etc.
		"http://www.loc.gov/loc.terms/relators/AFT": "aft",
		// Bibliographic antecedent
		"http://www.loc.gov/loc.terms/relators/ANT": "ant",
		// Author in quotations or text abstracts
		"http://www.loc.gov/loc.terms/relators/AQT": "aqt",
		// Screenwriter
		"http://www.loc.gov/loc.terms/relators/AUS": "screenwriter",
		// Author, joint author
		"http://www.loc.gov/loc.terms/relators/AUT": "author",
		// Collaborator
		"http://www.loc.gov/loc.terms/relators/CLB": "collaborator",
		// Compiler
		"http://www.loc.gov/loc.terms/relators/COM": "compiler",
		// Contributor
		"http://www.loc.gov/loc.terms/relators/CTB": "contributor",
		// Directory
		"http://www.loc.gov/loc.terms/relators/DRT": "director",
		// Editor
		"http://www.loc.gov/loc.terms/relators/EDT": "editor",
		// Narrator
		"http://www.loc.gov/loc.terms/relators/NRT": "narrator",
		// Other
		"http://www.loc.gov/loc.terms/relators/OTH": "other",
		// Publishing director
		"http://www.loc.gov/loc.terms/relators/PBD": "publishing_director",
		// Programmer
		"http://www.loc.gov/loc.terms/relators/PRG": "programmer",
		// Reviewer
		"http://www.loc.gov/loc.terms/relators/REV": "reviewer",
		// Research team member
		"http://www.loc.gov/loc.terms/relators/RTM": "research_team",
		// Speaker
		"http://www.loc.gov/loc.terms/relators/SPK": "speaker",
		// Teacher
		"http://www.loc.gov/loc.terms/relators/TCH": "teacher",
		// Translator
		"http://www.loc.gov/loc.terms/relators/TRL": "translator",
	}
	if val, ok := roles[role_uri]; ok {
		return val
	}
	return "contributor"
}

func dateTypeFromTimestamp(dtType string, timestamp string, description string) *simplified.DateType {
	dt := new(simplified.DateType)
	dt.Type = new(simplified.Type)
	dt.Type.ID = dtType
	dt.Type.Title = dtType
	dt.Description = description
	if len(timestamp) > 9 {
		dt.Date = timestamp[0:10]
	} else {
		dt.Date = timestamp
	}
	return dt
}

func mkSimpleIdentifier(scheme, value string) *simplified.Identifier {
	identifier := new(simplified.Identifier)
	identifier.Scheme = strings.ToLower(scheme)
	identifier.Identifier = value
	return identifier
}

func funderFromItem(item *eprinttools.Item) *simplified.Funder {
	funder := new(simplified.Funder)
	if item.GrantNumber != "" {
		funder.Award = new(simplified.Identifier)
		funder.Award.Number = item.GrantNumber
		funder.Award.Scheme = "eprints_grant_number"
	}
	if item.Agency != "" {
		org := new(simplified.Identifier)
		org.Name = item.Agency
		org.Scheme = "eprints_agency"
		funder.Funder = org
	}
	return funder
}

// metadataFromEPrint extracts metadata from the EPrint record
func metadataFromEPrint(eprint *eprinttools.EPrint, rec *simplified.Record) error {
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
	if err := simplifyContributors(eprint, rec); err != nil {
		return err
	}
	if eprint.Abstract != "" {
		rec.Metadata.Description = eprint.Abstract
	}
	rec.Metadata.PublicationDate = eprint.PubDate()

	// Rights are scattered in several EPrints fields, they need to
	// be evaluated to create a "Rights" object used in DataCite/Invenio
	addRights := false
	rights := new(simplified.Right)
	if eprint.Rights != "" {
		addRights = true
		rights.Description = &simplified.Description{
			Description: eprint.Rights,
		}
	}
	// Figure out if our copyright information is in the Note field.
	if (eprint.Note != "") && (strings.Contains(eprint.Note, "©") || strings.Contains(eprint.Note, "copyright") || strings.Contains(eprint.Note, "(c)")) {
		addRights = true
		rights.Description = &simplified.Description{
			Description: fmt.Sprintf("%s", eprint.Note),
		}
	}
	if addRights {
		rec.Metadata.Rights = append(rec.Metadata.Rights, rights)
	}
	if eprint.CopyrightStatement != "" {
		rights := new(simplified.Right)
		rights.Description = &simplified.Description{
			Description: eprint.CopyrightStatement,
		}
		rec.Metadata.Rights = append(rec.Metadata.Rights, rights)
	}
	// FIXME: work with Tom to sort out how "Rights" and document level
	// copyright info should work.

	if (eprint.Subjects != nil) && (eprint.Subjects.Items != nil) {
		for _, item := range eprint.Subjects.Items {
			subject := new(simplified.Subject)
			subject.Subject = item.Value
			rec.Metadata.Subjects = append(rec.Metadata.Subjects, subject)
		}
	}

	// FIXME: Work with Tom to figure out correct mapping of rights from EPrints XML
	// FIXME: Language appears to be at the "document" level, not record level

	// Dates are scattered through the primary eprint table.
	if (eprint.DateType != "published") && (eprint.Date != "") {
		rec.Metadata.Dates = append(rec.Metadata.Dates, dateTypeFromTimestamp("pub_date", eprint.Date, "Publication Date"))
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
	if eprint.RevNumber != 0 {
		rec.Metadata.Version = fmt.Sprintf("v%d", eprint.RevNumber)
	}
	if eprint.Publisher != "" {
		rec.Metadata.Publisher = eprint.Publisher
	} else if eprint.Publication != "" {
		rec.Metadata.Publisher = eprint.Publication
	} else if eprint.DOI == "" {
		rec.Metadata.Publisher = "Caltech Library"
	}

	// Pickup EPrint ID as "external identifier" in .metadata.identifier
	if eprint.EPrintID > 0 {
		rec.Metadata.Identifiers = append(rec.Metadata.Identifiers, mkSimpleIdentifier("eprintid", fmt.Sprintf("%d", eprint.EPrintID)))
	}

	if eprint.DOI != "" {
		rec.Metadata.Identifiers = append(rec.Metadata.Identifiers, mkSimpleIdentifier("doi", eprint.DOI))
	}
	if eprint.ISBN != "" {
		rec.Metadata.Identifiers = append(rec.Metadata.Identifiers, mkSimpleIdentifier("isbn", eprint.ISBN))
	}
	if eprint.ISSN != "" {
		rec.Metadata.Identifiers = append(rec.Metadata.Identifiers, mkSimpleIdentifier("issn", eprint.ISSN))
	}
	if eprint.PMCID != "" {
		rec.Metadata.Identifiers = append(rec.Metadata.Identifiers, mkSimpleIdentifier("pmcid", eprint.PMCID))
	}
	if (eprint.Funders != nil) && (eprint.Funders.Items != nil) {
		for _, item := range eprint.Funders.Items {
			if item.Agency != "" {
				rec.Metadata.Funding = append(rec.Metadata.Funding, funderFromItem(item))
			}
		}
	}
	return nil
}

// filesFromEPrint extracts all the file specific metadata from the
// EPrint record
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
			if len(doc.Files) > 0 {
				for _, docFile := range doc.Files {
					addFiles = true
					entry := new(simplified.Entry)
					entry.FileID = docFile.URL
					entry.Size = docFile.FileSize
					entry.MimeType = docFile.MimeType
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
//	 eprintUsername := os.Getenv("EPRINT_USERNAME")
//	 eprintPassword := os.Getenv("EPRINT_PASSWORD")
//		eprintHost := "eprints.example.edu"
//	 eprintId := "11822"
//		src, err := app.Run(os.Stdin, os.Stdout, os.Stderr,
//		                     eprintUser, eprintPassword,
//	                      eprintHost, eprintId, debug)
//		if err != nil {
//		    // ... handle error ...
//		}
//		fmt.Printf("%s\n", src)
//
// ```
func (app *EPrint2Rdm) Run(in io.Reader, out io.Writer, eout io.Writer, username string, password string, host string, eprintId string, debug bool) error {
	buf := new(bytes.Buffer)
	getURL := fmt.Sprintf("https://%s/rest/eprint/%s.xml", host, eprintId)
	auth := "basic"
	options := map[string]bool{
		"debug": debug,
	}
	exitCode := eprinttools.RunEPrintsRESTClient(buf, getURL, auth, username, password, options)
	if exitCode != 0 {
		return fmt.Errorf("failed to retrieve %q\n", getURL)
	}
	src := buf.Bytes()
	eprints := new(eprinttools.EPrints)
	if err := xml.Unmarshal(src, &eprints); err != nil {
		return err
	}
	record := new(simplified.Record)
	err := CrosswalkEPrintToRecord(eprints.EPrint[0], record)
	if err != nil {
		return err
	}
	src, err = json.MarshalIndent(record, "", "   ")
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s\n", src)
	return nil
}
