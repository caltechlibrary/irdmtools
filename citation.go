package irdmtools

import (
	"fmt"
	"log"
	"path"
	"strings"

	// Caltech Library Packages
	"github.com/caltechlibrary/simplified"
	"github.com/caltechlibrary/eprinttools"
)

// irdmtools provides a means of turning an EPrint or RDM record into a datastructure suitable
// to process with CSL (Citation Style Language) implementations. It is intended as a lighter weight
// representation of a repository item.

// CiteProc* types is based on a very fast read of the following websites reviewed on 2024-03-06.
// My interpretations of what I read probably contains errors!
//
// - <https://github.com/Sett17/citeproc-js-go>
// - <https://github.com/Juris-M/citeproc-js>
// - <https://github.com/citation-style-language/schema>, <https://citeproc-js.readthedocs.io/en/latest/>
// - <https://juris-m.github.io/>,
// - <http://fbennett.github.io/z2csl/>
// - <https://aurimasv.github.io/z2csl/typeMap.xml>
// - <https://github.com/citation-style-language/schema>
// - <https://github.com/Juris-M/citeproc-js/blob/master/citeproc.js>
// - <https://github.com/citation-style-language/schema/blob/master/schemas/input/csl-citation.json>
// - <https://github.com/citation-style-language/schema/blob/master/schemas/input/csl-data.json>
//

// Citation implements the data structure for CiteProc's Item representing a single
// bibliographic citation.
type Citation struct {
	//
	// The first four properties are requred. Without them they don't make sense in a dataset collection of citations.
	//

	// ID holds a citation id. This is formed from the originating collection (e.g. repository) and that collection's id, i.e. {REPO_ID}:{RECORD_ID}.
	ID string `json:"id,required" xml:"id,required" yaml:"id,required"`

	// Collection is the dataset collection the citation came from or the Repository collection name (E.g. authors, caltechauthors)
	Collection string `json:"collection,required" xml:"collection,required" yaml:"collection,required"`

	// CollectionID is the id used in the originating collection
	CollectionID string `json:"collection_id,required" xml:"collection_id,required" yaml:"collection_id,required"`

	// CiteUsingURL holds the URL the citation will use to reach the object.
	// This is normally the URL of the item in your repository. You could map this to the DOI or other
	// resolver system.
	CiteUsingURL string `json:"cite_using_url,required" xml:"cite_using_url,required" yaml:"cite_using_url,required"`

	// ResourceType is the string from the repository that identifies the type of resource the record is about
	ResourceType string `json:"resource_type,omitempty" xml:"resource_type,omitempty" yaml:"resource_type,omitempty"`

	// AlternateId a list of Item identifies, not in the CiteProc spec but useful to me and likely useful
	// in fielded searching, e.g. looking up a citation with a given ISBN ir ISSN
	AlternateId []*CitationIdentifier `json:"alternate_id,omitempty" xml:"alternate_di,omitempty" yaml:"alternate_id,omitempty"`

	// Type holds the citeproc "type" of bibliographic record. In DataCite records
	// this would be found in `.access.types.citeproc`.
	Type string `json:"type,omitempty" xml:"type,omitempty" yaml:"type,omitempty"`

	// Title holds the title used for the citation.
	Title string `json:"title,omitempty" xml:"title,omitempty" yaml:"title,omitempty"`

	// BookTitle holds a book title when the citation is a chapter contribution
	BookTitle string `json:"book_title,omitempty" xml:"book_title,omitempty" yaml:"book_title,omitempty"`

	// AlternateTitle holds additional titles refering to this item. Not part of the CiteProc item description but
	// useful for search purposes.
	AlternateTitle []string `json:"alternate_title,omitempty" xml:"alternate_title,omitempty" yaml:"alternate_title,omitempty"`

	// Author holds a list of author as CitationAgent objects.
	Author []*CitationAgent `json:"author,omitempty" xml:"author,omitempty" yaml:"author,omitempty"`

	// Editor holds a list of editor as CitationAgent objects
	Editor []*CitationAgent `json:"editor,omitempty" xml:"editor,omitempty" yaml:"editor,omitempty"`

	// Reviewer holds a list of reviewer as CitationAgent objects
	Reviewer []*CitationAgent `json:"reviewer,omitempty" xml:"reviewer,omitempty" yaml:"reviewer,omitempty"`

	// ThesisAdvisors holds a list of thesis advisors as CitationAgent objects
	ThesisAdvisor []*CitationAgent `json:"thesis_advisor,omitempty" xml:"thesis_advisor,omitempty" yaml:"thesis_advisor,omitempty"`

	// ThesisCommittee holds a list of thesis committee members as CitationAgent objects
	ThesisCommittee []*CitationAgent `json:"thesis_committee,omitempty" xml:"thesis_committee,omitempty" yaml:"thesis_committee,omitempty"`

	// Contributor holds a list of contributors (people who contributed but are not authors, editors, reviewers, thesis advisors, etc)
	Contributor []*CitationAgent `json:"contributor,omitempty" xml:"contributor,omitempty" yaml:"contributor,omitempty"`

	// Translator  holds a list of people who translated the work
	Translator []*CitationAgent `json:"translator,omitempty" xml:"translator,omitempty" yaml:"translator,omitempty"`

	// LocalGroup holds information about Caltech affiliated groups
	LocalGroup []*CitationAgent `json:"local_group,omitempty" xml:"local_group,omitempty" yaml:"local_group,omitempty"`

	// Date holds a map to related citeproc item dates. Currently unused.
	Date map[string]*CitationDate `json:"dates,omitempty" xml:"dates,omitempty" yaml:"dates,omitempty"`

	// Abstract holds the abstract, useful for search applications, not needed fir CiteProc
	Abstract string `json:"abstract,omitempty" xml:"abstract,omitempty" yaml:"abstract,omitempty"`

	// DOI of object
	DOI string `json:"doi,omitempty" xml:"doi,omitempty" yaml:"doi,omitempty"`

	// PMCID
	PMCID string `json:"pmcid,omitempty" xml:"pmcid,omitempty" yaml:"pmcid,omitempty"`

	// ISSN
	ISSN string `json:"issn,omitempty" xml:"issn,omitempty" yaml:"issn,omitempty"`

	// ISBN
	ISBN string `json:"isbn,omitempty" xml:"isbn,omitempty" yaml:"isbn,omitempty"`

	// Publisher holds the publisher's name
	Publisher string `json:"publisher,omitempty" xml:"publisher,omitempty" yaml:"publisher,omitempty"`

	// PlaceOfPublication holds the address or location description of the publiser (e.g. Los Angeles, CA)
	PlaceOfPublication string `json:"place_of_publication,omitempty" xml:"place_of_publication,omitempty" yaml:"place_of_publication,omitempty"`

	// Publication holds the name of the journal or publication, e.g. "Journal of Olympic Thumb Wrestling"
	Publication string `json:"publication,omitempty" xml:"publication,omitempty" yaml:"publication,omitempty"`

	// PublicationDate is a string, can be an approximate date. It's the date used to sort citations by in terms of record availabilty
	// E.g. for Thesis this would be the graduation year, for monographs and internal reports this might be the date made publically
	// available.
	PublicationDate string `json:"publication_date,omitempty" xml:"publication_date,omitempty" yaml:"publication_date,omitempty"`

	// Book related

	// Edition of book
	Edition string `json:"edition,omitempty" xml:"edition,omitempty" yaml:"edition,omitempty"`

	// Chapters from book
	Chapters string `json:"chapters,omitempty" xml:"chapters,omitempty" yaml:"chapters,omitempty"`


	// Series/SeriesNumber values from CaltechAUTHORS (mapped from custom fields)
	Series       string `json:"series,omitempty" xml:"series,omitempty" yaml:"series,omitempty"`
	SeriesNumber string `json:"series_number,omitempty" xml:"series_number,omitempty" yaml:"series_number,omitempty"`

	// Volume/Issue values mapped from CrossRef/DataCite data models
	Volume string `json:"volume,omitempty" xml:"volume,omitempty" yaml:"volume,omitempty"`
	Issue  string `json:"issue,omitempty" xml:"issue,omitempty" yaml:"issue,omitempty"`

	// Pages
	Pages string `json:"pages,omitempty" xml:"pages,omitempty" yaml:"pages,omitempty"`


	// ThesisDegree for thesis types
	ThesisDegree string `json:"thesis_degree,omitempty" xml:"thesis_degree,omitempty" yaml:"thesis_degree,omitempty"`

	// Thesis Type
	ThesisType string `json:"thesis_type,omitempty" xml:"thesis_type,omitempty" yaml:"thesis_type,omitempty"`

	// ThesisYear for thesis types, year degree granted
	ThesisYear string `json:"thesis_year,omitempty" xml:"thesis_year,omitempty" yaml:"thesis_year,omitempty"`

	// Patent citation data
	PatentApplication string `json:"patent_applicant,omitempty" xml:"patent_applicatant,omitempty" yaml:"patent_applicant,omitempty"`

	// Patent Assignee
	PatentAssignee string `json:"patent_assignee,omitempty" xml:"patent_assignee,omitempty" yaml:"patent_assignee,omitempty"`

	// Patent Number
	PatentNumber string `json:"patent_number,omitempty" xml:"patent_number,omitempty" yaml:"patent_number,omitempty"`
}

// CitationIdentifier is a minimal object to identify a type of identifier, e.g. ISBN, ISSN, ROR, ORCID, etc.
type CitationIdentifier struct {
	// Type holds the identifier type, e.g. ISSN, ISBN, ROR, ORCID
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Value holds an identifier value.
	Value string `json:"id,omitempty" yaml:"id,omitempty"`
}

// CitationAgent this describes a person or organization for the purposes of CiteProc item data.
// This is based on https://citeproc-js.readthedocs.io/en/latest/csl-json/markup.html, reviewed 2024-03-06.
type CitationAgent struct {
	// FamilyName holds a person's family name
	FamilyName string `json:"family_name,omitempty" xml:"family_name,omitempty" yaml:"family_name,omitempty"`

	// LivedName holds a person's lived or given. It is express encoded as "given" in JSON, XML and YAML for
	// to becompatible with historical records not as a justication for that "given" implies
	// in 2024 in the United States.
	LivedName string `json:"given_name,omitempty" xml:"given_name,omitempty" yaml:"given_name,omitempty"`

	// NonDroppingParticle holds non dropping particles that should not be dropped from a name, e.g. "de las"
	NonDroppingParticle string `json:"non-dropping-particle,omitempty" xml:"non-dropping-particle,omitempty" yaml:"non-dropping-particle,omitempty"`

	// DroppingParticle holds the list of particles that can be dropped.
	DroppingParticle string `json:"dropping-particle,omitempty" xml:"dropping-particle,omitempty" yaml:"dropping-particle,omitempty"`

	// Prefix, e.g. Mr., Mrs, Prof.
	Prefix string `json:"prefix,omitempty" xml:"prefix,omitempty" yaml:"prefix,omitempty"`

	// Suffix, e.g. Jr., PhD. etc.
	Suffix string `json:"suffix,omitempty" xml:"suffix,omitempty" yaml:"suffix,omitempty"`

	// Literal would be use for a group or organization, e.g. "ACME Widgets and Gadgets, Inc."
	Literal string `json:"literal,omitempty" xml:"literal,omitempty" yaml:"literal,omitempty"`

	// ORCID identifier
	ORCID string `json:"orcid,omitempty" xml:"orcid,omitempty" yaml:"orcid,omitempty"`

	// ISNI
	ISNI string `json:"isni,omitempty" xml:"isni,omitempty" yaml:"isni,omitempty"`

	// clpid - Caltech Library Person Identifier
	CLpid string `json:"clpid,omitempty" xml:"clpid,omitempty" yaml:"clpid,omitempty"`

	// clgid - Caltech Library Group Indentifier
	CLgid string `json:"clgid,omitempty" xml:"clgid,omitempty" yaml:"clgid,omitempty"`
}

// CitationDate holds date information, this includes support for partial dates (e.g. year, year-month).
type CitationDate struct {
	// DateParts holds the data parts expressed as array of an array of integers
	DateParts [][]int `json:"date-parts,omitempty" xml:"date-parts,omitempty" yaml:"date-parts,omitempty"`
	// Raw holds the raw string from a bibiographic source, e.g. publisher
	Raw string `json:"raw,omitempty" xml:"raw,omitempty" yaml:"raw,omitempty"`
}

// CrosswalkRecord takes a simplified record and return maps the values into the Citation.
func (cite *Citation) CrosswalkRecord(cName string, cID string, citeUsingURL string, rec *simplified.Record) error {
	// map repository required fields, everything else is derived from crosswalk
	cName = path.Base(strings.TrimSuffix(cName, ".ds"))
	cite.ID = strings.ToLower(fmt.Sprintf("%s:%s", cName, cID))
	cite.Collection = cName
	cite.CollectionID = cID
	cite.CiteUsingURL = citeUsingURL

	// Now crosswalk the rest of the citation from the simplified record.
	if rec.Metadata != nil {
		// map title of simplified record
		cite.Title = rec.Metadata.Title
		// map resource type from simplified record
		if rec.Metadata.ResourceType != nil {
			if resourceType, ok := rec.Metadata.ResourceType["id"].(string); ok {
				cite.Type = resourceType
			}
		}
		// map authors, contributors, editors, thesis advisors, committee members from simplified record
		for i, creator := range rec.Metadata.Creators {
			agent, _, err := CrosswalkCreatorToCitationAgent(creator)
			if err != nil {
				log.Printf("skipping author (%d), %s", i, err)
				continue
			}
			cite.Author = append(cite.Author, agent)
		}
		for i, contributor := range rec.Metadata.Contributors {
			agent, role, err := CrosswalkCreatorToCitationAgent(contributor)
			if err != nil {
				log.Printf("skipping contributor (%d), %s", i, err)
				continue
			}
			switch role {
			case "editor":
				cite.Editor = append(cite.Editor, agent)
			case "reviewer":
				cite.Reviewer = append(cite.Reviewer, agent)
			case "thesis_advisor":
				cite.ThesisAdvisor = append(cite.ThesisAdvisor, agent)
			case "thesis_committee":
				cite.ThesisCommittee = append(cite.ThesisCommittee, agent)
			case "translator":
				cite.Translator = append(cite.Translator, agent)
			default:
				cite.Contributor = append(cite.Contributor, agent)
			}
		}
		// map publisher from simplified record
		cite.Publisher = rec.Metadata.Publisher
		// map publication date
		cite.PublicationDate = rec.Metadata.PublicationDate
	}
	// DataCite doesn't provide some commonly used fields. These are getting mapped
	// into the CustomFields namespace in the object.
	if rec.CustomFields != nil {
		if journalInfo, ok := rec.CustomFields["journal:journal"].(map[string]string); ok {
			if publication, ok := journalInfo["title"]; ok {
				// map publication from simplified record
				// FIXME: DataCite/RDM do not provide an explicit publication field for some reason.
				// We use a custom field, `.custom_fields["journal:joural"].title` to identify publishers ...
				cite.Publication = publication
			}
			// map series/number from simplified record
			if series, ok := journalInfo["series"]; ok {
				cite.Series = series
			}
			if seriesNumber, ok := journalInfo["number"]; ok {
				cite.SeriesNumber = seriesNumber
			}
			// map volume/issue from simplified record
			if volume, ok := journalInfo["volume"]; ok {
				cite.Volume = volume
			}
			if issue, ok := journalInfo["issue"]; ok {
				cite.Issue = issue
			}
			if pages, ok := journalInfo["pages"]; ok {
				cite.Pages = pages
			}
		}
		if imprintInfo, ok := rec.CustomFields["imprint:imprint"].(map[string]interface{}); ok {
			if title, ok := imprintInfo["title"].(string); ok {
				cite.BookTitle = title
			}
			if chapters, ok := imprintInfo["chapters"].(string); ok {
				cite.Chapters = chapters
			}
			if isbn, ok := imprintInfo["isbn"].(string); ok {
				cite.ISBN = isbn
			}
			if pages, ok := imprintInfo["pages"].(string); ok {
				cite.Pages = pages
			}
			if place, ok := imprintInfo["place"].(string); ok {
				cite.PlaceOfPublication = place
			}
			if edition, ok := imprintInfo["edition"].(string); ok {
				cite.Edition = edition
			}
		}
		if caltechPlaceOfPubs, ok := rec.CustomFields["caltech:place_of_publication"].(map[string]interface{}); ok {
			if place, ok := caltechPlaceOfPubs["place"]; ok {
				cite.PlaceOfPublication = place.(string)
			}
		}
		if caltechSeries, ok := rec.CustomFields["caltech:series"].(map[string]interface{}); ok {
			if series, ok := caltechSeries["series"]; ok {
				cite.Series = series.(string)
			}
		}

		if caltechGroups, ok := rec.CustomFields["caltech:groups"].([]interface{}); ok {
			if len(caltechGroups) > 0 {
				groupList := []*CitationAgent{}
				for _, groups := range caltechGroups {
					if group, ok := groups.(map[string]interface{}); ok {
						addItem := false
						agent := new(CitationAgent)
						if id, ok := group["id"]; ok {
							addItem = true
							agent.CLgid = id.(string)
						}
						// FIXME: title is populated from translating the vocabulary against
						// and group id attribute so this is always empty.
						if title, ok := group["title"].(map[string]interface{}); ok {
							if en, ok := title["en"]; ok {
								addItem = true
								agent.Literal = en.(string)
							}
						}
						//NOTE: Need to make sure we're not adding a duplicate groups
						for i := 0; i < len(groupList); i++ {
							grp := groupList[i]
							if grp.CLgid == agent.CLgid {
								addItem = false
								break
							}
							if agent.Literal != "" && (grp.Literal == agent.Literal) {
								addItem = false
							}
						}
						if addItem {
							groupList = append(groupList, agent)
						}
					}
				}
				if len(groupList) > 0 {
					cite.LocalGroup = groupList
				}
			}
		}
	}

	// map CiteUsingURL from simplified record
	// map doi from simplified record
	if rec.ExternalPIDs != nil {
		if doi, ok := rec.ExternalPIDs["doi"]; ok {
			cite.DOI = doi.Identifier
			if citeUsingURL == "" && cite.DOI != "" {
				cite.CiteUsingURL = "https://doi.org/" + strings.TrimPrefix(cite.DOI, "https://doi.org/")
			}
		}
		if pmcid, ok := rec.ExternalPIDs["pmcid"]; ok {
			cite.PMCID = pmcid.Identifier
		}
		if isbn, ok := rec.ExternalPIDs["isbn"]; ok {
			cite.ISBN = isbn.Identifier
		}
		if issn, ok := rec.ExternalPIDs["issn"]; ok {
			cite.ISSN = issn.Identifier
		}
	}
	return nil
}

// CrosswalkCreatorToCitationAgent takes a simplified.Cretor and returns a CitationAgent, role (e.g. "author", "editor", "thesis_advisor", "thesis_committee", "reviewer", "contributor"), and an error value
func CrosswalkCreatorToCitationAgent(creator *simplified.Creator) (*CitationAgent, string, error) {
	if creator.PersonOrOrg == nil {
		return nil, "", fmt.Errorf("create.PersonOrOrg is nil")
	}
	citationAgent, err := CrosswalkPersonOrOrgToCitationAgent(creator.PersonOrOrg)
	if err != nil {
		return nil, "", err
	}
	return citationAgent, "", nil
}

// CrosswalkPersonOrOrgToCitationAgent takes a simplified.PersonOrOrg and returns a CitationAgent
func CrosswalkPersonOrOrgToCitationAgent(personOrOrg *simplified.PersonOrOrg) (*CitationAgent, error) {
	citationAgent := new(CitationAgent)
	if personOrOrg.FamilyName != "" {
		citationAgent.FamilyName = personOrOrg.FamilyName
	}
	if personOrOrg.GivenName != "" {
		citationAgent.LivedName = personOrOrg.GivenName
	}
	if personOrOrg.Name != "" {
		citationAgent.Literal = personOrOrg.Name
	}
	// Map personOrOrg identifiers
	for _, identifier := range personOrOrg.Identifiers {
		switch identifier.Scheme {
		case "clpid":
			citationAgent.CLpid = identifier.Identifier
		case "orcid":
			citationAgent.ORCID = identifier.Identifier
		case "isni":
			citationAgent.ISNI = identifier.Identifier
		}
	}
	return citationAgent, nil
}

// ToString convert a CitationAgent to a string representation
func (ca *CitationAgent) ToString() string {
	return fmt.Sprintf("%s, %s", ca.FamilyName, ca.LivedName)
}


// CrosswalkEPrint takes an eprinttools.EPrint record and return maps the values into the Citation.
func (cite *Citation) CrosswalkEPrint(cName string, cID string, citeUsingURL string, eprint *eprinttools.EPrint) error {
	// map repository required fields, everything else is derived from crosswalk
	cName = path.Base(strings.TrimSuffix(cName, ".ds"))
	cite.ID = strings.ToLower(fmt.Sprintf("%s:%s", cName, cID))
	cite.Collection = cName
	cite.CollectionID = cID
	cite.CiteUsingURL = citeUsingURL

	// from the eprint table
	cite.Title = eprint.Title
	cite.Type = eprint.Type
	cite.Abstract = eprint.Abstract
	cite.Publisher = eprint.Publisher
	cite.Publication = eprint.Publication
	cite.PublicationDate = eprint.PubDate()
	cite.BookTitle = eprint.BookTitle
	// Not sure where to find the chapter information in EPrints record.
	//cite.Chapters = eprint.Chapters
	cite.PlaceOfPublication = eprint.PlaceOfPub
	cite.Edition = eprint.Edition
	cite.Series = eprint.Series
	cite.SeriesNumber = eprint.Number
	cite.Volume = eprint.Volume
	cite.Issue = eprint.Number
	cite.Pages = eprint.PageRange
	cite.ISBN = eprint.ISBN
	cite.ISSN = eprint.ISSN
	cite.DOI = eprint.DOI
	cite.PMCID = eprint.PMCID
	
	if eprint.ThesisType != "" {
		cite.ThesisType = eprint.ThesisType
	}
	if eprint.DateType == "degree" {
		cite.PublicationDate = eprint.Date
		cite.ThesisYear = eprint.Date[0:4]
	}
	if eprint.OfficialURL != "" {
		cite.CiteUsingURL = eprint.OfficialURL
	}

	// map authors, contributors, editors, thesis advisors, committee members from eprint_*
	if eprint.Creators.Length() > 0 {
		for i := 0; i < eprint.Creators.Length(); i++ {
			creator := eprint.Creators.IndexOf(i)
			if creator.Name != nil {
				agent := new(CitationAgent)
				agent.FamilyName = creator.Name.Family
				agent.LivedName = creator.Name.Given
				agent.CLpid = creator.Name.ID
				agent.ORCID = creator.Name.ORCID
				agent.Prefix = creator.Name.Honourific
				agent.Suffix = creator.Name.Lineage
				cite.Author = append(cite.Author, agent)
			}
		}
	}

	// Map in corporate authors
	if eprint.CorpCreators.Length() > 0 {
		for i := 0; i < eprint.CorpCreators.Length(); i++ {
			creator := eprint.CorpCreators.IndexOf(i)
			if creator.Value != "" {
				agent := new(CitationAgent)
				agent.Literal = creator.Value
				cite.Author = append(cite.Author, agent)
			}
		}
	}

	// Map in editors
	if eprint.Editors.Length() > 0 {
		for i := 0; i < eprint.Editors.Length(); i++ {
			creator := eprint.Editors.IndexOf(i)
			if creator.Name != nil {
				agent := new(CitationAgent)
				agent.FamilyName = creator.Name.Family
				agent.LivedName = creator.Name.Given
				agent.CLpid = creator.Name.ID
				agent.ORCID = creator.Name.ORCID
				agent.Prefix = creator.Name.Honourific
				agent.Suffix = creator.Name.Lineage
				cite.Editor = append(cite.Editor, agent)
			}
		}
	}

	// Map in contributors
	if eprint.Contributors.Length() > 0 {
		for i := 0; i < eprint.Contributors.Length(); i++ {
			creator := eprint.Contributors.IndexOf(i)
			if creator.Name != nil {
				agent := new(CitationAgent)
				agent.FamilyName = creator.Name.Family
				agent.LivedName = creator.Name.Given
				agent.CLpid = creator.Name.ID
				agent.ORCID = creator.Name.ORCID
				agent.Prefix = creator.Name.Honourific
				agent.Suffix = creator.Name.Lineage
				cite.Contributor = append(cite.Contributor, agent)
			}
		}
	}

	// Map in corporate contributors
	if eprint.CorpContributors.Length() > 0 {
		for i := 0; i < eprint.CorpContributors.Length(); i++ {
			creator := eprint.CorpContributors.IndexOf(i)
			if creator.Value != "" {
				agent := new(CitationAgent)
				agent.Literal = creator.Value
				cite.Contributor = append(cite.Contributor, agent)
			}
		}
	}	

	// map in Thesis Adivors 
	if eprint.ThesisAdvisor.Length() > 0 {
		for i := 0; i < eprint.ThesisAdvisor.Length(); i++ {
			creator := eprint.ThesisAdvisor.IndexOf(i)
			if creator.Name != nil {
				agent := new(CitationAgent)
				agent.FamilyName = creator.Name.Family
				agent.LivedName = creator.Name.Given
				agent.CLpid = creator.Name.ID
				agent.ORCID = creator.Name.ORCID
				agent.Prefix = creator.Name.Honourific
				agent.Suffix = creator.Name.Lineage
				cite.ThesisAdvisor = append(cite.ThesisAdvisor, agent)
			}
		}
	}

	// map in Thesis committee
	if eprint.ThesisCommittee.Length() > 0 {
		for i := 0; i < eprint.ThesisCommittee.Length(); i++ {
			creator := eprint.ThesisCommittee.IndexOf(i)
			if creator.Name != nil {
				agent := new(CitationAgent)
				agent.FamilyName = creator.Name.Family
				agent.LivedName = creator.Name.Given
				agent.CLpid = creator.Name.ID
				agent.ORCID = creator.Name.ORCID
				agent.Prefix = creator.Name.Honourific
				agent.Suffix = creator.Name.Lineage
				cite.ThesisCommittee = append(cite.ThesisCommittee, agent)
			}
		}
	}

	// map local groups from eprint_local_group table
	for i := 0; i < eprint.LocalGroup.Length(); i++ {
		group := eprint.LocalGroup.IndexOf(i)
		if group != nil && group.Value != "" {
			agent := new(CitationAgent)
			agent.Literal = group.Value
			///NOTE: CLgid can't be mapped directly from EPrints as there is no group id field of any type.
			cite.LocalGroup = append(cite.LocalGroup, agent)
		}
	}

	// If we are processing thesis then we can merge division list in with Groups
	if eprint.Divisions.Length() > 0 {
		for i := 0; i < eprint.Divisions.Length(); i++ {
			group := eprint.Divisions.IndexOf(i)
			if group != nil && group.Value != "" {
				agent := new(CitationAgent)
				agent.Literal = group.Value
				// NOTE: CLgid can't be mapped directly from EPrint Divisions
				cite.LocalGroup = append(cite.LocalGroup, agent)
			}
		}
	}
	return nil
}
