 package irdmtools

 import (
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
