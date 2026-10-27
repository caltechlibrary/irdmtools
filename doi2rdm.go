// / irdmtools is a package for working with institutional repositories and
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
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	// 3rd Party packages
	"gopkg.in/yaml.v3"

	// Caltech Library packages
	"github.com/caltechlibrary/crossrefapi"
	"github.com/caltechlibrary/simplified"
)

const (
	EXIT_OK = 0
	ENOENT = 2
	ENOEXEC = 8
	EAGAIN = 11
	// The workspace exit codes (DR-0014), used by the failures added with the
	// symbols option so that a script can tell them from "DOI not found" (2).
	// The codes above are the older ones and are not changed here.
	EX_NOINPUT = 66 // a named input, such as the symbols file, is missing
	EX_CONFIG = 78 // a configuration file is present but wrong
)

// Doi2Rdm holds the configuration for doi2rdm cli.
type Doi2Rdm struct {
	Cfg *Config

	// queryCrossRef fetches a CrossRef work. It is nil outside tests, which
	// leaves QueryCrossRefWork, the live query, in place.
	queryCrossRef func(cfg *Config, doi string, options *Doi2RdmOptions) (*crossrefapi.Works, error)

	// queryDataCite fetches a DataCite object. It is nil outside tests, which
	// leaves QueryDataCiteObject, the live query, in place.
	queryDataCite func(cfg *Config, doi string, options *Doi2RdmOptions) (map[string]interface{}, error)
}

type Doi2RdmOptions struct {
	MailTo              string            `json:"mailto,omitempty" yaml:"mailto,omitempty"`
	Download            bool              `json:"download,omitempty" yaml:"download,omitempty"`
	DotInitials         bool              `json:"dot_initials,omitempty" yaml:"dot_initials,omitempty"`
	ContributorTypes    map[string]string `json:"contributor_types,omitemptpy" yaml:"contributor_types,omitempty"`
	ResourceTypes       map[string]string `json:"resource_types,omitempty" yaml:"resource_types,omitempty"`
	DoiPrefixPublishers map[string]string `json:"doi_prefix_publishers,omitempty" yaml:"doi_prefix_publishers,omitempty"`
	ISSNJournals        map[string]string `json:"issn_journals,omitempty" yaml:"issn_journals,omitempty"`
	ISSNPublishers      map[string]string `json:"issn_publishers,omitempty" yaml:"issn_publishers,omitempty"`
	Debug               bool              `json:"debug,omitempty" yaml:"debug,omitempty"`
	// Symbols is the path of a crossrefapi override file for the table that
	// turns CrossRef markup, entities and MathML into plain text. A relative
	// path is read relative to the options file. Empty means the built in table.
	Symbols string `json:"symbols,omitempty" yaml:"symbols,omitempty"`

	// normalizer is built from Symbols the first time it is needed.
	normalizer *crossrefapi.Normalizer
}

// crossRefNormalizer returns the normalizer for these options, building it
// from the Symbols file on first use (the built in table when Symbols is
// empty). A built normalizer never changes, but this method is not safe to
// call from several goroutines before the first call has returned.
//
// @returns {*crossrefapi.Normalizer} the normalizer to convert CrossRef text with
// @returns {error} an error naming the Symbols file if it is missing or invalid
// @example
//
//	n, err := options.crossRefNormalizer()
func (options *Doi2RdmOptions) crossRefNormalizer() (*crossrefapi.Normalizer, error) {
	if options.normalizer == nil {
		n, err := crossrefapi.NewNormalizer(options.Symbols)
		if err != nil {
			return nil, err
		}
		options.normalizer = n
	}
	return options.normalizer, nil
}

var (
	DefaultDoi2RdmOptionsYAML = []byte(`# This YAML file controls the mappings of 
# CrossRef records to RDM records values. It is based on the practice
# of Caltech Library in the development of CaltechAUTHORS and CaltechTHESIS
# over the last decades.
#
# Set the mail to used when connecting to CrossRef. This is usually the
# email address for our organization but could be for a person.
# It is the email address CrossRef will use if you're causing a problem
# and they need you to stop.
#
#mailto: jane.doe@example.edu
mailto: helpdesk@library.caltech.edu
# Add a period after initials is missing
dot_initials: true
# Map the CrossRef type to RDM type
contributor_types:
  author: author
  editor: editor
  reviewer: reviewer
  review-assistent: other
  stats-reviewer: other
  reader: other
  translator: translator
# Map the CrossRef and DataCite resource type to the RDM type
resource_types:
  article: publication-article
  Preprint: publication-preprint
  preprint: publication-preprint
  journal-article: publication-article
  book: publication-book
  book_section: publication-section
  book-chapter: publication-section
  conference_item: conference-paper
  proceedings-article: conference-paper
  dataset: dataset
  experiment: publication-deliverable
  journal_issue: publication-issue
  lab_notes: labnotebook
  monograph: publication-report
  oral_history: publication-oralhistory
  patent: publication-patent
  software: software
  teaching_resource: teachingresource
  thesis: publication-thesis
  video: video
  website: other
  other: other
  image: other
  report: publication-workingpaper
  report-component: publication-workingpaper
  posted-content: publication-preprint
  DataPaper: publication-datapaper
  Text: publication-other
# Mapping DOI prefixes to Publisher names (used to normalize publisher names)
doi_prefix_publishers:
# Mapping ISSN prefixes to Journals (used to normalize journal titles names)
issn_journals:
# Mapping ISSN prefixes to Publishers (used to normalize publisher names)
issn_publishers:
# CrossRef titles and abstracts hold markup, character entities and MathML.
# They are converted to plain text with a built in table of symbols. To change
# the table, name a crossrefapi override file here. A relative path is read
# relative to this options file. Leave it out to use the built in table.
#symbols: symbols.json
`)
)

// Configure reads the configuration file and environtment
// initialing the Cfg attribute of a Doi2Rdm object. It returns an error
// if problem were encounter.
//
// ```
//
//	app := new(irdmtools.Doi2Rdm)
//	if err := app.Configure("irdmtools.yaml", "TEST_"); err != nil {
//	   // ... handle error ...
//	}
//	fmt.Printf("Invenio RDM API UTL: %q\n", app.Cfg.IvenioAPI)
//	fmt.Printf("Invenio RDM token: %q\n", app.Cfg.InvenioToken)
//
// ```
func (app *Doi2Rdm) Configure(configFName string, envPrefix string, debug bool) error {
	if app == nil {
		app = new(Doi2Rdm)
	}
	cfg := NewConfig()
	// Load the config file if name isn't an empty string
	if configFName != "" {
		err := cfg.LoadConfig(configFName)
		if err != nil {
			return err
		}
	}
	// Merge settings from the environment
	if err := cfg.LoadEnv(envPrefix); err != nil {
		return err
	}
	app.Cfg = cfg
	if debug {
		app.Cfg.Debug = true
	}
	// Make sure we have a minimal useful configuration
	if app.Cfg.InvenioAPI == "" || app.Cfg.InvenioToken == "" {
		return fmt.Errorf("RDM_URL or RDM_TOK available")
	}
	return nil
}

// loadCrossRefOptions reads the options YAML (the built in default when
// optionFName is empty) for a CrossRef run and builds the normalizer it names,
// so a bad options or symbols file stops the run before any record is fetched.
// A relative Symbols path is read relative to the options file.
//
// @param optionFName {string} the options YAML file name, or "" for the default
// @returns {*Doi2RdmOptions} the options, nil on error
// @returns {int} the exit code: ENOENT or ENOEXEC for the options file, EX_NOINPUT for a missing symbols file, EX_CONFIG for an invalid one
// @returns {error} an error naming the file that could not be read or parsed
// @example
//
//	options, exitCode, err := app.loadCrossRefOptions("options.yaml")
func (app *Doi2Rdm) loadCrossRefOptions(optionFName string) (*Doi2RdmOptions, int, error) {
	var (
		err error
		src []byte
	)
	src = DefaultDoi2RdmOptionsYAML
	if optionFName != "" {
		src, err = os.ReadFile(optionFName)
		if err != nil {
			return nil, ENOENT, err
		}
	}
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(src, &options); err != nil {
		return nil, ENOEXEC, err
	}
	if app.Cfg.Debug {
		options.Debug = app.Cfg.Debug
	}
	if options.MailTo == "" {
		//mailTo = fmt.Sprintf("%s@%s", os.Getenv("USER"), os.Getenv("HOSTNAME"))
		options.MailTo = "helpdesk@library.caltech.edu"
	}
	if options.Symbols != "" && optionFName != "" && !filepath.IsAbs(options.Symbols) {
		options.Symbols = filepath.Join(filepath.Dir(optionFName), options.Symbols)
	}
	if _, err := options.crossRefNormalizer(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, EX_NOINPUT, err
		}
		return nil, EX_CONFIG, err
	}
	return options, EXIT_OK, nil
}

// RunCrossRefToRdm implements the doi2rdm cli behaviors using the CrossRef service.
// With the exception of the "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//	app := new(irdmtools.Doi2Rdm)
//	// Load irdmtools settings
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//	// If options are provided then we need to set the filename
//	optionsFName := "doi2rdm.yaml"
//	doi := "10.3847/1538-3881/ad2765"
//	src, exitCode, err := app.Run(os.Stdin, os.Stdout, os.Stderr, optionFName, doi, "", false)
//	if err != nil {
//	    // ... handle error ...
//      os.Exit(exitCode)
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *Doi2Rdm) RunCrossRefToRdm(in io.Reader, out io.Writer, eout io.Writer, optionFName, doi string, diffFName string) (int, error) {
	options, exitCode, err := app.loadCrossRefOptions(optionFName)
	if err != nil {
		return exitCode, err
	}
	var oRecord *simplified.Record
	if diffFName != "" {
		oWork := new(crossrefapi.Works)
		src, err := os.ReadFile(diffFName)
		if err != nil {
			return ENOENT, err
		}
		if err := JSONUnmarshal(src, &oWork); err != nil {
			return ENOEXEC, err
		}
		oRecord, err = CrosswalkCrossRefWork(app.Cfg, oWork, options)
		if err != nil {
			return ENOEXEC, err
		}
	}
	query := QueryCrossRefWork
	if app.queryCrossRef != nil {
		query = app.queryCrossRef
	}
	nWork, err := query(app.Cfg, doi, options)
	if err != nil {
		return ENOENT, err
	}
	return app.renderCrossRefWork(out, eout, doi, nWork, options, oRecord, diffFName != "")
}

// renderCrossRefWork crosswalks a CrossRef work that has been fetched and
// writes the record as JSON to out, or in diff mode the difference from the old
// record, and nothing else. What the text normalizer could not convert cleanly
// is written to eout, one line per warning: "doi2rdm: DOI: field: code:
// fragment". Warnings never change out or the result. In diff mode only the
// new work's warnings are written.
//
// @param out {io.Writer} where the record JSON is written
// @param eout {io.Writer} where the warning lines are written, may be nil
// @param doi {string} the DOI of the work, named in each warning line
// @param nWork {*crossrefapi.Works} the work to crosswalk
// @param options {*Doi2RdmOptions} the crosswalk options
// @param oRecord {*simplified.Record} the old record to diff against, used when diff is true
// @param diff {bool} true to write the diff of oRecord and the new record
// @returns {int} EXIT_OK, or ENOEXEC if the work cannot be crosswalked or the JSON made
// @returns {error} an error describing the failure
// @example
//
//	exitCode, err := app.renderCrossRefWork(os.Stdout, os.Stderr, doi, work, options, nil, false)
func (app *Doi2Rdm) renderCrossRefWork(out io.Writer, eout io.Writer, doi string, nWork *crossrefapi.Works, options *Doi2RdmOptions, oRecord *simplified.Record, diff bool) (int, error) {
	var (
		src []byte
		err error
	)
	nRecord, warnings, err := CrosswalkCrossRefWorkWarnings(app.Cfg, nWork, options)
	if err != nil {
		return ENOEXEC, err
	}
	if diff {
		src, err = oRecord.DiffAsJSON(nRecord)
	} else {
		src, err = JSONMarshalIndent(nRecord, "", "    ")
	}
	if err != nil {
		return ENOEXEC, err
	}
	if eout != nil {
		for _, w := range warnings {
			fmt.Fprintln(eout, w.Line(doi))
		}
	}
	fmt.Fprintf(out, "%s\n", src)
	return EXIT_OK, nil
}

// RunDataCiteToRdm implements the doi2rdm cli behaviors using the DataCite service.
// With the exception of the "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//		app := new(irdmtools.Doi2Rdm)
//	 // Load irdmtools settings
//		if err := app.LoadConfig("irdmtools.json"); err != nil {
//		   // ... handle error ...
//		}
//	 // If options are provided then we need to set the filename
//	 optionsFName := "doi2rdm.yaml"
//		doi := "10.48550/arXiv.2104.02480"
//		src, err := app.RunDataCiteToRdm(os.Stdin, os.Stdout, os.Stderr, optionFName, doi, "", false)
//		if err != nil {
//		    // ... handle error ...
//		}
//		fmt.Printf("%s\n", src)
//
// ```
func (app *Doi2Rdm) RunDataCiteToRdm(in io.Reader, out io.Writer, eout io.Writer, optionFName, doi string, diffFName string) (int, error) {
	var (
		err error
		src []byte
	)
	src = DefaultDoi2RdmOptionsYAML
	if optionFName != "" {
		src, err = os.ReadFile(optionFName)
		if err != nil {
			return ENOENT, err
		}
	}
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(src, &options); err != nil {
		return ENOEXEC, err
	}
	if app.Cfg.Debug {
		options.Debug = app.Cfg.Debug
	}
	if options.MailTo == "" {
		//mailTo = fmt.Sprintf("%s@%s", os.Getenv("USER"), os.Getenv("HOSTNAME"))
		options.MailTo = "helpdesk@library.caltech.edu"
	}
	var (
		oRecord *simplified.Record
		nRecord *simplified.Record
	)
	if diffFName != "" {
		object := map[string]interface{}{}
		src, err := os.ReadFile(diffFName)
		if err != nil {
			return ENOENT, err
		}
		if err := JSONUnmarshal(src, &object); err != nil {
			return ENOEXEC, err
		}
		oRecord, err = CrosswalkDataCiteObject(app.Cfg, object, options)
		if err != nil {
			return ENOEXEC, err
		}
	}
	query := QueryDataCiteObject
	if app.queryDataCite != nil {
		query = app.queryDataCite
	}
	nWork, err := query(app.Cfg, doi, options)
	if err != nil {
		return ENOENT, err
	}
	if len(nWork) == 0 {
		return ENOENT, fmt.Errorf("not data received for %q", doi)
	}
	nRecord, err = CrosswalkDataCiteObject(app.Cfg, nWork, options)
	if err != nil {
		return ENOEXEC, err
	}
	if diffFName != "" {
		src, err = oRecord.DiffAsJSON(nRecord)
	} else {
		src, err = JSONMarshalIndent(nRecord, "", "    ")
	}
	if err != nil {
		return ENOEXEC, err
	}
	fmt.Fprintf(out, "%s\n", src)
	return EXIT_OK, nil
}

// RunDoiToRDMCombined implements the doi2rdm cli behaviors using the CrossRead and DataCite service.
// With the exception of the "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//		app := new(irdmtools.Doi2Rdm)
//	 // Load irdmtools settings
//		if err := app.LoadConfig("irdmtools.json"); err != nil {
//		   // ... handle error ...
//		}
//	 // If options are provided then we need to set the filename
//	 optionsFName := "doi2rdm.yaml"
//		doi := "10.48550/arXiv.2104.02480"
//		src, err := app.RunDoiToRdmCombined(os.Stdin, os.Stdout, os.Stderr, optionFName, doi, "", false)
//		if err != nil {
//		    // ... handle error ...
//		}
//		fmt.Printf("%s\n", src)
//
// ```
func (app *Doi2Rdm) RunDoiToRdmCombined(in io.Reader, out io.Writer, eout io.Writer, optionFName, doi string, diffFName string) (int, error) {
	// Do we have an arXiv id?
	if strings.HasPrefix(strings.ToLower(doi), "arxiv:") {
		return app.RunDataCiteToRdm(in, out, eout, optionFName, doi, diffFName)
	}
	if crCode, crErr := app.RunCrossRefToRdm(in, out, eout, optionFName, doi, diffFName); crErr != nil  {
		// A symbols file that is missing or invalid is a problem with the
		// configuration, not with the DOI. Report it now: DataCite does not
		// read that file and would answer "not found", hiding it.
		if crCode == EX_NOINPUT || crCode == EX_CONFIG {
			return crCode, crErr
		}
		// Then try DataCiteToRdm
		if exitCode, dcErr := app.RunDataCiteToRdm(in, out, eout, optionFName, doi, diffFName); dcErr != nil {
			return exitCode, fmt.Errorf("crossref: %s, datacite: %s", crErr, dcErr)
		}
	}
	return EXIT_OK, nil
}
