package irdmtools

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"


	// Caltech Library Packages
	"github.com/caltechlibrary/crossrefapi"
	"github.com/caltechlibrary/simplified"
)

// crossRefAPIURL is the base URL of the CrossRef API. It is empty outside
// tests, which leaves the client's own address, https://api.crossref.org.
var crossRefAPIURL string

func QueryCrossRefWork(cfg *Config, doi string, options *Doi2RdmOptions) (*crossrefapi.Works, error) {
	appName := path.Base(os.Args[0])
	client, err := crossrefapi.NewCrossRefClient(appName, options.MailTo)
	if err != nil {
		return nil, ConfigError(err)
	}
	if crossRefAPIURL != "" {
		client.API = crossRefAPIURL
	}
	works, err := client.Works(doi)
	if err != nil {
		return nil, queryError("CrossRef", doi, err)
	}
	// The client returns no error and no record for a failure status.
	if err := classifyHTTPStatus("CrossRef", doi, client.StatusCode); err != nil {
		return nil, err
	}
	if works == nil {
		return nil, DataError(fmt.Errorf("CrossRef answered %q with an empty response", doi))
	}
	if cfg.Debug {
		src, _ := JSONMarshalIndent(works, "", "    ")
		fmt.Fprintf(os.Stderr, "works JSON:\n\n%s\n\n", src)
	}
	return works, nil
}

// getWorksResourceType retrives the resource type from works.message.type
// runs normalize
func getWorksResourceType(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.Type != "" {
		return work.Message.Type
	}
	return ""
}

// CrossRefWarning is something the text normalizer could not convert cleanly
// in one field of a CrossRef work. The record is still made, with the best
// text available; the warning says where a person may want to look.
type CrossRefWarning struct {
	// Field is the part of the record the warning is about: "title",
	// "additional title" or "description".
	Field string
	// Warning is the normalizer's warning, with its code and the fragment of
	// the source it is about.
	Warning crossrefapi.Warning
}

// Line returns the warning as one line of text for standard error:
// "doi2rdm: DOI: field: code: fragment". Line breaks in the fragment are
// replaced by spaces so a warning is always one line.
//
// @param doi {string} the DOI of the work the warning is about
// @returns {string} the warning as one line without a trailing newline
// @example
//
//	w := CrossRefWarning{Field: "title", Warning: crossrefapi.Warning{Code: "math-linear", Fragment: "Z_N"}}
//	fmt.Println(w.Line("10.1/x")) // doi2rdm: 10.1/x: title: math-linear: Z_N
func (w CrossRefWarning) Line(doi string) string {
	fragment := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(w.Warning.Fragment)
	return fmt.Sprintf("doi2rdm: %s: %s: %s: %s", doi, w.Field, w.Warning.Code, fragment)
}

// fieldWarnings labels normalizer warnings with the field they came from.
func fieldWarnings(field string, warnings []crossrefapi.Warning) []CrossRefWarning {
	var result []CrossRefWarning
	for _, w := range warnings {
		result = append(result, CrossRefWarning{Field: field, Warning: w})
	}
	return result
}

// getWorksTitles returns the title of a CrossRef Works object as plain text
// with its subtitle appended, the other titles that add something to it, and
// the warnings from converting them. With no title the title is "" and there
// are no other titles.
//
// @param work {*crossrefapi.Works} a CrossRef work, which may have no message
// @param options {*Doi2RdmOptions} the options, which name the symbol table to convert with
// @returns {string} the full title as plain text
// @returns {[]string} the alternate titles as plain text
// @returns {[]CrossRefWarning} what could not be converted cleanly
// @returns {error} an error if the normalizer cannot be built
// @example
//
//	title, others, warnings, err := getWorksTitles(work, options)
func getWorksTitles(work *crossrefapi.Works, options *Doi2RdmOptions) (string, []string, []CrossRefWarning, error) {
	if work == nil || work.Message == nil {
		return "", nil, nil, nil
	}
	n, err := options.crossRefNormalizer()
	if err != nil {
		return "", nil, nil, err
	}
	title, titleWarnings := n.FullTitle(work.Message)
	others, otherWarnings := n.AlternateTitles(work.Message)
	warnings := fieldWarnings("title", titleWarnings)
	warnings = append(warnings, fieldWarnings("additional title", otherWarnings)...)
	return title, others, warnings, nil
}

// getWorksDescription returns the abstract of a CrossRef Works object as HTML
// that Invenio RDM's sanitizer keeps: paragraphs, headings as bold paragraphs
// and inline markup, with MathML as TeX between \( and \), and the leading
// "Abstract" heading and the copyright text dropped, and the warnings from
// converting it. A work with no abstract gives "".
//
// @param work {*crossrefapi.Works} a CrossRef work, which may have no message
// @param options {*Doi2RdmOptions} the options, which name the symbol table to convert with
// @returns {string} the description as HTML, empty when there is no abstract
// @returns {[]CrossRefWarning} what could not be converted cleanly
// @returns {error} an error if the normalizer cannot be built
// @example
//
//	description, warnings, err := getWorksDescription(work, options)
func getWorksDescription(work *crossrefapi.Works, options *Doi2RdmOptions) (string, []CrossRefWarning, error) {
	if work == nil || work.Message == nil {
		return "", nil, nil
	}
	n, err := options.crossRefNormalizer()
	if err != nil {
		return "", nil, err
	}
	description, warnings := n.AbstractHTML(work.Message)
	return description, fieldWarnings("description", warnings), nil
}

// getWorksPublisher
func getWorksPublisher(work *crossrefapi.Works) string {
	// FIXME: Need to know if publisher holds the publisher and container type holds publication based on work.Message.Type
	if work.Message != nil && work.Message.Publisher != "" {
		return work.Message.Publisher
	}
	return ""
}

// getWorksPublication
func getWorksPublication(work *crossrefapi.Works) string {
	if work.Message != nil && len(work.Message.ContainerTitle) > 0 {
		return work.Message.ContainerTitle[0]
	}
	return ""
}

// getWorksWorksSeries
func getWorksSeries(work *crossrefapi.Works) string {
	// FIXME: Need to know if publisher holds the publisher and container type holds publication based on work.Message.Type
	if work.Message != nil && work.Message.ShortContainerTitle != nil && len(work.Message.ShortContainerTitle) > 0 {
		return work.Message.ShortContainerTitle[0]
	}
	return ""
}

// getWorksWorksVolume
func getWorksVolume(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.Volume != "" {
		return work.Message.Volume
	}
	return ""
}

// getWorksWorksIssue
func getWorksIssue(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.Issue != "" {
		return work.Message.Issue
	}
	return ""
}

// getWorksWorksPublisherLocation
func getWorksPublisherLocation(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.PublisherLocation != "" {
		return work.Message.PublisherLocation
	}
	return ""
}

// getWorksWorksPageRange
func getWorksPageRange(work *crossrefapi.Works) string {
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

// getWorksArticleNumber
func getWorksArticleNumber(work *crossrefapi.Works) string {
	/* FIXME: Not sure where article numbers map from in the CrossRef API
	- ComponentNumber
	- PartNumber
	*/
	if work.Message != nil && work.Message.ArticleNumber != "" {
		return work.Message.ArticleNumber
	}
	return ""
}

// getWorksISBNs
func getWorksISBNs(work *crossrefapi.Works) []string {
	isbns := []string{}
	if work.Message != nil && work.Message.ISBN != nil {
		for _, value := range work.Message.ISBN {
			isbns = append(isbns, value)
		}
	}
	return isbns
}

// getWorksISSNs
func getWorksISSNs(work *crossrefapi.Works) []string { 
	issns := []string{}
	if work.Message != nil && work.Message.ISSN != nil {
		for _, value := range work.Message.ISSN {
			issns = append(issns, value)
		}
	}
	return issns
}

// getWorksFunding
func getWorksFunding(work *crossrefapi.Works) []*simplified.Funder {
	funding := []*simplified.Funder{}
	suffixToROR := map[string]string{}
	if work.Message != nil && work.Message.Funder != nil && len(work.Message.Funder) > 0 {
		var (
			suffix string
			ror string
			ok bool
		)
		for _, funder := range work.Message.Funder {
			agency := &simplified.FunderIdentifier {}
			if funder.Name != "" {
				agency.Name = funder.Name
			}
			ror = ""
			if funder.Identifiers != nil && len(funder.Identifiers) > 0 {
				for _, identifier := range funder.Identifiers {
					if identifier.IdType == "ROR" && identifier.AssertedBy == "publisher" {
						ror = strings.TrimPrefix(identifier.Id, "https://ror.org/")
					}
				}
			}
			if ror == "" && funder.DOI != "" && funder.DoiAssertedBy == "publisher" {
				parts := strings.SplitN(funder.DOI, "/", 2)
				if len(parts) == 2 {
					suffix = strings.TrimSpace(parts[1])
					ror, ok = suffixToROR[suffix]
					if ! ok {
						ror, ok = lookupROR(suffix, true)
						if ok {
							suffixToROR[suffix] = ror
						}
					}
				}
			}
			if ror != "" {
				agency.Identifier = ror
			}
			if len(funder.Award) > 0 {
				for _, award := range funder.Award {
					grant := &simplified.Funder{
						Award: &simplified.AwardIdentifier{
							Number: award,
						},
					}
					if agency.Name != "" || agency.Identifier != "" {
						grant.Funder = agency
					} 
					funding = append(funding, grant)
				}
			} else {
				funding = append(funding, &simplified.Funder{
					Funder: agency,
				})
			}
		}
	}
	return funding
}

// getWorksDOI
func getWorksDOI(work *crossrefapi.Works) string {
	if work.Message != nil && work.Message.DOI != "" {
		return work.Message.DOI
	}
	return ""
}

// getWorksLinks
func getWorksLinks(work *crossrefapi.Works) []*simplified.Identifier {
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

func makeSimpleRole(role string) *simplified.Role {
	return &simplified.Role{
		ID: role,
	}
}

func makeSimpleTitleDetail(title string) *simplified.TitleDetail {
	// NOTE: Per issue 87, additional titles need a title type set
	// or RDM will reject the record. CrossRef doesn't distinguish
	// why a work has multiple titles (e.g. translated title vs.
	// alternative title) so "other" is used as a safe default.
	return &simplified.TitleDetail{
		Title: title,
		Type: &simplified.Type{
			ID: "other",
		},
	}
}

func crosswalkWorksAuthorAffiliationToCreatorAffiliation(crAffiliation *crossrefapi.Organization) *simplified.Affiliation {
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

func crosswalkWorksPersonToCreator(author *crossrefapi.Person, role string) *simplified.Creator {
	po := new(simplified.PersonOrOrg)
	po.FamilyName = author.Family
	po.GivenName = author.Given
	if author.Family != "" { //&& author.Given != "" { // removed per issue #80
		po.Type = "personal"
		po.Name = fmt.Sprintf("%s, %s", po.FamilyName, po.GivenName)
	} else {
		po.Type = "organizational"
		po.Name = author.Name
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
		creator.Role = makeSimpleRole(role)
	}

	if author.Affiliation != nil && len(author.Affiliation) > 0 {
		for _, crAffiliation := range author.Affiliation {

			affiliation := crosswalkWorksAuthorAffiliationToCreatorAffiliation(crAffiliation)
			if affiliation != nil && creator.HasAffiliation(affiliation) == false {
				creator.Affiliations = append(creator.Affiliations, affiliation)
			}
		}
	}
	return creator
}

func crosswalkWorksLicenseToRight(license *crossrefapi.License) *simplified.Right {
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

func crosswalkWorksDateObjectToDateType(do *crossrefapi.DateObject, description string) *simplified.DateType {
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

func getWorksCreators(work *crossrefapi.Works) []*simplified.Creator {
	creators := []*simplified.Creator{}
	if work.Message != nil && work.Message.Author != nil {
		for _, person := range work.Message.Author {
			creators = append(creators, crosswalkWorksPersonToCreator(person, ""))
		}
	}
	return creators
}

func getWorksContributors(work *crossrefapi.Works) []*simplified.Creator {
	creators := []*simplified.Creator{}
	// NOTE: The works message object containers the related contributors as
	// separate entries.
	// .message.translator
	// .message.editor
	// .message.chair
	// There is a reference to .contributor and .reviewer but not sure if they really exists in the scheme.
	if work.Message != nil && work.Message.Translator != nil {
		for _, person := range work.Message.Translator {
			creators = append(creators, crosswalkWorksPersonToCreator(person, "translator"))
		}
	}
	if work.Message != nil && work.Message.Editor != nil {
		for _, person := range work.Message.Editor {
			creators = append(creators, crosswalkWorksPersonToCreator(person, "editor"))
		}
	}
	if work.Message != nil && work.Message.Chair != nil {
		for _, person := range work.Message.Chair {
			creators = append(creators, crosswalkWorksPersonToCreator(person, "chair"))
		}
	}
	return creators
}

func getWorksLicenses(work *crossrefapi.Works) []*simplified.Right {
	if work.Message != nil && work.Message.License != nil {
		rights := []*simplified.Right{}
		for _, license := range work.Message.License {
			right := crosswalkWorksLicenseToRight(license)
			if right != nil {
				rights = append(rights, right)
			}
		}
		return rights
	}
	return nil
}

func getWorksSubjects(work *crossrefapi.Works) []*simplified.Subject {
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


func getWorksPublishedPrint(work *crossrefapi.Works) *simplified.DateType {
	if work.Message != nil && work.Message.PublishedPrint != nil {
		return crosswalkWorksDateObjectToDateType(work.Message.PublishedPrint, "published print")
	}
	return nil
}

func getWorksPublishedOnline(work *crossrefapi.Works) *simplified.DateType {
	if work.Message != nil && work.Message.PublishedOnline != nil {
		return crosswalkWorksDateObjectToDateType(work.Message.PublishedOnline, "published online")
	}
	return nil
}

func getWorksPublicationDate(work *crossrefapi.Works) string {
	var pubDate *simplified.DateType
	if work.Message != nil  && work.Message.Published != nil {
		pubDate = crosswalkWorksDateObjectToDateType(work.Message.Published, "published date")
	}
	if pubDate != nil && pubDate.Date != "" {
		return pubDate.Date
	}
	// If pubDate isn't available then guess what it should be.
	printDate := getWorksPublishedPrint(work)
	onlineDate := getWorksPublishedOnline(work)
	acceptedDate := getWorksAccepted(work)
	if (printDate == nil || printDate.Date == "") && (onlineDate == nil || onlineDate.Date == "") && (acceptedDate == nil || acceptedDate.Date == ""){
		return ""
	}
	if (printDate != nil && printDate.Date != "") && (onlineDate != nil && onlineDate.Date != "") {
		// NOTE: If we get this far we need to compare dates' date strings.
		// This is a naive compare it assumes the date string formats are
		// alphabetical.
		i := strings.Compare(printDate.Date, onlineDate.Date)
		if i < 0 || i == 0 {
			return printDate.Date
		}
	}
	if printDate != nil && printDate.Date != "" {
		return printDate.Date
	}
	if onlineDate != nil && onlineDate.Date != "" {
		return onlineDate.Date
	}
	if acceptedDate != nil && acceptedDate.Date != "" {
		return acceptedDate.Date
	}
	return ""
}

func getWorksAccepted(work *crossrefapi.Works) *simplified.DateType {
	if work.Message != nil && work.Message.Accepted != nil {
		return crosswalkWorksDateObjectToDateType(work.Message.Accepted, "accepted")
	}
	return nil
}

func getWorksApproved(work *crossrefapi.Works) *simplified.DateType {
	if work.Message != nil && work.Message.Approved != nil {
		return crosswalkWorksDateObjectToDateType(work.Message.Approved, "approved")
	}
	return nil
}

// normalizeWorksJournalName will check the ISSN to see if we have
// a preferred name in our options. If so it will return that.
func normalizeWorksJournalName(val string, work *crossrefapi.Works, options *Doi2RdmOptions) string {
	for _, issn := range getWorksISSNs(work) {
		if issn != "" {
			if value, ok := options.ISSNJournals[issn]; ok {
				return value
			}
		}
	}
	return val
}

// normalizeWorksPublisherName will check the publisher DOI and ISSN to see if we have
// a preferred name in our options. If so it will return that.
func normalizeWorksPublisherName(val string, work *crossrefapi.Works, options *Doi2RdmOptions) string {
	for _, issn := range getWorksISSNs(work) {
		if issn != "" {
			if value, ok := options.ISSNPublishers[issn]; ok {
				return value
			}
		}
	}
	doi := getWorksDOI(work)
	if doi != "" {
		doiPrefix, _ := DoiPrefix(doi)
		if value, ok := options.DoiPrefixPublishers[doiPrefix]; ok {
			return value
		}
	}
	return val
}

// CrosswalkCrossRefWork takes a Works object from the CrossRef API
// and maps the fields into an simplified Record struct return a
// new struct or error.
func CrosswalkCrossRefWork(cfg *Config, work *crossrefapi.Works, options *Doi2RdmOptions) (*simplified.Record, error) {
	rec, _, err := CrosswalkCrossRefWorkWarnings(cfg, work, options)
	return rec, err
}

// CrosswalkCrossRefWorkWarnings does what CrosswalkCrossRefWork does and also
// returns what the text normalizer could not convert cleanly in the title,
// the additional titles and the description. A warning never fails the
// crosswalk: the record holds the best text available.
//
// @param cfg {*Config} the configuration
// @param work {*crossrefapi.Works} the CrossRef work to map
// @param options {*Doi2RdmOptions} the crosswalk options
// @returns {*simplified.Record} the new record, nil on error
// @returns {[]CrossRefWarning} the conversion warnings, in field order
// @returns {error} an error if the work is nil or cannot be mapped
// @example
//
//	rec, warnings, err := CrosswalkCrossRefWorkWarnings(cfg, work, options)
//	for _, w := range warnings {
//		fmt.Fprintln(os.Stderr, w.Line(doi))
//	}
func CrosswalkCrossRefWorkWarnings(cfg *Config, work *crossrefapi.Works, options *Doi2RdmOptions) (*simplified.Record, []CrossRefWarning, error) {
	if work == nil {
		return nil, nil, fmt.Errorf("crossref api works not populated")
	}
	rec := new(simplified.Record)
	// .message.type -> .record.metadata.resource_type (via controlled vocabulary)
	if value := getWorksResourceType(work); value != "" {
		if err := SetResourceType(rec, value, options.ResourceTypes); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksDOI(work); value != "" {
		if err := SetDOI(rec, value); err != nil {
			return nil, nil, err
		}
	}
	title, otherTitles, titleWarnings, err := getWorksTitles(work, options)
	if err != nil {
		return nil, nil, err
	}
	if title != "" {
		if err := SetTitle(rec, title); err != nil {
			return nil, nil, err
		}
	}
	for _, val := range otherTitles {
		if err := AddAdditionalTitles(rec, makeSimpleTitleDetail(val)); err != nil {
			return nil, nil, err
		}
	}
	// NOTE: Abstract becomes Description in simplified records
	description, descriptionWarnings, err := getWorksDescription(work, options)
	if err != nil {
		return nil, nil, err
	}
	if description != "" {
		if err := SetDescription(rec, description); err != nil {
			return nil, nil, err
		}
	}
	if values := getWorksCreators(work); values != nil && len(values) > 0 {
		if err := SetCreators(rec, values); err != nil {
			return nil, nil, err
		}
	}
	if values := getWorksContributors(work); values != nil && len(values) > 0 {
		if err := SetContributors(rec, values); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksPublisher(work); value != "" {
		// FIXME: Setting the publisher name is going to be normalized via DOI prefix, maybe ISSN?
		value := normalizeWorksPublisherName(value, work, options)
		if err := SetPublisher(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksPublication(work); value != "" {
		value := normalizeWorksJournalName(value, work, options)
		if err := SetPublication(rec, value); err != nil {
			return nil, nil, err
		}
	}
	/* FIXME: Need to know where this it's assignted in simplified model.
	Also the data I fetch from CrossRef now looks like an alternate short
	title so works.message["short-container-title"] may not be the right
	place to fetch this data.
	if value := getWorksSeries(work); value != "" {
		if err := SetSeries(rec, value); err != nil {
			return nil, nil, err
		}
	}
	*/
	if value := getWorksVolume(work); value != "" {
		if err := SetVolume(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksIssue(work); value != "" {
		if err := SetIssue(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksPublisherLocation(work); value != "" {
		if err := SetPublisherLocation(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksPageRange(work); value != "" {
		if err := SetPageRange(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksArticleNumber(work); value != "" {
		if err := SetArticleNumber(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if values := getWorksISBNs(work); values != nil && len(values) > 0 {
		if err := SetImprintField(rec, "isbn", values); err != nil {
			return nil, nil, err
		}
	}
	if values := getWorksISSNs(work); len(values) > 0 {
		if err := SetJournalField(rec, "issn", values[0]); err != nil {
			return nil, nil, err
		}
		if len(values) > 1 {
			for i := 1; i < len(values); i++ {
				AddIdentifier(rec, "issn", values[i])
			}
		}
	}
	if values := getWorksFunding(work); values != nil && len(values) > 0 {
		if err := SetFunding(rec, values); err != nil {
			return nil, nil, err
		}
	}
	if values := getWorksLicenses(work); values != nil {
		if err := AddRights(rec, values); err != nil {
			return nil, nil, err
		}
	}
	if values := getWorksSubjects(work); values != nil {
		if err := AddSubjects(rec, values); err != nil {
			return nil, nil, err
		}
	}
	// NOTE: Crossref has many dates, e.g. publised print, published online
	if value := getWorksPublishedPrint(work); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksPublishedOnline(work); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksAccepted(work); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, nil, err
		}
	}
	if value := getWorksApproved(work); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, nil, err
		}
	}
	// NOTE: Publication Date should be the earlier of print or online
	if value := getWorksPublicationDate(work); value != "" {
		if err := SetPublicationDate(rec, value); err != nil {
			return nil, nil, err
		}
	}

	// Default language to US English
	if err := SetLanguages(rec, "id", "eng"); err != nil {
		return nil, nil, err
	}

	// NOTE: We need to set the creation and updated time.
	now := time.Now()
	rec.Created = now
	rec.Updated = now
	warnings := append(titleWarnings, descriptionWarnings...)
	return rec, warnings, nil
}
