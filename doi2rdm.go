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
	"fmt"
	"io"
	"os"

	// 3rd Party packages
	"gopkg.in/yaml.v3"

	// Caltech Library packages
	"github.com/caltechlibrary/crossrefapi"
	//"github.com/caltechlibrary/dataciteapi"
	"github.com/caltechlibrary/simplified"
)

// Doi2Rdm holds the configuration for doi2rdm cli.
type Doi2Rdm struct {
	Cfg *Config
}

type Doi2RdmOptions struct {
	MailTo string `json:"mailto,omitempty" yaml:"mailto"`
	Download bool `json:"download,omitempty" yaml:"download"`
	DotInitials bool `json:"dot_initials,omitempty" yaml:"dot_initials"`
	ContributorTypes map[string]string `json:"contributor_types,omitemptpy" yaml:"contributor_types"`
	ResourceTypes map[string]string `json:"resource_types,omitempty" yaml:"resource_types"`
	DoiPrefixPublishers map[string]string `json:"doi_prefix_publishers,omitempty" yaml:"doi_prefix_publishers"`
	ISSNPublishers map[string]string `json:"issn_publishers,omitempty" yaml:"issn_publishers"`
	Debug bool `json:"debug,omitempty" yaml:"debug"`
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
# Map the CrossRef resource type to the RDM type
resource_types:
  article: publication-article
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
# Mapping DOI prefixes to Publisher names (used to normalize publisher names)
doi_prefix_publishers:
  10.1103: American Physical Society
  10.1063: American Institute of Physics
  10.1039: Royal Society of Chemistry
  10.1242: Company of Biologists
  10.1073: PNAS
  10.1109: IEEE
  10.2514: AIAA
  10.1029: AGU (pre-Wiley hosting)
  10.1093: MNRAS
  10.1046: Geophysical Journal International
  10.1175: American Meteorological Society
  10.1083: Rockefeller University Press
  10.1084: Rockefeller University Press
  10.1085: Rockefeller University Press
  10.26508: Rockefeller University Press
  10.1371: PLOS
  10.5194: European Geosciences Union
  10.1051: EDP Sciences
  10.2140: Mathematical Sciences Publishers
  10.1074: ASBMB
  10.1091: ASCB
  10.1523: Society for Neuroscience
  10.1101: Cold Spring Harbor
  10.1128: American Society for Microbiology
  10.1115: ASME
  10.1061: ASCE
  10.1038: Nature
  10.1126: Science
  10.1021: American Chemical Society
  10.1002: Wiley
  10.1016: Elsevier
# Mapping ISSN prefixes to Publisher names (used to normalize publisher names)
issn_publishers:
  0141-8130: Elsevier
  0266-2671: Cambridge University Press
  0166-5316: Elsevier
  1610-1928: S. Hirzel Verlag
  2157-6564: AlphaMed Press
  0166-218X: Elsevier
  0288-5514: Yōdosha
  0003-2654: Royal Society of Chemistry
  0093-6405: American Society of Civil Engineers
  0181-0529: Presses de l'ecole nationale des ponts et chaussées
  1047-8477: Elsevier
  1868-2529: Springer
  1935-8237: Now Publishers
  2041-8205: American Astronomical Society
  0103-9733: Sociedade Brasileira de Física
  1063-5203: Elsevier
  1541-1672: IEEE
  2405-6014: Elsevier
  0036-0279: Turpion
  0264-410X: Elsevier
  1811-5209: Mineralogical Society of America
  2055-7434: Nature Publishing Group
  0033-5533: MIT Press
  1536-1276: IEEE
  1949-3584: Wiley
  0037-6604: Sky Publishing Corp.
  0145-2126: Elsevier
  1010-6030: Elsevier
  1751-6161: Elsevier
  0021-8723: Oxford University Press
  0026-8232: University of Chicago Press
  0018-2370: Wiley-Blackwell
  0026-2692: Elsevier
  0885-8950: IEEE
  1532-3978: Association of Moving Image Archivists
  0002-2470: Air Pollution Control Association
  1941-4889: American Institute of Mathematical Sciences
  0095-2583: John Wiley
  1389-1723: Elsevier
  0013-0117: Wiley
  0042-6636: Virginia Historical Society
  0091-6749: Elsevier
  2058-5985: Oxford University Press
  0021-3640: American Institute of Physics
  1064-8275: SIAM
  0300-9580: Royal Society of Chemistry
  1465-5411: BioMed Central
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

// RunCrossRefToRdm implements the doi2rdm cli behaviors using the CrossRef service.
// With the exception of the "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//	app := new(irdmtools.Doi2Rdm)
//  // Load irdmtools settings
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//  // If options are provided then we need to set the filename
//  optionsFName := "doi2rdm.yaml"
//	doi := "10.48550/arXiv.2104.02480"
//	src, err := app.Run(os.Stdin, os.Stdout, os.Stderr, optionFName, doi, "", false)
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *Doi2Rdm) RunCrossRefToRdm(in io.Reader, out io.Writer, eout io.Writer, optionFName, doi string, diffFName string) error {
	var (
		err error
		src []byte
	)
	src = DefaultDoi2RdmOptionsYAML
	if optionFName != "" {
		src, err = os.ReadFile(optionFName)
		if err != nil {
			return err
		}
	}
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(src, &options); err != nil {
		return err
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
		oWork := new(crossrefapi.Works)
		src, err := os.ReadFile(diffFName)
		if err != nil {
			return err
		}
		if err := JSONUnmarshal(src, &oWork); err != nil {
			return err
		}
		oRecord, err = CrosswalkCrossRefWork(app.Cfg, oWork, options)
		if err != nil {
			return err
		}
	}
	nWork, err := QueryCrossRefWork(app.Cfg, doi, options)
	if err != nil {
		return err
	}
	nRecord, err = CrosswalkCrossRefWork(app.Cfg, nWork, options)
	if err != nil {
		return err
	}
	if diffFName != "" {
		src, err = oRecord.DiffAsJSON(nRecord)
	} else {
		src, err = JSONMarshalIndent(nRecord, "", "    ")
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s\n", src)
	return nil
}

/*
// RunDataCiteToRdm implements the doi2rdm cli behaviors using the DataCite service.
// With the exception of the "setup" action you should call `app.LoadConfig()` before execute
// Run.
//
// ```
//
//	app := new(irdmtools.Doi2Rdm)
//  // Load irdmtools settings
//	if err := app.LoadConfig("irdmtools.json"); err != nil {
//	   // ... handle error ...
//	}
//  // If options are provided then we need to set the filename
//  optionsFName := "doi2rdm.yaml"
//	doi := "10.48550/arXiv.2104.02480"
//	src, err := app.RunDataCiteToRdm(os.Stdin, os.Stdout, os.Stderr, optionFName, doi, "", false)
//	if err != nil {
//	    // ... handle error ...
//	}
//	fmt.Printf("%s\n", src)
//
// ```
func (app *Doi2Rdm) RunDataCiteToRdm(in io.Reader, out io.Writer, eout io.Writer, optionFName, doi string, diffFName string) error {
	var (
		err error
		src []byte
	)
	src = DefaultDoi2RdmOptionsYAML
	if optionFName != "" {
		src, err = os.ReadFile(optionFName)
		if err != nil {
			return err
		}
	}
	options := new(Doi2RdmOptions)
	if err := yaml.Unmarshal(src, &options); err != nil {
		return err
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
		oWork := new(dataciteapi.Works)
		src, err := os.ReadFile(diffFName)
		if err != nil {
			return err
		}
		if err := JSONUnmarshal(src, &oWork); err != nil {
			return err
		}
		oRecord, err = CrosswalkDataCiteWork(app.Cfg, oWork, options)
		if err != nil {
			return err
		}
	}
	nWork, err := QueryDataCiteWork(app.Cfg, doi, options)
	if err != nil {
		return err
	}
	nRecord, err = CrosswalkDataCiteWork(app.Cfg, nWork, options)
	if err != nil {
		return err
	}
	if diffFName != "" {
		src, err = oRecord.DiffAsJSON(nRecord)
	} else {
		src, err = JSONMarshalIndent(nRecord, "", "    ")
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s\n", src)
	return nil
}
*/
