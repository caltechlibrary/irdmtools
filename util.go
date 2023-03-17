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

// LoadConfig reads the configuration file and initializes
// the Cfg attribute of a Util object. It returns an error
// if problem were encounter.
//
// ```
//    util := new(irdmtools.Util)
//    if err := util.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    fmt.Printf("Invenio RDM API UTL: %q\n", util.Cfg.IvenioAPI)   
//    fmt.Printf("Invenio RDM token: %q\n", util.Cfg.InvenioToken)   
// ```
func (util *Util) LoadConfig(configFName string) error {
	if util.Cfg == nil {
		util.Cfg = new(Config)
	}
	// Read in the configuration from the environment
	api := os.Getenv("INVENIO_API")
	token := os.Getenv("INVENIO_TOKEN")
	if api != "" && token != "" {
		util.Cfg.InvenioAPI = api
		util.Cfg.InvenioToken = token
	}
	if configFName != "" {
		if _, err := os.Stat(configFName); os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist", configFName)
		}
		src, err := os.ReadFile(configFName)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(src, &util.Cfg); err != nil {
			return err
		}
	}
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
	invenioAPI := os.Getenv("INVENIO_API")
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
	case "get_record_ids":
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
