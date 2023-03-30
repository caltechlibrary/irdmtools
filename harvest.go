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
	"log"
	"os"
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
	if err := json.Unmarshal(src, &recordIds); err != nil {
		return err
	}
	l := log.New(os.Stderr, "", 1)
	const maxErrors = 100
	eCnt, hCnt, tot := 0, 0, len(recordIds)
	if debug {
		l.Printf("%d records ids", tot)
	}
	t0 := time.Now()
	for i, id := range recordIds {
		rec, rl, err := GetRecord(cfg, id)
		// The rest API seems to have two rate limits, 5000 requests per hour and 500 requests per minute
		if debug && ((i % 10) == 0) {
			l.Printf("retrieved record %q (%d/%d), %s", id, i, tot, rl.String())
		}
		// NOTE: We need to respect rate limits of RDM API
		rl.Throttle(i, tot)
		if err != nil {
			l.Printf("failed to get (%d) %q, %s", i, id, err)
			eCnt++
		} else {
			if c.HasKey(id) {
				if err := c.Update(id, rec); err != nil {
					l.Printf("failed to write %q to %s, %s", id, cName, err)
					eCnt++
				} else {
					hCnt++
				}
			} else {
				if err := c.Create(id, rec); err != nil {
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
		if (i % 250) == 0 {
			l.Printf("%d/%d (%q) records processed to %s %s", i, tot, id, cName, time.Since(t0))
		}
	}
	l.Printf("%d harvested, %d errors, running time %s", hCnt, eCnt, time.Since(t0))
	return nil
}