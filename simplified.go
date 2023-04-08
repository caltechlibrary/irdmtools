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
	// FIXME: Need to verify that we want to save the publication name as publisher
	if rec.CLAnnotations == nil {
		rec.CLAnnotations = make(map[string]interface{})
	}
	rec.CLAnnotations["publication"] = publication
	rec.Metadata.Publisher = publication
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
	// FIXME: Need to figure out where this goes
	if rec.CLAnnotations == nil {
		rec.CLAnnotations = make(map[string]interface{})
	}
	rec.CLAnnotations["series"] = series
	return nil
}

func SetVolume(rec *simplified.Record, volume string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// FIXME: Need to figure out where this goes
	if rec.CLAnnotations == nil {
		rec.CLAnnotations = make(map[string]interface{})
	}
	rec.CLAnnotations["volume"] = volume
	return nil
}

func SetPageRange(rec *simplified.Record, pageRange string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// FIXME: Need to figure out where this goes
	if rec.CLAnnotations == nil {
		rec.CLAnnotations = make(map[string]interface{})
	}
	rec.CLAnnotations["page_range"] = pageRange
	return nil
}

func SetArticleNumber(rec *simplified.Record, articleNo string) error {
	if rec.Metadata == nil {
		rec.Metadata = new(simplified.Metadata)
	}
	// FIXME: Need to figure out where this goes
	if rec.CLAnnotations == nil {
		rec.CLAnnotations = make(map[string]interface{})
	}
	rec.CLAnnotations["article_number"] = articleNo
	return nil
}

func AddBookTitle(rec *simplified.Record, bookTitle string) error {
	return fmt.Errorf("AddBookTitle() not implemented.")
}

func AddFunder(rec *simplified.Record, funder string, ror string, award string) error {
	return fmt.Errorf("AddFunder() not implemented.")
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
	return fmt.Errorf("SetFunding() not implemented")
}
