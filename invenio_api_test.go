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
	"encoding/json"
	"fmt"
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
		src, err := json.MarshalIndent(s, "", "    ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(fName, src, 0664); err != nil {
			return err
		}
	}
	return nil
}

func TestConfig(t *testing.T) {
	if cfg == nil {
		t.Errorf("tests are not configured")
		t.FailNow()
	}
	if cfg.InvenioAPI == "" {
		t.Errorf("missing an InvenioAPI URL")
	}
	if cfg.InvenioToken == "" {
		t.Errorf("missing an Invenio Token")
	}
}

func TestQuery(t *testing.T) {
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

func TestGetModifiedRecordIds(t *testing.T) {
	if cfg == nil {
		t.Skipf("Not configured for testing")
	}
	today := time.Now()
	end := today.Format("2006-01-02")
	start := today.AddDate(0, 0, -7).Format("2006-01-02")
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

func TestGetRecordIds(t *testing.T) {
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

func TestGetRecord(t *testing.T) {
	if cfg == nil || idsFName == "" {
		t.Skipf("Not configured for testing")
	}
	fmt.Fprintf(os.Stderr, "DEBUG idsFName -> %q\n", idsFName)
	src, err := os.ReadFile(idsFName)
	if err != nil {
		t.Errorf("failed to read ids from file %q, %s", idsFName, err)
		t.FailNow()
	}
	ids := []string{}
	if err := json.Unmarshal(src, &ids); err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Fprintf(os.Stderr, "DEBUG len(ids) -> %d\n", len(ids))
	for i, id := range ids {
		_, rl, err := GetRecord(cfg, id)
		if err != nil {
			t.Errorf("(%d) GetRecord(cfg, %q) %s\n%s", i, id, err, rl.String())
		}
	}
}
