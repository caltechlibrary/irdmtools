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
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
)

// RdmUtil holds the configuration for rdmutil cli.
type RdmUtil struct {
	Cfg *Config
	Debug bool
}

// Configure reads the configuration file and environtment
// initialing the Cfg attribute of a RdmUtil object. It returns an error
// if problem were encounter.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.Configure("irdmtools.json", "TEST_"); err != nil {
//	   // ... handle error ...
//	}
//	fmt.Printf("Invenio RDM API UTL: %q\n", app.Cfg.IvenioAPI)
//	fmt.Printf("Invenio RDM token: %q\n", app.Cfg.InvenioToken)
//
// ```
func (app *RdmUtil) Configure(configFName string, envPrefix string, debug bool) error {
	if app == nil {
		app = new(RdmUtil)
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
	if (app.Cfg.InvenioAPI == "" || app.Cfg.InvenioToken == "") && app.Cfg.InvenioDbHost == "" {
		return fmt.Errorf("RDM_URL or RDMTOK not set, RDM_DB_HOST not set")
	}
	return nil
}

// OpenDB takes a configured RdmUtil struct and opens the described database connection.
func (app *RdmUtil) OpenDB() error {
	if app.Cfg == nil {
		return fmt.Errorf("application not configured")
	}
	connstr := app.Cfg.MakeDSN()
	if connstr == "" {
		return fmt.Errorf("unable to form db connection string")
	}
	db, err := sql.Open("postgres", connstr)
	if err != nil {
		return err
	}
	if db != nil {
		app.Cfg.pgDB = db
	}
	return nil
}

// CloseDB() closes the Postgres connection
func (app *RdmUtil) CloseDB() error {
	if app.Cfg != nil {
		if app.Cfg.pgDB != nil {
			return app.Cfg.pgDB.Close()
		}
	}
	return fmt.Errorf("postgres db connection not found")
}

// CheckDOI checks the .pids.doi.identifier and returns a record from
// the match DOI.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//  doi := '10.1126/science.82.2123.219'
//	src, err := app.CheckDOI(doi)
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) CheckDOI(doi string) ([]byte, error) {
	records, err := CheckDOI(app.Cfg, doi)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(records, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}


// GetModified returns a byte slice for a JSON encode list
// of record ids modified (created, updated, deleted) in
// the given time range. If a problem occurs an error is returned.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	src, err := app.GetModifiedIds("2020-01-01", "2020-12-31")
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetModifiedIds(start string, end string) ([]byte, error) {
	ids, err := GetModifiedRecordIds(app.Cfg, start, end)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(ids, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetRecordIds returns a byte slice for a JSON encode list
// of record ids or an error. The record ids are for the latest
// pbulished verison of the records.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	src, err := app.GetRecordIds()
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetRecordIds() ([]byte, error) {
	ids, err := GetRecordIds(app.Cfg)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(ids, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}


// GetRecordStaleIds returns a byte slice for a JSON encode list
// of record ids or an error. The record ids are for the stale
// versions of published records.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	src, err := app.GetRecordStaleIds()
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetRecordStaleIds() ([]byte, error) {
	ids, err := GetRecordStaleIds(app.Cfg)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(ids, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetRecord returns a byte slice for a JSON encoded record
// or an error.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	recordId := "woie-x0121"
//	src, err := app.GetRecord(recordId)
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetRecord(id string) ([]byte, error) {
	rec, err := GetRecord(app.Cfg, id, false)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(rec, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetRawRecord returns a byte slice for a JSON encoded record
// as a `map[string]interface{}` retrieved from the RDM API.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	recordId := "woie-x0121"
//	src, err := app.GetRawRecord(recordId)
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetRawRecord(id string) ([]byte, error) {
	rec, err := GetRawRecord(app.Cfg, id)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(rec, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetRecordVersions returns a byte slice fron JSON encoded list
// of record versions for a given RDM record id.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	recordId := "5wh3x-cj477"
//	src, err := app.GetRecordVersions(recordId)
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetRecordVersions(id string) ([]byte, error) {
	records, err := GetRecordVersions(app.Cfg, id)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(records, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}


// GetDraftFiles returns the metadata for a draft's files
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId := "woie-x0121"
// src, err := app.GetFiles(recordId)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetDraftFiles(recordId string) ([]byte, error) {
	data, err := GetDraftFiles(app.Cfg, recordId, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// GetFiles returns the metadata for working with files
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId := "woie-x0121"
// src, err := app.GetFiles(recordId)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetFiles(recordId string) ([]byte, error) {
	data, err := GetFiles(app.Cfg, recordId, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}


// GetFile returns the metadata for a file
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId := "woie-x0121"
// fName := "article.pdf"
// src, err := app.GetFile(recordId, fName) 
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetFile(id string, fName string) ([]byte, error) {
	obj, err := GetFile(app.Cfg, id, fName)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(obj, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// RetrieveFile retrieves the file from an RDM instance.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId := "woie-x0121"
// fName := "article.pdf"
// data, err := app.RetrieveFile(recordId, fName) 
// if err != nil {
//   // ... handle error ...
// }
// os.WriteFile("article.pdf", data, 0664)
//
// ```
func (app *RdmUtil) RetrieveFile(id string, fName string) ([]byte, error) {
	data, err := RetrieveFile(app.Cfg, id, fName)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetVersions returns the versioning metadata for a record.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId := "woie-x0121"
// src, err := app.GetVersions(recordId)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetVersions(id string) ([]byte, error) {
	obj, err := GetVersions(app.Cfg, id)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(obj, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetVersionLatest returns the latest version metadata for a record.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId := "woie-x0121"
// src, err := app.GetVersionLatest(recordId)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetVersionLatest(id string) ([]byte, error) {
	obj, err := GetVersionLatest(app.Cfg, id)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(obj, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// NewRecord create a new record from JSON source. It returns a created
// record including a record id.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// jsonSrc, _ := os.ReadFile("new_record.json")
// src, err := app.NewRecord(jsonSrc)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) NewRecord(src []byte) ([]byte, error) {
	data, err := NewRecord(app.Cfg, src)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// NewRecordVersion create a new record version using record id.
// It returns a created record including a record id.
//
// ```
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId = "woie-x0121"
// src, err := app.NewRecordVersion(recordId)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func (app *RdmUtil) NewRecordVersion(recordId string) ([]byte, error) {
	data, err := NewRecordVersion(app.Cfg, recordId)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// PublishRecordVersion publish a new version draft using the 
// version's record id.
//
// ```
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId = "woie-x0121"
// version, pubDate := "internal", "2022-08"
// src, err := app.PublishRecordVersion(recordId)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func (app *RdmUtil) PublishRecordVersion(recordId string, version string, pubDate string) ([]byte, error) {
	data, err := PublishRecordVersion(app.Cfg, recordId, version, pubDate, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}



// NewDraft returns the a new draft of an existing record.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId = "woie-x0121"
// src, err := app.NewDraft(recordId)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) NewDraft(recordId string) ([]byte, error) {
	data, err := NewDraft(app.Cfg, recordId)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// GetDraft returns an existing draft of a record.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId := "woie-x0121"
// src, err := app.GetDraft(recordId)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) GetDraft(id string) ([]byte, error) {
	obj, err := GetDraft(app.Cfg, id)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(obj, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// UpdateDraft returns takes a record id and returns a draft record.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// id := "woie-x0121"
// jsonSrc, _ := os.ReadFile("draft.json")
// src, err := app.UpdateDraft(id, jsonSrc)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) UpdateDraft(recordId string, src []byte) ([]byte, error) {
	data, err := UpdateDraft(app.Cfg, recordId, src, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// SendToCommunity takes a RDM record id and community UUID. It populates the
// the parent element approriately for draft to be submitted to a specific community.
func (app *RdmUtil) SendToCommunity(recordId string, communityId string) ([]byte, error) {
	data, err := SendToCommunity(app.Cfg, recordId, communityId, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// SetFilesEnable takes a RDM record id and either boolean setting
// the files.enabled value in a draft record. Returns the draft record
// and an error value.
func (app *RdmUtil) SetFilesEnable(recordId string, enable bool) ([]byte, error) {
	m, err := SetFilesEnable(app.Cfg, recordId, enable, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(m, "", "     ")
}

// SetVersion takes a RDM record id and version string
// setting the .metadata.version in a draft record.
func (app *RdmUtil) SetVersion(recordId string, version string) ([]byte, error) {
	m, err := SetVersion(app.Cfg, recordId, version, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(m, "", "     ")
}

// SetPubDate takes a RDM record id and publication date string
// setting the .metadata.publication_date in a draft record.
func (app *RdmUtil) SetPubDate(recordId string, pubDate string) ([]byte, error) {
	m, err := SetPubDate(app.Cfg, recordId, pubDate, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(m, "", "     ")
}

// UploadFiles takes a RDM record id an list of files and uploads them to a draft.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// id := "woie-x0121"
// filenames := []string{ "article.pdf", "charts.zip", "data.zip" }
// src, err := app.UploadFiles(id, filenames)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) UploadFiles(recordId string, filenames []string) ([]byte, error) {
	data, err := UploadFiles(app.Cfg, recordId, filenames, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}



// DeleteFiles takes a RDM record id an list of files and removes them from a draft.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// id := "woie-x0121"
// filenames := []string{ "article.pdf", "charts.zip", "data.zip" }
// src, err := app.DeleteFiles(id, filenames)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) DeleteFiles(recordId string, filenames []string) ([]byte, error) {
	data, err := DeleteFiles(app.Cfg, recordId, filenames, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}


// DiscardDraft takes a record id and delete the draft.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// id := "woie-x0121"
// src, err := app.DiscardDraft(id)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) DiscardDraft(recordId string) ([]byte, error) {
	data, err := DiscardDraft(app.Cfg, recordId, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// GetReview takes a record id and returns information about
// review requests.
//
// ```
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// id := "woie-x0121"
// src, err := app.GetReview(id, "accept", "")
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func (app *RdmUtil) GetReview(recordId string) ([]byte, error) {
	data, err := GetReview(app.Cfg, recordId, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// ReviewRequest takes a record id, a decision and a comment and
// submits it to the review process.
//
// ```
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// id := "woie-x0121"
// src, err := app.ReviewRequest(id, "accept", "")
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func (app *RdmUtil) ReviewRequest(recordId string, decision string, comment string) ([]byte, error) {
	data, err := ReviewRequest(app.Cfg, recordId, decision, comment, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// ReviewComment takes a record id and a comment and
// submits the comment to the review process.
//
// ```
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// id := "woie-x0121"
// src, err := app.ReviewComment(id, "Not sure about this one, but it is exciting")
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func (app *RdmUtil) ReviewComment(recordId string, comment string) ([]byte, error) {
	data, err := ReviewRequest(app.Cfg, recordId, "comment", comment, app.Debug)
	if err != nil {
		return nil, err
	}
	return JSONMarshalIndent(data, "", "    ")
}

// GetAccess returns the JSON for the access attribute in a record if
// accessType parameter is an empty string or the specific access
// requested if not (e.g. "files", "record"). An error value is also
// returned.
//
// ```
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// recordId := "woie-x0121"
// accessType := "" // accessType = "record" // accessType := "files"
// src, err := app.GetAccess(recordId, accessType)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func (app *RdmUtil) GetAccess(id string, accessType string) ([]byte, error) {
	var src []byte
	src, err := GetAccess(app.Cfg, id, accessType)
	if err != nil {
		return nil, err
	}
	return src, nil
}

// SetAccess sets the access attribute for a record. The access type can
// be either record or files. The value can be either "public" or 
// "restricted". An error value is also returned with the function.
//
// ```
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	recordId := "woie-x0121"
//  accessType := "record" 
//	src, err := app.SetAccess(recordId, accessType, "public")
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//  accessType = "files"
//	src, err := app.SetAccess(recordId, accessType, "restricted")
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
// ```
func (app *RdmUtil) SetAccess(id string, accessType string, accessValue string) ([]byte, error) {
	var src []byte
	if accessType != "record" && accessType != "files" && accessType != "embargo" {
		return nil, fmt.Errorf("%q is not a supported access type (e.g. files, record)", accessType)
	}
	if accessValue != "public" && accessValue != "restricted" {
		return nil, fmt.Errorf("%q is not a supported access value (e.g. public, restricted)", accessValue)
	}
	// FIXME: I don't need to support embargo for migration but should
	// added later when we've migrated authors and before we migrate
	// thesis.
	src, err := SetAccess(app.Cfg, id, accessType, accessValue, app.Debug)
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetEndpoint performs a GET on the endpoint indicated by PATH provided.
func (app *RdmUtil) GetEndpoint(p string) ([]byte, error) {
	return GetEndpoint(app.Cfg, p)
}

// PostEndpoint performs a POST on the endpoint indicated by PATH provided.
func (app *RdmUtil) PostEndpoint(p string, data []byte) ([]byte, error) {
	return PostEndpoint(app.Cfg, p, data)
}

// PutEndpoint performs a PUT on the endpoint indicated by PATH provided.
func (app *RdmUtil) PutEndpoint(p string, data []byte) ([]byte, error) {
	return PutEndpoint(app.Cfg, p, data)
}

// PatchEndpoint performs a PATCH on the endpoint indicated by 
// PATH provided.
func (app *RdmUtil) PatchEndpoint(p string, data []byte) ([]byte, error) {
	return PatchEndpoint(app.Cfg, p, data)
}

// DeleteEndpoint performs a DELETE on the endpoint indicated by 
// PATH provided.
func (app *RdmUtil) DeleteEndpoint(p string) ([]byte, error) {
	return DeleteEndpoint(app.Cfg, p)
}


// Harvest takes a JSON file contianing a list of record ids and
// harvests them into a dataset v2 collection. The dataset collection
// must exist and be configured in either the environment or
// configuration file.
func (app *RdmUtil) Harvest(fName string) error {
	return Harvest(app.Cfg, fName, app.Cfg.Debug)
}

// getRecordParams parse the command parameters for record id oriented
// actions.
func getRecordParams(params []string, requireRecordId bool, requireInName bool, requireOutName bool) (string, string, string, error) {
	var (
		recordId string
		inName string
		outName string
	)
	i := 0
	if len(params) > i {
		recordId = params[i]
		i++
	} else if requireRecordId {
		return "", "", "", fmt.Errorf("(%d) Missing record id", i)
	}
	if len(params) > i {
		inName = params[i]
		i++
	} else if requireInName {
		return recordId, "", "", fmt.Errorf("(%d) Missing input filename", i)
	}
	if len(params) > i {
		outName = params[i]
		i++
	} else if requireOutName {
		return recordId, inName, "", fmt.Errorf("(%d) Missing output filename", i)
	}
	return recordId, inName, outName, nil
}

// getVersionParams parse the command parameters for record id and
// version oriented values.
func getVersionParams(params []string, requireRecordId bool, requireVersion bool, requirePubDate bool) (string, string, string, error) {
	var (
		recordId string
		version string
		pubDate string
	)
	i := 0
	if len(params) > i {
		recordId = params[i]
		i++
	} else if requireRecordId {
		return "", "", "", fmt.Errorf("(%d) Missing record id", i)
	}
	if len(params) > i {
		version = params[i]
		i++
	} else if requireVersion {
		return recordId, "", "", fmt.Errorf("(%d) Missing version label", i)
	}
	if len(params) > i {
		pubDate = params[i]
		// FIXME: Should vet the format of the date ...
		i++
	} else if requirePubDate {
		return recordId, version, "", fmt.Errorf("(%d) Missing publication date", i)
	}
	return recordId, version, pubDate, nil
}

// getIOParams parse the command parameters where the only options
// are setting input and output (i.e. no record id involved).
func getIOParams(params []string, requireInName bool, requireOutName bool) (string, string, error) {
	var (
		inName string
		outName string
	)
	i := 0
	if len(params) > i {
		inName = params[i]
		i++
	} else if requireInName {
		return "", "", fmt.Errorf("(%d) Missing input filename", i)
	}
	if len(params) > i {
		outName = params[i]
		i++
	} else if requireOutName {
		return inName, "", fmt.Errorf("(%d) Missing output filename", i)
	}
	return inName, outName, nil
}


func getFileParams(params []string, requireRecordId bool, requireFilenames bool) (string, []string, error) {
	recordId := ""
	filenames := []string{}
	i := 0
	if len(params) > i {
		recordId = params[i]
		i++
	} else if requireRecordId {
		return "", filenames, fmt.Errorf("Missing record id")
	}
	if len(params) > i {
		for j := i; j < len(params); j++ {
			filenames = append(filenames, params[j])
		}
	} else if requireFilenames {
		return "", filenames, fmt.Errorf("Missing filenames to upload")
	}
	return recordId, filenames, nil
}

func getAccessParams(params []string, requireRecordId bool, requireType, requireValue bool) (string, string, string, error) {
	recordId, accessType, accessVal := "", "", ""
	i := 0
	if len(params) > i {
		recordId = params[i]
		i++
	} else if requireRecordId {
		return "", "", "", fmt.Errorf("Missing record id")
	}
	if len(params) > i {
		accessType = params[i]
		i++
	} else if requireType {
		return recordId, "", "", fmt.Errorf("Missing access type")
	}
	if len(params) > i {
		accessVal = params[i]
		i++
	} else if requireValue {
		return recordId, accessType, "", fmt.Errorf("Missing access value")
	}
	return recordId, accessType, accessVal, nil
}


func getEndpointParams(params []string, requirePath bool, requireInName bool) (string, string, error) {
	p, inName := "", ""
	i := 0
	if len(params) > i {
		p = params[i]
		i++
	} else if requirePath {
		return "", "", fmt.Errorf("Missing path for endpoint")
	}
	if len(params) > i {
		inName = params[i]
		i++
	} else if requireInName {
		return "", "", fmt.Errorf("Missing input filename for endpoint")
	}
	return p, inName, nil
}

func getReviewCommentParams(params []string, requireRecordId bool, requireComment bool) (string, string, error) {
	recordId, comment := "", ""
	i := 0
	if len(params) > i {
		recordId = params[i]
		i++
	} else if requireRecordId {
		return "", "", fmt.Errorf("Missing record id")
	}
	if len(params) > i {
		comment = params[i]
		i++
	} else if requireComment {
		return recordId, "", fmt.Errorf("Missing comment")
	}
	return recordId, comment, nil
}

func getReviewParams(params []string, requireRecordId bool, requireDecision bool, requireComment bool) (string, string, string, error) {
	recordId, decision, comment := "", "", ""
	i := 0
	if len(params) > i {
		recordId = params[i]
		i++
	} else if requireRecordId {
		return "", "", "", fmt.Errorf("Missing record id")
	}
	if len(params) > i {
		decision = params[i]
		i++
	} else if requireDecision {
		return recordId, "", "", fmt.Errorf("Missing decision")
	}
	if len(params) > i {
		comment = params[i]
		i++
	} else if requireComment {
		return recordId, decision, "", fmt.Errorf("Missing comment")
	}
	return recordId, decision, comment, nil
}


// Run implements the irdmapp cli behaviors. With the exception of the
// Run.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	recordId := "wx0w-2231"
//	src, err := app.Run(os.Stdin, os.Stdout, os.Stderr,
//	                     "get_record", []string{recordId})
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) Run(in io.Reader, out io.Writer, eout io.Writer, action string, params []string) error {
	var (
		src []byte
		err error
		recordId string
		communityId string
		inName string
		outName string
		enable bool
		decision string
		comment string
		accessType string
		val string
		p string
		data []byte
		filenames []string
		version string
		pubDate string
	)
	switch action {
	case "setup":
		if len(params) == 0 {
			return fmt.Errorf("missing configuration name")
		}
		src, err = SampleConfig(params[0])
	case "check_doi":
		if len(params) == 0 {
			return fmt.Errorf("missing DOI to match")
		}
		doi := params[0]
		src, err = app.CheckDOI(doi)
	case "get_modified_ids":
		if len(params) == 0 {
			return fmt.Errorf("missing start and end dates")
		}
		start, end := params[0], ""
		if len(params) > 1 {
			end = params[1]
		}
		if err := app.OpenDB(); err != nil {
			return err
		}
		defer app.CloseDB()
		src, err = app.GetModifiedIds(start, end)
	case "get_all_ids":
		if err := app.OpenDB(); err != nil {
			return err
		}
		defer app.CloseDB()
		src, err = app.GetRecordIds()
	case "get_all_stale_ids":
		if err := app.OpenDB(); err != nil {
			return err
		}
		defer app.CloseDB()
		src, err = app.GetRecordStaleIds()
	case "get_raw_record":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.GetRawRecord(recordId)
	case "get_record":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		if err := app.OpenDB(); err != nil {
			return err
		}
		defer app.CloseDB()
		src, err = app.GetRecord(recordId)
	case "get_record_versions":
	    recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.GetRecordVersions(recordId)
	case "get_draft_files":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.GetDraftFiles(recordId)
	case "get_files":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.GetFiles(recordId)
	case "get_file":
		recordId, inName, _, err = getRecordParams(params, true, true, false)
		if err != nil {
			return err
		}
		src, err = app.GetFile(recordId, inName)
	case "retrieve_file":
		recordId, inName, outName, err = getRecordParams(params, true, true, true)
		if err != nil {
			return err
		}
		data, err = app.RetrieveFile(recordId, inName)
		if err != nil {
			return err
		}
		if err := os.WriteFile(outName, data, 0664); err != nil {
			return err
		}
		fmt.Fprintf(out, "Wrote %s %d bytes\n", outName, len(data))
		return nil
	case "get_versions":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.GetVersions(recordId)
	case "get_latest_version":
		recordId, _, outName, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.GetVersionLatest(recordId)
		if err != nil {
			return err
		}
		if outName != "" && outName != "-" {
			if err := os.WriteFile(outName, src, 0664); err != nil {
				return err
			}
			return nil
		}
	case "new_record":
		inName, outName, err = getIOParams(params, false, false)
		if inName != "" && inName != "-" {
			src, err = os.ReadFile(inName)
		} else {
			src, err = io.ReadAll(in)
		}
		if err != nil {
			return err
		}
		src, err = app.NewRecord(src)
		if err != nil {
			return err
		}
		if outName != "" && outName != "-" {
			if err := os.WriteFile(outName, src, 0664); err != nil {
				return err
			}
			return nil
		}
	case "new_version":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}	
		src, err = app.NewRecordVersion(recordId)
	case "publish_version":
		recordId, version, pubDate, err = getVersionParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.PublishRecordVersion(recordId, version, pubDate)
	case "new_draft":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.NewDraft(recordId)
	case "get_draft":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err !=nil {
			return err
		}
		src, err = app.GetDraft(recordId)
	case "update_draft":
		recordId, inName, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		if inName != "" && inName != "-" {
			src, err = os.ReadFile(inName)
		} else {
			src, err = io.ReadAll(in)
		}
		if err != nil {
			return err
		}
		src, err = app.UpdateDraft(recordId, src)
	case "set_files_enable":
		if len(params) != 2 {
			return fmt.Errorf("expected record id and either true or false")
		}
		if params[1] != "true" && params[1] != "false" {
			return fmt.Errorf("files.enable can be set to true or false")
		}
		recordId, enable = params[0], (params[1] == "true")
		src, err = app.SetFilesEnable(recordId, enable)
	case "set_version":
		if len(params) != 2 {
			return fmt.Errorf("expected record id and version string")
		}
		recordId, version = params[0], params[1]
		src, err = app.SetVersion(recordId, version)
	case "set_publication_date":
		if len(params) != 2 {
			return fmt.Errorf("expected record id and publication date")
		}
		recordId, pubDate = params[0], params[1]
		src, err = app.SetPubDate(recordId, pubDate)
	case "upload_files":
		recordId, filenames, err = getFileParams(params, true, true)
		if err != nil {
			return err
		}
		src, err = app.UploadFiles(recordId, filenames)
	case "delete_files":
		recordId, filenames, err = getFileParams(params, true, true)
		if err != nil {
			return err
		}
		src, err = app.DeleteFiles(recordId, filenames)
	case "discard_draft":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.DiscardDraft(recordId)
	case "send_to_community":
		if len(params) != 2 {
			return fmt.Errorf("expected record id and RDM comminity id")
		}
		recordId, communityId = params[0], params[1]
		src, err = app.SendToCommunity(recordId, communityId)
	case "get_review":
		if len(params) != 1 {
			return fmt.Errorf("expected record id")
		}
		recordId := params[0]
		src, err = app.GetReview(recordId)
	case "review_comment":
		recordId, comment, err = getReviewCommentParams(params, true, true)
		if err != nil {
			return err
		}
		src, err = app.ReviewComment(recordId, comment)
	case "review_request":
		recordId, decision, comment, err = getReviewParams(params, true, true, false)
		if err != nil {
			return err
		}
		src, err = app.ReviewRequest(recordId, decision, comment)
	case "get_access":
		recordId, accessType, _, err = getAccessParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.GetAccess(recordId, accessType)
	case "set_access":
		recordId, accessType, val, err = getAccessParams(params, true, true, true)
		if err != nil {
			return err
		}
		src, err = app.SetAccess(recordId, accessType, val)
	case "get_endpoint":
		if len(params) != 1 {
			return fmt.Errorf("get_endpoint requires a PATH value")
		}
		src, err = app.GetEndpoint(params[0])
	case "post_endpoint":
		p, inName, err = getEndpointParams(params, true, false)
		if err != nil {
			return err
		}
		data := []byte{}
		if inName != "" && inName != "-" {
			data, err = os.ReadFile(inName)
		} else {
			data, err = io.ReadAll(os.Stdin)
		}
		src, err = app.PostEndpoint(p, data)
	case "put_endpoint":
		p, inName, err = getEndpointParams(params, true, false)
		if err != nil {
			return err
		}
		data := []byte{}
		if inName != "" && inName != "-" {
			data, err = os.ReadFile(inName)
		} else {
			data, err = io.ReadAll(os.Stdin)
		}
		src, err = app.PutEndpoint(p, data)
	case "patch_endpoint":
		p, inName, err = getEndpointParams(params, true, false)
		if err != nil {
			return err
		}
		data := []byte{}
		if inName != "" && inName != "-" {
			data, err = os.ReadFile(inName)
		} else {
			data, err = io.ReadAll(os.Stdin)
		}
		src, err = app.PatchEndpoint(p, data)
	case "delete_endpoint":
		p, _, err = getEndpointParams(params, true, false)
		if err != nil {
			return err
		}
		src, err = app.DeleteEndpoint(p)
	case "harvest":
		if len(params) != 1 {
			return fmt.Errorf("JSON Identifier file required")
		}
		if err := app.OpenDB(); err != nil {
			return err
		}
		defer app.CloseDB()
		if err := app.Harvest(params[0]); err != nil {
			return err
		}

	default:
		err = fmt.Errorf("%q action is not supported", action)
	}
	if err != nil {
		return err
	}
	if src != nil {
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
	}
	return nil
}
