package irdmtools

import (
	"fmt"
	"os"
	"path"
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
	objects, err := client.Works(doi)
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

// getObjectTitle retrieves `.data.attributes["title"]`
func getObjectTitle(object map[string]interface{}) string {
	attrs, ok := getObjectDataAttributes(object)
	if ok {
		if title, ok := attrs["title"]; ok {
			return fmt.Sprintf("%s", title)
		}
	}
	return ""
}

// getObjectCiteProcType retrieves the `.access.types.citeproc` value if exists.
func getObjectCiteProcType(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if types, ok := attrs["types"].(map[string]string); ok {
			if citeproc, ok := types["citeproc"]; ok {
				return citeproc
			}
		}
	}
	return ""
}

// getObjectResourceType retrives the resource type from objects.message.type
// runs normalize
func getObjectResourceType(object map[string]interface{}) string {
	// The path to type informaiton is .access.types, this has a map of types
	// for various formats (e.g. .bibtex, .citeproc, .schemeOrg, .resourceType, .resourceTypeGeneral.
	// I think using the .citeproc value makes the most sense here.
	if data, ok := getObjectData(object); ok {
		return getObjectCiteProcType(data)
	}
	return ""
}

// getObjectTitleList get a title list from `.data.attributes["titles"]`.
func getObjectTitleList(object map[string]interface{}) ([]map[string]string, bool) {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if titleList, ok := attrs["titles"].([]map[string]string); ok {
			return titleList, ok
		}
	}
	return nil, false
}

// getObjectTitles retrieves an ordered list of titles from a DataCite Object object.
// The zero index is the primary document title, the remaining are alternative titles.
// If no titles are found then the slice of string will be empty.
func getObjectTitles(object map[string]interface{}) []string {
	if titleList, ok := getObjectTitleList(object); ok {
		titles := []string{}
		for _, tObj := range titleList {
			if title, ok := tObj["title"]; ok {
				titles = append(titles, title)
			}
		}
		return titles
	}
	return []string{}
}

// getObjectDescription retrieves the description (a.k.a. abstract) from the DataCite Object
// See example JSON <https://api.test.datacite.org/dois/10.82433/q54d-pf76?publisher=true&affiliation=true>
func getObjectDescription(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if description, ok := attrs["description"].(string); ok {
			return description
		}
	}
	return ""
}

// getObjectPublisher
// See example JSON <https://api.test.datacite.org/dois/10.82433/q54d-pf76?publisher=true&affiliation=true>
func getObjectPublisher(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
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
			if items, ok := attributes["relatedItems"].([]map[string]interface{}); ok {
				for _, item := range items {
					if relationType, ok := item["relationType"]; ok && relationType == "IsPublishedIn" {
						if titles, ok := item["titles"].([]map[string]interface{}); ok {
							for _, title := range titles {
								if val, ok := title["title"].(string); ok {
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
		if items, ok := attrs["relatedItems"].([]map[string]interface{}); ok {
			for _, item := range items {
				if issue, ok := item["issue"].(string); ok {
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
		if items, ok := attrs["relatedItems"].([]map[string]interface{}); ok {
			for _, item := range items {
				if issue, ok := item["volume"].(string); ok {
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
		if items, ok := attrs["relatedItems"].([]map[string]interface{}); ok {
			for _, item := range items {
				if issue, ok := item["issue"].(string); ok {
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
		if items, ok := attrs["relatedItems"].([]map[string]interface{}); ok {
			for _, item := range items {
				if firstPage, ok := item["firstPage"]; ok {
					if lastPage, ok := item["lastPage"]; ok {
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
			for _, identifier := range identifiers.([]map[string]interface{}) {
				if identifierType, ok := identifier["relatedIdentifierType"]; ok && identifierType == "ISBN" {
					if val, ok := identifier["relatedIdentifier"].(string); ok {
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
		if identifiers, ok := attrs["relatedIdentifiers"]; ok {
			for _, identifier := range identifiers.([]map[string]interface{}) {
				if identifierType, ok := identifier["relatedIdentifierType"]; ok && identifierType == "ISSN" {
					if val, ok := identifier["relatedIdentifier"].(string); ok {
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
	/* FIXME: Need to find example of where this is in DataCite JSON */
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
	return nil
}

func crosswalkObjectAuthorAffiliationToCreatorAffiliation(object map[string]interface{}) *simplified.Affiliation {
	/* FIXME: Need to find an example of where this is in DataCite JSON */
	return nil
}

func crosswalkObjectPersonToCreator(object map[string]interface{}) *simplified.Creator {
	/* FIXME: Need to figure this in DataCite JSON */
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
				if literal, ok := entity["literal"].(string); ok {
					agent.PersonOrOrg.Name = literal
				}
				if family, ok := entity["family_name"].(string); ok {
					agent.PersonOrOrg.FamilyName = family
				}
				if given, ok := entity["given_name"].(string); ok {
					agent.PersonOrOrg.GivenName = given
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
	return getObjectAgents(object, "author")
}

func getObjectContributors(object map[string]interface{}) []*simplified.Creator {
	return getObjectAgents(object, "contributors")
}

func getObjectLicenses(object map[string]interface{}) []*simplified.Right {
	/* FIXME: Need ot figure this out in DataCite JSON */
	return nil
}

func getObjectSubjects(object map[string]interface{}) []*simplified.Subject {
	/* FIXME: Need to figure this out in DataCite JSON */
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

func getObjectPublicationDate(object map[string]interface{}) string {
	if attrs, ok := getObjectDataAttributes(object); ok {
		if publicationDate, ok := attrs["published"].(string); ok {
			return publicationDate
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
	doi := getObjectDOI(object)
	if doi != "" {
		doiPrefix, _ := DoiPrefix(doi)
		if value, ok := options.DoiPrefixPublishers[doiPrefix]; ok {
			return value
		}
	}
	for _, issn := range getObjectISSNs(object) {
		if issn != "" {
			if value, ok := options.ISSNPublishers[issn]; ok {
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
	// NOTE: Description becomes Abstract in EPrint records
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
	if val := getObjectPublisher(object); val != "" {
		// NOTE: Setting the publisher name is going to be normalized via DOI prefix and ISSN.
		val = normalizeObjectPublisherName(val, object, options)
		if err := SetPublisher(rec, val); err != nil {
			return nil, err
		}
	}
	if value := getObjectPublication(object); value != "" {
		if err := SetPublication(rec, value); err != nil {
			return nil, err
		}
	}
	/* FIXME: Need to know where this it's assigned in simplified model.
	   Also the data I fetch from DataCite now looks like an alternate short
	   title so objects.message["short-container-title"] may not be the right
	   place to fetch this data.
	   if value := getObjectSeries(object); value != "" {
	       if err := SetSeries(rec, value); err != nil {
	           return nil, err
	       }
	   }
	*/
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
	if values := getObjectFunding(object); values != nil && len(values) > 0 {
		if err := SetFunding(rec, values); err != nil {
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
