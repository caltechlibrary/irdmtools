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
	"flag"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Ep3Util holds the configuration for ep3util cli.
type Ep3Util struct {
	Cfg *Config
	Debug bool
}

// Configure reads the configuration file and environtment
// initialing the Cfg attribute of a Ep3Util object. It returns an error
// if problem were encounter.
//
// ```
//
//	app := new(irdmtools.Ep3Util)
//	if err := app.Configure("irdmtools.json", "TEST_"); err != nil {
//	   // ... handle error ...
//	}
//	fmt.Printf("Invenio RDM API UTL: %q\n", app.Cfg.IvenioAPI)
//	fmt.Printf("Invenio RDM token: %q\n", app.Cfg.InvenioToken)
//
// ```
func (app *Ep3Util) Configure(configFName string, envPrefix string, debug bool) error {
	if app == nil {
		app = new(Ep3Util)
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
	if app.Cfg.EPrintHost == "" {
		return fmt.Errorf("EPRINT_HOST, EPRINT_USER, EPRINT_PASSWORD are not available")
	}
	return nil
}

// GetRecordIds returns a byte slice for a JSON encode list
// of record ids or an error based on the records listed in the EPrints.
//
// ```
//
//	app := new(irdmtools.Ep3Util)
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
func (app *Ep3Util) GetRecordIds() ([]byte, error) {
	timeout := time.Duration(timeoutSeconds)
	ids, err := GetKeys(app.Cfg, timeout, 3)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(ids, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetModifiedRecordIds returns a byte slice for a JSON encode list
// of record ids or an error based on the records listed in EPrints.
//
// ```
//
//	app := new(irdmtools.Ep3Util)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	src, err := app.GetModifiedRecordIds("2023-09-01", "2023-09-30")
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *Ep3Util) GetModifiedRecordIds(start string, end string) ([]byte, error) {
	ids, err := GetModifiedKeys(app.Cfg, start, end)
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
//	app := new(irdmtools.Ep3Util)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...

// GetRecord returns a byte slice for a JSON encoded record
// or an error.
//
// ```
//
//	app := new(irdmtools.Ep3Util)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	recordId := "23808"
//	src, err := app.GetRecord(recordId)
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *Ep3Util) GetRecord(id string) ([]byte, error) {
	eprintid, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(timeoutSeconds)
	rec, err := GetEPrint(app.Cfg, eprintid, timeout, 3)
	if err != nil {
		return nil, err
	}
	src, err := JSONMarshalIndent(rec, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// Harvest takes a JSON file contianing a list of record ids and
// harvests them into a dataset v2 collection. The dataset collection
// must exist and be configured in either the environment or
// configuration file.
func (app *Ep3Util) RunHarvest(in io.Reader, out io.Writer, eout io.Writer, all bool, modified bool, params []string) error {
	switch {
	case all:
		timeout := time.Duration(timeoutSeconds)
		ids, err := GetKeys(app.Cfg, timeout, 3)
		if err != nil {
			return err
		}
		return HarvestEPrintRecords(app.Cfg, ids, app.Cfg.Debug)	
	case modified:
		// FIXME: need to harvest modified eprints ...
		today := time.Now().Format("2006-01-02")
		start, end := today, today
		if len(params) < 1 {
			return fmt.Errorf("missing start and end date")
		}
		start = params[0]
		if len(params) > 1 {
			end = params[1]
		}
		ids, err := GetModifiedKeys(app.Cfg, start, end)
		if err != nil {
			return err
		}
		return HarvestEPrintRecords(app.Cfg, ids, app.Cfg.Debug)	
	default:
		return HarvestEPrints(app.Cfg, params[0], app.Cfg.Debug)
	}
}

// Run implements the irdmapp cli behaviors. With the exception of the
// Run.
//
// ```
//
//	app := new(irdmtools.Ep3Util)
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	recordId := "23808"
//	src, err := app.Run(os.Stdin, os.Stdout, os.Stderr,
//	                     "get_record", []string{recordId})
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *Ep3Util) Run(in io.Reader, out io.Writer, eout io.Writer, action string, params []string) error {
	var (
		src []byte
		err error
		recordId string
	)
	switch action {
	case "setup":
		if len(params) == 0 {
			return fmt.Errorf("missing configuration name")
		}
		src, err = SampleConfig(params[0])
	case "get_all_ids":
		src, err = app.GetRecordIds()
	case "get_modified_ids":
		today := time.Now().Format("2006-01-02")
		start, end := today, today
		if len(params) < 1 {
			return fmt.Errorf("missing a start and end date")
		}
		if len(params) > 0 {
			start = params[0]
		}
		if len(params) > 1 {
			end = params[1]
		}
		src, err = app.GetModifiedRecordIds(start, end)
	case "get_record":
		recordId, _, _, err = getRecordParams(params, true, false, false)
		if err != nil {
			return err
		}
		src, err = app.GetRecord(recordId)
	case "harvest":
		all, modified := false, false
		flagSet := flag.NewFlagSet("harvest", flag.ContinueOnError)
		flagSet.BoolVar(&all, "all", all, "harvest all records")
		flagSet.BoolVar(&modified, "modified", modified, "harvest records between start and optional end date")
		flagSet.Parse(params)
		params = flagSet.Args()
		if (! all) && len(params) < 1 {
			if modified {
				return fmt.Errorf("Missing a start and optional end date for harvest")
			}
			return fmt.Errorf("JSON Identifier file required")
		}
		return app.RunHarvest(in, out, eout, all, modified, params)
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
