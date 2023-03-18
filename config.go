 package irdmtools

 import (
 	"encoding/json"
 	"fmt"
 	"os"

	// Caltech Library packages
 	"github.com/caltechlibrary/dataset/v2"
 )

// Config holds the common configuration used by all irdmtools
type Config struct {
	 // Repository Name, e.g. CaltechAUTHORS, CaltechTHESIS, CaltechDATA
	 RepoName string `json:"repo_name,omitempty"`
	 // InvenioAPI holds the URL to the InvenioAPI
	 InvenioAPI string `json:"invenio_api,omitempty"`
	 // InvenioToken is holds the token string to access the API
	 InvenioToken string `json:"invenio_token,omitempty"`
	 // Invenio DSN holds the data source name for the Postgres database storing the invenio records
	 InvenioDSN string `json:"invenio_dsn,omitempty"`
	 // InvenioStorage holds the URI to the default storage of Invenio RDM objects, e.g. local file system or S3 bucket
	 InvenioStorage string `json:"invenio_storage,omitempty"`
	 // CName holds the dataset collection name used when harvesting content
	 CName string `json:"c_name,omitempty"`
	 // ds his a non-public point to an dataset collection structure
	 ds *dataset.Collection
 }

// prefixVar applies returns an variable name apply
// a prefix is one is given.
func prefixVar(varname string, prefix string) string {
	if prefix == "" {
		return varname
	}
	return fmt.Sprintf("%s%s", prefix, varname)
}

// LoadEnv checks the environment for configuration values if not 
// previusly sets them. It will apply a prefix to the expected 
// environment variable names if one is provided.
//
// ```
//     cfg := new(Config)
//     if err := cfg.LoadEnv("TEST_"); err != nil {
//           // ... error handle ...
//     }
// ```
func (cfg *Config) LoadEnv(prefix string) error {
	if cfg == nil {
		cfg = new(Config)
	}
	// Read in the configuration from the environment
	if api := os.Getenv(prefixVar("INVENIO_API", prefix)); api != "" && cfg.InvenioAPI == "" {
		cfg.InvenioAPI = api
	}
	if token := os.Getenv(prefixVar("INVENIO_TOKEN", prefix)); token != "" && cfg.InvenioToken == "" {
		cfg.InvenioToken = token
	}
	if cName := os.Getenv(prefixVar("COLLECTION_NAME", prefix)); cName != "" && cfg.CName == "" {
		cfg.CName = cName
	}
	return nil
}

// LoadConfig reads the configuration file and initializes
// the attributes in the Config struct. It returns an error
// if problem were encounter. NOTE: It does NOT merge the
// settings in the environment.
//
// ```
//    cfg := new(Config)
//    if err := cfg.LoadConfig("irdmtools.json"); err != nil {
//       // ... handle error ...
//    }
//    fmt.Printf("Invenio RDM API UTL: %q\n", cfg.IvenioAPI)   
//    fmt.Printf("Invenio RDM token: %q\n", cfg.InvenioToken)   
// ```
func (cfg *Config) LoadConfig(configFName string) error {
	if cfg == nil {
		cfg = new(Config)
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
	if err := json.Unmarshal(src, &cfg); err != nil {
		return err
	}
	return nil
}

