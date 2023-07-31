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
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// RdmUtil holds the configuration for rdmutil cli.
type RdmUtil struct {
	Cfg *Config
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
	if app.Cfg.InvenioAPI == "" || app.Cfg.InvenioToken == "" {
		return fmt.Errorf("RDM_URL or RDMTOK not available")
	}
	return nil
}

// Query returns a byte slice for a JSON encode list
// of record summaries or an error.
//
// ```
//
//	app := new(irdmtools.RdmUtil)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	src, err := app.Query("My favorite book", -1, "newest")
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) Query(q string, sort string) ([]byte, error) {
	records, err := Query(app.Cfg, q, sort)
	if err != nil {
		return nil, err
	}
	src, err := json.MarshalIndent(records, "", "    ")
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
	src, err := json.MarshalIndent(ids, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetRecordIds returns a byte slice for a JSON encode list
// of record ids or an error.
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
	src, err := json.MarshalIndent(ids, "", "    ")
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
	rec, err := GetRecord(app.Cfg, id)
	if err != nil {
		return nil, err
	}
	src, err := json.MarshalIndent(rec, "", "    ")
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
	src, err := json.MarshalIndent(rec, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
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
func (app *RdmUtil) GetFiles(id string) ([]byte, error) {
	obj, err := GetFiles(app.Cfg, id)
	if err != nil {
		return nil, err
	}
	src, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
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
	src, err := json.MarshalIndent(obj, "", "    ")
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
	src, err := json.MarshalIndent(obj, "", "    ")
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
	src, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// CreateDraft returns the a new draft of a record.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// jsonSrc, _ := os.ReadFile("draft.json")
// src, err := app.CreateDraft(jsonSrc)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) CreateDraft(src []byte) ([]byte, error) {
	data, err := CreateDraft(app.Cfg, src)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "    ")
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
	src, err := json.MarshalIndent(obj, "", "    ")
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
	data, err := UpdateDraft(app.Cfg, recordId, src)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "    ")
}

// DeleteDraft takes a record id and delete the draft.
//
// ```
//
// app := new(irdmtools.RdmUtil)
// if err := app.LoadConfig("irdmtools.json"); err != nil {
//   // ... handle error ...
// }
// id := "woie-x0121"
// src, err := app.DeleteDraft(id)
// if err != nil {
//   // ... handle error ...
// }
// fmt.Printf("%s\n", src)
//
// ```
func (app *RdmUtil) DeleteDraft(recordId string) ([]byte, error) {
	data, err := DeleteDraft(app.Cfg, recordId)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "    ")
}


// Harvest takes a JSON file contianing a list of record ids and
// harvests them into a dataset v2 collection. The dataset collection
// must exist and be configured in either the environment or
// configuration file.
func (app *RdmUtil) Harvest(fName string) error {
	return Harvest(app.Cfg, fName, app.Cfg.Debug)
}

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
		return "", "", "", fmt.Errorf("Missing record id")
	}
	if len(params) > i {
		inName = params[i]
		i++
	} else if requireInName {
		return recordId, "", "", fmt.Errorf("Missing input filename")
	}
	if len(params) > i {
		outName = params[i]
		i++
	} else if requireOutName {
		return recordId, inName, "", fmt.Errorf("Missing output filename")

	}
	return recordId, inName, outName, nil
}


// Run implements the irdmapp cli behaviors. With the exception of the
// "setup" action you should call `app.LoadConfig()` before execute
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
	switch action {
	case "setup":
		if len(params) == 0 {
			return fmt.Errorf("missing configuration name")
		}
		src, err := SampleConfig(params[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "query":
		if len(params) == 0 {
			return fmt.Errorf("missing query string")
		}
		q, sort := params[0], ""
		if len(params) > 1 {
			sort = params[1]
		}
		src, err := app.Query(q, sort)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_modified_ids":
		if len(params) == 0 {
			return fmt.Errorf("missing start and end dates")
		}
		start, end := params[0], ""
		if len(params) > 1 {
			end = params[1]
		}
		src, err := app.GetModifiedIds(start, end)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_all_ids":
		src, err := app.GetRecordIds()
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_raw_record":
		recordId, _, _, err := getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err := app.GetRawRecord(recordId)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_record":
		recordId, _, _, err := getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err := app.GetRecord(recordId)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_files":
		recordId, _, _, err := getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err := app.GetFiles(recordId)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_file":
		recordId, inName, _, err := getRecordParams(params, true, true, false)
		if err != nil {
			return err
		}
		src, err := app.GetFile(recordId, inName)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "retrieve_file":
		recordId, inName, outName, err := getRecordParams(params, true, true, true)
		if err != nil {
			return err
		}
		data, err := app.RetrieveFile(recordId, inName)
		if err != nil {
			return err
		}
		if err := os.WriteFile(outName, data, 0664); err != nil {
			return err
		}
		fmt.Fprintf(out, "Wrote %s %d bytes\n", outName, len(data))
		return nil
	case "get_versions":
		recordId, _, _, err := getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err := app.GetVersions(recordId)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_version_latest":
		recordId, _, _, err := getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err := app.GetVersionLatest(recordId)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "create_draft":
		_, inName, _, err := getRecordParams(params, false, true, false)
		if err != nil {
			return err
		}
		src := []byte{}
		if inName != "" && inName != "-" {
			src, err = os.ReadFile(inName)
		} else {
			src, err = io.ReadAll(in)
		}
		if err != nil {
			return err
		}
		src, err = app.CreateDraft(src)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_draft":
		recordId, _, _, err := getRecordParams(params, true, false, false)
		if err !=nil {
			return err
		}
		src, err := app.GetDraft(recordId)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "update_draft":
		recordId, inName, _, err := getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src := []byte{}
		if inName != "" && inName != "-" {
			src, err = os.ReadFile(inName)
		} else {
			src, err = io.ReadAll(in)
		}
		if err != nil {
			return err
		}
		src, err = app.UpdateDraft(recordId, src)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "delete_draft":
		recordId, _, _, err := getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err := app.DeleteDraft(recordId)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "harvest":
		if len(params) != 1 {
			return fmt.Errorf("JSON Identifier file required")
		}
		return app.Harvest(params[0])
	default:
		return fmt.Errorf("%q action is not supported", action)
	}
}
