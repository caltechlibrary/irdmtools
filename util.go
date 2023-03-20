package irdmtools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)


// Util holds the configuration for irdmutil cli.
type Util struct {
	Cfg *Config
}

// Configure reads the configuration file and environtment
// initialing the Cfg attribute of a Util object. It returns an error
// if problem were encounter.
//
// ```
//    util := new(irdmtools.Util)
//    if err := util.Configure("irdmtools.json", "TEST_"); err != nil {
//       // ... handle error ...
//    }
//    fmt.Printf("Invenio RDM API UTL: %q\n", util.Cfg.IvenioAPI)   
//    fmt.Printf("Invenio RDM token: %q\n", util.Cfg.InvenioToken)   
// ```
func (util *Util) Configure(configFName string, envPrefix string, debug bool) error {
	if util == nil {
		util = new(Util)
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
	util.Cfg = cfg
	if debug {
		util.Cfg.Debug = true
	}
	// Make sure we have a minimal useful configuration
	if util.Cfg.InvenioAPI == "" || util.Cfg.InvenioToken == "" {
		return fmt.Errorf("invenio API URL or token not available")
	}
	return nil
}

// SampleConfig display a minimal configuration for the irdmutil
// cli.  The minimal values in the configuration are "invenio_api"
// url and "invenio_token" holding the access token.
//
// ```
//    src, err := SampleConfig("irdmtools.json")
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
// ```
func SampleConfig(configFName string) ([]byte, error) {
	if configFName == "" {
		configFName = "irdmtools.json"
	}
	//fmt.Printf("DEBUG configFName %q\n", configFName)
	if _, err := os.Stat(configFName); err == nil {
		src, err := os.ReadFile(configFName)
		if err != nil {
			return nil, fmt.Errorf("%s already exists, failed to read file %s", configFName, err)
		}
		// NOTE: If we're reading the file from disk about copying the 
		// Invenio access token.
		if s := bytes.TrimSpace(src); len(s) > 0 {
			config := new(Config)
			err := json.Unmarshal(src, &config)
			if err != nil {
				return nil, err
			}
			config.InvenioToken = `__INVENIO_TOKEN_GOES_HERE__`
			src, err = json.MarshalIndent(config, "", "    ")
			return src, err
		}
	}
	invenioAPI := os.Getenv("RDM_INVENIO_API")
	if invenioAPI == "" {
		invenioAPI = "http://localhost:5000"
	}
	//invenioToken := os.Getenv("INVENIO_TOKEN")
	config := new(Config)
	// By default we look for Invenio-RDM as installed with
	// docker on localhost:5000
	config.InvenioAPI = invenioAPI
	config.InvenioToken = `__INVENIO_TOKEN_GOES_HERE__`
	src, err := json.MarshalIndent(config, "", "    ")
	return src, err
}


// Query returns a byte slice for a JSON encode list
// of record summaries or an error.
//
// ```
//    util := new(irdmtools.Util)
//    if err := util.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    src, err := util.Query("My favorite book", -1, "newest")
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
// ```
func (util *Util) Query(q string, sort string) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "DEBUG q %q, sort %q\n", q, sort)
	records, err := Query(util.Cfg, q, sort)
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
//    util := new(irdmtools.Util)
//    if err := util.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    src, err := util.GetModifiedIds("2020-01-01", "2020-12-31")
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
// ```
func (util *Util) GetModifiedIds(start string, end string) ([]byte, error) {
	ids, err := GetModifiedRecordIds(util.Cfg, start, end)
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
//    util := new(irdmtools.Util)
//    if err := util.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    src, err := util.GetRecordIds()
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
// ```
func (util *Util) GetRecordIds() ([]byte, error) {
	ids, err := GetRecordIds(util.Cfg)
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
//    util := new(irdmtools.Util)
//    if err := util.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    recordId := "woie-x0121"
//    src, err := util.GetRecord(recordId)
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
// ```
func (util *Util) GetRecord(id string) ([]byte, error) {
	rec, err := GetRecord(util.Cfg, id)
	if err != nil {
		return nil, err
	}
	src, err := json.MarshalIndent(rec, "", "    ")
	if err != nil {
		return nil, err
	}
	return src, nil
}

// Run implements the irdmutil cli behaviors. With the exception of the
// "setup" action you should call `util.LoadConfig()` before execute
// Run.
//
// ```
//    util := new(irdmtools.Util)
//    if err := util.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    recordId := "wx0w-2231"
//    src, err := util.Run(os.Stdin, os.Stdout, os.Stderr, 
//                         "get_record", []string{recordId})
//    if err != nil {
//        // ... handle error ...
//    }
//    fmt.Printf("%s\n", src)
// ```
func (util *Util) Run(in io.Reader, out io.Writer, eout io.Writer, action string, params []string) error {
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
		src, err := util.Query(q, sort)
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
		src, err := util.GetModifiedIds(start, end)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	case "get_all_ids":
		src, err := util.GetRecordIds()
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
		src, err := util.GetRecord(params[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "%s\n", bytes.TrimSpace(src))
		return nil
	default:
		return fmt.Errorf("%q action is not supported", action)
	}
}
