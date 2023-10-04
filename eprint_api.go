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
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	// Caltech Library Packages
	"github.com/caltechlibrary/eprinttools"
)

const (
	// Basic configurations that don't need to change much.
	timeoutSeconds               = 10
	maxConsecutiveFailedRequests = 10
)

func restClient(urlEndPoint string, timeout time.Duration, retryCount int) ([]byte, error) {
	var (
		username string
		password string
		auth     string
		src      []byte
	)
	if timeout == 0 {
		timeout = timeoutSeconds
	}
	u, err := url.Parse(urlEndPoint)
	if err != nil {
		return nil, fmt.Errorf("%q, %s,", urlEndPoint, err)
	}
	username, password, auth = "", "", "basic"
	if userinfo := u.User; userinfo != nil {
		username = userinfo.Username()
		if secret, isSet := userinfo.Password(); isSet {
			password = secret
		}
	}

	// NOTE: We build our client request object so we can
	// set authentication if necessary.
	req, err := http.NewRequest("GET", urlEndPoint, nil)
	switch strings.ToLower(auth) {
	case "basic":
		req.SetBasicAuth(username, password)
	}
	appName := path.Base(os.Args[0])
	req.Header.Set("User-Agent", fmt.Sprintf("%s %s", appName, Version))
	client := &http.Client{
		Timeout: timeout * time.Second,
	}
	var (
		res *http.Response
	)
	for i := 0; i <= retryCount && res == nil; i++ {
		res, err = client.Do(req)
		if err != nil {
			if strings.Contains(fmt.Sprintf("%s", err), "deadline exceeded") {
				if i < retryCount {
					fmt.Fprintf(os.Stderr, "will retry, ")
				}
				fmt.Fprintf(os.Stderr, "%s\n", err)
				time.Sleep(timeout * time.Second)
			} else {
				return nil, err
			}
		}
	}
	// NOTE: Make sure we're have handled final retry error.
	if res == nil && err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		src, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
	} else {
		//FIXME: urlEndPoint is exposing the password, need to replace with asterisks
		return nil, fmt.Errorf("%s for %s", res.Status, strings.Replace(urlEndPoint, ":"+password, ":***", 1))
	}
	if len(bytes.TrimSpace(src)) == 0 {
		return nil, fmt.Errorf("No data")
	}
	return src, nil
}

// getEPrintFromMySQL builds the EPrint record directly from the EPrint MySQL database.
func getEPrintFromMySQL(cfg *Config, eprintID int) (*eprinttools.EPrints, error) {
	var dsn string  
    if cfg.EPrintDbHost == "localhost" {
         dsn = fmt.Sprintf("%s:%s@/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.RepoID)
    } else {
         dsn = fmt.Sprintf("%s:%s@%s/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.EPrintDbHost, cfg.RepoID)
    }
    db, err := sql.Open("mysql", dsn)
    if err != nil {
         return nil, err
    }
    defer db.Close()
	eprint, err := SQLReadEPrint(db, cfg.EPrintHost, eprintID)
	if err != nil {
		return nil, err
	}
	eprints := new(eprinttools.EPrints)
	eprints.EPrint = append(eprints.EPrint, eprint)
	return eprints, nil
}

// GetEPrint fetches a single EPrint record via the EPrint REST API or MySQL database if configured.
func GetEPrint(cfg *Config, eprintID int, timeout time.Duration, retryCount int) (*eprinttools.EPrints, error) {
	if cfg == nil {
		return nil, fmt.Errorf("application is not configured to retrieve data from the EPrint REST API or MySQL database")
	}
	if cfg.EPrintDbHost != "" && cfg.EPrintDbUser != "" && cfg.EPrintDbPassword != "" {
		return getEPrintFromMySQL(cfg, eprintID)
	}
	baseURL := fmt.Sprintf("https://%s:%s@%s", cfg.EPrintUser, cfg.EPrintPassword, cfg.EPrintHost)
	endPoint := fmt.Sprintf("%s/rest/eprint/%d.xml", baseURL, eprintID)
	src, err := restClient(endPoint, timeout, retryCount)
	if err != nil {
		return nil, err
	}
	data := eprinttools.NewEPrints()
	err = xml.Unmarshal(src, &data)
	if err != nil {
		return nil, err
	}
	for _, e := range data.EPrint {
		e.SyntheticFields()
	}
	return data, nil
}

func getKeysFromMySQL(cfg *Config) ([]int, error) {
	var dsn string  
    if cfg.EPrintDbHost == "localhost" {
         dsn = fmt.Sprintf("%s:%s@/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.RepoID)
    } else {
         dsn = fmt.Sprintf("%s:%s@%s/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.EPrintDbHost, cfg.RepoID)
    }
    db, err := sql.Open("mysql", dsn)
    if err != nil {
         return nil, err
    }
    defer db.Close()
	ids, err := GetAllEPrintIDs(db)
	return ids, err
}

// GetKeys returns a list of eprint record ids from the EPrints REST API
func GetKeys(cfg *Config, timeout time.Duration, retryCount int) ([]int, error) {
	if cfg.EPrintDbHost != "" && cfg.EPrintDbUser != "" && cfg.EPrintDbPassword != "" {
		return getKeysFromMySQL(cfg)
	}
	type eprintKeyPage struct {
		XMLName xml.Name `xml:"html"`
		Anchors []string `xml:"body>ul>li>a"`
	}

	baseURL := fmt.Sprintf("https://%s:%s@%s", cfg.EPrintUser, cfg.EPrintPassword, cfg.EPrintHost)
	endPoint := fmt.Sprintf("%s/rest/eprint/", baseURL)
	src, err := restClient(endPoint, timeout, retryCount)
	if err != nil {
		return nil, err
	}
	keysPage := new(eprintKeyPage)
	err = xml.Unmarshal(src, &keysPage)
	if err != nil {
		return nil, err
	}
	// Build a list of Unique IDs in a map, then convert unique querys to results array
	results := []int{}
	for _, val := range keysPage.Anchors {
		if strings.HasSuffix(val, ".xml") == true {
			eprintID, err := strconv.Atoi(strings.TrimSuffix(val, ".xml"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not extract eprint ID from %s\n", val)
				continue
			}
			results = append(results, eprintID)
		}
	}
	return results, nil
}

// GetModifiedKeys returns a list of eprint record ids from the EPrints MySQL database.
// The REST API is just too slow to process.
func GetModifiedKeys(cfg *Config, start string, end string) ([]int, error) {
	if cfg.EPrintDbHost == "" || cfg.EPrintDbUser == "" || cfg.EPrintDbPassword == "" {
		return nil, fmt.Errorf("database connection not defined")
	}
	_, err := time.Parse("2006-01-02", start)
	if err != nil {
		return nil, err
	}
	_, err = time.Parse("2006-01-02", end)
	if err != nil {
		return nil, err
	}
	var dsn string  
    if cfg.EPrintDbHost == "localhost" {
         dsn = fmt.Sprintf("%s:%s@/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.RepoID)
    } else {
         dsn = fmt.Sprintf("%s:%s@%s/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.EPrintDbHost, cfg.RepoID)
    }
    db, err := sql.Open("mysql", dsn)
    if err != nil {
         return nil, err
    }
    defer db.Close()
	stmt := `SELECT eprintid 
FROM eprint
WHERE (lastmod_year >= ? AND lastmod_month >= ? AND lastmod_day >= ?) AND
      (lastmod_year <= ? AND lastmod_month <= ? AND lastmod_day <= ?)
ORDER BY lastmod_year DESC, lastmod_month DESC, lastmod_day DESC;
`
	// Note we validated the format so we will have three elements if we split on "-"
	p := strings.SplitN(start, "-", 3)
	sYear, sMonth, sDay := p[0], p[1], p[2]
	p = strings.SplitN(end, "-", 3)
	eYear, eMonth, eDay :=  p[0], p[1], p[2]

	rows, err := db.Query(stmt, sYear, sMonth, sDay, eYear, eMonth, eDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []int{}
	for rows.Next() {
		var eprintid int
		if err := rows.Scan(&eprintid); err != nil {
			return nil, err
		}
		ids = append(ids, eprintid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
