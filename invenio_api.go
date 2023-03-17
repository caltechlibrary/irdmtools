// invenio_api.go implements method to retreive data from an 
// Invenio RDM instance using the JSON API provided by Invenio.
package irdmtools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"net/http"
	"os"

	// Caltech Library Packages
//	"github.com/caltechlibrary/simplified"
)

// QueryResponse holds the response to /api/records?q=...
type QueryResponse struct {
	// 
	Hits *Hits `json:"hits,omitepmty"`
	Links *Links `json:"links,omitempty"`
	SortBy string `json:"sortBy,omitempty"`
}

type Hits struct {
	Hits []map[string]interface{} `json:"hits,omitempty"`
	Total int `json:"total,omitempty"`
}

type Links struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
	Prev string `json:"prev,omitempty"`
}

// getJSON sends a request to the InvenioAPI using
// a token, url and values as parameters. It return a 
// JSON encoded response as byte slice
func getJSON(token string, uri string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	src, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return src, nil
}

// GetRecordIds takes a configuration object, contacts am RDM
// instance and returns a list of ids and error.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
func GetRecordIds(cfg *Config) ([]string, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup our query parameters, i.e. q=*
	uri := fmt.Sprintf("%s/api/records?q=*&size=250", u.String())

	src, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	tot := 0
	results := new(QueryResponse)
	// NOTE: Need to unparse the response structure and 
	// then extract the IDs from the individual Hits results
	if err := json.Unmarshal(src, &results); err != nil {
		return nil, err
	}
	if results != nil && results.Hits != nil  && 
		results.Hits.Hits != nil && len(results.Hits.Hits) > 0{
		for _, hit := range results.Hits.Hits {
			if val, ok := hit["id"]; ok {
				id := val.(string)
				ids = append(ids, id)
			}
		}
		tot = results.Hits.Total		
		fmt.Fprintf(os.Stderr, "Total: %d\n", tot)
	}
	// FIXME need to handle paged response from Invenio
	if results.Links != nil {
		for results.Links != nil && results.Links.Next != "" {
			src, err := getJSON(cfg.InvenioToken, results.Links.Next)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(src, &results); err != nil {
				return nil, err
			}
			if results != nil && results.Hits != nil  && 
					results.Hits.Hits != nil && len(results.Hits.Hits) > 0 {
				for _, hit := range results.Hits.Hits {
					if val, ok := hit["id"]; ok {
						id := val.(string)
						ids = append(ids, id)
					}
				}
			}
			fmt.Fprintf(os.Stderr, "self %s\n", results.Links.Self)
			fmt.Fprintf(os.Stderr, "next %s\n", results.Links.Next)
		}
	}

	//fmt.Printf("%s\n", src)
	return ids, nil
}

// GetRecord takes a configuration object and record id, 
// contacts an RDM instance and returns a simplified record
// and error.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
func GetRecord(cfg *Config, id string) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s", u.String(), id)

	// FIXME need to handle paged response from Invenio
	src, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	rec := map[string]interface{}{}
	if err := json.Unmarshal(src, &rec); err != nil {
		return nil, err
	}
	return rec, nil
}
