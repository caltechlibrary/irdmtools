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
	"fmt"
	"math/rand"
	"database/sql"
	"os"
	"path"
	"testing"
	"time"
)

const (
	maxIdListSize = 1000
)

func saveIdsFile(fName string, ids []string, maxLength int) error {
	if _, err := os.Stat(idsFName); os.IsNotExist(err) {
		dName := path.Dir(idsFName)
		if _, err := os.Stat(dName); os.IsNotExist(err) {
			os.MkdirAll(dName, 0775)
		}
		s := ids[:]
		if len(s) > maxLength {
			s = s[0:maxLength]
		}
		if len(s) == 0 {
			return fmt.Errorf("no ids to save")
		}
		src, err := JSONMarshalIndent(s, "", "    ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(fName, src, 0664); err != nil {
			return err
		}
	}
	return nil
}

func Test01Config(t *testing.T) {
	if cfg == nil {
		t.Errorf("tests are not configured")
		t.FailNow()
	}
	if err := cfg.LoadEnv("TEST_"); err != nil  {
		t.Error(err)
	}
	// For go tooling all Invenio RDM through DB.
	/*
	if cfg.InvenioAPI != "" && cfg.InvenioToken == "" {
		t.Errorf("missing an Invenio API Token")
	}
	*/
	if cfg.InvenioDbHost == "" && cfg.InvenioDSN == "" {
		src, _ := JSONMarshalIndent(cfg, "", "    ")
		t.Errorf("expected either Invenio Db Host or DSN to be set via TEST_* environment variables, %s", src)
	}
}

/* NOTE: I am abondoning the Invenio RDM API in favor of direct Postgres access do to rate limiting challenges.
func Test01Query(t *testing.T) {
	if useQuery == "" {
		useQuery = "gravity"
	}
	if cfg == nil {
		t.Skipf("not configured for testing")
	}
	t.Logf("using useQuery %q", useQuery)
	records, err := Query(cfg, useQuery, "bestmatch")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(records) == 0 {
		t.Errorf("failed to find any records")
		t.FailNow()
	}
	ids := []string{}
	for i, rec := range records {
		if id, ok := rec["id"]; ok == false {
			t.Errorf("record (%d) is missing id, %+v", i, rec)
			t.FailNow()
		} else {
			ids = append(ids, fmt.Sprintf("%s", id))
		}
	}
	if idsFName != "" {
		if err := saveIdsFile(idsFName, ids, maxIdListSize); err != nil {
			t.Errorf("failed to write %q, %s", idsFName, err)
			t.FailNow()
		}
	}
}
*/

func Test01GetModifiedRecordIds(t *testing.T) {
	if cfg == nil {
		t.Skipf("Not configured for testing")
	}
	today := time.Now()
	end := today.Format("2006-01-02")
	start := today.AddDate(0, 0, -3).Format("2006-01-02")
	ids, err := GetModifiedRecordIds(cfg, start, end)
	if err != nil {
		t.Error(err)
	}
	for i, id := range ids {
		s := fmt.Sprintf("%T", id)
		if s != "string" {
			t.Errorf("expected (%d) a string, got %s", i, s)
			t.FailNow()
		}
	}
	if idsFName != "" {
		if err := saveIdsFile(idsFName, ids, maxIdListSize); err != nil {
			t.Errorf("failed to write %q, %s", idsFName, err)
			t.FailNow()
		}
	}
}

func Test01GetRecordIds(t *testing.T) {
	if cfg == nil {
		t.Skipf("Not configured for testing")
	}
	ids, err := GetRecordIds(cfg)
	if err != nil {
		t.Error(err)
	}
	for i, id := range ids {
		s := fmt.Sprintf("%T", id)
		if s != "string" {
			t.Errorf("expected (%d) a string, got %s", i, s)
			t.FailNow()
		}
	}
	if idsFName != "" {
		if err := saveIdsFile(idsFName, ids, maxIdListSize); err != nil {
			t.Errorf("failed to write %q, %s", idsFName, err)
			t.FailNow()
		}
	}
}

func Test02GetRecord(t *testing.T) {
	if cfg == nil || idsFName == "" {
		t.Skipf("Not configured for testing")
	}
	cfg := new(Config)
	if err := cfg.LoadEnv("TEST_"); err != nil {
		t.Error(err)
	}
	if cfg.RepoID == "" {
		t.Errorf("Missing repo id, aborting")
		t.FailNow()
	}
	if cfg.InvenioDbHost == "" {
		t.Errorf("Missing Invenio Db Hostname")
	}
	connstr := cfg.MakeDSN()
	if connstr == "" {
		t.Errorf("cfg.MakeDSN() returned empty dsn")
		t.FailNow()
	}
	db, err := sql.Open("postgres", connstr)
	if err != nil {
		t.Errorf("failed to open postgres, %s", err)
		t.FailNow()
	}
	defer db.Close()
	cfg.pgDB = db
	if connstr == "" {
		t.Errorf("cfg.MakeDSN() returned empty connection string, aborting test")
		t.FailNow()
	}
	src, err := os.ReadFile(idsFName)
	if err != nil {
		t.Errorf("failed to read ids from file %q, %s", idsFName, err)
		t.FailNow()
	}
	ids := []string{}
	if err := JSONUnmarshal(src, &ids); err != nil {
		t.Error(err)
		t.FailNow()
	}
	// Randomize the order of the ids before running GetRecord test.
	rand.Shuffle(len(ids), func(i int, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})
	// Take the first hundred to test
	test_ids := ids[0:100]
	t0 := time.Now()
	iTime := time.Now()
	tot := len(test_ids)
	reportProgress := false
	for i, id := range test_ids {
		_, err := GetRecord(cfg, id, false)
		if err != nil {
			t.Errorf("(%d) GetRecord(cfg, %q, false) %s for %s", i, id, err, connstr)
			t.FailNow()
		}
		if iTime, reportProgress = CheckWaitInterval(iTime, (15 * time.Second)); reportProgress || i == 0 {
			fmt.Fprintf(os.Stderr, "%s %s\n", ProgressIPS(t0, i, time.Second), ProgressETA(t0, i, tot))
		}
	}
}
