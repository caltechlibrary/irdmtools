// invenio_api.go implements method to retreive data from an
// Invenio RDM instance using the JSON API provided by Invenio.
package irdmtools

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	// Caltech Library Packages
	//	"github.com/caltechlibrary/simplified"
)

const (
	// pageSize sets the number of responses when accessing the
	// Invenio JSON API.
	pageSize = 250
)

// OAIListIdendifiersResponse
type OAIListIdentifiersResponse struct {
	XMLName         xml.Name            `xml:"OAI-PMH" json:"-"`
	XMLNS           string              `xml:"xmlns,attr,omitempty" json:"xmlns,omitempty"`
	ResponseDate    string              `xml:"responseDate,omitempty" json:"response_date,omitempty"`
	Request         string              `xml:"request,omitempty" json:"request,omitempty"`
	RequestAttr     map[string]string   `xml:"request,attr,omitempty" json:"request_attr,omitempty"`
	ListIdentifiers *OAIListIdentifiers `xml:"ListIdentifiers,omitempty" json:"list_identifiers,omitempty"`
}

type OAIListIdentifiers struct {
	Headers         []OAIHeader `xml:"header,omitempty" json:"header,omitempty"`
	ResumptionToken string      `xml:"resumptionToken,omitempty" json:"resumption_token,omitempty"`
}

// OAIHeader holds the response items for
type OAIHeader struct {
	Status     string   `xml:"status,attr,omitempty" json:"status,omitempty"`
	Identifier string   `xml:"identifier,omitempty" json:"identifier,omitempty"`
	DateStamp  string   `xml:"datestamp,omitempty" json:"datestamp,omitempty"`
	SetSpec    []string `xml:"setSpec,omitempty" json:"set_spec,omitempty"`
}

// QueryResponse holds the response to /api/records?q=...
type QueryResponse struct {
	//
	Hits   *Hits  `json:"hits,omitepmty"`
	Links  *Links `json:"links,omitempty"`
	SortBy string `json:"sortBy,omitempty"`
}

type Hits struct {
	Hits  []map[string]interface{} `json:"hits,omitempty"`
	Total int                      `json:"total,omitempty"`
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
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", resp.Status)
	}
	src, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return src, nil
}

// getXML sends a request to the Invenio API (e.g. OAI-PMH) using
// a token, url and values as parameters. It returns an
// xml encoded response as byte slice
func getXML(token string, uri string) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-type", "application/xml")
	resp, err := client.Do(req)
	if err != nil {
		return nil, resp.Header, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, resp.Header, fmt.Errorf("%s %s", resp.Status, uri)
	} 
	src, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return src, resp.Header, nil
}

// Query takes a query string and returns the paged object
// results as a slice of map[string]interface{}
//
// ```
//
//		   records, err := Query(cfg, "Geological History in Southern California", 250, "updated")
//	    if err != nil {
//	        // ... handle error ...
//	    }
//	    for _, rec := ranges {
//	        // ... process results ...
//	    }
//
// ```
func Query(cfg *Config, q string, size int, sortBy string) ([]map[string]interface{}, error) {
	if size == 0 {
		size = 25
	}
	if sortBy == "" {
		sortBy = "updated"
	}
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup our query parameters, i.e. q=*
	uri := fmt.Sprintf("%s/api/records?size=%d&sort=%s&q=%s", u.String(), size, sortBy, q)

	src, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	records := []map[string]interface{}{}
	tot := 0
	maxPages := 0
	results := new(QueryResponse)
	// NOTE: Need to unparse the response structure and
	// then extract the IDs from the individual Hits results
	if err := json.Unmarshal(src, &results); err != nil {
		return nil, err
	}
	if results != nil && results.Hits != nil &&
		results.Hits.Hits != nil && len(results.Hits.Hits) > 0 {
		for _, hit := range results.Hits.Hits {
			records = append(records, hit)
		}
		tot = results.Hits.Total
		maxPages = int(tot/size) + 1
		fmt.Fprintf(os.Stderr, "Total: %d, Pages %d, Query: %q\n", tot, maxPages, q)
	}
	// FIXME need to handle paged response from Invenio
	if results.Links != nil {
		for results.Links != nil &&
			results.Links.Self != results.Links.Next {
			src, err := getJSON(cfg.InvenioToken, results.Links.Next)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(src, &results); err != nil {
				return nil, err
			}
			if results != nil && results.Hits != nil &&
				results.Hits.Hits != nil && len(results.Hits.Hits) > 0 {
				for _, hit := range results.Hits.Hits {
					records = append(records, hit)
				}
			}
			fmt.Fprintf(os.Stderr, "prev %+v\n", results.Links.Prev)
			fmt.Fprintf(os.Stderr, "self %s\n", results.Links.Self)
			fmt.Fprintf(os.Stderr, "next %s\n", results.Links.Next)
		}
	}
	return records, nil
}

// GetRecordIds takes a configuration object, contacts am RDM
// instance and returns a list of ids and error.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// FIXME: The maxim result set from the API with paging in 10K items,
// this is too small for this function's purpose of returning all ids.
//
// The correct way to do this without API support is to directly
// query the PostgreSQL for the ids. Our production instances have
// PostgreSQL running in a container. That makes that appoach problematic
// inless we figure out way to replicate that instance. Another
// approach is ask the Invenio RDM developer core to provide an end point
// for returning a list of all keys.
//
// NOTE: Using the records API you get back a "bestmatch" result set.
// What are the other options (e.g. can we get results returned by
// in timestamp order)
//
// Using another end point, e.g. OAI-PMH might also solve problem
func GetRecordIds(cfg *Config) ([]string, error) {
	ids := []string{}
	resumptionToken := "     "
	uri := fmt.Sprintf("%s/oai2d?verb=ListIdentifiers&metadataPrefix=oai_dc", cfg.InvenioAPI)
	for i := 0; resumptionToken != ""; i++ {
		if i > 0 {	
			v := url.Values{}
			v.Set("resumptionToken", resumptionToken)
			uri = fmt.Sprintf("%s/oai2d?verb=ListIdentifiers&%s", cfg.InvenioAPI, v.Encode())
		}
		src, headers, err := getXML(cfg.InvenioToken, uri)
		xRateLimitLimit := headers.Values("X-RateLimit-Limit")
		xRateLimitRemaining := headers.Values("X-RateLimit-Remaining")
		if err != nil {
			return nil, err
		}

		rateLimit, err := strconv.Atoi(xRateLimitLimit[0])
		if err != nil {
			rateLimit = 60
		}
		remaining, err := strconv.Atoi(xRateLimitRemaining[0])
		if err != nil {
			remaining = 0
		}
		secondsToWait := time.Duration(int(rateLimit/60)) * time.Second

		// Pause to respect rate limits
		//FIXME: Need to calculate when I can request more ids
		if float64(remaining)/float64(rateLimit) <= 0.5 {
			fmt.Fprintf(os.Stderr, "DEBUG waiting (%d/%d), secondsToWait %s\n", remaining,rateLimit, secondsToWait)
   			time.Sleep(secondsToWait)
		} 
		//os.WriteFile("oai-pmh-list-identifiers.xml", src, 0660) // DEBUG
		if bytes.HasPrefix(src, []byte("<html")) {
			os.WriteFile("oai-pmh-error.html", src, 0660) // DEBUG
			resumptionToken = ""
		} else {
    		oai := new(OAIListIdentifiersResponse)
    		if err := xml.Unmarshal(src, oai); err != nil {
				os.WriteFile("oai-pmh-error.html", src, 0660) // DEBUG
				resumptionToken = ""
    			return nil, err
    		}
    		if oai.ListIdentifiers != nil {
    			resumptionToken = oai.ListIdentifiers.ResumptionToken
    			//fmt.Fprintf(os.Stderr, "DEBUG (%d) resumptionToken=%s\n", i, resumptionToken)
    			if oai.ListIdentifiers.Headers != nil {
    				for _, obj := range oai.ListIdentifiers.Headers {
    					if obj.Identifier != "" {
    						parts := strings.Split(obj.Identifier, ":")
    						last := len(parts) - 1
    						if last < 0 {
    							last = 0
    						}
    						id := parts[len(parts)-1]
    						//fmt.Fprintf(os.Stderr, "DEBUG id %q\n", id)
    						ids = append(ids, id)
    					}
    				}
    			}
    		} else {
    			resumptionToken = ""
    		}
		}
		if (len(ids) % 500) == 0 {
			fmt.Fprintf(os.Stderr, "%d ids harvested\n", len(ids))// DEBUG
		}
	}
	fmt.Fprintf(os.Stderr, "%d ids harvested (total)\n", len(ids))// DEBUG
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
