package irdmtools

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	// Caltech Library Packages
	"github.com/caltechlibrary/dataciteapi"
	"github.com/caltechlibrary/simplified"
)

func QueryDataCiteObject(cfg *Config, doi string, options *Doi2RdmOptions) (map[string]interface{}, error) {
	appName := path.Base(os.Args[0])
	client, err := dataciteapi.NewDataCiteClient(appName, options.MailTo)
	if err != nil {
		return nil, err
	}
	objects, err := client.Dois(doi)
	if err != nil {
		return nil, err
	}
	if len(objects) == 0 {
		return nil, fmt.Errorf("no data returned for %q", doi)
	}
	src, _ := JSONMarshalIndent(objects, "", "    ")
	if cfg.Debug {
		fmt.Fprintf(os.Stderr, "objects JSON:\n\n%s\n\n", src)
	}
	m := map[string]interface{}{}
	if err := JSONUnmarshal(src, &m); err != nil {
		return nil, fmt.Errorf("problem encoding/decoding DataCite object, %s", err)
	}
	return m, nil
}

// getObjectData retrieves the `.data` from the DateCite `.object`
func getObjectData(object map[string]interface{}) (map[string]interface{}, bool) {
	if data, ok := object["data"].(map[string]interface{}); ok {
		return data, ok
	}
	return nil, false
}

func getObjectDataAttributes(object map[string]interface{}) (map[string]interface{}, bool) {
	if data, ok := getObjectData(object); ok {
		attr, ok := data["attributes"].(map[string]interface{})
		return attr, ok
	}
	return nil, false
}

// getObjectTitle retrieves `.data.attributes["titles"]`
func getObjectTitle(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if title, ok := attrs["title"].(string); ok && (title != "") {
			return title
		}
		if values, ok := attrs["titles"].([]interface{}); ok {
			for _, val := range values {
				m := val.(map[string]interface{})
				if title, ok := m["title"].(string); ok {
					return title
				}
			}
		}
	}
	return ""
}

// getObjectResourceType retrieves the `.data.types.resourceTypeGeneral` value if exists.
//
// ```
// resourceType := getObjectResourceType(object)
// ```
func getObjectResourceType(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if types, ok := attrs["types"].(map[string]interface{}); ok {
			// The path to type informaiton is .access.types, this has a map of types
			// for various formats (e.g. .bibtex, .citeproc, .schemeOrg, .resourceType, .resourceTypeGeneral.
			// Tom says `.resourceTypeGeneral` makes the most sense here.
			if resourceType, ok := types["resourceTypeGeneral"].(string); ok {
				return resourceType
			}
		}
	}
	return ""
}

// getObjectDescription retrieves the description (a.k.a. abstract) from the DataCite Object
// See example JSON <https://api.test.datacite.org/dois/10.82433/q54d-pf76?publisher=true&affiliation=true>
func getObjectDescription(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if descriptions, ok := attrs["descriptions"]; ok {
			for _, item := range descriptions.([]interface{}) {
				m := item.(map[string]interface{})
				if values, ok := m["description"]; ok {
					return fmt.Sprintf("%s", values)
				}
			}
		}
	}
	return ""
}

// getObjectPublisher
// See example JSON <https://api.test.datacite.org/dois/10.82433/q54d-pf76?publisher=true&affiliation=true>
func getObjectPublisher(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if publisher, ok := attrs["publisher"].(string); ok {
			return publisher
		}
		if publisher, ok := attrs["publisher"].(map[string]string); ok {
			if name, ok := publisher["name"]; ok {
				return name
			}
		}
	}
	return ""
}

// getObjectPublication
// See example JSON <https://api.test.datacite.org/dois/10.82433/q54d-pf76?publisher=true&affiliation=true>
func getObjectPublication(object map[string]interface{}) string {
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if items, ok := attributes["relatedItems"].([]interface{}); ok {
				for _, item := range items {
					m := item.(map[string]interface{})
					if relationType, ok := m["relationType"].(string); ok && relationType == "IsPublishedIn" {
						if titles, ok := m["titles"].([]interface{}); ok {
							for _, title := range titles {
								m := title.(map[string]interface{})
								if val, ok := m["title"].(string); ok {
									return val
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}

// getObjectSeries
func getObjectSeries(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if items, ok := attrs["relatedItems"].([]interface{}); ok {
			for _, item := range items {
				m := item.(map[string]interface{})
				if issue, ok := m["issue"].(string); ok {
					return issue
				}
			}
		}
	}
	return ""
}

// getObjectVolume
func getObjectVolume(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if items, ok := attrs["relatedItems"].([]interface{}); ok {
			for _, item := range items {
				m := item.(map[string]interface{})
				if issue, ok := m["volume"].(string); ok {
					return issue
				}
			}
		}
	}
	return ""
}

// getObjectIssue
func getObjectIssue(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if items, ok := attrs["relatedItems"].([]interface{}); ok {
			for _, item := range items {
				m := item.(map[string]interface{})
				if issue, ok := m["issue"].(string); ok {
					return issue
				}
			}
		}
	}
	return ""
}

// getObjectPublisherLocation
func getObjectPublisherLocation(object map[string]interface{}) string {
	/* FIXME: Not sure where to find this.  */
	return ""
}

// getObjectPageRange
func getObjectPageRange(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if items, ok := attrs["relatedItems"].([]interface{}); ok {
			for _, item := range items {
				m := item.(map[string]interface{})
				if firstPage, ok := m["firstPage"].(string); ok {
					if lastPage, ok := m["lastPage"].(string); ok {
						return fmt.Sprintf("%s - %s", firstPage, lastPage)
					}
					return fmt.Sprintf("%s - %s", firstPage, firstPage)
				}
			}
		}
	}
	return ""
}

// getObjectArticleNumber
func getObjectArticleNumber(object map[string]interface{}) string {
	/* FIXME: Not sure where article numbers map from in the DataCite API */
	return ""
}

// getObjectISBNs
func getObjectISBNs(object map[string]interface{}) []string {
	isbns := []string{}
	if attrs, ok := getObjectDataAttributes(object); ok {
		if identifiers, ok := attrs["relatedIdentifiers"]; ok {
			for _, item := range identifiers.([]interface{}) {
				m := item.(map[string]interface{})
				if identifierType, ok := m["relatedIdentifierType"].(string); ok && identifierType == "ISBN" {
					if val, ok := m["relatedIdentifier"].(string); ok {
						isbns = append(isbns, val)
					}
				}
			}
		}
	}
	return isbns
}

// getObjectISSNs
func getObjectISSNs(object map[string]interface{}) []string {
	issns := []string{}
	if attrs, ok := getObjectDataAttributes(object); ok {
		if identifiers, ok := attrs["relatedIdentifiers"].([]interface{}); ok {
			for _, item := range identifiers {
				m := item.(map[string]interface{})
				if identifierType, ok := m["relatedIdentifierType"].(string); ok && identifierType == "ISSN" {
					if val, ok := m["relatedIdentifier"].(string); ok {
						issns = append(issns, val)
					}
				}
			}
		}
	}
	return issns
}

// getObjectFunding
func getObjectFunding(object map[string]interface{}) []*simplified.Funder {
	if attrs, ok := getObjectDataAttributes(object); ok {
		funders := []*simplified.Funder{}
		if fundingReferences, ok := attrs["fundingReferences"].([]interface{}); ok {
			for _, item := range fundingReferences {
				m := item.(map[string]interface{})
				funder := new(simplified.Funder)
				if funderName, ok := m["funderName"].(string); ok {
					funder.Funder = new(simplified.FunderIdentifier)
					funder.Funder.Name = funderName
				}
				if awardNumber, ok := m["awardNumber"].(string); ok {
					funder.Award = new(simplified.AwardIdentifier)
					funder.Award.Number = awardNumber
				}
				if funder.Funder != nil || funder.Award != nil {
					funders = append(funders, funder)
				}
			}
		}
		if len(funders) > 0 {
			return funders
		}
	}
	return nil
}

// getObjectDOI
func getObjectDOI(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if doi, ok := attrs["doi"].(string); ok {
			return doi
		}
	}
	return ""
}

// getObjectLinks
func getObjectLinks(object map[string]interface{}) []*simplified.Identifier {
	/* FIXME: Need to find an example of where this is in DataCite JSON */
	if attrs, ok := getObjectDataAttributes(object); ok {
		links := []*simplified.Identifier{}
		if url, ok := attrs["contentUrl"].(string); ok {
			identifier := new(simplified.Identifier)
			identifier.Scheme = "url"
			identifier.Identifier = url
		}
		if len(links) > 0 {
			return links
		}
	}
	return nil
}

func getObjectAgents(object map[string]interface{}, agentType string) []*simplified.Creator {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if peopleOrGroups, ok := attrs[agentType].([]interface{}); ok {
			agents := []*simplified.Creator{}
			for _, item := range peopleOrGroups {
				entity := item.(map[string]interface{})
				agent := new(simplified.Creator)
				agent.PersonOrOrg = new(simplified.PersonOrOrg)
				if nameType, ok := entity["nameType"].(string); ok {
					agent.PersonOrOrg.Type = strings.ToLower(nameType)
				}
				if name, ok := entity["name"].(string); ok {
					agent.PersonOrOrg.Name = name
				}
				if family, ok := entity["familyName"].(string); ok {
					agent.PersonOrOrg.FamilyName = family
				}
				if given, ok := entity["givenName"].(string); ok {
					agent.PersonOrOrg.GivenName = given
				}
				// This is a fallback to make sure the Person.OrOrg.Type gets set.
				if agent.PersonOrOrg.Type == "" && agent.PersonOrOrg.GivenName != "" && agent.PersonOrOrg.FamilyName != "" {
					agent.PersonOrOrg.Type = "personal"
				}
				if nameIdentifiers, ok := entity["nameIdentifiers"].([]interface{}); ok {
					agent.PersonOrOrg.Identifiers = []*simplified.Identifier{}
					for _, value := range nameIdentifiers {
						if m, ok := value.(map[string]interface{}); ok {
							id := &simplified.Identifier{}
							if val, ok := m["nameIdentifier"].(string); ok {
								if scheme, ok := m["nameIdentifierScheme"].(string); ok {
									id.Scheme = scheme
									if scheme == "ROR" {
										id.ID = val
									} else {
										id.Identifier = val
									}
									agent.PersonOrOrg.Identifiers = append(agent.PersonOrOrg.Identifiers, id)
								}
							}
						}
					}

				}
				if agent.PersonOrOrg.Name != "" || agent.PersonOrOrg.FamilyName != "" {
					agents = append(agents, agent)
				}
			}
			return agents			
		}		
	}
	return nil
}

func getObjectCreators(object map[string]interface{}) []*simplified.Creator {
	return getObjectAgents(object, "creators")
}

func getObjectContributors(object map[string]interface{}) []*simplified.Creator {
	return getObjectAgents(object, "contributors")
}

func getObjectLicenses(object map[string]interface{}) []*simplified.Right {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if rightsList, ok := attrs["rightsList"].([]interface{}); ok {
			licenses := []*simplified.Right{}
			for _, item := range rightsList {
				license := new(simplified.Right)
				license.Title = map[string]string{}
				m := item.(map[string]interface{})
				if title, ok := m["rights"].(string); ok {
					license.Title["en"] = title
				}
				if identifier, ok := m["rightsIdentifier"].(string); ok {
					license.ID = identifier
				}
				if license.ID != "" || len(license.Title) > 0 {
					licenses = append(licenses, license)
				}
			}
			if len(licenses) > 0 {
				return licenses
			}
		}
	}
	return nil
}

func isDuplicateSubject(subject *simplified.Subject, subjectList []*simplified.Subject) bool {
	for _, item := range subjectList {
		if subject.Subject == item.Subject {
			return true
		}
	}
	return false
}

func getObjectSubjects(object map[string]interface{}) []*simplified.Subject {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if items, ok := attrs["subjects"].([]interface{}); ok {
			subjects := []*simplified.Subject{}
			for _, item := range items {
				m := item.(map[string]interface{})
				if s, ok := m["subject"]; ok {
					subject := new(simplified.Subject)
					subject.Subject = s.(string)
					if ! isDuplicateSubject(subject, subjects) {
						subjects = append(subjects, subject)		
					}
				}
			}
			return subjects
		}
	}
	return nil
}

func getObjectPublishedPrint(object map[string]interface{}) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

func getObjectPublishedOnline(object map[string]interface{}) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

// take a list of dates and map by their types.
func mapDatesToType(items []interface{}) map[string]string {
	dtMap := map[string]string{}
	for _, item := range items {
		m := item.(map[string]interface{})
		if dateType, ok := m["dateType"].(string); ok {
			dateType = strings.ToLower(dateType)
			if val, ok := m["date"].(string); ok {
				// Always take the first version of the provides types (no overwriting)
				if _, conflict := m[dateType]; ! conflict {
					dtMap[dateType] = val
				}
			}
		}
	}
	return dtMap
}
func getObjectPublicationDate(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if published, ok := attrs["published"].(string); ok {
			return published
		}
		if items, ok := attrs["dates"].([]interface{}); ok {
			// First look for "published" date
			dates := mapDatesToType(items)
			for _, dateType := range []string{"published", "issued", "accepted"} {
				if val, ok := dates[dateType]; ok {
					return val
				}
			}
		}
		if publicationYear, ok := attrs["publicationYear"].(int); ok {
			return fmt.Sprintf("%d", publicationYear)
		}
	}
	return ""
}

func getObjectAccepted(object map[string]interface{}) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

func getObjectApproved(object map[string]interface{}) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

// normalizeObjectPublisherName will check the publisher DOI and ISSN to see if we have
// a preferred name in our options. If so it will return that.
func normalizeObjectPublisherName(val string, object map[string]interface{}, options *Doi2RdmOptions) string {
	for _, issn := range getObjectISSNs(object) {
		if issn != "" {
			if value, ok := options.ISSNPublishers[issn]; ok {
				return value
			}
		}
	}
	doi := getObjectDOI(object)
	if doi != "" {
		doiPrefix, _ := DoiPrefix(doi)
		if value, ok := options.DoiPrefixPublishers[doiPrefix]; ok {
			return value
		}
	}
	return val
}

// normalizeObjectJournalName will check the ISSN to see if we have
// a preferred name in our options. If so it will return that.
func normalizeObjectJournalName(val string, object map[string]interface{}, options *Doi2RdmOptions) string {
	for _, issn := range getObjectISSNs(object) {
		if issn != "" {
			if value, ok := options.ISSNJournals[issn]; ok {
				return value
			}
		}
	}
	return val
}

func getObjectIdentifier(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if identifier, ok := attrs["identifier"].(string); ok {
			return identifier
		}
	}
	return ""
}

// CrosswalkDataCiteObject takes a Object object from the DataCite API
// and maps the fields into an simplified Record struct return a
// new struct or error.
func CrosswalkDataCiteObject(cfg *Config, object map[string]interface{}, options *Doi2RdmOptions) (*simplified.Record, error) {
	if object == nil {
		return nil, fmt.Errorf("crossref api objects not populated")
	}
	rec := new(simplified.Record)
	rec.Metadata = new(simplified.Metadata)

	// .message.type -> .record.metadata.resource_type (via controlled vocabulary)
	if value := getObjectResourceType(object); value != "" {
		if err := SetResourceType(rec, value, options.ResourceTypes); err != nil {
			return nil, err
		}
	}
	if value := getObjectDOI(object); value != "" {
		if err := SetDOI(rec, value); err != nil {
			return nil, err
		}
	}
	if val := getObjectTitle(object); val != "" {
		if err := SetTitle(rec, val); err != nil {
			return nil, err
		}
	}
	if val := getObjectDescription(object); val != "" {
		if err := SetDescription(rec, val); err != nil {
			return nil, err
		}
	}
	if values := getObjectCreators(object); values != nil && len(values) > 0 {
		if err := SetCreators(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getObjectContributors(object); values != nil && len(values) > 0 {
		if err := SetContributors(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getObjectLicenses(object); values != nil {
		if err := AddRights(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getObjectSubjects(object); values != nil {
		if err := AddSubjects(rec, values); err != nil {
			return nil, err
		}
	}
	if values := getObjectFunding(object); values != nil && len(values) > 0 {
		if err := SetFunding(rec, values); err != nil {
			return nil, err
		}
	}
	if val := getObjectPublisher(object); val != "" {
		// NOTE: Setting the publisher name is going to be normalized via DOI prefix for records with ISSN.
		val = normalizeObjectPublisherName(val, object, options)
		if err := SetPublisher(rec, val); err != nil {
			return nil, err
		}
	}
	if val := getObjectPublication(object); val != "" {
		// NOTE: Setting the publisher name is going to be normalized via DOI prefix for records with ISSN.
		val = normalizeObjectJournalName(val, object, options)
		if err := SetPublisher(rec, val); err != nil {
			return nil, err
		}
	}
	if values := getObjectISBNs(object); values != nil && len(values) > 0 {
		if err := SetImprintField(rec, "isbn", values); err != nil {
			return nil, err
		}
	}
	if values := getObjectISSNs(object); len(values) > 0 {
		if err := SetJournalField(rec, "issn", values[0]); err != nil {
			return nil, err
		}
		if len(values) > 1 {
			for i := 1; i < len(values); i++ {
				AddIdentifier(rec, "issn", values[i])
			}
		}
	}
	if value := getObjectPublication(object); value != "" {
		if err := SetPublication(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getObjectVolume(object); value != "" {
		if err := SetVolume(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getObjectIssue(object); value != "" {
		if err := SetIssue(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getObjectPublisherLocation(object); value != "" {
		if err := SetPublisherLocation(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getObjectPageRange(object); value != "" {
		if err := SetPageRange(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getObjectArticleNumber(object); value != "" {
		if err := SetArticleNumber(rec, value); err != nil {
			return nil, err
		}
	}
	// NOTE: Crossref has many dates, e.g. publised print, published online
	if value := getObjectPublishedPrint(object); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getObjectPublishedOnline(object); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getObjectAccepted(object); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, err
		}
	}
	if value := getObjectApproved(object); value != nil {
		if err := AddDate(rec, value); err != nil {
			return nil, err
		}
	}
	// NOTE: Publication Date should be the earlier of print or online
	if value := getObjectPublicationDate(object); value != "" {
		if err := SetPublicationDate(rec, value); err != nil {
			return nil, err
		}
	}

	// Default language to US English
	if err := SetLanguages(rec, "id", "eng"); err != nil {
		return nil, err
	}

	// NOTE: We need to set the creation and updated time.
	now := time.Now()
	rec.Created = now
	rec.Updated = now
	return rec, nil
}
