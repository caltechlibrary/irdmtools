package irdmtools

import (
	// Caltech Library Package
	"github.com/caltechlibrary/simplified"
)

// Wrapps the simplified package with crosswalks

func SetResourceType(rec *simplified.Record, resourceType string) error {
	return fmt.Errorf("SetResourceType() not implemented")
}

func SetTitles(rec *simplified.Record, titles []string) error {
	return fmt.Errorf("SetTitle() not implemented")
}

func SetPublication(rec *simplified.Record, publication string) error {
	return fmt.Errorf("SetPublication() not implemented")
}

func SetPublisher(rec *simplified.Record, publisher string) error {
	return fmt.Errorf("SetPublisher() not implemented")
}

func SetPublisherLocation(rec *simplified.Record, publisherLocation string) error {
	return fmt.Errorf("SetPublisherLocation() not implemented")
}

func SetSeries(rec *simplified.Record, series string) error {
	return fmt.Errorf("SetSeries() not implemented")
}

func SetVolume(rec *simplified.Record, volume string) error {
	return fmt.Errorf("SetVolume() not implemented")
}

func SetPageRange(rec *simplified.Record, pageRange string) error {
	return fmt.Errorf("SetPageRange() not implemented")
}

func AddISBN(rec *simplified.Record, isbn string, isbnType string) error {
	return fmt.Errorf("AddISBN() not implemented")
}

func AddISSN(rec *simplified.Record, issn string, issnType string) error {
	return fmt.Errorf("AddISSN() not implemented")
}

func AddBookTitle(rec *simplified.Record, bookTitle string) error {
	return fmt.Errorf("AddBookTitle() not implemented.")
}

func AddFunder(rec *simplified.Record, funder string, ror string, award string) error {
	return fmt.Errorf("AddFunder() not implemented.")
}

func SetDOI(rec *simplified.Record, doi string) error {
	return fmt.Errorf("SetDOI() not implemented.")
}

func AddRelatedDOI(rec *simplified.Record, doi string) error {
	return fmt.Errorf("AddRelatedDOI() not implemented.")
}

func AddRelatedURL(rec *simplified.Record, url string) error {
	return fmt.Errorf("AddRelatedURL() not implemented.")
}

func AddPublicationDate(rec *simplified.Record, dt string, publicationType string) error {
	return fmt.Errorf("AddPublicationDate() not implemented.")
}

func AddCreator(rec *simplified.Record, creator *simplified.PersonOrOrg, role *simplified.Role) error {
	return fmt.Errorf("AddCreator() not implemented.")
}

func AddContributors(rec *simplified.Record, contributor *simplified.PersonOrOrg, role *simplified.Role) error {
	return fmt.Errorf("AddContributor() not implemented.")
}

func SetDescription(rec *simplified.Record, description string) error {
	return fmt.Errorf("SetDescription() not implemented.")
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

func PresentationType(rec *simplified.Record, presentationType string) error {
	return fmt.Errorf("SetPresentationType() not implemented")
}

