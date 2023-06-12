package irdmtools

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

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
	//FIXME: Ideally this should take a resource type map.
	switch strings.ToLower(s) {
	//	case "proceedings-article":
	//		//FIXME: this mapping may not be correct, was book_section in EPrints CaltechAUTHORS
	//		return "publication-section"
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
	if work.Message != nil && work.Message.Type != "" {
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

// getAbstract retrieves the abstract from the CrossRef Works
func getAbstract(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.Abstract != "" {
		return work.Message.Abstract
	}
	return ""
}

// getPublisher
func getPublisher(work *crossrefapi.Works) string {
	// FIXME: Need to know if publisher holds the publisher and container type holds publication based on work.Message.Type
	if work.Message != nil && work.Message.Publisher != "" {
		return work.Message.Publisher
	}
	return ""
}

// getPublication
func getPublication(work *crossrefapi.Works) string {
	if work.Message != nil && len(work.Message.ContainerTitle) > 0 {
		return work.Message.ContainerTitle[0]
	}
	return ""
}

// getSeries
func getSeries(work *crossrefapi.Works) string {
	// FIXME: Need to know if publisher holds the publisher and container type holds publication based on work.Message.Type
	if work.Message != nil && work.Message.ShortContainerTitle != nil && len(work.Message.ShortContainerTitle) > 0 {
		return work.Message.ShortContainerTitle[0]
	}
	return ""
}

// getVolume
func getVolume(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.Volume != "" {
		return work.Message.Volume
	}
	return ""
}

// getIssue
func getIssue(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.Issue != "" {
		return work.Message.Issue
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
			isbns = append(isbns, mkSimpleIdentifier("isbn", value))
		}
	}
	return isbns
}

// getISSNs
func getISSNs(work *crossrefapi.Works) []*simplified.Identifier {
	issns := []*simplified.Identifier{}
	if work.Message != nil && work.Message.ISSN != nil {
		for _, value := range work.Message.ISSN {
			issns = append(issns, &simplified.Identifier{Scheme: "issn", Identifier: value})
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
					Award: &simplified.AwardIdentifier{
						Number: award,
						Title: &simplified.TitleDetail{
							Encoding: " ",
						},
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
				Scheme:     "url",
				Identifier: link.URL,
				Name:       link.ContentType,
			})
		}
	}
	return identifiers
}

func makeIdentifiers(scheme string, identifierList []string) []*simplified.Identifier {
	identifiers := []*simplified.Identifier{}
	for _, val := range identifierList {
		identifiers = append(identifiers, mkSimpleIdentifier(scheme, val))
	}
	return identifiers
}

func mkSimpleRole(role string) *simplified.Role {
	return &simplified.Role{
		ID: role,
	}
}

func mkSimpleTitleDetail(title string) *simplified.TitleDetail {
	return &simplified.TitleDetail{
		Title: title,
	}
}

func crosswalkAuthorAffiliationToCreatorAffiliation(crAffiliation *crossrefapi.Organization) *simplified.Affiliation {
	if crAffiliation.IDs != nil {
		for _, id := range crAffiliation.IDs {
			if id.IdType == "ROR" && id.AssertedBy == "publisher" {
				affiliation := new(simplified.Affiliation)
				ror := strings.TrimPrefix(id.Id, "https://ror.org/")
				affiliation.ID = ror
				return affiliation
			}
		}
	}
	return nil
}

func crossrefPersonToCreator(author *crossrefapi.Person, role string) *simplified.Creator {
	po := new(simplified.PersonOrOrg)
	po.FamilyName = author.Family
	po.GivenName = author.Given
	if author.Family != "" && author.Given != "" {
		po.Type = "personal"
		po.Name = fmt.Sprintf("%s, %s", po.FamilyName, po.GivenName)
	} else {
		po.Type = "organization"
	}
	if author.ORCID != "" {
		po.Identifiers = append(po.Identifiers, &simplified.Identifier{
			Scheme:     "orcid",
			Identifier: strings.TrimPrefix(author.ORCID, "http://orcid.org/"),
		})
	}
	creator := new(simplified.Creator)
	creator.PersonOrOrg = po
	if role != "" {
		creator.Role = mkSimpleRole(role)
	}

	if author.Affiliation != nil && len(author.Affiliation) > 0 {
		for _, crAffiliation := range author.Affiliation {

			affiliation := crosswalkAuthorAffiliationToCreatorAffiliation(crAffiliation)
			if affiliation != nil && creator.HasAffiliation(affiliation) == false {
				creator.Affiliations = append(creator.Affiliations, affiliation)
			}
		}
	}
	return creator
}

func crossrefLicenseToRight(license *crossrefapi.License) *simplified.Right {
	if license.URL == "" {
		return nil
	}
	right := new(simplified.Right)
	right.Link = license.URL
	if license.ContentVersion != "" {
		cv := map[string]string {
			"en": license.ContentVersion,
		}
		right.Description = cv
	} else {
		d := map[string]string{
			"en": "url to license",
		}
		right.Description = d
	}
	t := map[string]string{
		"en": "url",
	}
	right.Title = t
	return right
}

func crosswalkDateObjectToDateType(do *crossrefapi.DateObject, description string) *simplified.DateType {
	dt := new(simplified.DateType)
	ymd := []string{}
	for _, aVal := range do.DateParts {
		for _, val := range aVal {
			ymd = append(ymd, fmt.Sprintf("%02d", val))
		}
	}
	dt.Date = strings.Join(ymd, "-")
	dt.Description = description
	return dt
}

func getCreators(work *crossrefapi.Works) []*simplified.Creator {
	creators := []*simplified.Creator{}
	if work.Message != nil && work.Message.Author != nil {
		for _, person := range work.Message.Author {
			creators = append(creators, crossrefPersonToCreator(person, ""))
		}
	}
	return creators
}

func getContributors(work *crossrefapi.Works) []*simplified.Creator {
	creators := []*simplified.Creator{}
	// NOTE: The works message object containers the related contributors as
	// separate entries.
	// .message.translator
	// .message.editor
	// .message.chair
	// There is a reference to .contributor and .reviewer but not sure if they really exists in the scheme.
	if work.Message != nil && work.Message.Translator != nil {
		for _, person := range work.Message.Translator {
			creators = append(creators, crossrefPersonToCreator(person, "translator"))
		}
	}
	if work.Message != nil && work.Message.Editor != nil {
		for _, person := range work.Message.Editor {
			creators = append(creators, crossrefPersonToCreator(person, "editor"))
		}
	}
	if work.Message != nil && work.Message.Chair != nil {
		for _, person := range work.Message.Chair {
			creators = append(creators, crossrefPersonToCreator(person, "chair"))
		}
	}
	return creators
}

func getLicenses(work *crossrefapi.Works) []*simplified.Right {
	if work.Message != nil && work.Message.License != nil {
		rights := []*simplified.Right{}
		for _, license := range work.Message.License {
			right := crossrefLicenseToRight(license)
			if right != nil {
				rights = append(rights, right)
			}
		}
		return rights
	}
	return nil
}

func getSubjects(work *crossrefapi.Works) []*simplified.Subject {
	if work.Message != nil && work.Message.Subject != nil {
		subjects := []*simplified.Subject{}
		for _, s := range work.Message.Subject {
			if s != "" {
				subjects = append(subjects, &simplified.Subject{
					Subject: s,
				})
			}
		}
		return subjects
	}
	return nil
}

func getPublishedPrint(work *crossrefapi.Works) *simplified.DateType {
	if work.Message != nil && work.Message.PublishedPrint != nil {
		return crosswalkDateObjectToDateType(work.Message.PublishedPrint, "published print")
	}
	return nil
}

func getPublishedOnline(work *crossrefapi.Works) *simplified.DateType {
	if work.Message != nil && work.Message.PublishedOnline != nil {
		return crosswalkDateObjectToDateType(work.Message.PublishedOnline, "published online")
	}
	return nil
}

func getPublicationDate(work *crossrefapi.Works) string {
	printDate := getPublishedPrint(work)
	onlineDate := getPublishedOnline(work)
	if (printDate == nil || printDate.Date == "") && (onlineDate == nil || onlineDate.Date == "") {
		return ""
	}
	if printDate == nil || printDate.Date == "" {
		return onlineDate.Date
	}
	if onlineDate == nil || onlineDate.Date == "" {
		return printDate.Date
	}
	// NOTE: If we get this far we need to compare dates' date strings.
	// This is a naive compare it assumes the date string formats are
	// alphabetical.
	i := strings.Compare(printDate.Date, onlineDate.Date)
	if i < 0 || i == 0 {
		return printDate.Date
	}
	return onlineDate.Date
}

func getAccepted(work *crossrefapi.Works) *simplified.DateType {
	if work.Message != nil && work.Message.Accepted != nil {
		return crosswalkDateObjectToDateType(work.Message.Accepted, "accepted")
	}
	return nil
}

func getApproved(work *crossrefapi.Works) *simplified.DateType {
	if work.Message != nil && work.Message.Approved != nil {
		return crosswalkDateObjectToDateType(work.Message.Approved, "approved")
	}
	return nil
}

// CrosswalkCrossRefWork takes a Works object from the CrossRef API
// and maps the fields into an simplified Record struct return a
// new struct or error.
func CrosswalkCrossRefWork(cfg *Config, work *crossrefapi.Works, resourceTypeMap map[string]string, contributorTypeMap map[string]string) (*simplified.Record, error) {
	rec := new(simplified.Record)
	// .message.type -> .record.metadata.resource_type (via controlled vocabulary)
	if value := getResourceType(work); value != "" {
		if err := SetResourceType(rec, value, resourceTypeMap); err != nil {
			return nil, err
		}
	}
	if value := getDOI(work); value != "" {
		if err := SetDOI(rec, value); err != nil {
			return nil, err
		}
	}
	if values := getTitles(work); len(values) > 0 {
		for i, val := range values {
			if i == 0 {
				if err := SetTitle(rec, val); err != nil {
					return nil, err
				}
			} else {
				if err := AddAdditionalTitles(rec, mkSimpleTitleDetail(val)); err != nil {
					return nil, err
				}
			}
		}
	}
	// NOTE: Abstract becomes Description in simplified records
	if value := getAbstract(work); value != "" {
		if err := SetDescription(rec, value); err != nil {
			return nil, err
		}
	}
	if values := getCreators(work); values != nil && len(values) > 0 {
		if err := SetCreators(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getContributors(work); values != nil && len(values) > 0 {
		if err := SetContributors(rec, values); err != nil {
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
	/* FIXME: Need to know where this it's assignted in simplified model.
	Also the data I fetch from CrossRef now looks like an alternate short
	title so works.message["short-container-title"] may not be the right
	place to fetch this data.
	if value := getSeries(work); value != "" {
		if err := SetSeries(rec, value); err != nil {
			return nil, err
		}
	}
	*/
	if value := getVolume(work); value != "" {
		if err := SetVolume(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getIssue(work); value != "" {
		if err := SetIssue(rec, value); err != nil {
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
		if err := AddRelatedIdentifiers(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getISSNs(work); values != nil && len(values) > 0 {
		if err := AddRelatedIdentifiers(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getFunding(work); values != nil && len(values) > 0 {
		if err := SetFunding(rec, values); err != nil {
			return nil, err
		}
	}
	/* Removed per issue #10
	if values := getLinks(work); values != nil && len(values) > 0 {
		if err := AddRelatedIdentifiers(rec, values); err != nil {
			return nil, err
		}
	}
	*/
	if values := getLicenses(work); values != nil {
		if err := AddRights(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getSubjects(work); values != nil {
		if err := AddSubjects(rec, values); err != nil {
			return nil, err
		}
	}
	// NOTE: Crossref has many dates, e.g. publised print, published online
	if value := getPublishedPrint(work); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getPublishedOnline(work); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getAccepted(work); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getApproved(work); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, err
		}
	}
	// NOTE: Publication Date should be the earlier of print or online
	if value := getPublicationDate(work); value != "" {
		if err := SetPublicationDate(rec, value); err != nil {
			return nil, err
		}
	}

	// NOTE: We need to set the creation and updated time.
	now := time.Now()
	rec.Created = now
	rec.Updated = now
	return rec, nil
}
