// rdmds2citations is a command line program that will convert a dataset collection of EPrint 3.3 records into a citations dataset collection
//
// @author R. S. Doiel, <rsdoiel@caltech.edu>
// @author Tom Morrell, <tmorrell@caltech.edu>
//
// Copyright (c) 2024, Caltech
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
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	// Caltech Library packages
	"github.com/caltechlibrary/irdmtools"
)

var (
	helpText = `%{app_name}(1) irdmtools user manual | version {version} {release_hash}
% R. S. Doiel and Tom Morrell
% {release_date}

# NAME

{app_name}

# SYNOPSIS

{app_name} [OPTIONS] RDM_DS CITATION_DS [RECORD_ID]

# DESCRIPTION

{app_name} is a Caltech Library oriented command line application
that takes an dataset collection of RDM records and converts then
to a citations dataset collection. It can do so for a single record id
or read a JSON list of record ids to migrate.


RDM_DS is the dataset collection holding the eprint records.

CITATION_DS is the dataset collection where the citation formatted
objects will be written.

# OPTIONS

-help
: display help

-license
: display license

-version
: display version

-config
: provide a path to an alternate configuration file (e.g. "irdmtools.json")

-ids JSON_ID_FILE
: read ids from a file.

-keys
: works from a key list, one per line. Maybe file or standard input (use filename as "-")

-prefix
: Applies a prefix before the provided key when saving a record. E.g. `+"`"+`-prefix caltechauthors`+`"`+`

-host
: Set the hostname of base url to for reference records (e.g. authors.library.caltech.edu). Can also be set via the environment as RDM_URL.

# ENVIRONMENT 

Some settings can be picked from the environment.

RDM_URL
: Sets the URL to the RDM is installed (e.g. "https://authors.library.caltech.edu").

# EXAMPLE

Example of a dataset collection called "authors.ds" and "data.ds"
RDM records aggregated into a "citation.ds" dataset
collection using prefixes and the source repository ids.

~~~shell
{app_name} -prefix caltechauthors \
           -host authors.library.caltech.edu \
		   authors.ds citation.ds k3tpc-ga970
{app_name} -prefix caltechdata \
           -host data.caltech.edu \
		   data.ds citation.ds zzj7r-61978
~~~

`
)

// getDSIds will read in a JSON list of RDM ids from either standard
// input or a JSON file.
func getDSIds(idsFName string) ([]string, error) {
	var err error
	in := os.Stdin
  	if idsFName != "-" {
   		in, err = os.Open(idsFName)
   		if err != nil {
   			return nil, err
   		}
   		defer in.Close()
   	}
   	src, err := io.ReadAll(in)
   	if err != nil {
		return nil, err
   	}
   	ids := []string{}
   	if err := irdmtools.JSONUnmarshal(src, &ids); err != nil {
		return nil, err
   	}
	return ids, nil
}

// getKeyList will read in a list if keys one per line and return an array of keys from
// a dataset collectiojn to migrate.
func getKeyList(keysFName string) ([]string, error) {
	var err error
	in := os.Stdin
  	if keysFName != "-" {
   		in, err = os.Open(keysFName)
   		if err != nil {
   			return nil, err
   		}
   		defer in.Close()
   	}
	keys := []string{}
	r := bufio.NewReader(in)
	i := 0
	for {
		s, err := r.ReadString('\n')
		if s != "" {
			keys = append(keys, strings.TrimSpace(s))
		}
		i++
		if err != nil {
			break
		}
	}
	log.Printf("%d keys to process, %d lines read", len(keys), i)
	return keys, nil
}


func main() {
	appName := path.Base(os.Args[0])
	// NOTE: The following are set when version.go is generated
	version := irdmtools.Version
	releaseDate := irdmtools.ReleaseDate
	releaseHash := irdmtools.ReleaseHash
	fmtHelp := irdmtools.FmtHelp

	showHelp, showVersion, showLicense := false, false, false
	idsFName, keysFName, repoHost, prefix := "", "", "", ""
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.StringVar(&idsFName, "ids", idsFName, "read ids from a file")
	flag.StringVar(&keysFName, "keys", keysFName, "read keys from a file or standard input")
	flag.StringVar(&prefix, "prefix", prefix, "apply this prefix to keys being imported before saving")
	flag.StringVar(&repoHost, "host", repoHost, "repository hostname, used to form URL to original record")

	flag.Parse()
	args := flag.Args()

	in := os.Stdin
	out := os.Stdout
	eout := os.Stderr

	if showHelp {
		fmt.Fprintf(out, "%s\n", fmtHelp(helpText, appName, version, releaseDate, releaseHash))
		os.Exit(0)
	}
	if showVersion {
		fmt.Fprintf(out, "%s %s %s\n", appName, version, releaseHash)
		os.Exit(0)
	}
	if showLicense {
		fmt.Fprintf(out, "%s\n", irdmtools.LicenseText)
		os.Exit(0)
	}
	var (
		dsIds []string
		err error
	)
	if idsFName != ""  {
		dsIds, err = getDSIds(idsFName)
		if err != nil {
   			fmt.Fprintf(eout, "%s\n", err)
   			os.Exit(1)
		}
	}
	if keysFName != "" {
		dsIds, err = getKeyList(keysFName)
		if err != nil {
   			fmt.Fprintf(eout, "%s\n", err)
   			os.Exit(1)
		}
	}
	os.Exit(irdmtools.RunRdmDSToCitationDS(in, out, eout, args, repoHost, prefix, dsIds))
}
