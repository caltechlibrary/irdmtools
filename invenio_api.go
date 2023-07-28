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
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	// Caltech Library Packages
	"github.com/caltechlibrary/simplified"
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
// JSON encoded response as byte slice, the response header and error
func getJSON(token string, uri string) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-type", "application/json")
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

// getXML sends a request to the Invenio API (e.g. OAI-PMH) using
// a token, url and values as parameters. It returns an
// xml encoded response as byte slice, the response header and error
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

// getRawFile sends a request to the Invenio API using a token, url
// and values as parameters. It retrieves the file contents and returns
// it as a byte array along with response header and error.
func getRawFile(token string, uri string, contentType string) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", contentType)
	resp, err := client.Do(req)
	if err != nil {
		return nil, resp.Header, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, resp.Header, fmt.Errorf("%s %s", resp.Status, uri)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return data, resp.Header, err
}

// postJSON takes a token, uri and JSON source as byte slice 
// and sends it to the RDM instance for processing.
func postJSON(token string, uri string, src []byte) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(src))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, resp.Header, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 201 {
		return nil, resp.Header, fmt.Errorf("%s %s", resp.Status, uri)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return data, resp.Header, err
}

// putJSON takes a token, uri and JSON source as byte slice
// and sends it to RDM instance for processing.
func putJSON(token string, uri string, src []byte) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(src))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, resp.Header, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 201 {
		return nil, resp.Header, fmt.Errorf("%s %s", resp.Status, uri)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return data, resp.Header, err
}

// delJSON takes a token, uri and sends it to RDM instance 
// for processing.
func delJSON(token string, uri string) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", uri, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		return nil, resp.Header, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, resp.Header, fmt.Errorf("%s %s", resp.Status, uri)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return data, resp.Header, err
}


// Query takes a query string and returns the paged object
// results as a slice of `map[string]interface{}`
//
// ```
// records, err := Query(cfg, "Geological History in Southern California", "newest")
//
//	if err != nil {
//	    // ... handle error ...
//	}
//
//	for _, rec := ranges {
//	    // ... process results ...
//	}
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
	t0 := time.Now()
	reportProgress := false
	iTime := time.Now()
	results := new(QueryResponse)
	records := []map[string]interface{}{}
	for i := 0; uri != ""; i++ {
		if iTime, reportProgress = CheckWaitInterval(iTime, time.Minute); reportProgress || (i == 0) {
			log.Printf("(%d/%d) %s", len(records), tot, ProgressETR(t0, len(records), tot))
		}
		dbgPrintf(cfg, "requesting %s", uri)
		src, headers, err := getJSON(cfg.InvenioToken, uri)
		if err != nil {
			return nil, err
		}
		cfg.rl.FromHeader(headers)
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
			cfg.rl.Throttle(i, tot)
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
	t0 := time.Now()
	iTime, reportProgress := time.Now(), false
	uri := fmt.Sprintf("%s/oai2d?verb=ListIdentifiers&metadataPrefix=oai_dc", cfg.InvenioAPI)
	for i := 0; resumptionToken != ""; i++ {
		if iTime, reportProgress = CheckWaitInterval(iTime, (1 * time.Minute)); reportProgress || (len(ids) == 0) {
			var lastId string
			if len(ids) > 0 {
				lastId = ids[len(ids)-1]
			}
			log.Printf("GetRecordIds(cfg) last id %q: %s", lastId, ProgressIPS(t0, len(ids), time.Minute))
		}
		if i > 0 {
			v := url.Values{}
			v.Set("resumptionToken", resumptionToken)
			uri = fmt.Sprintf("%s/oai2d?verb=ListIdentifiers&%s", cfg.InvenioAPI, v.Encode())
		}
		src, headers, err := getXML(cfg.InvenioToken, uri)
		if err != nil {
			return ids, err
		}
		cfg.rl.FromHeader(headers)
		// NOTE: We need to respect rate limits of the RDM API
		cfg.rl.Throttle(i, 1)
		if bytes.HasPrefix(src, []byte("<html")) {
			dbgPrintf(cfg, "\n%s\n", src)
			resumptionToken = ""
		} else {
			oai := new(OAIListIdentifiersResponse)
			if err := xml.Unmarshal(src, oai); err != nil {
				dbgPrintf(cfg, "\n%s\n", src)
				resumptionToken = ""
				return ids, err
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
	}
	log.Printf("%d total retrieved in %s", len(ids), time.Since(t0).Round(time.Second))
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
	debug := cfg.Debug
	ids := []string{}
	resumptionToken := "     "
	uri := fmt.Sprintf("%s/oai2d?verb=ListIdentifiers&metadataPrefix=oai_dc&from=%s&until=%s", cfg.InvenioAPI, start, end)
	dbgPrintf(cfg, "requesting %s", uri)
	t0, iTime, reportProgress := time.Now(), time.Now(), false
	for i := 0; resumptionToken != ""; i++ {
		if iTime, reportProgress = CheckWaitInterval(iTime, (1 * time.Minute)); reportProgress || (len(ids) == 0) {
			var lastId string
			if len(ids) > 0 {
				lastId = ids[len(ids)-1]
			}
			if lastId != "" {
				log.Printf("GetModifiedRecordIDs(cfg, %q, %q) last id %q: %s", start, end, lastId, ProgressIPS(t0, len(ids), time.Minute))
			} else {
				log.Printf("GetModifiedRecordIDs(cfg, %q, %q) %s", start, end, ProgressIPS(t0, len(ids), time.Minute))
			}
		}
		if i > 0 {
			v := url.Values{}
			v.Set("resumptionToken", resumptionToken)
			uri = fmt.Sprintf("%s/oai2d?verb=ListIdentifiers&%s", cfg.InvenioAPI, v.Encode())
		}
		src, headers, err := getXML(cfg.InvenioToken, uri)
		if err != nil {
			return nil, err
		}
		cfg.rl.FromHeader(headers)
		// NOTE: We need to respect rate limits for RDM API, unfortunately we don't know the total number of keys from this API request ...
		cfg.rl.Throttle(i, 1)

		if bytes.HasPrefix(src, []byte("<html")) {
			// FIXME: Need to display error contained in the HTML response
			if debug {
				dbgPrintf(cfg, "\n%s\n", src)
			}
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
	}
	log.Printf("%d total retrieved in %s", len(ids), time.Since(t0).Round(time.Second))
	return ids, nil
}

// GetRawRecord takes a configuration object and record id,
// contacts an RDM instance and returns a map[string]interface{} record
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// mapRecord, err := GetRawRecord(cfg, id)
// if err != nil {
//     // ... handle error ...
// }
// ```
func GetRawRecord(cfg *Config, id string) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s", u.String(), id)

	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	rec := map[string]interface{}{}
	if err := json.Unmarshal(src, &rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// GetRecord takes a configuration object and record id,
// contacts an RDM instance and returns a simplified record
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// record, err := GetRecord(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
// ```
func GetRecord(cfg *Config, id string) (*simplified.Record, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s", u.String(), id)

	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	rec := new(simplified.Record)
	if err := json.Unmarshal(src, &rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// GetFiles takes a configuration object and record id,
// contacts an RDM instance and returns the files metadata
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// entries, err := GetFiles(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
//
// ```
func GetFiles(cfg *Config, id string) (*simplified.Files, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/files", u.String(), id)

	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := new(simplified.Files)
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// GetFile takes a configuration object, record id and filename,
// contacts an RDM instance and returns the specific file metadata
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// fName := "article.pdf"
// entry, err := GetFile(cfg, id, fName)
//
// if err != nil {
//    // ... handle error ...
// }
//
// ```
func GetFile(cfg *Config, id string, fName string) (*simplified.Entry, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/files/%s", u.String(), id, fName)

	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := new(simplified.Entry)
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// RetrieveFile takes a configuration object, record id and filename,
// contacts an RDM instance and returns the specific file 
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// fName := "article.pdf"
// data, err := RetrieveFile(cfg, id, fName)
// if err != nil {
//    // ... handle error ...
// }
// os.WriteFile(fName, data, 0664)
// ```
func RetrieveFile(cfg *Config, id string, fName string) ([]byte, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// NOTE: We need to get the metadata to know the mime-type to request.
	obj, err := GetFile(cfg, id, fName)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	//uri := fmt.Sprintf("%s/api/records/%s/files/%s/content", u.String(), id, fName)
	uri := fmt.Sprintf("%s/records/%s/files/%s?download=1", u.String(), id, fName)

	data, headers, err := getRawFile(cfg.InvenioToken, uri, obj.MimeType)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return data, nil
}

// GetVersions takes a configuration object and record id,
// contacts an RDM instance and returns the versons metadata
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// versions, err := GetVersions(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
//
// ```
func GetVersions(cfg *Config, id string) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/versions", u.String(), id)

	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// GetVersionLatest takes a configuration object and record id,
// contacts an RDM instance and returns the versons metadata
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// versions, err := GetVersionLatest(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
//
// ```
func GetVersionLatest(cfg *Config, id string) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/versions/latest", u.String(), id)

	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// CreateDraft takes a configuration object and record id,
// contacts an RDM instance and create a draft of a record 
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// fName := "drft.json" // A draft record in JSON
// src, _ := os.ReadFile(fName)
// draft, err := CreateDraft(cfg, src)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", draft)
// ```
func CreateDraft(cfg *Config, src []byte) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records", u.String())
	src, headers, err := postJSON(cfg.InvenioToken, uri, src)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// GetDraft takes a configuration object and record id,
// contacts an RDM instance retrieves an existing draft
// of a record and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// draft, err := GetDraft(cfg, id)
//
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", draft)
// ```
func GetDraft(cfg *Config, id string) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft", u.String(), id)

	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// UpdateDraft takes a configuration object and record id,
// contacts an RDM instance and create a draft of a record 
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// fName := "draft.json" // An updated draft record in JSON
// src, _ := os.ReadFile(fName)
// draft, err := UpdateDraft(cfg, id, src)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", draft)
// ```
func UpdateDraft(cfg *Config, recordId string, src []byte) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft", u.String(), recordId)
	src, headers, err := putJSON(cfg.InvenioToken, uri, src)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// DeleteDraft takes a configuration object and record id,
// contacts an RDM instance and deletes a draft of a record 
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// _, err := DeleteDraft(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
// ```
func DeleteDraft(cfg *Config, recordId string) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft", u.String(), recordId)
	src, headers, err := delJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := json.Unmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}


