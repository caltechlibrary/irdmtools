package irdmtools

import (
	"fmt"
	"os"

	// Caltech Library Package
	"github.com/caltechlibrary/simplified"
)

// Wraps the simplified package with crosswalks
func SetDOI(rec *simplified.Record, doi string) error {
	pid := new(simplified.PersistentIdentifier)
	pid.Identifier = doi
	// NOTE: This makes this mapping Caltech Specific, should really check who the provider is.
	pid.Provider = "datacite"
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
		fmt.Fprintf(os.Stderr, "DEBUG resourceTypeMap -> %+v\n", resourceTypeMap)
		return fmt.Errorf("resource type %q not mapped", resourceType)
	}
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	if rec.Metadata.ResourceType == nil {
		rec.Metadata.ResourceType = make(map[string]string)
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
	rec.Metadata.Contributors = creators
	return nil
}

func SetContributors(rec *simplified.Record, creators []*simplified.Creator) error {
	return fmt.Errorf("SetContributors() not implemented")
}

func AddRelatedIdentifiers(rec *simplified.Record, identifiers []*simplified.Identifier) error {
	for _, identifier := range identifiers {
		rec.Metadata.Identifiers = append(rec.Metadata.Identifiers, identifier)
	}
	return nil
}

func SetPublication(rec *simplified.Record, publication string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal fields are going under the "custom_fields" off the root
	// metadata object in RDM 12.
	if rec.CustomFields == nil {
		rec.CustomFields = make(map[string]interface{})
	}
	rec.CustomFields["rdm:publication"] = publication
	return nil
}

func SetPublisher(rec *simplified.Record, publisher string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	rec.Metadata.Publisher = publisher
	return nil
}

func SetPublisherLocation(rec *simplified.Record, publisherLocation string) error {
	return fmt.Errorf("SetPublisherLocation() not implemented")
}

func SetSeries(rec *simplified.Record, series string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal content goes in rdm:journal custom fields.
	if rec.CustomFields == nil {
		rec.CustomFields = make(map[string]interface{})
	}
	_, ok := rec.CustomFields["rdm:journal"]
	if ! ok {
		rec.CustomFields["rdm:journal"] = make(map[string]interface{})
	}
	if journal, ok := rec.CustomFields["rdm:journal"]; ok {
		if journal == nil {
			journal = make(map[string]interface{})
		}
		m := journal.(map[string]interface{})
		m["series"] = series
		rec.CustomFields["rdm:journal"] = journal
	}
	return nil
}

func SetVolume(rec *simplified.Record, volume string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal content goes in rdm:journal custom fields.
	if rec.CustomFields == nil {
		rec.CustomFields = make(map[string]interface{})
	}
	_, ok := rec.CustomFields["rdm:journal"]
	if ! ok {
		rec.CustomFields["rdm:journal"] = make(map[string]interface{})
	}
	if journal, ok := rec.CustomFields["rdm:journal"]; ok {
		if journal == nil {
			journal = make(map[string]interface{})
		}
		m := journal.(map[string]interface{})
		m["volume"] = volume
		rec.CustomFields["rdm:journal"] = journal
	}
	return nil
}

func SetPageRange(rec *simplified.Record, pageRange string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal content goes in rdm:journal custom fields.
	if rec.CustomFields == nil {
		rec.CustomFields = make(map[string]interface{})
	}
	_, ok := rec.CustomFields["rdm:journal"]
	if ! ok {
		rec.CustomFields["rdm:journal"] = make(map[string]interface{})
	}
	if journal, ok := rec.CustomFields["rdm:journal"]; ok {
		if journal == nil {
			journal = make(map[string]interface{})
		}
		m := journal.(map[string]interface{})
		m["pages"] = pageRange
		rec.CustomFields["rdm:journal"] = journal
	}
	return nil
}

func SetArticleNumber(rec *simplified.Record, articleNo string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// NOTE: Journal content goes in rdm:journal custom fields.
	if rec.CustomFields == nil {
		rec.CustomFields = make(map[string]interface{})
	}
	_, ok := rec.CustomFields["rdm:journal"]
	if ! ok {
		rec.CustomFields["rdm:journal"] = map[string]interface{}{}
	}
	if journal, ok := rec.CustomFields["rdm:journal"]; ok {
		if journal == nil {
			journal = make(map[string]interface{})
		}
		m := journal.(map[string]interface{})
		m["article_number"] = articleNo
		rec.CustomFields["rdm:journal"] = journal
	}
	return nil
}

func AddBookTitle(rec *simplified.Record, bookTitle string) error {
	return fmt.Errorf("AddBookTitle() not implemented.")
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

func AddPublicationDate(rec *simplified.Record, dt string, publicationType string) error {
	return fmt.Errorf("AddPublicationDate() not implemented.")
}

func SetEdition(rec *simplified.Record, edition string) error {
	return fmt.Errorf("SetEdition() not implemented.")
}

func AddSubject(rec *simplified.Record, subject string) error {
	return fmt.Errorf("AddSubject() not implemented.")
}

func AddKeyword(rec *simplified.Record, keyword string) error {
	return fmt.Errorf("AddKeyword() not implemented")
}

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

func SetFunding(rec *simplified.Record, funding []*simplified.Funder) error {
	for _, funder := range funding {
		if err := AddFunder(rec, funder); err != nil {
			return err
		}
	}
	return nil
}
