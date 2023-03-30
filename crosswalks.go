package irdmtools

import (
	"encoding/json"
	"fmt"
	"strings"

	// Caltech Library Packages
	"github.com/caltechlibrary/crossrefapi"
	"github.com/caltechlibrary/simplified"
)

// normalizeCrossRefType converts content type from CrossRef
// to Authors (e.g. "journal-article" to "publication-article")
func normalizeCrossRefType(s string) string {
	switch strings.ToLower(s) {
	case "proceedings-article":
		//FIXME: this mapping may not be correct, was book_section in EPrints CaltechAUTHORS
		return "publication-section"
	case "journal-article":
		return "publication-article"
	case "book-chapter":
		return "publication-section"
	default:
		return s
	}
}

// getResourceType retrives the resource type from works.message.type
// runs normalize
func getResourceType(work *crossrefapi.Works) string {
	if works.Message != nil {
		return works.Message.Type
	}
	return ""
}

// CrosswalkCrossRefWork takes a Works object from the CrossRef API
// and maps the fields into an simplified Record struct return a new struct or
// error.
func CrosswalkCrossRefWork(work crossrefapi.Works) (*Record, error) {
	rec := new(simplified.Record)
	obj := work.Message
	// .message.type -> .record.metadata.resource_type (via controlled vocabulary)
	if resourceType := getResourceType(work); resourceType != "" {
		if err := rec.SetRecourseType(resourceType); err != nil {
			return nil, err
		}
	}
	/*
	if s, ok := indexInto(obj, "message", "type"); ok == true {
		rec.Type = normalizeCrossRefType(fmt.Sprintf("%s", s))
	} else {
		return nil, fmt.Errorf("Can't find type in object")
	}
	// Title
	if a, ok := indexInto(obj, "message", "title"); ok == true {
		if len(a.([]interface{})) > 0 {
			eprint.Title = fmt.Sprintf("%s", a.([]interface{})[0].(string))
		}
	}

	// NOTE: Assuming IsPublished is true given that we're talking to
	// CrossRef API which holds published content.
	// IsPublished
	eprint.IsPublished = "pub"

	// Publisher
	if s, ok := indexInto(obj, "message", "publisher"); ok == true {
		eprint.Publisher = fmt.Sprintf("%s", s)
	}
	// Publication
	if eprint.Type == "article" {
		if l, ok := indexInto(obj, "message", "container-title"); ok == true {
			if len(l.([]interface{})) > 0 {
				eprint.Publication = l.([]interface{})[0].(string)
			}
		}
	}
	// Series
	if eprint.Type == "book" {
		if l, ok := indexInto(obj, "message", "container-title"); ok == true {
			if len(l.([]interface{})) > 0 {
				eprint.Series = l.([]interface{})[0].(string)
			}
		}
	}
	if l, ok := indexInto(obj, "message", "short-container-title"); ok == true {
		if len(l.([]interface{})) > 0 {
			eprint.Series = l.([]interface{})[0].(string)
		}
	}

	// Volume
	if eprint.Type == "article" {
		if s, ok := indexInto(obj, "message", "volume"); ok == true {
			eprint.Volume = fmt.Sprintf("%s", s)
		}
		// Number
		if s, ok := indexInto(obj, "message", "journal-issue", "issue"); ok == true {
			eprint.Number = fmt.Sprintf("%s", s)
		}

	}

	// PlaceOfPub taken from publisher-location in CrossRef
	if s, ok := indexInto(obj, "message", "publisher-location"); ok == true {
		eprint.PlaceOfPub = s.(string)
	}

	// FIXME: this needs to be article number, per migration committee
	// PageRange
	if s, ok := indexInto(obj, "message", "page"); ok == true {
		eprint.PageRange = s.(string)

	}

	// ISBN
	if a, ok := indexInto(obj, "message", "ISBN"); ok == true {
		if len(a.([]interface{})) > 0 {
			s := a.([]interface{})[0]
			eprint.ISBN = s.(string)
		}
	}

	// ISSN
	if a, ok := indexInto(obj, "message", "ISSN"); ok == true {
		if len(a.([]interface{})) > 0 {
			s := a.([]interface{})[0]
			eprint.ISSN = s.(string)
		}
	}

	// NOTE: This doesn't appear to be used by CaltechAUTHORS for full book
	// BookTitle
	if eprint.Title != "" && eprint.Type == "book" {
		eprint.BookTitle = eprint.Title
	}

	// Funders
	if a, ok := indexInto(obj, "message", "funder"); ok == true {
		eprint.Funders = new(FunderItemList)
		for _, entry := range a.([]interface{}) {
			var agency string
			m := entry.(map[string]interface{})
			if name, ok := indexInto(m, "name"); ok == true && name != "N/A" {
				agency = name.(string)
			}
			if a2, ok := indexInto(m, "award"); ok == true && a2 != "N/A" {
				for _, number := range a2.([]interface{}) {
					item := new(Item)
					item.Agency = agency
					item.GrantNumber = number.(string)
					eprint.Funders.Append(item)
				}
			} else {
				item := new(Item)
				item.Agency = agency
				if item.Agency != "" || item.GrantNumber != "" {
					eprint.Funders.Append(item)
				}
			}
		}
	}

	// NOTE: Caltech Library puts the DOI in the related URL field rather than
	// in EPrint's default location. This code puts the DOI in the default
	// location. If you need Caltech Library's bahavior use clsrules.Apply()
	// to conform to that regime.
	if doi, ok := indexInto(obj, "message", "DOI"); ok == true {
		eprint.DOI = doi.(string)
	}
	if l, ok := indexInto(obj, "message", "update-to"); ok == true {
		for _, o := range l.([]interface{}) {
			m := o.(map[string]interface{})
			if newDoi, ok := indexInto(m, "DOI"); ok == true && newDoi != "" {
				dt, _ := indexInto(m, "updated", "date-time")
				when := dt.(string)
				l, _ := indexInto(m, "label")
				label := l.(string)
				if len(when) > 10 {
					when = when[0:10]
				}
				entry := new(Item)
				entry.Type = "doi"
				entry.URL = fmt.Sprintf("https://doi.org/%s", newDoi)
				entry.Description = fmt.Sprintf("%s, %s", label, when)
				if eprint.RelatedURL == nil {
					eprint.RelatedURL = new(RelatedURLItemList)
				}
				eprint.RelatedURL.Append(entry)
			}
		}
	}

	// RelatedURLs (links in message of CrossRef works object)
	if l, ok := indexInto(obj, "message", "link"); ok == true {
		if eprint.RelatedURL == nil {
			eprint.RelatedURL = new(RelatedURLItemList)
		}
		for _, o := range l.([]interface{}) {
			entry := new(Item)
			if s, ok := indexInto(o.(map[string]interface{}), "URL"); ok == true {
				entry.URL = s.(string)
			}
			// NOTE: Related URL Type is not related to mime-type,
			// import related URLs without type information.
			if s, ok := indexInto(o.(map[string]interface{}), "type"); ok == true {
				entry.Type = s.(string)
			}
			if len(entry.URL) > 0 { //&& len(entry.Type) > 0 {
				eprint.RelatedURL.Append(entry)
			}
		}
	}

	// NOTE: We prefer the publication date of published-print and
	// fallback to issued date then finally created date.
	eprint.DateType = "published"
	if published, ok := indexInto(obj, "message", "published-print", "date-parts"); ok == true {
		var l1, l2 []interface{}
		if len(published.([]interface{})) == 1 {
			l1 = published.([]interface{})
			l2 = l1[0].([]interface{})
			ymd := []string{}
			for _, v := range l2 {
				n := v.(json.Number).String()
				if len(n) < 2 {
					n = "0" + n
				}
				ymd = append(ymd, n)
			}
			eprint.Date = strings.Join(ymd, "-")
		}
	} else if issued, ok := indexInto(obj, "message", "issued", "date-time"); ok == true {
		// DateType
		eprint.Date = fmt.Sprintf("%s", issued)
	} else if created, ok := indexInto(obj, "message", "created", "date-time"); ok == true {
		// DateType
		eprint.Date = fmt.Sprintf("%s", created)
	}
	if len(eprint.Date) > 10 {
		eprint.Date = eprint.Date[0:10]
	}

	// Authors list
	if l, ok := indexInto(obj, "message", "author"); ok == true {
		creators := new(CreatorItemList)
		corpCreators := new(CorpCreatorItemList)
		for _, entry := range l.([]interface{}) {
			author := entry.(map[string]interface{})
			item := new(Item)
			item.Name = new(Name)
			if orcid, ok := author["ORCID"]; ok == true {
				item.ORCID = orcid.(string)
				if strings.HasPrefix(orcid.(string), "http://orcid.org/") {
					item.ORCID = strings.TrimPrefix(orcid.(string), "http://orcid.org/")
				}
				if strings.HasPrefix(orcid.(string), "https://orcid.org/") {
					item.ORCID = strings.TrimPrefix(orcid.(string), "https://orcid.org/")
				}
			}
			if family, ok := author["family"]; ok == true {
				item.Name.Family = family.(string)
			}
			if given, ok := author["given"]; ok == true {
				item.Name.Given = given.(string)
			}
			//NOTE: if as have a 'name' then we'll add it to
			// as a corp_creators
			if name, ok := author["name"]; ok == true {
				item.Name.Value = strings.TrimSpace(name.(string))
				if strings.HasPrefix(item.Name.Value, "(") && strings.HasSuffix(item.Name.Value, ")") {
					item.Name.Value = strings.TrimSuffix(strings.TrimPrefix(item.Name.Value, "("), ")")
				}
			}
			if item.Name.Given != "" && item.Name.Family != "" {
				creators.Append(item)
			}
			if item.Name.Value != "" {
				corpCreators.Append(item)
			}
		}
		if len(creators.Items) > 0 {
			eprint.Creators = creators
		}
		if len(corpCreators.Items) > 0 {
			eprint.CorpCreators = corpCreators
		}
	}

	// Editors (*EditorItemList)
	if l, ok := indexInto(obj, "message", "editor"); ok == true {
		editors := new(EditorItemList)
		for _, entry := range l.([]interface{}) {
			editor := entry.(map[string]interface{})
			item := new(Item)
			item.Name = new(Name)
			if orcid, ok := editor["ORCID"]; ok {
				item.ORCID = orcid.(string)
				if strings.HasPrefix(orcid.(string), "http://orcid.org/") {
					item.ORCID = strings.TrimPrefix(orcid.(string), "http://orcid.org/")
				}
				if strings.HasPrefix(orcid.(string), "https://orcid.org/") {
					item.ORCID = strings.TrimPrefix(orcid.(string), "https://orcid.org/")
				}
				if family, ok := editor["family"]; ok == true {
					item.Name.Family = family.(string)
				}
				if given, ok := editor["given"]; ok == true {
					item.Name.Given = given.(string)
				}
				//NOTE: if as have a 'name' then we'll add it to
				// as a corp_creators
				if name, ok := editor["name"]; ok == true {
					item.Name.Value = strings.TrimSpace(name.(string))
					if strings.HasPrefix(item.Name.Value, "(") && strings.HasSuffix(item.Name.Value, ")") {
						item.Name.Value = strings.TrimSuffix(strings.TrimPrefix(item.Name.Value, "("), ")")
					}
				}
				if item.Name.Given != "" && item.Name.Family != "" {
					editors.Append(item)
				}
			}
		}
		if len(editors.Items) > 0 {
			eprint.Editors = editors
		}
	}

	// Abstract
	if abstract, ok := indexInto(obj, "message", "abstract"); ok {
		eprint.Abstract = fmt.Sprintf("%s", abstract)
	}

	// Edition
	if edition, ok := indexInto(obj, "message", "edition-number"); ok {
		eprint.Abstract = fmt.Sprintf("%s", edition)
	}

	// Subjects
	if l, ok := indexInto(obj, "message", "subject"); ok {
		subjects := new(SubjectItemList)
		for _, entry := range l.([]interface{}) {
			item := new(Item)
			item.SetAttribute("value", entry.(string))
			subjects.Append(item)
		}
		if subjects.Length() > 0 {
			eprint.Subjects = subjects
		}
	}
	//FIXME: Need to find value in CrossRef works metadata for this

	// Keywords
	//FIXME: Need to find value in CrossRef works metadata for this

	// FullTextStatus
	//FIXME: Need to find value in CrossRef works metadata for this

	// Note
	//FIXME: Need to find value in CrossRef works metadata for this

	//FIXME: Need to find value in CrossRef works metadata for this

	// Refereed
	//FIXME: Need to find value in CrossRef works metadata for this

	// Projects
	//FIXME: Need to find value in CrossRef works metadata for this

	// Contributors (*ContriborItemList)
	//FIXME: Need to find value in CrossRef works metadata for this

	// MonographType
	//FIXME: Need to find value in CrossRef works metadata for this

	// PresType (presentation type)
	//FIXME: Need to find value in CrossRef works metadata for this
	return eprint, nil
	*/
	return rec, fmt.ErrorF("CrosswalkCrossRefWorks() not implemented")
}
