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
	"path"
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
	var (
		req *http.Request
		err error
	)
	client := &http.Client{}
	if src == nil || len(src) == 0 {
		req, err = http.NewRequest("POST", uri, nil)
	} else {
		req, err = http.NewRequest("POST", uri, bytes.NewBuffer(src))
	}
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		if resp.Header != nil {
			return nil, resp.Header, err
		}
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 201 {
		return nil, resp.Header, fmt.Errorf("%s %s", resp.Status, uri)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		if resp.Header != nil {
			return nil, resp.Header, err
		}
		return nil, nil, err
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

// patchJSON takes a token, a uri and JSON source as byte slice
// and sends it to RDM instance for processing.
func patchJSON(token string, uri string, src []byte) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", uri, bytes.NewBuffer(src))
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

func putFile(token string, uri string, fName string) ([]byte, http.Header, error) {
	client := &http.Client{}
	in, err := os.Open(fName)
	if err != nil {
		return nil, nil, err
	}
	defer in.Close()

	req, err := http.NewRequest("PUT", uri, in)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		return nil, resp.Header, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, resp.Header, fmt.Errorf("%s %s", resp.Status, uri)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return data, resp.Header, err
}

func deleteFile(token string, uri string, fName string) ([]byte, http.Header, error) {
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
	if resp.StatusCode != 204 { 
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
// record, err := GetRecord(cfg, id, false)
// if err != nil {
//    // ... handle error ...
// }
// ```
func GetRecord(cfg *Config, id string, draft bool) (*simplified.Record, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s", u.String(), id)
	if draft {
		uri = fmt.Sprintf("%s/api/records/%s/draft", u.String(), id)
	}
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
func GetFiles(cfg *Config, id string) (*simplified.FileListing, error) {
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
	obj := new(simplified.FileListing)
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

// CreateRecord takes a configuration object and JSON record values.
// It contacts an RDM instance and create a new record return the 
// JSON for the newly created record with a record id.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// fName := "new_record.json" // A new record in JSON
// src, _ := os.ReadFile(fName)
// record, err := CreateRecord(cfg, src)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", record)
// ```
func CreateRecord(cfg *Config, src []byte) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a new record, the JSON returned is supposed
	// to contain the record id and rest of record.
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
func CreateDraft(cfg *Config, recordId string, src []byte) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft", u.String(), recordId)
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

// DiscardDraft takes a configuration object and record id,
// contacts an RDM instance and deletes a draft of a record 
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// _, err := DiscardDraft(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
// ```
func DiscardDraft(cfg *Config, recordId string) (map[string]interface{}, error) {
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
	if len(src) > 0 {
		obj := map[string]interface{}{}
		if err := json.Unmarshal(src, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	}
	return nil, nil
}


// getRecordLink retrieve a the record link uri for attribute name
// from a map[string]interface{} representation of a record under 
// links attribute.
func getRecordLink(m map[string]interface{}, attr string) (string, bool) {
	if elem, ok := m["links"]; ok {
		links := elem.(map[string]interface{})
		if link, hasLink := links[attr]; hasLink {
			return link.(string), true
		}
	}
	return "", false
}

// PublishDraft takes a configuration object and record id,
// contacts an RDM instance and publishes the draft record 
// and returns an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// _, err := PublishDraft(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
// ```
func PublishDraft(cfg *Config, recordId string) (map[string]interface{}, error) {
	m, err := GetDraft(cfg, recordId)
	if err != nil {
		return nil, err
	}
	// First see if we're submitting this for review
	link, ok := getRecordLink(m, "publish")
	if ! ok {
		return nil, fmt.Errorf("cannot publish %q, no link", recordId)
	}
	// Setup API request for a record
	// Make sure we have a valid URL
	src, headers, err := postJSON(cfg.InvenioToken, link, nil)
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

// SubmitDraft takes a configuration object and record id,
// contacts an RDM instance and submits a draft record 
// for review. It returns JSON results and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// _, err := SubmitDraft(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
// ```
func SubmitDraft(cfg *Config, recordId string) (map[string]interface{}, error) {
	m, err := GetDraft(cfg, recordId)
	if err != nil {
		return nil, err
	}
	// First see if we're submitting this for review
	link, ok := getRecordLink(m, "submit-review")
	if ! ok {
		link, ok = getRecordLink(m, "publish")
		if ! ok {
			return nil, fmt.Errorf("cannot submit for review %q, no link to review or submit-review", recordId)
		}
	}
	appName := path.Base(os.Args[0])
	payload := map[string]string{
			"content": fmt.Sprintf("This record is submitted automatically with %s", appName),
			"format": "html",
		}
	payloadSrc, err := json.MarshalIndent(payload, "", "     ")
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	// Make sure we have a valid URL
	src, headers, err := postJSON(cfg.InvenioToken, link, payloadSrc)
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


// getReviewLink retrieve a the record review link uri for 
// attribute name from a map[string]interface{} representation of a record.
func getReviewLink(m map[string]interface{}, attr string) (string, bool) {
	if elem, ok := m["links"]; ok {
		links := elem.(map[string]interface{})
		if attr == "accept" || attr == "cancel" || attr == "decline" {
			if elem, ok = links["actions"]; ok {
				links = elem.(map[string]interface{})
				if link, hasLink := links[attr]; hasLink {
					return link.(string), true
				}
			}
		} else {
			if elem, ok = links[attr]; ok {
				return elem.(string), true
			}
		}
	}
	return "", false
}

// getReviewCommunity takes a review record as map[string]interface{} and
// returns the comminuty uuid is found.
func getReviewCommunity(m map[string]interface{}) (string, bool) {
	if elem, ok := m["reciever"]; ok {
		data := elem.(map[string]interface{})
		if elem, ok = data["community"]; ok {
			return elem.(string), true
		}
	}
	return "", false
}

// ReviewDraft takes a configuration object and record id, a decision,
// and optional comment contacts an RDM instance and updates the 
// review status for the submitted draft record.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// _, err := ReviewDraft(cfg, id, "accept", "")
// if err != nil {
//    // ... handle error ...
// }
// ```
func ReviewDraft(cfg *Config, recordId string, decision string, comment string) (map[string]interface{}, error) {
	m, err := GetDraft(cfg, recordId)
	if err != nil {
		return nil, err
	}

	// Setup for API review request for a record
	link, ok := getRecordLink(m, "review")
	if ! ok {
		return nil, fmt.Errorf("cannot find review link for %q", recordId)
	}
	src, headers, err := getJSON(cfg.InvenioToken, link)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)

	reviewObj := map[string]interface{}{}
	err = json.Unmarshal(src, &reviewObj)
	if err != nil {
		return nil, err
	}
	// Get community uuid
	//community, _ := getReviewCommunity(reviewObj)
	// Get Review links
	commentLink, _ := getReviewLink(reviewObj, "comments")
	acceptLink, _ := getReviewLink(reviewObj, "accept")
	cancelLink, _ := getReviewLink(reviewObj, "cancel")
	declineLink, _ := getReviewLink(reviewObj, "decline")

	// Pick link to do update
	switch decision {
		case "accept":
			link = acceptLink
		case "cancel":
			link = cancelLink
		case "decline":
			link = declineLink
		case "":
			link = commentLink
		default:
			return nil, fmt.Errorf("unsupported decision type %q", decision)
	}

	// Setup payload for update
	payload := map[string]interface{}{}
	if comment != "" {
		payload["content"] = comment
	} else {
		appName := path.Base(os.Args[0])
		payload["content"] = fmt.Sprintf(`this record was processed by %s`, appName)
	}
	payload["format"] = "html"
	payloadSrc, err := json.MarshalIndent(payload, "", "    ")
	if err != nil {
		return nil, err
	}

	// Make review request with Payload
	src, headers, err = postJSON(cfg.InvenioToken, link, payloadSrc)
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



// UploadFiles takes a configuration object and record id,
// and a map to filename and paths contacts an RDM instance 
// and adds the files to a draft record.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// fNames := ["article.pdf", "data.zip" ] // add files to draft record in JSON
// src, _ := os.ReadFile(fName)
// draft, err := UploadFiles(cfg, id, fNames)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", draft)
// ```
func UploadFiles(cfg *Config, recordId string, filenames []string) ([]byte, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Take our list of files and turn it into a request
	uploadInfo := []map[string]string{}
	for _, fName := range filenames {
		key := path.Base(fName)
		uploadInfo = append(uploadInfo, map[string]string{ "key": key })
	}
	// Now turn uploadInfo into an array of objects and do POST
	srcInfo, err := json.MarshalIndent(uploadInfo, "", "    ")
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft/files", u.String(), recordId)
	src, headers, err := postJSON(cfg.InvenioToken, uri, srcInfo)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	filesInfo := new(simplified.FileListing)
	if err := json.Unmarshal(src, &filesInfo); err != nil {
		return nil, err
	}
	if filesInfo == nil || filesInfo.Entries == nil {
		return nil, fmt.Errorf("not file info returned")
	}
	// NOTE: Figure out what the content URL is and post to it.
	for _, fName := range filenames {
		key := path.Base(fName)
		uri = fmt.Sprintf("%s/api/records/%s/draft/files/%s/content", u.String(), recordId, key)
		if _, _, err := putFile(cfg.InvenioToken, uri, fName); err != nil {
			return nil, err
		}
		// Commit the upload
		uri = fmt.Sprintf("%s/api/records/%s/draft/files/%s/commit", u.String(), recordId, key)
		if _, _, err := postJSON(cfg.InvenioToken, uri, nil); err != nil {
			return nil, err
		}
	}
	uri = fmt.Sprintf("%s/api/records/%s/draft/files", u.String(), recordId)
	src, _, err = getJSON(cfg.InvenioToken, uri)
	return src, nil
}


// DeleteFiles takes a configuration object and record id,
// and list of files and removes from a draft.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// fNames := ["article.pdf", "data.zip" ] // add files to draft record in JSON
// src, _ := os.ReadFile(fName)
// draft, err := DeleteFiles(cfg, id, fNames)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", draft)
// ```
func DeleteFiles(cfg *Config, recordId string, filenames []string) ([]byte, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	var uri string
	// NOTE: Figure out what the content URL is and post to it.
	for _, fName := range filenames {
		key := path.Base(fName)
		uri = fmt.Sprintf("%s/api/records/%s/draft/files/%s", u.String(), recordId, key)
		if _, _, err := deleteFile(cfg.InvenioToken, uri, fName); err != nil {
			return nil, err
		}
	}
	uri = fmt.Sprintf("%s/api/records/%s/draft/files", u.String(), recordId)
	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}

// GetAccess takes an acces token, a record id and optionally a 
// access type. Returns either the access object or 
// attribute if type is specified. Also returns an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// src, err := GetAccess(cfg.InvenioToken, id, "")
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func GetAccess(cfg *Config, recordId string, accessType string) ([]byte, error) {
	var src []byte
	rec, err := GetRecord(cfg, recordId, false)
	if err != nil {
		rec, err = GetRecord(cfg, recordId, true)
		if err != nil {
			return nil, err
		}
	}
	switch accessType {
	case "files":
		src, err = json.MarshalIndent(rec.RecordAccess.Files, "", "    ")
	case "record":
		src, err = json.MarshalIndent(rec.RecordAccess.Record, "", "    ")
	case "embargo":
		src, err = json.MarshalIndent(rec.RecordAccess.Embargo, "", "    ")
	case "":
		src, err = json.MarshalIndent(rec.RecordAccess, "", "    ")
	default:
		return nil, fmt.Errorf("%q is not a supported access type", accessType)
	}
	if err != nil {
		return nil, err
	}
	return src, nil
}

// SetAccess takes an access token, record id, a access type and value.
// Returns the updated access object and error value.
//
// FIXME: Current this method only supports setting record and files 
// attributes to "public" and "restricted". Future implementations may
// add support to set record embargos.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// src, err := GetAccess(cfg.InvenioToken, id, "")
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func SetAccess(cfg *Config, recordId string, accessType string, accessValue string) ([]byte, error) {
	var (
		src []byte
	)
		
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// First check if we can get a record, then check if there is a draft record.
	uri := fmt.Sprintf("%s/records/%s", u.String(), recordId)
	rec, err := GetRecord(cfg, recordId, false)
	if err != nil {
		draft, err := GetDraft(cfg, recordId)
		if err != nil {
			return nil, err
		}
		if draft == nil {
			return nil, fmt.Errorf("unable to find record or draft for %s\n", recordId)
		}
		access := map[string]interface{}{
			"files": "public",
			"record": "public",
		}
		if _, ok := draft["access"]; ok {
			access = draft["access"].(map[string]interface{})
		}
		switch accessType {
			case "files":
				access["files"] = accessValue
			case "record":
				access["record"] = accessValue
			default:
				return nil, fmt.Errorf("%q is not a supported access type", accessType)
		}
		draft["access"] = access
		src, err := json.MarshalIndent(draft, "", "    ")
		if err != nil {
			return nil, err
		}
		draft, err =  UpdateDraft(cfg, recordId, src)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(draft, "", "    ")
	} 

	switch accessType {
	case "files":
		rec.RecordAccess.Files = accessValue
	case "record":
		rec.RecordAccess.Record = accessValue
	default:
		return nil, fmt.Errorf("%q is not a supported access type", accessType)
	}
	src, err = json.MarshalIndent(rec, "", "    ")
	if err != nil {
		return nil, err
	}
	src, headers, err := postJSON(cfg.InvenioToken, uri, src)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}


// GetEndpoint takes an access token and endpoint path and returns
// JSON source and error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// p := "api/records/qez01-2309a/draft"
// src, err := GetEndpoint(cfg.InvenioToken, p)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func GetEndpoint(cfg *Config, p string) ([]byte, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s/%s", u.String(), p)
	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}

// PostEndpoint takes an access token and endpoint path along with
// JSON source as payload and returns JSON source and error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// p := "api/records/qez01-2309a/draft"
// data := os.ReadFile("draft.json")
// src, err := PostEndpoint(cfg.InvenioToken, p, data)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func PostEndpoint(cfg *Config, p string, payload []byte) ([]byte, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s/%s", u.String(), p)
	src, headers, err := postJSON(cfg.InvenioToken, uri, payload)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}

// PutEndpoint takes an access token and endpoint path along with
// JSON source as payload and returns JSON source and error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// p := "api/records/qez01-2309a/draft"
// data := os.ReadFile("draft.json")
// src, err := PutEndpoint(cfg.InvenioToken, p, data)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func PutEndpoint(cfg *Config, p string, payload []byte) ([]byte, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s/%s", u.String(), p)
	src, headers, err := putJSON(cfg.InvenioToken, uri, payload)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}


// PatchEndpoint takes an access token and endpoint path along with
// JSON source as payload and returns JSON source and error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// p := "api/records/qez01-2309a/draft"
// data := os.ReadFile("draft.json")
// src, err := PatchEndpoint(cfg.InvenioToken, p, data)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func PatchEndpoint(cfg *Config, p string, payload []byte) ([]byte, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s/%s", u.String(), p)
	src, headers, err := patchJSON(cfg.InvenioToken, uri, payload)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}


