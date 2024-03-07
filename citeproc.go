package irdmtools

import (
	"github.com/gofrs/uuid"
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

// CiteProcItem implements the data structure for CiteProc's Item representing a single
// bibliographic citation.
type CiteProcItem struct {
	// Uuid isn't part of CiteProc's spec but it can be handly when dealing with records aggregated from various
	// sources that don't have something like a DOI.
	Uuid uuid.UUID `json:"uuid,omitempty" xml:"uuid,omitempty" yaml:"uuid,omitempty"`

	// Id holds the primary unique identifier for a given citation. In BibTeX this is the string after the opening "{".
	Id string `json:"id,omitempty" xml:"id,omitempty" yaml:"id,omitempty"`

	// AlternateId a list of Item identifies, not in the CiteProc spec but useful to me and likely useful
	// in fielded searching, e.g. looking up a citation with a given ISBN ir ISSN
	AlternateId []*CiteProcIdentifier `json:"alternate_id,omitempty" xml:"alternate_di,omitempty" yaml:"alternate_id,omitempty"`
	
	// Type holds the citeproc "type" of bibliographic record. In DataCite records
	// this would be found in `.access.types.citeproc`.
	Type string `json:"type,omitempty" xml:"type,omitempty" yaml:"type,omitempty"`

	// Title holds the title used for the citation.
	Title string `json:"title,omitempty" xml:"title,omitempty" yaml:"title,omitempty"`

	// AlternateTitle holds additional titles refering to this item. Not part of the CiteProc item description but
	// useful for search purposes.
	AlternateTitle []string `json:"alternate_title,omitempty" xml:"alternate_title,omitempty" yaml:"alternate_title,omitempty"`

	// Author holds a list of "author" objects. 
	Author []*CiteProcAgent `json:"author,omitempty" xml:"author,omitempty" yaml:"author,omitempty"`

	// Date holds a map to related citeproc item dates.
	Date map[string]*CiteProcDate `json:"dates,omitempty" xml:"dates,omitempty" yaml:"dates,omitempty"`

	// Abstract holds the abstract, useful for search applications, not needed fir CiteProc
	Abstract string `json:"abstract,omitempty" xml:"abstract,omitempty" yaml:"abstract,omitempty"`

	// Prefix would appear before the citation item.
	Prefix string `json:"prefix,omitempty" xml:"prefix,omitempty" yaml:"prefix,omitempty"`

	// Suffix would appear after the citation, e.g. "see ..."
	Suffix string `json:"suffix,omitempty" xml:"suffix,omitempty" yaml:"suffix,omitempty"`
}

// CiteProcAgent this describes a person or organization for the purposes of CiteProc item data.
// This is based on https://citeproc-js.readthedocs.io/en/latest/csl-json/markup.html, reviewed 2024-03-06.
type CiteProcAgent struct {
	// Uuid isn't part of CiteProc's person object but if you're storing lists of
	// people like when you aggregate publications by people then a UUID is handy. Not
	// everyone has an ORCID, ISNI, Viaf, etc.
	Uuid uuid.UUID `json:"uuid,omitemtpy" xml:"uuid,omitempty" yaml:"uuid,omitempty"`

	// Family holds a person's family name
	Family string `json:"family,omitempty" xml:"family,omitempty" yaml:"family,omitempty"`

	// Lived holds a person's given or lived name. It is express encoded as "given" for
	// to becompatible with historical records not at a justication for that "given" implies
	// in 2024 in the United States.
	Lived string `json:"given,omitempty" xml:"given,omitempty" yaml:"given,omitempty"`

	// NonDroppingParticle holds non dropping particles that should not be dropped from a name, e.g. "de las"
	NonDroppingParticle string `json:"non-dropping-particle,omitempty" xml:"non-dropping-particle,omitempty" yaml:"non-dropping-particle,omitempty"`

	// DroppingParticle holds the list of particles that can be dropped.
	DroppingParticle string `json:"dropping-particle,omitempty" xml:"dropping-particle,omitempty" yaml:"dropping-particle,omitempty"`

	// Prefix (FIXME: verify this exists in a CiteProc person reference), because they can exist ... 
	Prefix string `json:"prefix,omitempty" xml:"prefix,omitempty" yaml:"prefix,omitempty"`

	// Suffix, e.g. Jr., PhD. etc.
	Suffix string `json:"suffix,omitempty" xml:"suffix,omitempty" yaml:"suffix,omitempty"`

	// Literal would be use for a group or organization, e.g. "ACME Widgets and Gadgets, Inc."
	Literal string `json:"literal,omitempty" xml:"literal,omitempty" yaml:"literal,omitempty"`
}

// CiteProcDate holds date information, this includes support for partial dates (e.g. year, year-month).
type CiteProcDate struct {
	// DateParts holds the data parts expressed as array of an array of integers
	DateParts [][]int `json:"date-parts,omitempty" xml:"date-parts,omitempty" yaml:"date-parts,omitempty"`
	// Raw holds the raw string from a bibiographic source, e.g. publisher
	Raw string `json:"raw,omitempty" xml:"raw,omitempty" yaml:"raw,omitempty"`
}
