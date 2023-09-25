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
	"database/sql"
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

	// 3rd Party packages
	_ "github.com/lib/pq"

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

// errorToString
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%s", err)
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
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.Header == nil {
		return nil, nil, fmt.Errorf("nil response header")
	}
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
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.Header == nil {
		return nil, nil, fmt.Errorf("nil response header")
	}
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
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.Header == nil {
		return nil, nil, fmt.Errorf("nil response header")
	}
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
func postJSON(token string, uri string, src []byte, expectedStatusCode int, debug bool) ([]byte, http.Header, error) {
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
		if resp != nil && resp.Header != nil {
			return nil, resp.Header, err
		}
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.Header == nil {
		return nil, nil, fmt.Errorf("nil response header")
	}
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG postJSON(token, %q, src, true) -> %d, %s\n", uri, resp.StatusCode, resp.Status)
	}
	if resp.StatusCode != expectedStatusCode {
		return nil, resp.Header, fmt.Errorf("POST %s %s, expected %d\n\tpayload\n%s\n", resp.Status, uri, expectedStatusCode, src)
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
func putJSON(token string, uri string, src []byte, expectedStatusCode int, debug bool) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(src))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.Header == nil {
		return nil, nil, fmt.Errorf("nil response header")
	}
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG putJSON(token, %q, src, true) -> %d, %s\n", uri, resp.StatusCode, resp.Status)
	}
	if resp.StatusCode != expectedStatusCode {
		return nil, resp.Header, fmt.Errorf("PUT %s %s, expected %d\n\tpayload\n%s\n", resp.Status, uri, expectedStatusCode, src)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return data, resp.Header, err
}

// patchJSON takes a token, a uri and JSON source as byte slice
// and sends it to RDM instance for processing.
func patchJSON(token string, uri string, src []byte, expectedStatusCode int, debug bool) ([]byte, http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("PATCH", uri, bytes.NewBuffer(src))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.Header == nil {
		return nil, nil, fmt.Errorf("nil response header")
	}
	if resp.StatusCode != expectedStatusCode {
		return nil, resp.Header, fmt.Errorf("PATCH %s %s, expected %d\n\tpayload\n%s\n", resp.Status, uri, expectedStatusCode, src)
	}
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG patchJSON(token, %q, %s, %d, true) ->%d %s\n", uri, src, expectedStatusCode, resp.StatusCode, resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return data, resp.Header, err
}


// delJSON takes a token, uri and sends it to RDM instance 
// for processing.
func delJSON(token string, uri string, expectedStatusCode int, debug bool) ([]byte, http.Header, error) {
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
	if resp.StatusCode != expectedStatusCode {
		return nil, resp.Header, fmt.Errorf("%s %s, expected %d", resp.Status, uri, expectedStatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	return data, resp.Header, err
}

func putFile(token string, uri string, fName string, expectedStatusCode int, debug bool) ([]byte, http.Header, error) {
	client := &http.Client{}
	src, err := os.ReadFile(fName)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("PUT", uri, bytes.NewBuffer(src))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/octet-stream")
	resp, err := client.Do(req)
	if err != nil {
		return nil, resp.Header, err
	}
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG putFile(tokan, %q, %q, %d, true) -> %d, %s", uri, fName, expectedStatusCode, resp.StatusCode, resp.Status)
	}
	if resp.StatusCode != expectedStatusCode {
		return nil, resp.Header, fmt.Errorf("%s %s, expected %d", resp.Status, uri, expectedStatusCode)
	}
	return nil, resp.Header, err
}

func deleteFile(token string, uri string, fName string, expectedStatusCode int, debug bool) ([]byte, http.Header, error) {
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
	if resp.StatusCode != expectedStatusCode { 
		return nil, resp.Header, fmt.Errorf("%s %s, expected %d", resp.Status, uri, expectedStatusCode)
	}
	if debug {
		fmt.Fprintf(os.Stderr, "DEBUG deleteFile(token, %q, %q, %d, true) -> %d, %s\n", uri, fName, expectedStatusCode, resp.StatusCode, resp.Status)
	}
	return nil, resp.Header, err
}


// CheckDOI takes a DOI and does a lookup to see if there are any
// matching .pids.doi.indentifier values.
//
// ```
// doi := "10.1126/science.82.2123.219"
// records, err := CheckDOI(cfg, doi)
// if err != nil {
//    // ... handle error ...
// }
// for _, rec := ranges {
//    // ... process results ...
// }
// ```
func CheckDOI(cfg *Config, doi string) ([]map[string]interface{}, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup our query parameters, i.e. q=*
	// ?q=pids.doi.identifier:"10.1126/science.82.2123.219"&allversions=true
	u.Path = "/api/records"
	
	q := url.Values{}
	q.Set("q", fmt.Sprintf("pids.doi.identifier:%q", doi))
	q.Set("allversions", "true")
	uri := fmt.Sprintf("%s?%s", u.String(), q.Encode())
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
		if err := JSONUnmarshal(src, &results); err != nil {
			return nil, err
		}
		if results != nil && results.Hits != nil &&
			results.Hits.Hits != nil && len(results.Hits.Hits) > 0 {
			for _, hit := range results.Hits.Hits {
				records = append(records, hit)
			}
			tot = results.Hits.Total
			dbgPrintf(cfg, "(%d/%d) %s\n", len(records), tot, doi)
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
		if err := JSONUnmarshal(src, &results); err != nil {
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

// getRecordIdsFromPg will return all record ids found by querying Invenio RDM's Postgres
// database.
func getRecordIdsFromPg(cfg *Config) ([]string, error) {
	
	sslmode := "?sslmode=require"
	if strings.HasPrefix(cfg.InvenioDbHost, "localhost") {
		sslmode = "?sslmode=disable"
	}
	connStr := fmt.Sprintf("postgres://%s@%s/%s%s", 
		cfg.InvenioDbUser, cfg.InvenioDbHost, cfg.RepoID, sslmode)
	if cfg.InvenioDbPassword != "" {
		connStr = fmt.Sprintf("postgres://%s:%s@%s/%s%s", 
			cfg.InvenioDbUser, cfg.InvenioDbPassword, cfg.InvenioDbHost, cfg.RepoID, sslmode)
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	keys := []string{}
	stmt := `SELECT json->>'id' AS rdmid FROM rdm_records_metadata WHERE json->'access'->>'record' = 'public'`
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var rdmid string
		if err := rows.Scan(&rdmid); err != nil {
			return nil, err
		}
		keys = append(keys, rdmid)
	}
	err = rows.Err()
	return keys, err
}

// getModifiedRecordIdsFromPg will return of record ids found in date range by querying
// Invenio RDM's Postgres database.
func getModifiedRecordIdsFromPg(cfg *Config, startDate string, endDate string) ([]string, error) {
	sslmode := "?sslmode=require"
	if strings.HasPrefix(cfg.InvenioDbHost, "localhost") {
		sslmode = "?sslmode=disable"
	}
	connStr := fmt.Sprintf("postgres://%s@%s/%s%s", 
		cfg.InvenioDbUser, cfg.InvenioDbHost, cfg.RepoID, sslmode)
	if cfg.InvenioDbPassword != "" {
		connStr = fmt.Sprintf("postgres://%s:%s@%s/%s%s", 
			cfg.InvenioDbUser, cfg.InvenioDbPassword, cfg.InvenioDbHost, cfg.RepoID, sslmode)
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	keys := []string{}
	stmt := fmt.Sprintf(`with t as (
    select json->>'id' as rdmid, 
	       json->'access'->>'record' as status,
	       jsonb_path_query(json->'metadata'->'dates', '$.type')::jsonb->>'id' as date_type,
	       to_date(jsonb_path_query(json->'metadata'->'dates', '$.date') #>> '{}', 'YYYY-MM-DD') as dt
    from rdm_records_metadata
    where json->'access'->>'record' = 'public'
) select rdmid
from t 
where date_type = 'updated'
  and (dt between '%s' and '%s')
order by dt;`, startDate, endDate)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var rdmid string
		if err := rows.Scan(&rdmid); err != nil {
			return nil, err
		}
		keys = append(keys, rdmid)
	}
	err = rows.Err()
	return keys, err
}
// GetRecordIds takes a configuration object, contacts am RDM
// instance and returns a list of ids and error. If the RDM
// database connection is included in the configuration the faster
// method of querrying Postgres is used, otherwise OAI-PMH is used
// to get the id list.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set. It is highly recommended that the
// InvenioDbUser, InvenioDbPassword and InvenioDbHost is configured.
//
// NOTE: This method relies on OAI-PMH, this is a rate limited process
// so results can take quiet some time.
func GetRecordIds(cfg *Config) ([]string, error) {
	if cfg.InvenioDbHost != "" && cfg.InvenioDbUser != "" {
		return getRecordIdsFromPg(cfg)
	}
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
	if cfg.InvenioDbHost != "" && cfg.InvenioDbUser != "" {
		return getModifiedRecordIdsFromPg(cfg, start, end)
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
	if err := JSONUnmarshal(src, &rec); err != nil {
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
	if err := JSONUnmarshal(src, &rec); err != nil {
		return nil, err
	}
	return rec, nil
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
	if err := JSONUnmarshal(src, &obj); err != nil {
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
	if err := JSONUnmarshal(src, &obj); err != nil {
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
	if err := JSONUnmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// NewRecord takes a configuration object and JSON record values.
// It contacts an RDM instance and create a new record return the 
// JSON for the newly created record with a record id. When records
// are created they are in "draft" state.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// fName := "new_record.json" // A new record in JSON
// src, _ := os.ReadFile(fName)
// record, err := NewRecord(cfg, src)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", record)
// ```
func NewRecord(cfg *Config, src []byte) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a new record, the JSON returned is supposed
	// to contain the record id and rest of record.
	uri := fmt.Sprintf("%s/api/records", u.String())
	src, headers, err := postJSON(cfg.InvenioToken, uri, src, http.StatusCreated, false)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := JSONUnmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// NewRecordVersion takes a configuration object and record id to
// create the new version draft. The returns JSON record values includes
// the new record id identifying the new version.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id = "38rg4-36m04" 
// record, err := NewRecordVersion(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", record)
// ```
func NewRecordVersion(cfg *Config, recordId string) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a new record, the JSON returned is supposed
	// to contain the record id and rest of record.
	uri := fmt.Sprintf("%s/api/records/%s/versions", u.String(), recordId)
	src, headers, err := postJSON(cfg.InvenioToken, uri, nil, http.StatusCreated, false)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	if len(src) > 0 {
		obj := map[string]interface{}{}
		if err := JSONUnmarshal(src, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	}
	return nil, nil
}

// PublishRecordVersion takes a configuration object and record id of
// a new version draft and publishes it. NOTE: creating a new
// version will clear .metadata.publication_date, version label and DOI.
// These can be replace when publishing in the version and pubDate
// parameter. If those values are empty string no change is made
// to the draft before publishing.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id = "38rg4-36m04" 
// version, pubDate := "internal", "2022-08"
// record, err := PublicRecordVersion(cfg, id, version, pubDate)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", record)
// ```
func PublishRecordVersion(cfg *Config, recordId string, version string, pubDate string, debug bool) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	if version != "" || pubDate != "" {
		// We need fetch and update the draft before publising it.
		m, err := GetDraft(cfg, recordId)
		if err != nil {
			return nil, err
		}
		metadata, ok := m["metadata"].(map[string]interface{})
		if ! ok {
			return nil, fmt.Errorf("missing metadata element in draft record %s", recordId)
		}
		if version != "" {
			metadata["version"] = version
		}
		if pubDate != "" {
			metadata["publication_date"] = pubDate
		}
		m["metadata"] = metadata
		payload, err := JSONMarshalIndent(m, "", "     ")
		if err != nil {
			return nil, err
		}
		_, err = UpdateDraft(cfg, recordId, payload, debug)
		if err != nil {
			return nil, err
		}
	}
	// Setup API request for a new record, the JSON returned is supposed
	// to contain the record id and rest of record.
	uri := fmt.Sprintf("%s/api/records/%s/draft/actions/publish", u.String(), recordId)
	src, headers, err := postJSON(cfg.InvenioToken, uri, nil, http.StatusAccepted, false)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	if len(src) > 0 {
		obj := map[string]interface{}{}
		if err := JSONUnmarshal(src, &obj); err != nil {
			return nil, err
		}
		return obj, nil
	}
	return nil, nil
}

// NewDraft takes a configuration object and record id,
// contacts an RDM instance and create a draft of an existing record 
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id = "38rg4-36m04" 
// draft, err := NewDraft(cfg, id)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", draft)
// ```
func NewDraft(cfg *Config, recordId string) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft", u.String(), recordId)
	src, headers, err := postJSON(cfg.InvenioToken, uri, nil, http.StatusCreated, false)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := JSONUnmarshal(src, &obj); err != nil {
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
	if err := JSONUnmarshal(src, &obj); err != nil {
		return nil, err
	}
	// Sometimes .pids.doi comes back with missing indentifier value
	// but scheme is doi. If this is the case removed it.
	if elem, ok := obj["pids"]; ok {
		pids := elem.(map[string]interface{})
		if elem, ok := pids["doi"]; ok {
			doi := elem.(map[string]interface{})
			if identifier, ok := doi["identifier"]; ok && identifier.(string) == "" {
				delete(pids, "doi")
			}
		} 
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
// debug := true
// draft, err := UpdateDraft(cfg, id, src, debug)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%+v\n", draft)
// ```
func UpdateDraft(cfg *Config, recordId string, payloadSrc []byte, debug bool) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft", u.String(), recordId)
	src, headers, err := putJSON(cfg.InvenioToken, uri, payloadSrc, http.StatusOK, debug)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	obj := map[string]interface{}{}
	if err := JSONUnmarshal(src, &obj); err != nil {
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
// debug := true
// _, err := DiscardDraft(cfg, id, debug)
// if err != nil {
//    // ... handle error ...
// }
// ```
func DiscardDraft(cfg *Config, recordId string, debug bool) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft", u.String(), recordId)
	_, headers, err := delJSON(cfg.InvenioToken, uri, http.StatusNoContent, debug)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return nil, nil
}

// SetFilesEnable will set the metadata.files.enable value.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// enabled := true
// debug := true
// _, err := SetFilesEnable(cfg, id, enabled, debug)
// if err != nil {
//    // ... handle error ...
// }
// ```
func SetFilesEnable(cfg *Config, recordId string, enable bool, debug bool) (map[string]interface{}, error) {
	m, err := GetDraft(cfg, recordId)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("unabled to find draft for %q", recordId)
	}
	updateDraft := false
	if elem, ok := m["files"]; ok {
		data := elem.(map[string]interface{})
		if elem, ok = data["enabled"]; ok {
			setting := elem.(bool)
			if setting != enable {
				data["enabled"] = enable
				m["files"] = data
				updateDraft = true
			}
		}
	} else {
		// NOTE: We don't have a files element in our record. We create one
		// and update the draft
		m["files"] = map[string]interface{}{
			"enabled": enable,
		}
		updateDraft = true	
	}
	if updateDraft {
		src, err := JSONMarshalIndent(m, "", "    ")
		if err != nil {
			return m, err
		}
		return UpdateDraft(cfg, recordId, src, debug)
	}
	return m, nil
}

// SetVersion will set the metadata.version value.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// versoin := 'public files'
// debug := true
// _, err := SetVersion(cfg, id, version, debug)
// if err != nil {
//    // ... handle error ...
// }
// ```
func SetVersion(cfg *Config, recordId string, version string, debug bool) (map[string]interface{}, error) {
	m, err := GetDraft(cfg, recordId)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("unabled to find draft for %q", recordId)
	}
	if elem, ok := m["metadata"]; ok {
		data := elem.(map[string]interface{})
		data["version"] = version
		m["metadata"] = data
		src, err := JSONMarshalIndent(m, "", "    ")
		if err != nil {
			return m, err
		}
		return UpdateDraft(cfg, recordId, src, debug)
	}
	return nil, fmt.Errorf("could not find .metadata.version")
}

// SetPubDate will set the metadata.publication_date value.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// pubDate := "2016"
// debug := true
// _, err := SetPublicationDate(cfg, id, pubDate, debug)
// if err != nil {
//    // ... handle error ...
// }
// ```
func SetPubDate(cfg *Config, recordId string, pubDate string, debug bool) (map[string]interface{}, error) {
	m, err := GetDraft(cfg, recordId)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("unabled to find draft for %q", recordId)
	}
	if elem, ok := m["metadata"]; ok {
		data := elem.(map[string]interface{})
		data["publication_date"] = pubDate
		m["metadata"] = data
		src, err := JSONMarshalIndent(m, "", "    ")
		if err != nil {
			return m, err
		}
		return UpdateDraft(cfg, recordId, src, debug)
	}
	return nil, fmt.Errorf("could not find .metadata.publication_date")
}

// GetDraftFiles takes a configuration object and record id,
// contacts an RDM instance and returns the files metadata
// and an error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// debug := true
// entries, err := GetDraftFiles(cfg, id, debug)
// if err != nil {
//    // ... handle error ...
// }
// ```
func GetDraftFiles(cfg *Config, recordId string, debug bool) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s/%s/draft/files", u.String(), recordId)
	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	m := map[string]interface{}{}
	if err := JSONUnmarshal(src, &m); err != nil {
		return nil, err
	}
	return m, nil
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
// debug := true
// entries, err := GetFiles(cfg, id, debug)
// if err != nil {
//    // ... handle error ...
// }
// ```
func GetFiles(cfg *Config, recordId string, debug bool) (map[string]interface{}, error) {
	// Make sure we have a valid URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s/%s/files", u.String(), recordId)
	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	m := map[string]interface{}{}
	if err := JSONUnmarshal(src, &m); err != nil {
		return nil, err
	}
	return m, nil
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
func UploadFiles(cfg *Config, recordId string, filenames []string, debug bool) (map[string]interface{}, error) {
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
	payloadSrc, err := JSONMarshalIndent(uploadInfo, "", "    ")
	if err != nil {
		return nil, err
	}
	// Setup API request for a record
	uri := fmt.Sprintf("%s/api/records/%s/draft/files", u.String(), recordId)
	src, headers, err := postJSON(cfg.InvenioToken, uri, payloadSrc, http.StatusCreated, debug)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	filesInfo := new(simplified.FileListing)
	if err := JSONUnmarshal(src, &filesInfo); err != nil {
		return nil, err
	}
	if filesInfo == nil || filesInfo.Entries == nil {
		return nil, fmt.Errorf("not file info returned")
	}
	// NOTE: Figure out what the content URL is and post to it.
	for _, fName := range filenames {
		key := path.Base(fName)
		uri = fmt.Sprintf("%s/api/records/%s/draft/files/%s/content", u.String(), recordId, key)
		_, headers, err := putFile(cfg.InvenioToken, uri, fName, http.StatusOK, debug)
		if err != nil {
			return nil, err
		}
		cfg.rl.FromHeader(headers)
		// Commit the upload
		uri = fmt.Sprintf("%s/api/records/%s/draft/files/%s/commit", u.String(), recordId, key)
		_, headers, err = postJSON(cfg.InvenioToken, uri, nil, http.StatusOK, debug)
		if err != nil {
			return nil, err
		}
		cfg.rl.FromHeader(headers)
	}
	return GetDraft(cfg, recordId)
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
func DeleteFiles(cfg *Config, recordId string, filenames []string, debug bool) ([]byte, error) {
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
		_, headers, err := deleteFile(cfg.InvenioToken, uri, fName, http.StatusNoContent, debug)
		if err != nil {
			return nil, err
		}
		cfg.rl.FromHeader(headers)
	}
	return nil, nil
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
		src, err = JSONMarshalIndent(rec.RecordAccess.Files, "", "    ")
	case "record":
		src, err = JSONMarshalIndent(rec.RecordAccess.Record, "", "    ")
	case "embargo":
		src, err = JSONMarshalIndent(rec.RecordAccess.Embargo, "", "    ")
	case "":
		src, err = JSONMarshalIndent(rec.RecordAccess, "", "    ")
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
// debug := true
// src, err := SetAccess(cfg.InvenioToken, id, "", debug)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func SetAccess(cfg *Config, recordId string, accessType string, accessValue string, debug bool) ([]byte, error) {
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
		src, err := JSONMarshalIndent(draft, "", "    ")
		if err != nil {
			return nil, err
		}
		draft, err = UpdateDraft(cfg, recordId, src, debug)
		if err != nil {
			return nil, err
		}
		return JSONMarshalIndent(draft, "", "    ")
	} 

	switch accessType {
	case "files":
		rec.RecordAccess.Files = accessValue
	case "record":
		rec.RecordAccess.Record = accessValue
	default:
		return nil, fmt.Errorf("%q is not a supported access type", accessType)
	}
	src, err = JSONMarshalIndent(rec, "", "    ")
	if err != nil {
		return nil, err
	}
	src, headers, err := postJSON(cfg.InvenioToken, uri, src, http.StatusOK, debug)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}


// SendToCommunity sends a draft to an RDM community. This will trigger
// the review step need for publication. You need the record id and a
// community id (looks like a UUID). Returns a map[string]interface{}
// and error values.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// comminityId := ... // this is a UUID like value
// debug := true
// src, err := SendToCommunity(cfg, id, communityId, debug)
// if err != nil {
//    // ... handle error ...
// }
// ```
func SendToCommunity(cfg *Config, recordId string, communityId string, debug bool) (map[string]interface{}, error) {
	appName := path.Base(os.Args[0])
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}

	// Get draft so we have the links we need to do submissions
	m, err := GetDraft(cfg, recordId)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("unabled to find draft for %q", recordId)
	}

	// Get review and submit links
	reviewLink := fmt.Sprintf("%s/api/records/%s/draft/review", u.String(), recordId)

	payload := fmt.Sprintf(`{
  "receiver": {
    "community": %q
  },
  "type": "community-submission"
}
`, communityId)
	src, headers, err := putJSON(cfg.InvenioToken, reviewLink, []byte(payload), http.StatusOK, debug)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)

	data := map[string]interface{}{}
	err = JSONUnmarshal(src, &data)
	
	requestId, ok := data["id"]
	if !ok {
		return nil, fmt.Errorf("failed to get request id from PUT, payload received\n%s", payload)
	}

	// Figure out the submitLink for the comment now that we have a 
	// request id.
	submitLink := fmt.Sprintf("%s/api/requests/%s/comments", u.String(), requestId)

	// Add Submit Review and comment with a POST and submit-review link
	comment := fmt.Sprintf("This record is submitted automatically with %s", appName)
	payload = fmt.Sprintf(`{
  "payload": {
    "content": %q,
    "format": "html"
  }
}`, comment)
	src, headers, err = postJSON(cfg.InvenioToken, submitLink, []byte(payload), http.StatusCreated, debug)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)

	data = map[string]interface{}{}
	err = JSONUnmarshal(src, &data)
	if err != nil {
		return nil, err
	}

	// Now we need to "submit for review" to community.
	submitReviewLink := fmt.Sprintf("%s/api/records/%s/draft/actions/submit-review", u.String(), recordId)
	payload = `{
  "payload": {
    "content": "Thank you in advance for the review.",
    "format": "html"
  }
}`
	src, headers, err = postJSON(cfg.InvenioToken, submitReviewLink, []byte(payload), http.StatusAccepted, debug)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)

	// Get the updated record and return it.
	return GetDraft(cfg, recordId)
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


// GetReview takes a configuration object, record id and
// returns an review object (which includes a request id) and error code.
func GetReview(cfg *Config, recordId string, debug  bool) (map[string]interface{}, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s/api/records/%s/draft/review", u.String(), recordId)
	src, headers, err := getJSON(cfg.InvenioToken, uri)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)

	m := map[string]interface{}{}
	err = JSONUnmarshal(src, &m)
	if err != nil {
		return nil, err
	}
	return m, err	
}

// ReviewRequest takes a configuration object and record id, a decision,
// and optional comment contacts an RDM instance and updates the 
// review status for the submitted draft record.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// id := "qez01-2309a"
// debug := true
// _, err := ReviewDraft(cfg, id, "accept", "", debug)
// if err != nil {
//    // ... handle error ...
// }
// ```
func ReviewRequest(cfg *Config, recordId string, decision string, comment string, debug bool) (map[string]interface{}, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	// GetReview data
	m, err := GetReview(cfg, recordId, debug)
	if err != nil {
		return nil, err
	}
	// Figure out the request id.
	requestId, ok := m["id"]
	if ! ok {
		return nil, fmt.Errorf("could not retrieve review request for %q", recordId)
	}

	// Pick link and status code for update
	link, expectedStatusCode, payload := "", http.StatusOK, ""
	switch decision {
		case "accept":
			link = fmt.Sprintf("%s/api/requests/%s/actions/accept", u.String(), requestId)
			if comment == "" {
				payload = `{ "payload": {"content": "You are in!", "format": "html"} }`
			} else {
				payload = fmt.Sprintf(`{ "payload": { "content": %q, "format": "html"} }`, comment)
			}
		case "cancel":
			link = fmt.Sprintf("%s/api/requests/%s/actions/cancel", u.String(), requestId)
			if comment == "" {
				payload = `{ "payload": {"content": "Didn't mean to do that!", "format": "html"} }`
			} else {
				payload = fmt.Sprintf(`{ "payload": { "content": %q, "format": "html"} }`, comment)
			}
		case "decline":
			link = fmt.Sprintf("%s/api/requests/%s/actions/decline", u.String(), requestId)
			if comment == "" {
				payload = `{ "payload": {"content": "You are not in!", "format": "html"} }`
			} else {
				payload = fmt.Sprintf(`{ "payload": { "content": %q, "format": "html"} }`, comment)
			}
		case "comment":
			link = fmt.Sprintf("%s/api/requests/%s/comments", u.String(), requestId)
			payload = fmt.Sprintf(`{ "payload": { "content": %q, "format": "html"}`, comment)
			expectedStatusCode = http.StatusCreated
		default:
			return nil, fmt.Errorf("unsupported decision type %q", decision)
	}

	// Make review request with Payload
	src, headers, err := postJSON(cfg.InvenioToken, link, []byte(payload), expectedStatusCode, debug)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)

	obj := map[string]interface{}{}
	if err := JSONUnmarshal(src, &obj); err != nil {
		return nil, err
	}
	return obj, nil
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
	src, headers, err := postJSON(cfg.InvenioToken, uri, payload, http.StatusOK, false)
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
	src, headers, err := putJSON(cfg.InvenioToken, uri, payload, http.StatusOK, false)
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
	src, headers, err := patchJSON(cfg.InvenioToken, uri, payload, http.StatusOK, false)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}


// DeleteEndpoint takes an access token and endpoint path along with
// JSON source as payload and returns JSON source and error value.
//
// The configuration object must have the InvenioAPI and
// InvenioToken attributes set.
//
// ```
// cfg, _ := LoadConfig("config.json")
// p := "api/records/qez01-2309a/draft"
// src, err := DeleteEndpoint(cfg.InvenioToken, p)
// if err != nil {
//    // ... handle error ...
// }
// fmt.Printf("%s\n", src)
// ```
func DeleteEndpoint(cfg *Config, p string) ([]byte, error) {
	// Make sure we have a URL
	u, err := url.Parse(cfg.InvenioAPI)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s/%s", u.String(), p)
	src, headers, err := delJSON(cfg.InvenioToken, uri, http.StatusNoContent, false)
	if err != nil {
		return nil, err
	}
	cfg.rl.FromHeader(headers)
	return src, nil
}


