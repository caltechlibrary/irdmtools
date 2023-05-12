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
	"flag"
	"log"
	"os"
	"path"
	"testing"
)

var (
	cfg      *Config
	useQuery string
	idsFName string
)

func TestMain(m *testing.M) {
	var (
		configFName string
		envPrefix   string
	)
	envPrefix = "TEST_"
	l := log.New(os.Stderr, "", 1)
	// call flag.Parse() here if TestMain uses flags
	flag.StringVar(&configFName, "config", configFName, "config file for testing")
	flag.StringVar(&useQuery, "q", useQuery, "use this test query")
	flag.StringVar(&idsFName, "ids", idsFName, "use this JSON id list for testing")
	flag.Parse()

	if cfg == nil {
		cfg = NewConfig()
	}
	if configFName != "" {
		l.Printf("loading %s\n", configFName)
		if err := cfg.LoadConfig(configFName); err != nil {
			l.Fatal(err)
		}
	}
	if cfg.Debug {
		l.Printf("loading env using prefix %s\n", envPrefix)
	}
	if err := cfg.LoadEnv(envPrefix); err != nil {
		l.Fatal(err)
	}
	cfg.Debug = true
	if cfg.InvenioAPI == "" {
		l.Fatal("invenio api not configured")
	}
	if cfg.InvenioToken == "" {
		l.Fatal("invenio troken not configured")
	}
	if idsFName == "" {
		idsFName = path.Join("testdata", "test_record_ids.json")
	}

	if _, err := os.Stat(idsFName); os.IsNotExist(err) {
		l.Printf("skipping testing Test02GetRecord, no ids file")
		l.Printf("skipping testing Test03Harvest, no ids file")

	}
	os.Exit(m.Run())
}
