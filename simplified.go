package irdmtools

import (
	"fmt"
	//"os"
	"strings"

	// Caltech Library Package
	"github.com/caltechlibrary/simplified"
)

// Wraps the simplified package with crosswalks
func SetDOI(rec *simplified.Record, doi string) error {
	pid := new(simplified.PersistentIdentifier)
	pid.Identifier = doi
	// NOTE: Per issue 24, the provider should always be external.
	pid.Provider = "external"
	pid.Client = ""
	if rec.ExternalPIDs == nil {
		rec.ExternalPIDs = make(map[string]*simplified.PersistentIdentifier)
	}
	rec.ExternalPIDs["doi"] = pid
	return nil
}

func SetResourceType(rec *simplified.Record, resourceType string, resourceTypeMap map[string]string) error {
	val, ok := resourceTypeMap[resourceType]
	if !ok {
		return fmt.Errorf("resource type %q not mapped", resourceType)
	}
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.ResourceType == nil {
		rec.Metadata.ResourceType = make(map[string]interface{})
	}
	rec.Metadata.ResourceType["id"] = val
	return nil
}

func SetTitle(rec *simplified.Record, title string) error {
	rec.Metadata.Title = title
	return nil
}

func AddAdditionalTitles(rec *simplified.Record, title *simplified.TitleDetail) error {
	rec.Metadata.AdditionalTitles = append(rec.Metadata.AdditionalTitles, title)
	return nil
}

func SetDescription(rec *simplified.Record, description string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	rec.Metadata.Description = description
	return nil
}

func SetCreators(rec *simplified.Record, creators []*simplified.Creator) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	rec.Metadata.Creators = creators
	return nil
}

func SetContributors(rec *simplified.Record, contributors []*simplified.Creator) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	rec.Metadata.Contributors = contributors
	return nil
}

func AddIdentifiers(rec *simplified.Record, identifiers []*simplified.Identifier) error {
	for _, identifier := range identifiers {
		rec.Metadata.Identifiers = append(rec.Metadata.Identifiers, identifier)
	}
	return nil
}

func AddIdentifier(rec *simplified.Record, scheme string, identifier string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.Identifiers == nil {
		rec.Metadata.Identifiers = []*simplified.Identifier{}
	}
	obj := new(simplified.Identifier)
	obj.Scheme = scheme
	obj.Identifier = identifier
	rec.Metadata.Identifiers = append(rec.Metadata.Identifiers, obj)
	return nil
}

func AddRelatedIdentifiers(rec *simplified.Record, identifiers []*simplified.Identifier) error {
	for _, identifier := range identifiers {
		rec.Metadata.RelatedIdentifiers = append(rec.Metadata.RelatedIdentifiers, identifier)
	}
	return nil
}

func AddRelatedIdentifier(rec *simplified.Record, scheme string, relationType string, identifier string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.RelatedIdentifiers == nil {
		rec.Metadata.RelatedIdentifiers = []*simplified.Identifier{}
	}
	obj := new(simplified.Identifier)
	obj.Scheme = scheme
	if relationType != "" {
		obj.RelationType = &simplified.TypeDetail{
			ID: relationType,
		}
	}
	obj.Identifier = identifier
	rec.Metadata.RelatedIdentifiers = append(rec.Metadata.RelatedIdentifiers, obj)
	return nil
}




func SetCustomField(rec *simplified.Record, customField string, key string, value interface{}) error {
	if rec.CustomFields == nil {
		rec.CustomFields = make(map[string]interface{})
	}
	if key == "" {
		rec.CustomFields[customField] = value
	} else {
		_, ok := rec.CustomFields[customField]
		if ! ok {
			rec.CustomFields[customField] = make(map[string]interface{})
		}
		m := rec.CustomFields[customField].(map[string]interface{})
		m[key] = value
		rec.CustomFields[customField] = m
	}
	return nil
}

func SetImprintField(rec *simplified.Record, key string, value interface{}) error {
	// NOTE: Journal fields are going under the "custom_fields" off the root
	// metadata object in RDM v12.
	return SetCustomField(rec, "imprint:imprint", key, value)
}

func SetEdition(rec *simplified.Record, edition string) error {
	return SetImprintField(rec, "edition", edition)
}

func SetJournalField(rec *simplified.Record, key string, value interface{}) error {
	// NOTE: Journal fields are going under the "custom_fields" off the root
	// metadata object in RDM v12.
	return SetCustomField(rec, "journal:journal", key, value)
}

func SetPublication(rec *simplified.Record, publication string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	return SetJournalField(rec, "title", publication)
}

func SetPublisherLocation (rec *simplified.Record, place string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	return SetJournalField(rec, "place", place)
}

func SetPublicationDateByType(rec *simplified.Record, dt string, publicationType string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// FIXME: Shouldn't publication_date being the dates array with a type?
	rec.Metadata.PublicationDate = dt
	switch publicationType {
	case "article":
		return SetJournalField(rec, "publication_date", dt)
	case "book":
		return SetImprintField(rec, "publication_date", dt)
	case "book_section":
		return SetImprintField(rec, "publication_date", dt)
	case "monograph":
		return SetImprintField(rec, "publication_date", dt)
	}
	return nil
}

func SetPublisher(rec *simplified.Record, publisher string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	rec.Metadata.Publisher = publisher
	return nil
}

func SetPublicationDate(rec *simplified.Record, pubDate string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	rec.Metadata.PublicationDate = pubDate
	return nil
}

func SetSeries(rec *simplified.Record, series string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal content goes in journal:journal custom fields except
	// for series, which goes in caltech:series.series
	return SetCustomField(rec, "caltech:series", "series", series)
}

func SetVolume(rec *simplified.Record, volume string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal content goes in journal:journal custom fields.
	return SetJournalField(rec, "volume", volume)
}

func SetIssue(rec *simplified.Record, issue string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal content goes in journal:journal custom fields.
	return SetJournalField(rec, "issue", issue)
}


func SetPageRange(rec *simplified.Record, pageRange string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal content goes in journal:journal custom fields.
	return SetJournalField(rec, "pages", pageRange)
}

func SetArticleNumber(rec *simplified.Record, articleNo string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Per issue #37, Article Number should go in pages.
	//return SetJournalField(rec, "article_number", articleNo)
	return SetJournalField(rec, "pages", articleNo)
}

func AddRights(rec *simplified.Record, rights []*simplified.Right) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.Rights == nil {
		rec.Metadata.Rights = []*simplified.Right{}
	}
	rec.Metadata.Rights = append(rec.Metadata.Rights, rights...)
	return nil
}

func AddSubjects(rec *simplified.Record, subjects []*simplified.Subject) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.Subjects == nil {
		rec.Metadata.Subjects = []*simplified.Subject{}
	}
	rec.Metadata.Subjects = append(rec.Metadata.Subjects, subjects...)
	return nil
}

func AddSubject(rec *simplified.Record, subject string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.Subjects == nil {
		rec.Metadata.Subjects = []*simplified.Subject{}
	}
	obj := new(simplified.Subject)
	obj.Subject = strings.TrimSpace(subject)
	rec.Metadata.Subjects = append(rec.Metadata.Subjects, obj)
	return nil
}

func AddKeyword(rec *simplified.Record, keyword string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	obj := new(simplified.Subject)
	obj.Subject = strings.TrimSpace(keyword)
	rec.Metadata.Subjects = append(rec.Metadata.Subjects, obj)
	return nil
}


func AddDate(rec *simplified.Record, dt *simplified.DateType) error {
	if dt != nil {
		if rec.Metadata == nil {
			rec.Metadata = new(simplified.Metadata)
		}
		if rec.Metadata.Dates == nil {
			rec.Metadata.Dates = []*simplified.DateType{}
		}
		if dt.Type == nil {
			dt.Type = new(simplified.Type)
			dt.Type.ID = "accepted"
		}
		rec.Metadata.Dates = append(rec.Metadata.Dates, dt)
	}
	return nil
}


func AddFunder(rec *simplified.Record, funder *simplified.Funder) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.Funding == nil {
		rec.Metadata.Funding = []*simplified.Funder{}
	}
	rec.Metadata.Funding = append(rec.Metadata.Funding, funder)
	return nil
}

func SetFunding(rec *simplified.Record, funding []*simplified.Funder) error {
	for _, funder := range funding {
		if err := AddFunder(rec, funder); err != nil {
			return err
		}
	}
	return nil
}

// FIXME: Need to implement the following functions

func SetFullTextStatus(rec *simplified.Record, status bool) error {
	return fmt.Errorf("SetFullTextStatus() not implemented")
}

func SetReferred(rec *simplified.Record, referred bool) error {
	return fmt.Errorf("SetReferred() not implemented")
}

func SetProject(rec *simplified.Record, project string) error {
	return fmt.Errorf("SetProject() not implemented")
}

func SetMonographType(rec *simplified.Record, monographType string) error {
	return fmt.Errorf("SetMonographType() not implemented")
}

func SetPresentationType(rec *simplified.Record, presentationType string) error {
	return fmt.Errorf("SetPresentationType() not implemented")
}

func SetLanguages(rec *simplified.Record, key string, value interface{}) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.Languages == nil {
		rec.Metadata.Languages = []map[string]interface{}{}
	}
	m := map[string]interface{}{}
	m[key] = value
	rec.Metadata.Languages = append(rec.Metadata.Languages, m)
	return nil
}
