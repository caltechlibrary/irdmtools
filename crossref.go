package irdmtools

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	// Caltech Library Packages
	"github.com/caltechlibrary/crossrefapi"
	"github.com/caltechlibrary/simplified"
)

func QueryCrossRefWork(cfg *Config, doi string, mailTo string, dotInitials bool, downloadDocument bool) (*crossrefapi.Works, error) {
	appName := path.Base(os.Args[0])
	client, err := crossrefapi.NewCrossRefClient(appName, mailTo)
	if err != nil {
		return nil, err
	}
	works, err := client.Works(doi)
	if err != nil {
		return nil, err
	}
	if cfg.Debug {
		src, _ := json.MarshalIndent(works, "", "    ")
		fmt.Fprintf(os.Stderr, "works JSON:\n\n%s\n\n", src)
	}
	return works, nil
}

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
	if work.Message != nil {
		return work.Message.Type
	}
	return ""
}

// getTitles retrieves an ordered list of titles from a CrossRef Works object.
// The zero index is the primary document title, the remaining are alternative titles.
// If no titles are found then the slice of string will be empty.
func getTitles(work *crossrefapi.Works) []string {
	if work.Message != nil && work.Message.Title != nil && len(work.Message.Title) > 0 {
		return work.Message.Title[:]
	}
	return []string{}
}

// getPublisher
func getPublisher(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.Publisher != "" {
		return work.Message.Publisher
	}
	return ""
}

// getPublication
func getPublication(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.ContainerTitle != nil && len(work.Message.ContainerTitle) > 0 {
		return work.Message.ContainerTitle[0]
	}
	return ""
}

// getSeries
func getSeries(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.ShortContainerTitle != nil && len(work.Message.ShortContainerTitle) > 0 {
		return work.Message.ShortContainerTitle[0]
	}
	return ""
}

// getVolume
func getVolume(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.JournalIssue != nil && work.Message.JournalIssue.Issue != "" {
		return work.Message.JournalIssue.Issue
	}
	return ""
}

// getPublisherLocation
func getPublisherLocation(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.PublisherLocation != "" {
		return work.Message.PublisherLocation
	}
	return ""
}

// getPageRange
func getPageRange(work *crossrefapi.Works) string {
	/*
		// FIXME: this needs to be article number, per migration committee
		// PageRange
		if s, ok := indexInto(obj, "message", "page"); ok == true {
			eprint.PageRange = s.(string)
		}
	*/
	if work.Message != nil && work.Message.Page != "" {
		return work.Message.Page
	}
	return ""
}

// getArticleNumber
func getArticleNumber(work *crossrefapi.Works) string {
	/* FIXME: Not sure where article numbers map from in the CrossRef API
	- ComponentNumber
	- PartNumber
	*/
	if work.Message != nil && work.Message.ArticleNumber != "" {
		return work.Message.ArticleNumber
	}
	return ""
}

// getISBNs
func getISBNs(work *crossrefapi.Works) []*simplified.Identifier {
	isbns := []*simplified.Identifier{}
	if work.Message != nil && work.Message.ISBN != nil {
		for _, value := range work.Message.ISBN {
			isbns = append(isbns, &simplified.Identifier{Scheme: "ISBN", Identifier: value})
		}
	}
	return isbns
}

// getISSNs
func getISSNs(work *crossrefapi.Works) []*simplified.Identifier {
	issns := []*simplified.Identifier{}
	if work.Message != nil && work.Message.ISSN != nil {
		for _, value := range work.Message.ISSN {
			issns = append(issns, &simplified.Identifier{Scheme: "ISSN", Identifier: value})
		}
	}
	return issns
}

// getFunding
func getFunding(work *crossrefapi.Works) []*simplified.Funder {
	funding := []*simplified.Funder{}
	if work.Message != nil && work.Message.Funder != nil && len(work.Message.Funder) > 0 {
		for _, funder := range work.Message.Funder {
			for _, award := range funder.Award {
				funding = append(funding, &simplified.Funder{
					Funder: &simplified.Identifier{
						Name: funder.Name,
					},
					Award: &simplified.Identifier{
						Number: award,
					},
				})
			}
		}
	}
	return funding
}

// getDOI
func getDOI(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.DOI != "" {
		return work.Message.DOI
	}
	return ""
}

// getLinks
func getLinks(work *crossrefapi.Works) []*simplified.Identifier {
	identifiers := []*simplified.Identifier{}
	if work.Message != nil && work.Message.Link != nil && len(work.Message.Link) > 0 {
		for _, link := range work.Message.Link {
			identifiers = append(identifiers, &simplified.Identifier{
				Scheme:     "URL",
				Identifier: link.URL,
				Name:       link.ContentType,
			})
		}
	}
	return identifiers
}

// CrosswalkCrossRefWork takes a Works object from the CrossRef API
// and maps the fields into an simplified Record struct return a new struct or
// error.
func CrosswalkCrossRefWork(cfg *Config, work *crossrefapi.Works) (*simplified.Record, error) {
	rec := new(simplified.Record)
	// .message.type -> .record.metadata.resource_type (via controlled vocabulary)
	if value := getResourceType(work); value != "" {
		if err := SetResourceType(rec, value); err != nil {
			return nil, err
		}
	}
	if values := getTitles(work); len(values) > 0 {
		if err := SetTitles(rec, values); err != nil {
			return nil, err
		}
	}
	if value := getPublisher(work); value != "" {
		if err := SetPublisher(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getPublication(work); value != "" {
		if err := SetPublication(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getSeries(work); value != "" {
		if err := SetSeries(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getVolume(work); value != "" {
		if err := SetVolume(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getPublisherLocation(work); value != "" {
		if err := SetPublisherLocation(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getPageRange(work); value != "" {
		if err := SetPageRange(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getArticleNumber(work); value != "" {
		if err := SetArticleNumber(rec, value); err != nil {
			return nil, err
		}
	}
	if values := getISBNs(work); values != nil && len(values) > 0 {
		if err := SetISBNs(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getISSNs(work); values != nil && len(values) > 0 {
		if err := SetISSNs(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getFunding(work); values != nil && len(values) > 0 {
		if err := SetFunding(rec, values); err != nil {
			return nil, err
		}
	}
	if value := getDOI(work); value != "" {
		if err := SetDOI(rec, value); err != nil {
			return nil, err
		}
	}
	if values := getLinks(work); values != nil && len(values) > 0 {
		if err := SetRelatedIdentifiers(rec, values); err != nil {
			return nil, err
		}
	}
	// FIXME: Need to map related titles, e.g. when a section/article is part of a book (anthology) or proceedings
	// FIXME: Need to crosswalk any related identifiers into the simple model
	return rec, fmt.Errorf("CrosswalkCrossRefWorks() not implemented")
	/*
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

		// NOTE: Assuming IsPublished is true given that we're talking to
		// CrossRef API which holds published content.
		// IsPublished
		eprint.IsPublished = "pub"
	*/
}
