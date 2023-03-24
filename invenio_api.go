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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

func dbgPrintf(cfg *Config, s string, args ...interface{}) {
	if cfg.Debug {
		if strings.HasSuffix(s, "\n") {
			fmt.Fprintf(os.Stderr, s, args...)
		} else {
			fmt.Fprintf(os.Stderr, s+"\n", args...)
		}
	}
}

// getJSON sends a request to the InvenioAPI using
// a token, url and values as parameters. It return a
// JSON encoded response as byte slice
func getJSON(token string, uri string) ([]byte, *RateLimit, error) {
	rl := new(RateLimit)
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, rl, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-type", "application/json")
	resp, err := client.Do(req)
	rl.FromResponse(resp)
	if err != nil {
		return nil, rl, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		rl.Fprintf(os.Stderr)
	}
	if resp.StatusCode != 200 {
		return nil, rl, fmt.Errorf("%s %s", resp.Status, uri)
	}
	src, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, rl, err
	}
	return src, rl, nil
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
// results as a slice of `map[string]interface{}`
//
// ```
//
//		   records, err := Query(cfg, "Geological History in Southern California", "newest")
//	    if err != nil {
//	        // ... handle error ...
//	    }
//	    for _, rec := ranges {
//	        // ... process results ...
//	    }
//
// ```
func Query(cfg *Config, q string, sort string) ([]map[string]interface{}, error) {
	if sort == "" {
		sort = "bestmatch"
	}
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup our query parameters, i.e. q=*
	uri := fmt.Sprintf("%s/api/records?sort=%s&q=%s", u.String(), sort, q)
	tot := 0
	results := new(QueryResponse)
	records := []map[string]interface{}{}
	for i := 0; uri != ""; i++ {
		dbgPrintf(cfg, "requesting %s", uri)
		src, rl, err := getJSON(cfg.InvenioToken, uri)
		if err != nil {
			return nil, err
		}
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
			dbgPrintf(cfg, "(%d/%d) %s\n", len(records), tot, q)
		}
		if results.Links != nil && results.Links.Self != results.Links.Next {
			uri = results.Links.Next
		} else {
			uri = ""
		}
		if uri != "" {
			// NOTE: We need to respect the rate limits of RDM's API
			rl.Throttle(i, tot)
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
// NOTE: This method relies on OAI-PMH, this is a rate limited process
// so results can take quiet some time.
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
		if err != nil {
			return nil, err
		}
		rl := new(RateLimit)
		rl.FromHeader(headers)
		// NOTE: We need to respect rate limits of the RDM API
		rl.Throttle(i, 1)

		if cfg.Debug {
			os.WriteFile("oai-pmh-list-identifiers.xml", src, 0660)
		}
		if bytes.HasPrefix(src, []byte("<html")) {
			if cfg.Debug {
				os.WriteFile("oai-pmh-error.html", src, 0660)
			}
			resumptionToken = ""
		} else {
			oai := new(OAIListIdentifiersResponse)
			if err := xml.Unmarshal(src, oai); err != nil {
				if cfg.Debug {
					os.WriteFile("oai-pmh-error.html", src, 0660)
				}
				resumptionToken = ""
				return nil, err
			}
			if oai.ListIdentifiers != nil {
				resumptionToken = oai.ListIdentifiers.ResumptionToken
				if oai.ListIdentifiers.Headers != nil {
					for _, obj := range oai.ListIdentifiers.Headers {
						if obj.Identifier != "" {
							parts := strings.Split(obj.Identifier, ":")
							last := len(parts) - 1
							if last < 0 {
								last = 0
							}
							id := parts[len(parts)-1]
							ids = append(ids, id)
						}
					}
				}
			} else {
				resumptionToken = ""
			}
		}
		if (len(ids) % 500) == 0 {
			dbgPrintf(cfg, "%d ids retrieved", len(ids))
		}
	}
	dbgPrintf(cfg, "%d ids retrieved (total)", len(ids))
	return ids, nil
}

// GetModifiedRecordIds takes a configuration object, contacts am RDM
// instance and returns a list of ids created, deleted or updated in
// the time range specififed. I problem is encountered returns an error.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// NOTE: This method relies on OAI-PMH, this is a rate limited process
// so results can take quiet some time.
func GetModifiedRecordIds(cfg *Config, start string, end string) ([]string, error) {
	if start == "" {
		start = time.Now().Format("2006-01-02")
	}
	if end == "" {
		end = time.Now().Format("2006-01-02")
	}
	ids := []string{}
	resumptionToken := "     "
	uri := fmt.Sprintf("%s/oai2d?verb=ListIdentifiers&metadataPrefix=oai_dc&from=%s&until=%s", cfg.InvenioAPI, start, end)
	dbgPrintf(cfg, "requesting %s", uri)
	for i := 0; resumptionToken != ""; i++ {
		if i > 0 {
			v := url.Values{}
			v.Set("resumptionToken", resumptionToken)
			uri = fmt.Sprintf("%s/oai2d?verb=ListIdentifiers&%s", cfg.InvenioAPI, v.Encode())
		}
		src, headers, err := getXML(cfg.InvenioToken, uri)
		if err != nil {
			return nil, err
		}
		rl := new(RateLimit)
		rl.FromHeader(headers)
		// NOTE: We need to respect rate limits for RDM API, unfortunately we don't know the total number of keys from this API request ...
		rl.Throttle(i, 1)

		if bytes.HasPrefix(src, []byte("<html")) {
			// FIXME: Need to display error contained in the HTML response
			os.WriteFile("oai-pmh-error.html", src, 0660) // DEBUG
			resumptionToken = ""
		} else {
			oai := new(OAIListIdentifiersResponse)
			if err := xml.Unmarshal(src, oai); err != nil {
				resumptionToken = ""
				return nil, err
			}
			if oai.ListIdentifiers != nil {
				resumptionToken = oai.ListIdentifiers.ResumptionToken
				if oai.ListIdentifiers.Headers != nil {
					for _, obj := range oai.ListIdentifiers.Headers {
						if obj.Identifier != "" {
							parts := strings.Split(obj.Identifier, ":")
							last := len(parts) - 1
							if last < 0 {
								last = 0
							}
							id := parts[len(parts)-1]
							ids = append(ids, id)
						}
					}
				}
			} else {
				resumptionToken = ""
			}
		}
		if (len(ids) % 500) == 0 {
			dbgPrintf(cfg, "%d ids retrieved for %s - %s", len(ids), start, end)
		}
	}
	dbgPrintf(cfg, "%d ids retrieved (total)", len(ids))
	return ids, nil
}

// GetRecord takes a configuration object and record id,
// contacts an RDM instance and returns a simplified record, a rate limit struct
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// record, rateLimit, err := GetConfig(cfg, id)
//
//	if err != nil {
//		  // ... handle error ...
//	}
//
// ```
func GetRecord(cfg *Config, id string) (map[string]interface{}, *RateLimit, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s", u.String(), id)

	// NOTE: rl is necessary to handle repeat requests to Invenio
	src, rl, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, rl, err
	}
	rec := map[string]interface{}{}
	if err := json.Unmarshal(src, &rec); err != nil {
		return nil, rl, err
	}
	return rec, rl, nil
}
