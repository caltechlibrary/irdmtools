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
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	// Caltech Library packages
	"github.com/caltechlibrary/dataset/v2"
)

func Harvest(cfg *Config, fName string, debug bool) error {
	cName := cfg.CName
	if cName == "" {
		return fmt.Errorf("dataset collection not configured")
	}
	if fName == "" {
		return fmt.Errorf("JSON ids file not set")
	}
	c, err := dataset.Open(cName)
	if err != nil {
		return err
	}
	defer c.Close()
	src, err := os.ReadFile(fName)
	if err != nil {
		return err
	}
	recordIds := []string{}
	if err := JSONUnmarshal(src, &recordIds); err != nil {
		return err
	}
	l := log.New(os.Stderr, "", 1)
	const maxErrors = 100
	eCnt, hCnt, tot := 0, 0, len(recordIds)
	if debug {
		l.Printf("%d record ids", tot)
	}
	t0 := time.Now()
	iTime, reportProgress := time.Now(), false
	//fmt.Printf("DEBUG are we using the RDM REST API? %t\n", (cfg.InvenioDbHost == ""))
	if cfg.InvenioDbHost == "" {
		cfg.rl = new(RateLimit)
	} else {
		cfg.rl = nil
		// Need to open our Postgres connection and defer the closing of it.
		if cfg.pgDB == nil {
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
				return  err
			}
			defer db.Close()
			cfg.pgDB = db
		}
	}
	//fmt.Printf("DEBUG is cfg.rl nil? %t, is cfg.pgDB nil? %t\n", (cfg.rl == nil), (cfg.pgDB == nil))
	for i, id := range recordIds {
		rec, err := GetRecord(cfg, id, false)
		if err != nil {
			msg := fmt.Sprintf("%s", err)
			if strings.HasPrefix(msg, "429 ") {
				cfg.rl.Fprintf(os.Stderr)
			}
			log.Printf("failed to get (%d) %q, %s", i, id, err)
			eCnt++
		} else {
			if c.HasKey(id) {
				if err := c.UpdateObject(id, rec); err != nil {
					log.Printf("failed to write %q to %s, %s", id, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			} else {
				if err := c.CreateObject(id, rec); err != nil {
					log.Printf("failed to write %q to %s, %s", id, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			}
		}
		if eCnt > maxErrors {
			return fmt.Errorf("Stopped, %d errors encountered", eCnt)
		}
		// The rest API seems to have two rate limits, 5000 requests per hour and 500 requests per minute
		if iTime, reportProgress = CheckWaitInterval(iTime, time.Minute); reportProgress || i == 0 {
			log.Printf("%s last id %q (%d/%d) %s", cName, id, i, tot, ProgressETA(t0, i, tot))
		}
		// NOTE: We need to respect rate limits of RDM API if we're using it!
		if cfg.rl != nil {
			fmt.Println("DEBUG we are throttling API access")
			cfg.rl.Throttle(i, tot)
		}
	}
	log.Printf("%d harvested, %d errors, running time %s", hCnt, eCnt, time.Since(t0).Round(time.Second))
	return nil
}

func harvestEPrintRecordsFromMySQL(cfg *Config, recordIds []int, debug bool) error {
	cName := cfg.CName
	c, err := dataset.Open(cName)
	if err != nil {
		return err
	}
	defer c.Close()

	var dsn string	
	if cfg.EPrintDbHost == "localhost" {
		dsn = fmt.Sprintf("%s:%s@/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.RepoID)
	} else {
		dsn = fmt.Sprintf("%s:%s@%s/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.EPrintDbHost, cfg.RepoID)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	l := log.New(os.Stderr, "", 1)
	const maxErrors = 100
	eCnt, hCnt, tot := 0, 0, len(recordIds)
	if debug {
		l.Printf("%d record ids", tot)
	}
	t0 := time.Now()
	iTime, reportProgress := time.Now(), false
	for i, eprintid := range recordIds {
		id := fmt.Sprintf("%d", eprintid)
		if err != nil {
			l.Printf("failed to convert %d to string, %s", eprintid, err)
			continue
		}
		eprint, err := SQLReadEPrint(db, cfg.EPrintHost, eprintid)
		if err != nil {
			l.Printf("failed to get (%d) %d, %s", i, eprintid, err)
			eCnt++
		} else {
			if c.HasKey(id) {
				if err := c.UpdateObject(id, eprint); err != nil {
					l.Printf("failed to write %d to %s, %s", eprintid, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			} else {
				if err := c.CreateObject(id, eprint); err != nil {
					l.Printf("failed to write %d to %s, %s", eprintid, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			}
		}
		if eCnt > maxErrors {
			return fmt.Errorf("Stopped, %d errors encountered", eCnt)
		}
		// The rest API seems to have two rate limits, 5000 requests per hour and 500 requests per minute
		if iTime, reportProgress = CheckWaitInterval(iTime, time.Minute); reportProgress || i == 0 {
			l.Printf("%s last id %d (%d/%d) %s", cName, eprintid, i, tot, ProgressETA(t0, i, tot))
		}
	}
	l.Printf("%d harvested, %d errors, running time %s", hCnt, eCnt, time.Since(t0).Round(time.Second))
	return nil
}

func HarvestEPrintRecords(cfg *Config, recordIds []int, debug bool) error {
	// Check if we can harvest directly from EPrnits MySQL database.
	if cfg.EPrintDbHost != "" && cfg.EPrintDbUser != "" && cfg.EPrintDbPassword != "" {
		return harvestEPrintRecordsFromMySQL(cfg, recordIds, debug)
	}
	cName := cfg.CName
	if cName == "" {
		return fmt.Errorf("dataset collection not configured")
	}
	if len(recordIds) == 0 {
		return fmt.Errorf("no record ids to harvest")
	}
	c, err := dataset.Open(cName)
	if err != nil {
		return err
	}
	defer c.Close()
	l := log.New(os.Stderr, "", 1)
	const maxErrors = 100
	eCnt, hCnt, tot := 0, 0, len(recordIds)
	if debug {
		l.Printf("%d record ids", tot)
	}
	t0 := time.Now()
	iTime, reportProgress := time.Now(), false
	cfg.rl = new(RateLimit)
	timeout := time.Duration(timeoutSeconds)
	for i, eprintid := range recordIds {
		id := strconv.Itoa(eprintid)
		rec, err := GetEPrint(cfg, eprintid, timeout, 3)
		if err != nil {
			msg := fmt.Sprintf("%s", err)
			if strings.HasPrefix(msg, "429 ") {
				cfg.rl.Fprintf(os.Stderr)
			}
			l.Printf("failed to get (%d) %q, %s", i, id, err)
			eCnt++
		} else {
			if c.HasKey(id) {
				if err := c.UpdateObject(id, rec); err != nil {
					l.Printf("failed to write %q to %s, %s", id, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			} else {
				if err := c.CreateObject(id, rec); err != nil {
					l.Printf("failed to write %q to %s, %s", id, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			}
		}
		if eCnt > maxErrors {
			return fmt.Errorf("Stopped, %d errors encountered", eCnt)
		}
		// The rest API seems to have two rate limits, 5000 requests per hour and 500 requests per minute
		if iTime, reportProgress = CheckWaitInterval(iTime, time.Minute); reportProgress || i == 0 {
			l.Printf("%s last id %q (%d/%d) %s", cName, id, i, tot, ProgressETA(t0, i, tot))
		}
	}
	l.Printf("%d harvested, %d errors, running time %s", hCnt, eCnt, time.Since(t0).Round(time.Second))
	return nil
}

func HarvestEPrints(cfg *Config, fName string, debug bool) error {
	cName := cfg.CName
	if cName == "" {
		return fmt.Errorf("dataset collection not configured")
	}
	if fName == "" {
		return fmt.Errorf("JSON ids file not set")
	}
	c, err := dataset.Open(cName)
	if err != nil {
		return err
	}
	defer c.Close()
	src, err := os.ReadFile(fName)
	if err != nil {
		return err
	}
	recordIds := []int{}
	if err := JSONUnmarshal(src, &recordIds); err != nil {
		return err
	}
	l := log.New(os.Stderr, "", 1)
	const maxErrors = 100
	eCnt, hCnt, tot := 0, 0, len(recordIds)
	if debug {
		l.Printf("%d record ids", tot)
	}
	t0 := time.Now()
	iTime, reportProgress := time.Now(), false
	cfg.rl = new(RateLimit)
	timeout := time.Duration(timeoutSeconds)
	for i, eprintid := range recordIds {
		id := strconv.Itoa(eprintid)
		rec, err := GetEPrint(cfg, eprintid, timeout, 3)
		if err != nil {
			msg := fmt.Sprintf("%s", err)
			if strings.HasPrefix(msg, "429 ") {
				cfg.rl.Fprintf(os.Stderr)
			}
			log.Printf("failed to get (%d) %q, %s", i, id, err)
			eCnt++
		} else {
			if c.HasKey(id) {
				if err := c.UpdateObject(id, rec); err != nil {
					log.Printf("failed to write %q to %s, %s", id, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			} else {
				if err := c.CreateObject(id, rec); err != nil {
					log.Printf("failed to write %q to %s, %s", id, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			}
		}
		if eCnt > maxErrors {
			return fmt.Errorf("Stopped, %d errors encountered", eCnt)
		}
		// The rest API seems to have two rate limits, 5000 requests per hour and 500 requests per minute
		if iTime, reportProgress = CheckWaitInterval(iTime, time.Minute); reportProgress || i == 0 {
			log.Printf("%s last id %q (%d/%d) %s", cName, id, i, tot, ProgressETA(t0, i, tot))
		}
	}
	log.Printf("%d harvested, %d errors, running time %s", hCnt, eCnt, time.Since(t0).Round(time.Second))
	return nil
}
