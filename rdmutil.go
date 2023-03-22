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
//    app := new(irdmtools.RdmUtil)
//    if err := app.Configure("irdmtools.json", "TEST_"); err != nil {
//       // ... handle error ...
//    }
//    fmt.Printf("Invenio RDM API UTL: %q\n", app.Cfg.IvenioAPI)   
//    fmt.Printf("Invenio RDM token: %q\n", app.Cfg.InvenioToken)   
// ```
func (app *RdmUtil) Configure(configFName string, envPrefix string, debug bool) error {
	if app == nil {
		app = new(RdmUtil)
	}
	cfg := new(Config)
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
		return fmt.Errorf("invenio API URL or token not available")
	}
	return nil
}


// Query returns a byte slice for a JSON encode list
// of record summaries or an error.
//
// ```
//    app := new(irdmtools.RdmUtil)
//    if err := app.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    src, err := app.Query("My favorite book", -1, "newest")
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
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
//    app := new(irdmtools.RdmUtil)
//    if err := app.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    src, err := app.GetModifiedIds("2020-01-01", "2020-12-31")
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
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
//    app := new(irdmtools.RdmUtil)
//    if err := app.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    src, err := app.GetRecordIds()
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
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
//    app := new(irdmtools.RdmUtil)
//    if err := app.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    recordId := "woie-x0121"
//    src, err := app.GetRecord(recordId)
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
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

// Harvest takes a JSON file contianing a list of record ids and
// harvests them into a dataset v2 collection. The dataset collection
// must exist and be configured in either the environment or
// configuration file.
func (app *RdmUtil) Harvest(fName string) error {
	return Harvest(app.Cfg, fName)
}

// Run implements the irdmapp cli behaviors. With the exception of the
// "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//    app := new(irdmtools.RdmUtil)
//    if err := app.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    recordId := "wx0w-2231"
//    src, err := app.Run(os.Stdin, os.Stdout, os.Stderr, 
//                         "get_record", []string{recordId})
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
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
	case "get_record":
		if len(params) == 0 {
			return fmt.Errorf("missing record id")
		} else if len(params) > 1 {
			return fmt.Errorf("unexpected parameters, only expected on one record id")
		}
		src, err := app.GetRecord(params[0])
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
