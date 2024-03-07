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

func QueryDataCiteObject(cfg *Config, doi string, options *Doi2RdmOptions) (*dataciteapi.Object, error) {
	appName := path.Base(os.Args[0])
	client, err := dataciteapi.NewDataCiteClient(appName, options.MailTo)
	if err != nil {
		return nil, err
	}
	objects, err := client.Works(doi)
	if err != nil {
		return nil, err
	}
	if cfg.Debug {
		src, _ := JSONMarshalIndent(objects, "", "    ")
		fmt.Fprintf(os.Stderr, "objects JSON:\n\n%s\n\n", src)
	}
	return &objects, nil
}

// getObjectData retrieves the `.access` from the DateCite `.object`
func getObjectData(object *dataciteapi.Object) (map[string]interface{}, bool) {
	data, ok := object["data"].(map[string]interface{})
	return data, ok
}

func getObjectDataAttributes(data map[string]interface{}) (map[string]interface{}, bool) {
	attr, ok := data["attributes"].(map[string]interface{})
	return attr, ok
}

// getObjectCiteProcType retrieves the `.access.types.citeproc` value if exists.
func getObjectCiteProcType(data map[string]interface{}) string {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if types, ok := attributes["types"].(map[string]string); ok {
				if citeproc, ok := types["citeproc"].(string); ok {
					return citeproc
				}
			}
		}
		return ""
}

// getObjectResourceType retrives the resource type from objects.message.type
// runs normalize
func getObjectResourceType(object *dataciteapi.Object) string {
	// The path to type informaiton is .access.types, this has a map of types
	// for various formats (e.g. .bibtex, .citeproc, .schemeOrg, .resourceType, .resourceTypeGeneral.
	// I think using the .citeproc value makes the most sense here.
	if data, ok := getObjectData(object); ok {
		if resourceType, ok := getObjectCiteProcType(data); ok {
			return resourceType
		}
	}
	return ""
}

// getObjcetDataTitles extracts a list of titles from a list of title objects.
func getObjectDataTitles(data map[string]interface{}) ([]map[string]interface{}, bool) {
	if aatributes, ok := getObjectDataAttributes(data); ok {
		if titles, ok := data["titles"].([]map[string]string); ok {
			return titles, ok
		}
	}
	return nil, false
}

// getObjectTitles retrieves an ordered list of titles from a DataCite Object object.
// The zero index is the primary document title, the remaining are alternative titles.
// If no titles are found then the slice of string will be empty.
func getObjectTitles(object *dataciteapi.Object) []string {
	if data, ok := getObjectData(object); ok {
		if titleList, ok := getObjectDataTitles(data); ok {
			titles := []string {}
			for _, tObj := range titleList  {
				if title, ok := tObj["title"]; ok {
					titles = append(titles, title)
				}
			}
			return titles 	
		}
	}
	return []string{}
}

// getObjectAbstract retrieves the abstract from the DataCite Object
// See example JSON <https://api.test.datacite.org/dois/10.82433/q54d-pf76?publisher=true&affiliation=true>
func getObjectAbstract(object *dataciteapi.Object) string {
	/* abstract doesn't seem to exist in Schema
	if data, ok := getObjectData(object); ok {
		if abstract, ok := data["abstract"]; ok {
			return data.(string)
		}
	}
	*/
	return ""
}

// getObjectPublisher
// See example JSON <https://api.test.datacite.org/dois/10.82433/q54d-pf76?publisher=true&affiliation=true>
func getObjectPublisher(object *dataciteapi.Object) string {
	// FIXME: Need to know if publisher holds the publisher and container type holds publication based on object.Message.Type
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if publisher, ok := attributes["publisher"]; ok {
				if name, ok := publisher["name"]; ok {
					return publisher
				}
			}
		}
	}
	return ""
}

// getObjectPublication
// See example JSON <https://api.test.datacite.org/dois/10.82433/q54d-pf76?publisher=true&affiliation=true>
func getObjectPublication(object *dataciteapi.Object) string {
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if items, ok := attributes["relatedItems"].([]map[string]intereface{}); ok {
				for _, item := range items {
					if relationType, ok := items["relationType"]; ok && relationType == "IsPublishedIn" {
						if titles, ok := items["titles"].([]map[string]interface{}); ok {
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

// getObjectObjectSeries
func getObjectSeries(object *dataciteapi.Object) string {
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if items, ok := attributes["relatedItems"].([]map[string]interface{}); ok {
				for _, item := range items {
					if issue, ok := items["issue"].(string); ok {
						return issue
					}
				}
			}
		}
	}
	return ""
}

// getObjectObjectVolume
func getObjectVolume(object *dataciteapi.Object) string {
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if items, ok := attributes["relatedItems"].([]map[string]interface{}); ok {
				for _, item := range items {
					if issue, ok := items["volume"].(string); ok {
						return issue
					}
				}
			}
		}
	}
	return ""
}

// getObjectObjectIssue
func getObjectIssue(object *dataciteapi.Object) string {
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if items, ok := attributes["relatedItems"].([]map[string]interface{}); ok {
				for _, item := range items {
					if issue, ok := items["issue"].(string); ok {
						return issue
					}
				}
			}
		}
	}
	return ""
}

// getObjectObjectPublisherLocation
func getObjectPublisherLocation(object *dataciteapi.Object) string {
	/* Note sure where to find this.  */
	return ""
}

// getObjectObjectPageRange
func getObjectPageRange(object *dataciteapi.Object) string {
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if items, ok := attributes["relatedItems"].([]map[string]interface{}); ok {
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
	}
	return ""
}

// getObjectArticleNumber
func getObjectArticleNumber(object *dataciteapi.Object) string {
	/* FIXME: Not sure where article numbers map from in the DataCite API
	*/
	return ""
}

// getObjectISBNs
func getObjectISBNs(object *dataciteapi.Object) []string {
	isbns := []string{}
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if identifiers, ok := attributes["relatedIdentifiers"]; ok {
				for _, identifier := range identifiers {
					if identifierType, ok := identifier["relatedIdentifierType"]; ok && identifierType == "ISBN" {
						if val, ok := identifier["relatedIdentifier"]; ok {
							isbns = append(isbns, val)
						}
					}
				}
			}
		}
	}
	return isbns
}

// getObjectISSNs
func getObjectISSNs(object *dataciteapi.Object) []string { 
	issns := []string{}
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if identifiers, ok := attributes["relatedIdentifiers"]; ok {
				for _, identifier := range identifiers {
					if identifierType, ok := identifier["relatedIdentifierType"]; ok && identifierType == "ISSN" {
						if val, ok := identifier["relatedIdentifier"]; ok {
							isbns = append(isbns, val)
						}
					}
				}
			}
		}
	}
	return issns
}

// getObjectFunding
func getObjectFunding(object *dataciteapi.Object) []*simplified.Funder {
	/* FIXME: Need to find example of where this is in DataCite JSON */
	return nil
}

// getObjectDOI
func getObjectDOI(object *dataciteapi.Object) string {
	if data, ok := getObjectData(object); ok {
		if attributes, ok := getObjectDataAttributes(data); ok {
			if doi, ok := attributes["doi"]; ok {
				return doi
			}
		}
	}
	return ""
}

// getObjectLinks
func getObjectLinks(object *dataciteapi.Object) []*simplified.Identifier {
	/* FIXME: Need to find an example of where this is in DataCite JSON */
	return nil
}

func crosswalkObjectAuthorAffiliationToCreatorAffiliation(object *dataciteapi.Object) *simplified.Affiliation {
	/* FIXME: NEed to find an example of where this is in DataCite JSON */
	return nil
}

func crosswalkObjectPersonToCreator(object *dataciteapi.Object) *simplified.Creator {
	/* FIXME: Need to figure this in DataCite JSON */
	return nil
}

func crosswalkObjectLicenseToRight(license *dataciteapi.License) *simplified.Right {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

func crosswalkObjectDateObjectToDateType(do *dataciteapi.DateObject, description string) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

func getObjectCreators(object *dataciteapi.Object) []*simplified.Creator {
	creators := []*simplified.Creator{}
	/* FIXME: Need to figure this out in DataCite JSON */
	return creators
}

func getObjectContributors(object *dataciteapi.Object) []*simplified.Creator {
	creators := []*simplified.Creator{}
	/* FIXME: Need to figure this out in DataCITE JSON */
	return creators
}

func getObjectLicenses(object *dataciteapi.Object) []*simplified.Right {
	/* FIXME: Need ot figure this out in DataCite JSON */
	return nil
}

func getObjectSubjects(object *dataciteapi.Object) []*simplified.Subject {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

func getObjectPublishedPrint(object *dataciteapi.Object) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

func getObjectPublishedOnline(object *dataciteapi.Object) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

func getObjectPublicationDate(object *dataciteapi.Object) string {
	/* FIXME: Need to figure this out in DataCite JSON */
	return ""
}

func getObjectAccepted(object *dataciteapi.Object) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

func getObjectApproved(object *dataciteapi.Object) *simplified.DateType {
	/* FIXME: Need to figure this out in DataCite JSON */
	return nil
}

// normalizeObjectPublisherName will check the publisher DOI and ISSN to see if we have
// a preferred name in our options. If so it will return that.
func normalizeObjectPublisherName(val string, object *dataciteapi.Object, options *Doi2RdmOptions) string {
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

// CrosswalkDataCiteObject takes a Object object from the DataCite API
// and maps the fields into an simplified Record struct return a
// new struct or error.
func CrosswalkDataCiteObject(cfg *Config, object *dataciteapi.Object, options *Doi2RdmOptions) (*simplified.Record, error) {
	if object == nil {
		return nil, fmt.Errorf("crossref api objects not populated")
	}
	rec := new(simplified.Record)
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
	if values := getObjectTitles(object); len(values) > 0 {
		for i, val := range values {
			if i == 0 {
				if err := SetTitle(rec, val); err != nil {
					return nil, err
				}
			} else {
				if err := AddAdditionalTitles(rec, makeSimpleTitleDetail(val)); err != nil {
					return nil, err
				}
			}
		}
	}
	// NOTE: Abstract becomes Description in simplified records
	if value := getObjectAbstract(object); value != "" {
		if err := SetDescription(rec, value); err != nil {
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
	if value := getObjectPublisher(object); value != "" {
		// FIXME: Setting the publisher name is going to be normalized via DOI prefix, maybe ISSN?
		value := normalizeObjectPublisherName(value, object, options)
		if err := SetPublisher(rec, value); err != nil {
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
