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
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	// Caltech Library Packages
	"github.com/caltechlibrary/dataset/v2"
	"github.com/caltechlibrary/eprinttools"
	"github.com/caltechlibrary/simplified"
)

// Rdm2EPrint holds the configuration for rdmutil cli.
type Rdm2EPrint struct {
	Cfg *Config
}

var (

	// resourceMap maps a resource from RDM to EPRints.
	resourceMap = map[string]string{
		"publication-article":              "article",
		"publication-section":              "book_section",
		"publication-report":               "monograph",
		"publication-preprint":		        "monograph",
		"publication-book":                 "book",
		"conference-paper":                 "conference_item",
		"conference-poster":                "conference_item",
		"conference-presentation":          "conference_item",
		"publication-conferenceproceeding": "book",
		"publication-patent":               "patent",
		"publication-technicalnote":        "monograph",
		"publication-thesis":               "thesis",
		"teachingresource":                 "teaching_resource",
		"teachingresource-lecturenotes":    "teching_resource",
		"teachingresource-textbook":        "teaching_resource",
	}

	// communityMap maps community names to save repeated calls the API
	communityMap = map[string]string{}
)

// lookupCommunityName checks to see if a community id is in communityMap,
// if not queries the RDM API for the community name, stores it in the map
// and returns the value with the function call.
func lookupCommunityName(cfg *Config, communityID string) (string, bool) {
	if communityName, ok := communityMap[communityID]; ok {
		return communityName, true
	}
	// e.g. https://authors.library.caltech.edu/api/communities/aedd135f-227e-4fdf-9476-5b3fd011bac6
	apiURL := fmt.Sprintf("%s/api/communities/%s", cfg.InvenioAPI, communityID)
	src, headers, err := getJSON(cfg.InvenioToken, apiURL)
	if err != nil {
		return "", false
	}
	cfg.rl.FromHeader(headers)
	m := map[string]interface{}{}
	if err := JSONUnmarshal(src, &m); err != nil {
		return "", false
	}
	if metadata, ok := m["metadata"].(map[string]interface{}); ok {
		if title, ok := metadata["title"].(string); ok {
			communityMap[communityID] = title
			return title, true
		}
	}
	return "", false
}

// CrosswalkRdmToEPrint takes a public RDM record and
// converts it to an EPrint struct which can be rendered as
// JSON or XML.
//
// ```
// app := new(irdmtools.Rdm2EPrint)
//
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//
// recordId := "woie-x0121"
// src, err := app.GetRecord(cfg, recordId, false)
//
//	if err != nil {
//	   // ... handle error ...
//	}
//
// rec := new(simplified.Record)
// eprint := new (eprinttools.EPrint)
// eprints := new (eprinttools.EPrints)
//
//	if err := irdmtools.JSONUnmarshal(src, &rec); err != nil {
//	   // ... handle error ...
//	}
//
//	if err := CrosswalkRdmToEPrint(rec, eprint) {
//	   // ... handle error ...
//	}
//
// // Add eprint to outer EPrints struct before rendering
// eprints.EPrint = append(eprints.EPrint, eprint)
// // Output as JSON for single eprint record
// src, _ := irdmtools.JSONMarshalIndent(eprints)
// fmt.Printf("%s\n", src)
// ```
func CrosswalkRdmToEPrint(cfg *Config, rec *simplified.Record, eprint *eprinttools.EPrint) error {
	if rec.RecordAccess != nil {
		// We'll assume these are public records so we set eprint_status to "archive" if "open"
		// otherwise we'll assume these would map to the inbox.
		if rec.RecordAccess.Record == "public" {
			eprint.MetadataVisibility = "show"
			eprint.EPrintStatus = "archive"
			if rec.RecordAccess.Files == "public" {
				eprint.FullTextStatus = "public"
			} else {
				eprint.FullTextStatus = "restricted"
			}
		} else {
			eprint.EPrintStatus = "inbox"
			eprint.MetadataVisibility = "no_search"
			eprint.FullTextStatus = "restricted"
		}
	}

	// Default EPrint id is a URL, so we'll point at the RDM location.
	if rec.ID != "" {
		eprint.ID = fmt.Sprintf("%s/records/%s", cfg.InvenioAPI, rec.ID)
	}

	if rec.Metadata != nil {
		// get EPrint ID from rec if set
		if eprintid, ok := getMetadataIdentifier(rec, "eprintid"); ok {
			if eprintid != "" {
				eprint.EPrintID, _ = strconv.Atoi(eprintid)
			}
		}
		if doi, ok := getMetadataIdentifier(rec, "doi"); ok {
			eprint.DOI = doi
		} 
		if (rec.ExternalPIDs != nil) {
			if doi, ok := rec.ExternalPIDs["doi"]; ok {
				eprint.DOI = doi.Identifier
			}
		}
		if pmcid, ok := getMetadataIdentifier(rec, "pmcid"); ok {
			eprint.PMCID = pmcid
		}
		if rec.Metadata != nil {
			if rec.Metadata.PublicationDate != "" {
				eprint.Date = rec.Metadata.PublicationDate
				eprint.DateType = "published"
				eprint.IsPublished = "pub"
			} else {
				eprint.IsPublished = "unpub"
			}
			if rec.Metadata.Title != "" {
				eprint.Title = rec.Metadata.Title
			}
			if rec.Metadata.Description != "" {
				eprint.Abstract = rec.Metadata.Description
			}
		}
		eprint.Datestamp = rec.Created.Format(timestamp)
		eprint.LastModified = rec.Updated.Format(timestamp)
		if resourceType, ok := getMetadataResourceType(rec, resourceMap); ok {
			eprint.Type = resourceType
		}
		editors := &eprinttools.EditorItemList{}
		if rec.Metadata.Creators != nil && len(rec.Metadata.Creators) > 0 {
			creators := &eprinttools.CreatorItemList{}
			corpCreators := &eprinttools.CorpCreatorItemList{}
			for _, creator := range rec.Metadata.Creators {
				if creator.PersonOrOrg != nil {
					if item, ok := creatorPersonToEPrintItem(creator); ok {
						if creatorHasRole(creator.Role, "editor") {
							editors.Append(item)
						} else {
							creators.Append(item)
						}
					} else if item, ok := creatorCorpToEPrintItem(creator); ok {
						corpCreators.Append(item)
					}
				}
			}
			if creators.Length() > 0 {
				eprint.Creators = creators
			}
			if corpCreators.Length() > 0 {
				eprint.CorpCreators = corpCreators
			}
		}
		if rec.Metadata.Contributors != nil && len(rec.Metadata.Contributors) > 0 {
			contributors := &eprinttools.ContributorItemList{}
			for _, contributor := range rec.Metadata.Contributors {
				if contributor.PersonOrOrg != nil {
					if item, ok := creatorPersonToEPrintItem(contributor); ok {
						if creatorHasRole(contributor.Role, "editor") {
							editors.Append(item)
						} else {
							contributors.Append(item)
						}
					} else if item, ok := contributorCorpAsPersonToEPrintItem(contributor); ok {
						contributors.Append(item)
					}
				}
			}
			if contributors.Length() > 0 {
				eprint.Contributors = contributors
			}
			if editors.Length() > 0 {
				eprint.Editors = editors
			}
		}
		if rec.Metadata.PublicationDate != "" {
			eprint.Date = rec.Metadata.PublicationDate
			eprint.DateType = "published"
		}
		if rec.Metadata.Subjects != nil {
			// Note I am mapping sujects to keywords in EPrints given that each
			// EPrint repository has a hierarchy of subjects and URI are used to show that.
			keywords := []string{}
			for _, subject := range rec.Metadata.Subjects {
				if subject.Subject != "cls" {
					keywords = append(keywords, subject.Subject)
				}
			}
			if len(keywords) > 0 {
				eprint.Keywords = strings.Join(keywords, "; ")
			}
		}
		if rec.Metadata.Identifiers != nil {
			if resolverID, ok := getIdentifier(rec.Metadata.Identifiers, "resolverid"); ok {
				eprint.OfficialURL = fmt.Sprintf("https://resolver.caltech.edu/%s", resolverID)
				eprint.IDNumber = resolverID
			} 
		}
		if rec.Metadata.Rights != nil && len(rec.Metadata.Rights) > 0 {
			if rights, ok := rec.Metadata.Rights[0].Description["en"]; ok {
				eprint.Rights = rights
			}
		}
		if rec.Metadata.AdditionalDescriptions != nil && len(rec.Metadata.AdditionalDescriptions) > 0 {
			notes := []string{}
			for _, description := range rec.Metadata.AdditionalDescriptions {
				note := strings.TrimSpace(description.Description)
				if note != "" {
					notes = append(notes, note)
				}
			}
			eprint.Note = strings.Join(notes, "\n\n")
		}
		if rec.Metadata.Publisher != "" {
			eprint.Publisher = rec.Metadata.Publisher
		}
		if rec.Metadata.Funding != nil && len(rec.Metadata.Funding) > 0 {
			funders := &eprinttools.FunderItemList{}
			for _, funder := range rec.Metadata.Funding {
				var (
					agency string
					award  string
				)
				if funder.Funder != nil && funder.Funder.Name != "" {
					agency = funder.Funder.Name
				}
				if funder.Award != nil && funder.Award.Number != "" {
					award = funder.Award.Number
				}
				item := new(eprinttools.Item)
				if agency != "" {
					item.Agency = agency
				}
				if award != "" {
					item.GrantNumber = award
				}
				funders.Append(item)
			}
			if funders.Length() > 0 {
				eprint.Funders = funders
			}
		}
	}
	if rec.Parent != nil && rec.Parent.Communities != nil {
		communityID := rec.Parent.Communities.Default
		if collectionName, ok := lookupCommunityName(cfg, communityID); ok {
			eprint.Collection = collectionName
		}
	}
	if len(rec.CustomFields) > 0 {
		if imprintInfo, ok := rec.CustomFields["imprint:imprint"].(map[string]interface{}); ok {
			if title, ok := imprintInfo["title"].(string); ok {
				eprint.BookTitle = title
			}
			if isbn, ok := imprintInfo["isbn"].(string); ok {
				eprint.ISBN = isbn
			}
			if pages, ok := imprintInfo["pages"].(string); ok {
				eprint.PageRange = pages
			}
			if place, ok := imprintInfo["place"].(string); ok {
				eprint.PlaceOfPub = place
			}
			if edition, ok := imprintInfo["edition"].(string); ok {
				eprint.Edition = edition
			}
		}
		if journalInfo, ok := rec.CustomFields["journal:journal"].(map[string]interface{}); ok {
			if title, ok := journalInfo["title"].(string); ok {
				eprint.Publication = title
			}
			if volume, ok := journalInfo["volume"].(string); ok {
				eprint.Volume = volume
			}
			if issn, ok := journalInfo["issn"].(string); ok {
				eprint.ISSN = issn
			}
			if issueNo, ok := journalInfo["issue"].(string); ok {
				eprint.Number = issueNo
			}
			if pages, ok := journalInfo["pages"].(string); ok {
				eprint.PageRange = pages
			}
		}
		if caltechPlaceOfPubs, ok := rec.CustomFields["caltech:place_of_publication"].(map[string]interface{}); ok {
			if place, ok := caltechPlaceOfPubs["place"]; ok {
				eprint.PlaceOfPub = place.(string)
			}
		}
		if caltechSeries, ok := rec.CustomFields["caltech:series"].(map[string]interface{}); ok {
			if series, ok := caltechSeries["series"]; ok {
				eprint.Series = series.(string)
			}
		}
		
		if caltechGroups, ok := rec.CustomFields["caltech:groups"].([]interface{}); ok {
			if len(caltechGroups) > 0 {
				groupList := new(eprinttools.LocalGroupItemList)
				for _, groups := range caltechGroups {
					if group, ok := groups.(map[string]interface{}); ok {
						addItem := false
						item := new(eprinttools.Item)
						if id, ok := group["id"]; ok {
							addItem = true
							item.ID = id.(string)
						}
						// FIXME: title is populated from translating the vocabulary against
						// and group id attribute so this is always empty.
						if title, ok := group["title"].(map[string]interface{}); ok {
							if en, ok := title["en"]; ok {
								addItem = true
								item.Value = en.(string)
							}
						}
						//NOTE: Need to make sure we're not adding a duplicate groups
						for i := 0; i < groupList.Length(); i++ {
							grp := groupList.IndexOf(i)
							if grp.ID == item.ID {
								addItem = false
								break
							}
							if item.Value != "" && (grp.Value == item.Value) {
								addItem = false
							}
						}
						if addItem {
							groupList.Append(item)
						}
					}
				}
				if groupList.Length() > 0 {
					eprint.LocalGroup = groupList
				}
			}
		}
		otherNumberSystemItem := &eprinttools.Item{}
		if numName, ok := rec.CustomFields["caltech:other_num_name"].(string); ok {
			otherNumberSystemItem.Name = &eprinttools.Name{
				Value: numName,
			}
		}
		if numID, ok := rec.CustomFields["caltech:other_num_id"].(string); ok {
			otherNumberSystemItem.ID = numID
		}
		if otherNumberSystemItem.ID != "" || otherNumberSystemItem.Name != nil {
			eprint.OtherNumberingSystem = &eprinttools.OtherNumberingSystemItemList{}
			eprint.OtherNumberingSystem.Append(otherNumberSystemItem)
		}
	}
	if (rec.Files != nil) {
		// Finally we need to add our Related and Primary Objects
		defaultPreview := rec.Files.DefaultPreview
		if rec.Files.Entries != nil  {
			for _, entry := range rec.Files.Entries {
				if defaultPreview == "" {
					defaultPreview = entry.Key
				}
				if defaultPreview == entry.Key {
					eprint.PrimaryObject = map[string]interface{}{
						"basename": defaultPreview,
						"url": fmt.Sprintf("%s/records/%s/files/%s", cfg.InvenioAPI, rec.ID, defaultPreview),
					}
				} else {
					if eprint.RelatedObjects == nil {
						eprint.RelatedObjects = []map[string]interface{}{}
					}
					eprint.RelatedObjects = append(eprint.RelatedObjects, map[string]interface{}{
						"basename": entry.Key,
						"url": fmt.Sprintf("%s/records/%s/files/%s", cfg.InvenioAPI, rec.ID, entry.Key),
					})
				}
			}
		}
	}


	// Make sure we populate Official URL is populated if we don't have a resolver URL available
	if eprint.OfficialURL == "" && rec.ID != "" {
		//NOTE: We need to assemble an appropriate RDM url since resolver isn't available
		eprint.OfficialURL = fmt.Sprintf("%s/records/%s", cfg.InvenioAPI, rec.ID)
	}

	// Now that we have enough information the eprint structure we can answer some questions
	// and infer values.
	if eprint.Type != "article" && eprint.Publication == "" {
		eprint.IsPublished = "unpub"
	}
	return nil
}

// getIdentifier returns a related identifier with matching scheme
func getIdentifier(identifiers []*simplified.Identifier, scheme string) (string, bool) {
	for _, identifier := range identifiers {
		if identifier.Scheme == scheme {
			if identifier.Identifier != "" {
				return identifier.Identifier, true
			}
			if identifier.ID != "" {
				return identifier.ID, true
			}
		}
	}
	return "", false
}

func creatorHasRole(role *simplified.Role, roleType string) bool {
	if role != nil {
		if role.ID == roleType {
			return true
		}
		if role.Title != nil && len(role.Title) > 0 {
			if en, ok := role.Title["en"]; ok {
				if strings.Contains(strings.ToLower(en), roleType) {
					return true
				}
			}
		}
	}
	return false
}

// creatorPersonToEPrintItem takes a RDM .Metadata.Creators element and turns
// it into an eprintools.Item type for a person.
func creatorPersonToEPrintItem(creator *simplified.Creator) (*eprinttools.Item, bool) {
	if creator.PersonOrOrg == nil {
		return nil, false
	}
	if creator.PersonOrOrg.FamilyName == "" && creator.PersonOrOrg.GivenName == "" {
		return nil, false
	}
	item := new(eprinttools.Item)
	item.Name = &eprinttools.Name{
		Given:  creator.PersonOrOrg.GivenName,
		Family: creator.PersonOrOrg.FamilyName,
	}
	if clpid, ok := getPersonOrOrgIdentifier(creator.PersonOrOrg, "clpid"); ok {
		item.ID = clpid
	}
	if orcid, ok := getPersonOrOrgIdentifier(creator.PersonOrOrg, "orcid"); ok {
		item.ORCID = orcid
	}
	return item, true
}

// creatorCorpToEPrintItem takes a RDM .Metadata.Creators element and turns
// it into an eprintools.Item type for a organization.
func creatorCorpToEPrintItem(creator *simplified.Creator) (*eprinttools.Item, bool) {
	if creator.PersonOrOrg == nil {
		return nil, false
	}
	if creator.PersonOrOrg.FamilyName != "" && creator.PersonOrOrg.GivenName != "" {
		return nil, false
	}
	item := new(eprinttools.Item)
	item.Value = creator.PersonOrOrg.Name
	if ror, ok := getPersonOrOrgIdentifier(creator.PersonOrOrg, "ror"); ok {
		item.ID = ror
	}
	return item, true
}

func contributorCorpAsPersonToEPrintItem(creator *simplified.Creator) (*eprinttools.Item, bool) {
	if creator.PersonOrOrg == nil {
		return nil, false
	}
	if creator.PersonOrOrg.FamilyName != "" && creator.PersonOrOrg.GivenName != "" {
		return nil, false
	}
	item := new(eprinttools.Item)
	item.Name = &eprinttools.Name{}
	item.Name.Family = creator.PersonOrOrg.Name
	if ror, ok := getPersonOrOrgIdentifier(creator.PersonOrOrg, "ror"); ok {
		item.URI = ror
	}
	return item, true
}

// getPersonOrOrgIdentifier looks through the person or org identifier list for a maching scheme.
func getPersonOrOrgIdentifier(personOrOrg *simplified.PersonOrOrg, scheme string) (string, bool) {
	for _, identifier := range personOrOrg.Identifiers {
		if identifier.Scheme == scheme {
			return identifier.Identifier, true
		}
	}
	return "", false
}

// getMetadataIdentifier retrieves an indifier by scheme and returns the
// identifier value if available from .Metadata.Identifiers
func getMetadataIdentifier(rec *simplified.Record, scheme string) (string, bool) {
	if rec.Metadata != nil && rec.Metadata.Identifiers != nil {
		for _, identifier := range rec.Metadata.Identifiers {
			if identifier.Scheme == scheme {
				return identifier.Identifier, true
			}
		}
	}
	return "", false
}

// getMetadataResourceType returns a metadata resource type if found.
func getMetadataResourceType(rec *simplified.Record, resourceMap map[string]string) (string, bool) {
	if rec.Metadata != nil && rec.Metadata.ResourceType != nil {
		if val, ok := rec.Metadata.ResourceType["id"]; ok {
			resourceType := val.(string)
			if val, ok := resourceMap[resourceType]; ok {
				resourceType = val
			}
			return strings.ReplaceAll(resourceType, "-", "_"), true
		}
	}
	return "", false
}

// Configure reads the configuration file and environtment
// initialing the Cfg attribute of a RdmUtil object. It returns an error
// if problem were encounter.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.Configure("irdmtools.json", "TEST_"); err != nil {
//	   // ... handle error ...
//	}
//	fmt.Printf("Invenio RDM API UTL: %q\n", app.Cfg.IvenioAPI)
//	fmt.Printf("Invenio RDM token: %q\n", app.Cfg.InvenioToken)
//
// ```
func (app *Rdm2EPrint) Configure(configFName string, envPrefix string, debug bool) error {
	if app == nil {
		app = new(Rdm2EPrint)
	}
	cfg := NewConfig()
	// Load the config file if name isn't an empty string
	if configFName != "" {
		err := cfg.LoadConfig(configFName)
		if err != nil {
			return err
		}
	}
	// Merge settings from the environment
	if err := cfg.LoadEnv(envPrefix); err != nil {
		return err
	}
	app.Cfg = cfg
	if debug {
		app.Cfg.Debug = true
	}
	// Make sure we have a minimal useful configuration
	if app.Cfg.InvenioAPI == "" || (app.Cfg.InvenioToken == "" && app.Cfg.InvenioDbHost == "") {
		return fmt.Errorf("RDM_URL, RDMTOK or RDM_DB_HOST are missing")
	}
	return nil
}

// CusePostgresDB, if RDM's Postgres DB setup in the environment use it to
// handle record and key retrieval rather than the slower REST API.
func usePostgresDB(cfg *Config) bool {
	if (cfg != nil)  && (cfg.InvenioDbHost != "") && (cfg.InvenioDbUser != "") {
		return true
	}
	return false
}

func (app *Rdm2EPrint) Run(in io.Reader, out io.Writer, eout io.Writer, rdmids []string, asXML bool) error {
	eprints := new(eprinttools.EPrints)
	if usePostgresDB(app.Cfg) {
		cfg := app.Cfg
		sslmode := "?sslmode=require"
		if strings.HasPrefix(cfg.InvenioDbHost, "localhost") {
			sslmode = "?sslmode=disable"
		}
		connStr := fmt.Sprintf("postgres://%s@%s/%s%s", 
		cfg.InvenioDbUser, cfg.InvenioDbHost, cfg.RepoID, sslmode)
		if cfg.InvenioDbPassword != "" {
			connStr = fmt.Sprintf("postgres://%s:%s@%s/%s%s", 
				cfg.InvenioDbUser, cfg.InvenioDbPassword, cfg.InvenioDbHost, cfg.RepoID, sslmode)
		}
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		defer db.Close()
		app.Cfg.pgDB = db
	}
	for _, rdmid := range rdmids {
		rec, err := GetRecord(app.Cfg, rdmid, false)
		if err != nil {
			return err
		}
		eprint := new(eprinttools.EPrint)
		if err := CrosswalkRdmToEPrint(app.Cfg, rec, eprint); err != nil {
			return err
		}
		eprints.EPrint = append(eprints.EPrint, eprint)
	}
	var (
		src []byte
		err error
	)
	if asXML {
		src, err = xml.MarshalIndent(eprints, "", "  ")
	} else {
		src, err = JSONMarshalIndent(eprints, "", "     ")
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s\n", src)
	return nil
}

func (app *Rdm2EPrint) RunHarvest(in io.Reader, out io.Writer, eout io.Writer, cName string, rdmids []string) error {
	if len(rdmids) == 0 {
		return fmt.Errorf("no RDM ids to process")
	}
	ds, err := dataset.Open(cName)
	if err != nil {
		return err
	}
	defer ds.Close()
	if usePostgresDB(app.Cfg) {
		cfg := app.Cfg
		sslmode := "?sslmode=require"
		if strings.HasPrefix(cfg.InvenioDbHost, "localhost") {
			sslmode = "?sslmode=disable"
		}
		connStr := fmt.Sprintf("postgres://%s@%s/%s%s", 
				cfg.InvenioDbUser, cfg.InvenioDbHost, cfg.RepoID, sslmode)
		if cfg.InvenioDbPassword != "" {
			connStr = fmt.Sprintf("postgres://%s:%s@%s/%s%s", 
				cfg.InvenioDbUser, cfg.InvenioDbPassword, cfg.InvenioDbHost, cfg.RepoID, sslmode)
		}
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		defer db.Close()
		app.Cfg.pgDB = db
	}

	eCnt, cCnt, tot := 0, 0, len(rdmids)
	t0 := time.Now()
	rptTime := time.Now()
	reportProgress := false
	log.Printf("Started processing %d records into %s", len(rdmids), cName)
	for i, rdmid := range rdmids {
		rec, err := GetRecord(app.Cfg, rdmid, false)
		if err != nil {
			log.Printf("Aborting, failed to get record (%d) %s, %s", i, rdmid, err)
			return err
		}
		eprint := new(eprinttools.EPrint)
		if err := CrosswalkRdmToEPrint(app.Cfg, rec, eprint); err != nil {
			log.Printf("Aborting, failed to crosswalk record (%d) %s, %s", i, rdmid, err)
			return err
		}
		if ds.HasKey(rec.ID) {
			if err := ds.UpdateObject(rec.ID, eprint); err != nil {
				log.Printf("error (update): %q, %s", rec.ID, err)
				eCnt++
			} else {
				cCnt++
			}
		} else {
			if err := ds.CreateObject(rec.ID, eprint); err != nil {
				log.Printf("error (create): %q, %s", rec.ID, err)
				eCnt++
			} else {
				cCnt++
			}
		}
		if rptTime, reportProgress = CheckWaitInterval(rptTime, (30 * time.Second)); reportProgress || (i % 10000) == 0 {
			log.Printf("%s %s (%d/%d) %s", cName, time.Since(t0).Round(time.Second), i, tot, ProgressETA(t0, i, tot))
		}
	}
	log.Printf("Finished %s, processed %d records in %s", cName, tot, time.Since(t0).Round(time.Second))
	log.Printf("%d errors encountered, %d processsed successfully", eCnt, cCnt)
	return nil
}

// Run in pipline mode, e.g. `eprint2rdm XXXXX-XXXXX | rdm2eprint` should round trip the EPrint record
// to RDM then back again. It reads from standard input and writes to standard out.
func (app *Rdm2EPrint) RunPipeline(in io.Reader, out io.Writer, eout io.Writer, asXML bool) error {
	eprint := new(eprinttools.EPrint)
	eprints := new(eprinttools.EPrints)
	rec := new(simplified.Record)
	src, err := io.ReadAll(in)
	if err != nil {
		return err
	}
	if err := JSONUnmarshal(src, &rec); err != nil {
		return err
	}
	if err := CrosswalkRdmToEPrint(app.Cfg, rec, eprint); err != nil {
		return err
	}
	eprints.EPrint = append(eprints.EPrint, eprint)
	if asXML {
		src, err = xml.MarshalIndent(eprints, "", "  ")
	} else {
		src, err = JSONMarshalIndent(eprints, "", "    ")
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s\n", src)
	return nil
}
