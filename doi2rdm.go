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
	"encoding/json"
	"fmt"
	"io"
)

// Doi2Rdm holds the configuration for doi2rdm cli.
type Doi2Rdm struct {
	Cfg *Config
}

// Configure reads the configuration file and environtment
// initialing the Cfg attribute of a Doi2Rdm object. It returns an error
// if problem were encounter.
//
// ```
//
//	app := new(irdmtools.Doi2Rdm)
//	if err := app.Configure("irdmtools.json", "TEST_"); err != nil {
//	   // ... handle error ...
//	}
//	fmt.Printf("Invenio RDM API UTL: %q\n", app.Cfg.IvenioAPI)
//	fmt.Printf("Invenio RDM token: %q\n", app.Cfg.InvenioToken)
//
// ```
func (app *Doi2Rdm) Configure(configFName string, envPrefix string, debug bool) error {
	if app == nil {
		app = new(Doi2Rdm)
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

// Run implements the doi2rdm cli behaviors. With the exception of the
// "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//	app := new(irdmtools.Doi2Rdm)
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
func (app *Doi2Rdm) Run(in io.Reader, out io.Writer, eout io.Writer, options map[string]string, doi string) error {
	queryCrossRef, queryDataCite, debug := true, true, false
	dotInitials, downloadDocument := false, false
	mailTo, diffFName := "", ""
	for opt, val := range options {
		switch opt {
		case "diff":
			diffFName = val
		case "crossref_only":
			queryCrossRef = true
			queryDataCite = false
		case "datacite_only":
			queryCrossRef = false
			queryDataCite = true
		case "dot-initials":
			dotInitials = true
		case "download":
			downloadDocument = true
		case "mailto":
			mailTo = val
		case "debug":
			debug = true
		default:
			return fmt.Errorf("%q parameter is not supported", opt)
		}
	}

	var (
		oRecord *simplified.Record
		nRecord *simplified.Record
	)
	if queryCrossRef {
		if diffFName != "" {
			oWork := new(crossrefapi.Works)
			src, err := os.ReadFile(diffFName)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(src, &oWork); err != nil {
				return err
			}
			oRecord, err = CrosswalkCrosRef(app.Cfg, oWork, debug)
			if err != nil {
				return err
			}
		}
		nWork, err := QueryCrossRef(app.Cfg, doi, mailTo, dotInitials, downloadDocument, debug)
		if err != nil {
			return err
		}
		nRecord, err := CrosswalkCrossRef(app.Cfg, work, debug)
		if err != nil {
			return err
		}
		var src []byte
		if diffFName != "" {
			src, err = oRecord.DiffAsJSON(nRecord)
		} else {
			src, err = json.MarshalIndent(nRecord, "", "    ")
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", src)
		return nil
	}
	if queryDataCite {
		obj, err := QueryDataCite(app.Cfg, doi, mailTo, dotInitials, downloadDocument, debug)
		if err != nil {
			return err
		}
		nRecord, err = CrosswalkDataCite(app.Cfg, obj, debug)
		if err != nil {
			return err
		}
		src, err := json.MarshalIndent(nRecord, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", src)
		return nil
	}
	return fmt.Errorf("%s not found", doi)
}
