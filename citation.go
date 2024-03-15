package irdmtools

import (
	"fmt"
	"log"
	"path"
	"strings"

	// Caltech Library Packages
	"github.com/caltechlibrary/simplified"
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

	// Date holds a map to related citeproc item dates.
	Date map[string]*CitationDate `json:"dates,omitempty" xml:"dates,omitempty" yaml:"dates,omitempty"`

	// Abstract holds the abstract, useful for search applications, not needed fir CiteProc
	Abstract string `json:"abstract,omitempty" xml:"abstract,omitempty" yaml:"abstract,omitempty"`

	// DOI of object
	DOI string `json:"doi,omitempty" xml:"doi,omitempty" yaml:"doi,omitempty"`

	// Publisher holds the publisher's name
	Publisher string `json:"publisher,omitempty" xml:"publisher,omitempty" yaml:"publisher,omitempty"`

	// PublisherLocation holds the address or location description of the publiser (e.g. Los Angeles, CA)
	PublisherLocation string `json:"publisher_location,omitempty" xml:"publisher_location,omitempty" yaml:"publisher_location,omitempty"`

	// Publication holds the name of the journal or publication, e.g. "Journal of Olympic Thumb Wrestling"
	Publication string `json:"publication,omitempty" xml:"publication,omitempty" yaml:"publication,omitempty"`

	// PublicationDate is a string, can be an approximate date. It's the date used to sort citations by in terms of record availabilty
	// E.g. for Thesis this would be the graduation year, for monographs and internal reports this might be the date made publically
	// available.
	PublicationDate string `json:"publication_date,omitempty" xml:"publication_date,omitempty" yaml:"publication_date,omitempty"`

	// Series/SeriesNumber values from CaltechAUTHORS (mapped from custom fields)
	Series       string `json:"series,omitempty" xml:"series,omitempty" yaml:"series,omitempty"`
	SeriesNumber string `json:"series_number,omitempty" xml:"series_number,omitempty" yaml:"series_number,omitempty"`

	// Volume/Issue values mapped from CrossRef/DataCite data models
	Volume string `json:"volume,omitempty" xml:"volume,omitempty" yaml:"volume,omitempty"`
	Issue  string `json:"issue,omitempty" xml:"issue,omitempty" yaml:"issue,omitempty"`

	// Pages range
	Pages string `json:"pages,omitempty" xml:"pages,omitempty" yaml:"pages,omitempty"`
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
				cite.ResourceType = resourceType
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

