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
	"database/sql"
	"strings"

	// Caltech Library packages
	"github.com/caltechlibrary/dataset/v2"

	// 3rd Party Libraries
	"gopkg.in/yaml.v3"
	"github.com/joho/godotenv"
)

// Config holds the common configuration used by all irdmtools
type Config struct {
	// Debug is set true then methods with access to the Config obect
	// can use this flag to implement addition logging to standard err
	Debug bool `json:"-", yaml:"-"`
	// Repository Name, e.g. CaltechAUTHORS, CaltechTHESIS, CaltechDATA
	RepoName string `json:"repo_name,omitempty" yaml:"repo_name,omitempty"`
	// Repository ID, e.g. caltechauthors, caltechthesis, caltechdata (usually the db name for repository)
	// NOTE: It should also match the Postgres DB name used by RDM
	RepoID string `json:"repo_id,omitempty" yaml:"repo_id,omitempty"`
	// InvenioAPI holds the URL to the InvenioAPI
	InvenioAPI string `json:"rdm_url,omitempty" yaml:"rdm_url,omitempty"`
	// InvenioToken is holds the token string to access the API
	InvenioToken string `json:"rdmtok,omitempty" yaml:"rdmtok,omitempty"`
	// Invenio DSN holds the data source name for the Postgres database storing the invenio records
	InvenioDSN string `json:"rdm_dsn,omitempty" yaml:"rdm_dsn,omitempty"`
	// InvenioStorage holds the URI to the default storage of Invenio RDM objects, e.g. local file system or S3 bucket
	InvenioStorage string `json:"rdm_storage,omitempty" yaml:"rdm_storage,omitempty"`
	// InvenioCommunityID holds the community id for use with the API.
	InvenioCommunityID string `json:"rdm_community_id,omitempty" yaml:"rdm_community_id,omitempty"`
	// InvenioDbHost holds the name of the machine for the Postgres server
	InvenioDbHost string `json:"rdm_db_host,omitempty" yaml:"rdm_db_host,omitempty"`
	// InvenioDbUser holds the database username of the machine for the Postgres server
	InvenioDbUser string `json:"rdm_db_user,omitempty" yaml:"rdm_db_user,omitempty"`
	// InvenioDbPassword holds the database password of the machine for the Postgres server
	InvenioDbPassword string `json:"rdm_db_password,omitempty" yaml:"rdm_db_password,omitempty"`

	// CName holds the dataset collection name used when harvesting content
	CName string `json:"c_name,omitempty" yaml:"c_name,omitempty"`
	// MailTo holds an email address to use when an email (e.g. CrossRef API access) is needed
	MailTo string `json:"mailto,omitempty" yaml:"mailto,omitempty"`
	// ds his a non-public point to an dataset collection structure
	ds *dataset.Collection

	// EPrint configuration needed for migration related tools
	EPrintHost string `json:"eprint_host,omitempty" yaml:"eprint_host,omitempty"`
	EPrintUser string `json:"eprint_user,omitempty" yaml:"eprint_user,omitempty"`
	EPrintPassword string `json:"eprint_password,omitempty" yaml:"eprint_password,omitempty"`
	EPrintArchivesPath string `json:"eprint_archives_path,omitempty" yaml:"eprint_archives_path,omitempty"` 
	EPrintDbHost string `json:"eprint_db_host,omitempty" yaml:"eprint_db_host,omitempty"`
	EPrintDbUser string `json:"eprint_db_user,omitempty" yaml:"eprint_db_user,omitempty"`
	EPrintDbPassword string `json:"eprint_db_password,omitempty" yaml:"eprint_db_password,omitempty"`
	EPrintBaseURL string`json:"eprint_base_url,omitempty" yaml:"eprint_base_url,omitempty"`


	// rl holds rate limiter data for throttling API requests
	rl *RateLimit

	// pgDb holds a Postgres connection
	pgDB *sql.DB
	myDB *sql.DB
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

// MakeDSN will return the value set for cfg.InvenioDSN or set and return it if
// enough data is provided in the config.
func (cfg *Config) MakeDSN() string {
	if cfg.InvenioDSN == "" {
   		parts := []string{
   			"postgres://",
   		}
   		username := []string{}
   		if cfg.InvenioDbUser != "" {
   			username = append(username, cfg.InvenioDbUser)
   		}
   		if cfg.InvenioDbPassword != "" {
   			username = append(username, cfg.InvenioDbPassword)
   		}
   		if len(username) > 0 {
   			parts = append(parts, strings.Join(username, ":") + "@")
   		} else {
   			parts = append(parts, "")
   		}
   		if cfg.InvenioDbHost != "" {
   			parts = append(parts, cfg.InvenioDbHost)
   		}
   		if cfg.RepoID != "" {
   			parts = append(parts, "/" + cfg.RepoID)
   		}
   		if strings.HasPrefix(cfg.InvenioDbHost, "localhost") {
   			parts = append(parts, "?sslmode=disable")
   		} else {
   			parts = append(parts, "?sslmode=require")
   		}
   		if len(parts) > 1 {
   			return strings.Join(parts, "")
   		}
	}
	return cfg.InvenioDSN
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
	if err := godotenv.Load(); err != nil {
		fmt.Fprint(os.Stderr, "WARNING: failed to find or read .env file, %s\n", err)
	}

	if repoID := os.Getenv(prefixVar("REPO_ID", prefix)); repoID != "" {
		cfg.RepoID = repoID
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
	if eprintHost := os.Getenv(prefixVar("EPRINT_HOST", prefix)); eprintHost != "" && cfg.EPrintHost == "" {
		cfg.EPrintHost = eprintHost
	}
	if eprintUser := os.Getenv(prefixVar("EPRINT_USER", prefix)); eprintUser != "" && cfg.EPrintUser == "" {
		cfg.EPrintUser = eprintUser
	}
	if eprintPassword := os.Getenv(prefixVar("EPRINT_PASSWORD", prefix)); eprintPassword != "" && cfg.EPrintPassword == "" {
		cfg.EPrintPassword = eprintPassword
	}
	if eprintArchivesPath := os.Getenv(prefixVar("EPRINT_ARCHIVES_PATH", prefix)); eprintArchivesPath != "" && cfg.EPrintArchivesPath == "" {
		cfg.EPrintArchivesPath = eprintArchivesPath
	}
	if eprintDbHost := os.Getenv(prefixVar("EPRINT_DB_HOST", prefix)); eprintDbHost != "" {
		cfg.EPrintDbHost = eprintDbHost
	}
	if eprintDbUser := os.Getenv(prefixVar("EPRINT_DB_USER", prefix)); eprintDbUser != "" {
		cfg.EPrintDbUser = eprintDbUser
	}
	if eprintDbPassword := os.Getenv(prefixVar("EPRINT_DB_PASSWORD", prefix)); eprintDbPassword != "" {
		cfg.EPrintDbPassword = eprintDbPassword
	}
	if rdmDbHost := os.Getenv(prefixVar("RDM_DB_HOST", prefix)); rdmDbHost != "" {
		cfg.InvenioDbHost = rdmDbHost
	}
	if rdmDbUser := os.Getenv(prefixVar("RDM_DB_USER", prefix)); rdmDbUser != "" {
		cfg.InvenioDbUser = rdmDbUser
	}
	if rdmDbPassword := os.Getenv(prefixVar("RDM_DB_PASSWORD", prefix)); rdmDbPassword != "" {
		cfg.InvenioDbPassword = rdmDbPassword
	}
	// Build our InvenioDSN
	if rdmDSN := os.Getenv(prefixVar("RDM_DSN", prefix)); rdmDSN != "" {
		cfg.InvenioDSN = rdmDSN
	} else {
		cfg.InvenioDSN = cfg.MakeDSN()
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
	if err := godotenv.Load(); err != nil {
		fmt.Fprint(os.Stderr, "WARNING: failed to find or read .env file, %s\n", err)
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
	if strings.HasSuffix(configFName, ".yaml") {
		if err := yaml.Unmarshal(src, &cfg); err != nil {
			return err
		}
	} else {
		if err := JSONUnmarshal(src, &cfg); err != nil {
			return err
		}
	}
	// Build our DSN if not set.
	if cfg.InvenioDSN == "" {
		cfg.InvenioDSN = cfg.MakeDSN()
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
		// NOTE: If we're reading the file from disk avoid copying the
		// Invenio access token or EPrint user password.
		if s := bytes.TrimSpace(src); len(s) > 0 {
			config := new(Config)
			err := JSONUnmarshal(src, &config)
			if err != nil {
				return nil, err
			}
			config.InvenioToken = `__RDM_TOKEN_GOES_HERE__`
			config.EPrintPassword = `__EPRINT_USER_PASSWORD_HERE__`
			src, err = JSONMarshalIndent(config, "", "    ")
			return src, err
		}
	}
	repoID := os.Getenv("REPO_ID")
	if repoID == "" {
		repoID = "__REPO_ID__GOES_HERE__"
	}
	invenioAPI := os.Getenv("RDM_URL")
	if invenioAPI == "" {
		invenioAPI = "http://localhost:5000"
	}
	invenioDbHost := os.Getenv("RDM_DB_HOST")
	if invenioDbHost == "" {
		invenioDbHost = "__RDM_DB_HOST__GOES_HERE__"
	}
	invenioDbUser := os.Getenv("RDM_DB_USER")
	if invenioDbUser == "" {
		invenioDbUser = "__RDM_DB_USER__GOES_HERE__"
	}

	cName := os.Getenv("C_NAME")
	if cName == "" {
		cName = "__DATASET_COLLECTION_NAME_GOES_HERE__"
	}

	mailTo := os.Getenv("RDM_MAILTO")
	if mailTo == "" {
		mailTo = "__CROSSREF_API_MAILTO__GOES_HERE__"
	}

	eprintHost := os.Getenv("EPRINT_HOST")
	if eprintHost == "" {
		eprintHost = "__EPRINT_HOSTNAME__GOES_HERE__"
	}
	eprintUser := os.Getenv("EPRINT_USER")
	if eprintUser == "" {
		eprintUser = "__EPRINT_USER__GOES_HERE__"
	}
	eprintArchivesPath := os.Getenv("EPRINT_ARCHIVES_PATH")
	if eprintArchivesPath == "" {
		eprintArchivesPath = "__EPRINT_ARCHIVES_PATH__GOES_HERE__"
	}
	eprintDbHost := os.Getenv("EPRINT_DB_HOST")
	if eprintDbHost == "" {
		eprintDbHost = "__EPRINT_DB_HOST__GOES_HERE__"
	}
	eprintDbUser := os.Getenv("EPRINT_DB_USER")
	if eprintDbUser == "" {
		eprintDbUser = "__EPRINT_DB_USER__GOES_HERE__"
	}

	config := new(Config)
	// By default we look for Invenio-RDM as installed with
	// docker on localhost:5000
	config.RepoID = repoID
	config.InvenioAPI = invenioAPI
	config.InvenioToken = `__RDM_TOKEN__GOES_HERE__`
	config.InvenioDbHost = invenioDbHost
	config.InvenioDbUser = invenioDbUser
	config.InvenioDbPassword = `__RDM_DB_PASSWORD__GOES_HERE__`
	config.CName = cName
	config.MailTo = mailTo
	config.EPrintHost = eprintHost
	config.EPrintUser = eprintUser
	config.EPrintPassword = `__EPRINT_PASSWORD__GOES_HERE__`
	config.EPrintArchivesPath = eprintArchivesPath
	config.EPrintDbHost = eprintDbHost
	config.EPrintDbUser = eprintDbUser
	config.EPrintDbPassword = `__EPRINT_DB_PASSWORD__GOES_HERE__`
	src, err := JSONMarshalIndent(config, "", "    ")
	return src, err
}
