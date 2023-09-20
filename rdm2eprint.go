// irdmtools is a package for working with institutional repositories and
// data management systems. Current implementation targets Invenio-RDM.
//
// @author R. S. Doiel, <rsdoiel@caltech.edu>
// @author Tom Morrell, <tmorrell@caltech.edu>
//
// Copyright (c) 2023, Caltech
// All rights not granted herein are expressly reserved by Caltech.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
package irdmtools

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	// Caltech Library Packages
	"github.com/caltechlibrary/eprinttools"
	"github.com/caltechlibrary/simplified"
)


// Rdm2EPrint holds the configuration for rdmutil cli.
type Rdm2EPrint struct {
	Cfg *Config
}


var (

	// resourceMap maps a resource from RDM to EPRints.
	resourceMap = map[string]string {
		"publication-article": "article",
		"publication-section": "book_section",
		"publication-report": "monograph",
		"publication-book": "book",
		"conference-paper": "conference_item",
		"conference-poster": "conference_item",
		"conference-presentation": "conference_item",
		"publication-conferenceproceeding": "book",
		"publication-patent": "patent",
		"publication-technicalnote": "monograph",
		"publication-thesis": "thesis",
		"teachingresource": "teaching_resource",
		"teachingresource-lecturenotes": "teching_resource",
		"teachingresource-textbook": "teaching_resource",
	}
)


// CrosswalkRdmToEPrint takes a public RDM record and
// converts it to an EPrint struct which can be rendered as
// JSON or XML.
//
// ```
// app := new(irdmtools.Rdm2EPrint)
//
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//
// recordId := "woie-x0121"
// src, err := app.GetRecord(cfg, recordId, false)
//
//	if err != nil {
//	   // ... handle error ...
//	}
//
// rec := new(simplified.Record)
// eprint := new (eprinttools.EPrint)
// eprints := new (eprinttools.EPrints)
//
//	if err := irdmtools.JSONUnmarshal(src, &rec); err != nil {
//	   // ... handle error ...
//	}
//
//	if err := CrosswalkRdmToEPrint(rec, eprint) {
//	   // ... handle error ...
//	}
//
// // Add eprint to outer EPrints struct before rendering
// eprints.EPrint = append(eprints.EPrint, eprint)
// // Output as JSON for single eprint record
// src, _ := irdmtools.JSONMarshalIndent(eprints)
// fmt.Printf("%s\n", src)
// ```
func CrosswalkRdmToEPrint(rec *simplified.Record, eprint *eprinttools.EPrint) error {
	// get EPrint ID from rec if set
	if eprintid, ok := getMetadataIdentifier(rec, "eprintid"); ok {
		if eprintid != "" {
			eprint.EPrintID, _ = strconv.Atoi(eprintid)
		}
		eprint.ID = eprintid
	}
	if doi, ok := getMetadataIdentifier(rec, "doi"); ok {
		eprint.DOI = doi
	}
	// We'll assume these are public records so we set eprint_status to "archive" if "open"
	// otherwise we'll assume these would map to the inbox.
	if rec.RecordAccess != nil && rec.RecordAccess.Record == "public" {
		eprint.EPrintStatus = "archive"
	} else {
		eprint.EPrintStatus = "inbox"
	}
	/*
	t, err := time.Parse(time.RFC3339, rec.Created)
	if err != nil {
		return err
	}
	*/
	eprint.Datestamp = rec.Created.Format(timestamp)
	eprint.LastModified = rec.Updated.Format(timestamp)
	if resourceType, ok := getMetadataResourceType(rec, resourceMap); ok {
		fmt.Printf("DEBUG resourceType -> %T %+v\n", resourceType, resourceType)
		eprint.Type = resourceType
	}
	return nil
}

// getMetadataIdentifier retrieves an indifier by scheme and returns the
// identifier value if available from .Metadata.Identifiers
func getMetadataIdentifier(rec *simplified.Record, scheme string) (string, bool) {
	if rec.Metadata != nil && rec.Metadata.Identifiers != nil {
		for _, identifier := range rec.Metadata.Identifiers {
			if identifier.Scheme == scheme {
				return identifier.Identifier, true
			}
		}
	}
	return "", false
}

// getMetadataResourceType returns a metadata resource type if found.
func getMetadataResourceType(rec *simplified.Record, resourceMap map[string]string) (string, bool) {
	if rec.Metadata != nil && rec.Metadata.ResourceType != nil {
		fmt.Fprintf(os.Stderr, "DEBUG rec.Metadata.ResourceType -> %T %+v\n", rec.Metadata.ResourceType, rec.Metadata.ResourceType)
		if val, ok := rec.Metadata.ResourceType["id"]; ok {
			resourceType := val.(string)
			if val, ok := resourceMap[resourceType]; ok {
				resourceType = val
			}
			return strings.ReplaceAll(resourceType, "-", "_"), true
		}
	}
	return "", false
}

// Configure reads the configuration file and environtment
// initialing the Cfg attribute of a RdmUtil object. It returns an error
// if problem were encounter.
//
// ```
//
//  app := new(irdmtools.RdmUtil)
//  if err := app.Configure("irdmtools.json", "TEST_"); err != nil {
//     // ... handle error ...
//  }
//  fmt.Printf("Invenio RDM API UTL: %q\n", app.Cfg.IvenioAPI)
//  fmt.Printf("Invenio RDM token: %q\n", app.Cfg.InvenioToken)
//
// ```
func (app *Rdm2EPrint) Configure(configFName string, envPrefix string, debug bool) error {
    if app == nil {
        app = new(Rdm2EPrint)
    }
    cfg := NewConfig()
    // Load the config file if name isn't an empty string
    if configFName != "" {
        err := cfg.LoadConfig(configFName)
        if err != nil {
            return err
        }
    }
    // Merge settings from the environment
    if err := cfg.LoadEnv(envPrefix); err != nil {
        return err
    }
    app.Cfg = cfg
    if debug {
        app.Cfg.Debug = true
    }
    // Make sure we have a minimal useful configuration
    if app.Cfg.InvenioAPI == "" || app.Cfg.InvenioToken == "" {
        return fmt.Errorf("RDM_URL or RDMTOK not available")
    }
    return nil
}

func (app *Rdm2EPrint) Run(in io.Reader, out io.Writer, eout io.Writer, rdmids []string) error {
	eprints := new(eprinttools.EPrints)
	for _, rdmid := range rdmids {
		rec, err := GetRecord(app.Cfg, rdmid, false)
		if err != nil {
			return err
		}
		eprint := new(eprinttools.EPrint)
		if err := CrosswalkRdmToEPrint(rec, eprint); err != nil {
			return err
		}
		eprints.EPrint = append(eprints.EPrint, eprint)
	}
	src, err := JSONMarshalIndent(eprints, "", "     ")
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s\n", src)
	return nil
}
