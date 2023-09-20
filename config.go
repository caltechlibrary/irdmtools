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
	"fmt"
	"os"

	// Caltech Library packages
	"github.com/caltechlibrary/dataset/v2"
)

// Config holds the common configuration used by all irdmtools
type Config struct {
	// Debug is set true then methods with access to the Config obect
	// can use this flag to implement addition logging to standard err
	Debug bool `json:"-"`
	// Repository Name, e.g. CaltechAUTHORS, CaltechTHESIS, CaltechDATA
	RepoName string `json:"repo_name,omitempty"`
	// InvenioAPI holds the URL to the InvenioAPI
	InvenioAPI string `json:"rdm_url,omitempty"`
	// InvenioToken is holds the token string to access the API
	InvenioToken string `json:"rdmtok,omitempty"`
	// Invenio DSN holds the data source name for the Postgres database storing the invenio records
	InvenioDSN string `json:"rdm_dsn,omitempty"`
	// InvenioStorage holds the URI to the default storage of Invenio RDM objects, e.g. local file system or S3 bucket
	InvenioStorage string `json:"rdm_storage,omitempty"`
	// InvenioCommunityID holds the community id for use with the API.
	InvenioCommunityID string `json:"rdm_community_id,omitempty"`
	// CName holds the dataset collection name used when harvesting content
	CName string `json:"c_name,omitempty"`
	// MailTo holds an email address to use when an email (e.g. CrossRef API access) is needed
	MailTo string `json:"mailto,omitempty"`
	// ds his a non-public point to an dataset collection structure
	ds *dataset.Collection

	// rl holds rate limiter data for throttling API requests
	rl *RateLimit
}

// prefixVar applies returns an variable name apply
// a prefix is one is given.
func prefixVar(varname string, prefix string) string {
	if prefix == "" {
		return varname
	}
	return fmt.Sprintf("%s%s", prefix, varname)
}

// NewConfig generates an empty configuration struct.
func NewConfig() *Config {
	cfg := new(Config)
	cfg.rl = new(RateLimit)
	return cfg
}

// LoadEnv checks the environment for configuration values if not
// previusly sets them. It will apply a prefix to the expected
// environment variable names if one is provided.
//
// ```
//
//	cfg := new(Config)
//	if err := cfg.LoadEnv("TEST_"); err != nil {
//	      // ... error handle ...
//	}
//
// ```
func (cfg *Config) LoadEnv(prefix string) error {
	if cfg == nil {
		cfg = NewConfig()
	}
	// Read in the configuration from the environment
	if api := os.Getenv(prefixVar("RDM_URL", prefix)); api != "" && cfg.InvenioAPI == "" {
		cfg.InvenioAPI = api
	}
	if token := os.Getenv(prefixVar("RDMTOK", prefix)); token != "" && cfg.InvenioToken == "" {
		cfg.InvenioToken = token
	}
	if cName := os.Getenv(prefixVar("C_NAME", prefix)); cName != "" && cfg.CName == "" {
		cfg.CName = cName
	}
	if mailTo := os.Getenv(prefixVar("MAILTO", prefix)); mailTo != "" && cfg.MailTo == "" {
		cfg.MailTo = mailTo
	}
	return nil
}

// LoadConfig reads the configuration file and initializes
// the attributes in the Config struct. It returns an error
// if problem were encounter. NOTE: It does NOT merge the
// settings in the environment.
//
// ```
//
//	cfg := NewConfig()
//	if err := cfg.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	fmt.Printf("Invenio RDM API UTL: %q\n", cfg.IvenioAPI)
//	fmt.Printf("Invenio RDM token: %q\n", cfg.InvenioToken)
//	fmt.Printf("Dataset Collection: %q\n", cfg.CName)
//	fmt.Printf("MailTo: %q\n", cfg.MailTo)
//
// ```
func (cfg *Config) LoadConfig(configFName string) error {
	if cfg == nil {
		cfg = NewConfig()
	}
	if configFName == "" {
		return fmt.Errorf("configuration filename is an empty string")
	}
	if _, err := os.Stat(configFName); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", configFName)
	}
	src, err := os.ReadFile(configFName)
	if err != nil {
		return err
	}
	if err := JSONUnmarshal(src, &cfg); err != nil {
		return err
	}
	return nil
}

// SampleConfig display a minimal configuration for the rdmutil
// cli.  The minimal values in the configuration are "invenio_api"
// url and "invenio_token" holding the access token.
//
// ```
//
//	src, err := SampleConfig("irdmtools.json")
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func SampleConfig(configFName string) ([]byte, error) {
	if configFName == "" {
		configFName = "irdmtools.json"
	}
	if _, err := os.Stat(configFName); err == nil {
		src, err := os.ReadFile(configFName)
		if err != nil {
			return nil, fmt.Errorf("%s already exists, failed to read file %s", configFName, err)
		}
		// NOTE: If we're reading the file from disk about copying the
		// Invenio access token.
		if s := bytes.TrimSpace(src); len(s) > 0 {
			config := new(Config)
			err := JSONUnmarshal(src, &config)
			if err != nil {
				return nil, err
			}
			config.InvenioToken = `__RDM_TOKEN_GOES_HERE__`
			src, err = JSONMarshalIndent(config, "", "    ")
			return src, err
		}
	}
	invenioAPI := os.Getenv("RDM_URL")
	if invenioAPI == "" {
		invenioAPI = "http://localhost:5000"
	}
	//invenioToken := os.Getenv("RDMTOK")
	cName := os.Getenv("C_NAME")
	if cName == "" {
		cName = "__DATASET_COLLECTION_NAME_GOES_HERE__"
	}

	mailTo := os.Getenv("RDM_MAILTO")
	if mailTo == "" {
		mailTo = "__CROSSREF_API_MAILTO_GOES_HERE__"
	}
	config := new(Config)
	// By default we look for Invenio-RDM as installed with
	// docker on localhost:5000
	config.InvenioAPI = invenioAPI
	config.InvenioToken = `__RDM_TOKEN_GOES_HERE__`
	config.CName = cName
	config.MailTo = mailTo
	src, err := JSONMarshalIndent(config, "", "    ")
	return src, err
}
