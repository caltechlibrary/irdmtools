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
)

// Doi2Rdm holds the configuration for doi2rdm cli.
type Doi2Rdm struct {
	Cfg *Config
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
# Mapping ISSN prefixes to Journals (used to normalize journal titles names)
issn_journals:
  0002-9297: AJHG
  0004-6256: Astronomical Journal
  0004-637X: Astrophysical Journal
  0006-3495: Biophysical Journal
  0028-0836: Nature
  0035-2966: Monthly Notice of the Royal Astronomial Society
  0037-1106: Bulletin of the Seismological Society of America
  0037-1107: Bulletin of the Seismological Society of America
  0067-0049: Astrophysical Journal Supplement Series
  0092-8674: Cell
  0893-133X: Neuropsyhopharmacology
  0896-6273: Neuron
  0956-540X: Geophysical Journal International
  1061-4036: Nature Genetics
  1078-8956: Nature Medicine
  1087-0156: Nature Biotechnology
  1097-4164: Molecular Cell
  1097-4172: Cell
  1097-4180: Immunity
  1097-6256: Nature Neuroscience
  1362-4326: Trends in Biochemical Sciences
  1362-4555: Trends in Genetics
  1365-246X: Geophysical Journal International
  1365-2966: Monthly Notice of the Royal Astronomial Society
  1465-7392: Nature Cell Biology
  1471-003X: Nature Reviews Neuroscience
  1471-0048: Nature Reviews Neuroscience
  1471-0056: Nature Reviews Genetics
  1471-0064: Nature Reviews Genetics
  1471-0072: Nature Reviews Molecular Cell Biology
  1471-0080: Nature Reviews Molecular Cell Biology
  1471-4981: Trends in Immunology
  1471-499X: Trends in Molecular Medicine
  1471-5007: Trends in Parasitology
  1474-1733: Nature Reviews Immunology
  1474-1741: Nature Reviews Immunology
  1474-175X: Nature Reviews Cancer
  1474-1768: Nature Reviews Cancer
  1474-1776: Nature Reviews Drug Discovery
  1474-1784: Nature Reviews Drug Discovery
  1476-1122: Nature Materials
  1476-4660: Nature Materials
  1476-4679: Nature Cell Biology
  1476-4687: Nature
  1525-0016: Molecular Therapy
  1525-0024: Molecular Therapy
  1529-2908: Nature Immunoogy
  1529-2916: Nature Immunoogy
  1535-6108: Cancer Cell
  1537-6605: AJHG
  1538-3881: Astronomical Journal
  1538-4357: Astrophysical Journal
  1538-4365: Astrophysical Journal Supplement Series
  1542-0086: Biophysical Journal
  1545-9985: Nature Structural & Molecular Biology
  1545-9993: Nature Structural & Molecular Biology
  1546-170X: Nature Medicine
  1546-1718: Nature Genetics
  1546-1726: Nature Neuroscience
  1548-7091: Nature Methods
  1548-7105: Nature Methods
  1552-4450: Nature Chemical Biology
  1552-4469: Nature Chemical Biology
  1674-2052: Molecular Plant
  1740-1526: Nature Reviews Microbiology
  1740-1534: Nature Reviews Microbiology
  1745-2473: Nature Physics
  1745-2481: Nature Physics
  1745-3925: Monthly Notice of the Royal Astronomial Society Letters
  1745-3933: Monthly Notice of the Royal Astronomial Society Letters
  1748-3387: Nature Nanotechnology
  1748-3395: Nature Nanotechnology
  1749-4885: Nature Photonics
  1749-4893: Nature Photonics
  1750-2799: Nature Protocols
  1752-0894: Nature Geoscience
  1752-0908: Nature Geoscience
  1752-9867: Molecular Plant
  1754-2189: Nature Protocols
  1755-4330: Nature Chemistry
  1755-4349: Nature Chemistry
  1758-678X: Nature Climate Change
  1758-6798: Nature Climate Change
  1759-4758: Nature Reviews Neurology
  1759-4766: Nature Reviews Neurology
  1759-4774: Nature Reviews Clinical Oncology
  1759-4782: Nature Reviews Clinical Oncology
  1759-4790: Nature Reviews Rheumatology
  1759-4804: Nature Reviews Rheumatology
  1759-4812: Nature Reviews Urology
  1759-4820: Nature Reviews Urology
  1759-5002: Nature Reviews Cardiology
  1759-5010: Nature Reviews Cardiology
  1759-5029: Nature Reviews Endocrinology
  1759-5037: Nature Reviews Endocrinology
  1759-5045: Nature Reviews Gastroenterology & Hepatology
  1759-5053: Nature Reviews Gastroenterology & Hepatology
  1759-5061: Nature Reviews Nephrology
  1759-507X: Nature Reviews Nephrology
  1872-8383: Trends in Ecology & Evolution
  1873-3735: Trends in Pharmacological Sciences
  1875-9777: Cell Stem Cell
  1878-108X: Trends in Neurosciences
  1878-1551: Developmental Cell
  1878-3686: Cancer Cell
  1878-4186: Structure
  1878-4372: Trends in Plant Science
  1878-4380: Trends in Microbiology
  1879-0445: Current Biology
  1879-3061: Trends in Endocrinology & Metabolism
  1879-307X: Trends in Cognitive Sciences
  1879-3088: Trends in Cell Biology
  1879-3096: Trends in Biotechnology
  1884-4049: NPG Asia Materials
  1884-4057: NPG Asia Materials
  1932-7420: Cell Metabolism
  1934-6069: Cell Host & Microbe
  2041-1723: Nature Communications
  2041-8205: Astrophysical Journal Letters
  2041-8213: Astrophysical Journal Letters
  2044-4052: Nutrition & Diabetes
  2055-0278: Nature Plants
  2055-1010: npj Primary Care Respiratory Medicine
  2055-5008: npj Biofilms and Microbiomes
  2056-6387: npj Quantum Information
  2056-676X: Nature Reviews Disease Primers
  2056-7189: npj Systems Biology and Applications
  2056-7936: npj  Science of Learning
  2056-7944: npj Genomic Medicine
  2057-3960: npj Computational Materials
  2057-3995: npj Regenerative Medicine
  2058-5276: Nature Microbiology
  2058-7546: Nature Energy
  2058-8437: Nature Reviews Materials
  2059-0105: npj Vaccines
  2059-7037: npj Clean Water
  2157-846X: Nature Biomedical Engineering
  2162-2531: Moleclar Therapy - Nucleic Acids
  2211-1247: Cell Reports
  2213-6711: Stem Cell Reports
  2329-0501: Molecular Therapy Methods & Clinical Development
  2373-8057: npj Parkinson's Disease
  2373-8065: npj Microgravity
  2374-4677: npj Breast Cancer
  2396-8370: npj Science of Food
  2397-2106: npj Materials Degradation
  2397-334X: Nature Ecology & Evolution
  2397-3358: Nature Reviews Chemistry
  2397-3366: Nature Astronomy
  2397-3374: Nature Human Behaviour
  2397-3722: npj Climate and Atmospheric Science
  2397-4621: npj Fleible Electronics
  2397-4648: npj Quantum Materials
  2397-7132: npj 2D Materials and Applications
  2397-768X: npj Precision Oncology
  2398-6352: npj Digital Medicine
  2398-9629: Nature Sustainability
  2405-4720: Cell Systems
  2405-8025: Trends in Cancer
  2405-8440: Heliyon
  2451-9294: Chem
  2451-9448: Cell Chemical Biology
  2515-5172: Research Notes of the AAS
  2520-1131: Nature Electronics
  2520-1158: Nature Catalysis
  2522-5812: Nature Metabolism
  2522-5820: Nature Reviews Physics
  2522-5839: Nature Machine Intelligence
  2542-4351: Joule
  2589-0042: iScience
  2589-5974: Trends in Chemistry
  2590-2385: Metter
  2590-3322: One Earth
  2590-3462: Plant Communications
  2632-3338: Planetary Science Journal
  2661-8001: npj Urban Sustainability
  2662-1347: Nature Cancer
  2662-1355: Nature Food
  2662-138X: Nature Reviews Earth & Environment
  2662-8449: Nature Reviews Methods Primers
  2662-8457: Nature Computational Science
  2662-8465: Nature Aging
  2666-1667: STAR Protocols
  2666-2477: HGG Advances
  2666-3791: Cell Reports Medicine
  2666-3864: Cell Reports Physical Science
  2666-3899: Patterns
  2666-6340: Med
  2666-6758: The Innovation
  2666-979X: Cell Genomics
  2666-9986: Device
  2667-0747: Biophysical Reports
  2667-1093: Chem Catalysis
  2667-2375: Cell Reports Methods
  2730-9878: Nature Africa
  2731-0574: Nature Reviews Psychology
  2731-0582: Nature Synthesis
  2731-0590: Nature Cardiovascular Research
  2731-4243: npj Biodiversity
  2731-4251: npj Mental Health Research
  2731-426X: npj Ocean Sustainability
  2731-4278: npj Robotics
  2731-6068: npj Aging
  2731-6076: Nature Mental Health
  2731-6084: Nature Water
  2731-6092: Nature Reviews Bioengineering
  2731-8745: npj Antimicrobials and Resistance
  2731-8753: npj Complexity
  2731-9202: npj Sustainable Agriculture
  2731-9814: npj Climate Action
  2731-9997: Nature Cities
  2752-8200: RAS Techniques and Instruments
  2948-1198: Nature Chemical Engineering
  2948-1201: Nature Reviews Electrical Engineering
  2948-1570: NPP -- Digital Psychiatry and Neuroscience
  2948-1716: nph Women's Health
  2948-1767: npj Viruses
  2948-1775: npj Materials Sustainability
  2948-197X: npj Imaging
  2948-2100: npj Natural Hazards
  2948-2119: npj Spintronics
  2948-216X: npj Nanophotonics
  2948-281X: npj Biological TIming and Sleep
  2948-2828: npj Metabolic Health and Disease
  2948-2836: npj Cardiovascular Health
  2949-7906: Cell Reports Sustainability
  2950-1601: Nexus
  2950-3299: Molecular Therapy Oncology
  3004-8621: npj Advanced Manuscfacturing
  3004-863X: npj Biological Physics and Mechanics
  3004-8656: npj Biosensing
  3004-8664: npj Sustainable Mobility and Transport
  3004-8672: npj Unconventional Computing
  3004-9806: npj Gut and Liver
  3005-0677: Nature Reviews Biodiversity
  3005-0685: Nature Reviews Clean Technology
# Mapping ISSN prefixes to Publishers (used to normalize publisher names)
issn_publishers:
  0002-9297: Cell Press
  0004-6256: American Astronomical Society
  0004-637X: American Astronomical Society
  0006-3495: Cell Press
  0028-0836: Nature Publishing Group
  0035-2966: Royal Astronomical Society
  0037-1106: Seismological Society of America
  0037-1107: Seismological Society of America
  0067-0049: American Astronomical Society
  0092-8674: Cell Press
  0893-133X: Nature Publishing Group
  0896-6273: Cell Press
  0956-540X: Royal Astronomical Society
  1061-4036: Nature Publishing Group
  1078-8956: Nature Publishing Group
  1087-0156: Nature Publishing Group
  1097-4164: Cell Press
  1097-4172: Cell Press
  1097-4180: Cell Press
  1097-6256: Nature Publishing Group
  1362-4326: Cell Press
  1362-4555: Cell Press
  1365-246X: Royal Astronomical Society
  1365-2966: Royal Astronomical Society
  1465-7392: Nature Publishing Group
  1471-003X: Nature Publishing Group
  1471-0048: Nature Publishing Group
  1471-0056: Nature Publishing Group
  1471-0064: Nature Publishing Group
  1471-0072: Nature Publishing Group
  1471-0080: Nature Publishing Group
  1471-4981: Cell Press
  1471-499X: Cell Press
  1471-5007: Cell Press
  1474-1733: Nature Publishing Group
  1474-1741: Nature Publishing Group
  1474-175X: Nature Publishing Group
  1474-1768: Nature Publishing Group
  1474-1776: Nature Publishing Group
  1474-1784: Nature Publishing Group
  1476-1122: Nature Publishing Group
  1476-4660: Nature Publishing Group
  1476-4679: Nature Publishing Group
  1476-4687: Nature Publishing Group
  1525-0016: Cell Press
  1525-0024: Cell Press
  1529-2908: Nature Publishing Group
  1529-2916: Nature Publishing Group
  1535-6108: Cell Press
  1537-6605: Cell Press
  1538-3881: American Astronomical Society
  1538-4357: American Astronomical Society
  1538-4365: American Astronomical Society
  1542-0086: Cell Press
  1545-9985: Nature Publishing Group
  1545-9993: Nature Publishing Group
  1546-170X: Nature Publishing Group
  1546-1718: Nature Publishing Group
  1546-1726: Nature Publishing Group
  1548-7091: Nature Publishing Group
  1548-7105: Nature Publishing Group
  1552-4450: Nature Publishing Group
  1552-4469: Nature Publishing Group
  1674-2052: Cell Press
  1740-1526: Nature Publishing Group
  1740-1534: Nature Publishing Group
  1745-2473: Nature Publishing Group
  1745-2481: Nature Publishing Group
  1745-3925: Royal Astronomical Society
  1745-3933: Royal Astronomical Society
  1748-3387: Nature Publishing Group
  1748-3395: Nature Publishing Group
  1749-4885: Nature Publishing Group
  1749-4893: Nature Publishing Group
  1750-2799: Nature Publishing Group
  1752-0894: Nature Publishing Group
  1752-0908: Nature Publishing Group
  1752-9867: Cell Press
  1754-2189: Nature Publishing Group
  1755-4330: Nature Publishing Group
  1755-4349: Nature Publishing Group
  1758-678X: Nature Publishing Group
  1758-6798: Nature Publishing Group
  1759-4758: Nature Publishing Group
  1759-4766: Nature Publishing Group
  1759-4774: Nature Publishing Group
  1759-4782: Nature Publishing Group
  1759-4790: Nature Publishing Group
  1759-4804: Nature Publishing Group
  1759-4812: Nature Publishing Group
  1759-4820: Nature Publishing Group
  1759-5002: Nature Publishing Group
  1759-5010: Nature Publishing Group
  1759-5029: Nature Publishing Group
  1759-5037: Nature Publishing Group
  1759-5045: Nature Publishing Group
  1759-5053: Nature Publishing Group
  1759-5061: Nature Publishing Group
  1759-507X: Nature Publishing Group
  1872-8383: Cell Press
  1873-3735: Cell Press
  1875-9777: Cell Press
  1878-108X: Cell Press
  1878-1551: Cell Press
  1878-3686: Cell Press
  1878-4186: Cell Press
  1878-4372: Cell Press
  1878-4380: Cell Press
  1879-0445: Cell Press
  1879-3061: Cell Press
  1879-307X: Cell Press
  1879-3088: Cell Press
  1879-3096: Cell Press
  1884-4049: Nature Publishing Group
  1884-4057: Nature Publishing Group
  1932-7420: Cell Press
  1934-6069: Cell Press
  2041-1723: Nature Publishing Group
  2041-8205: American Astronomical Society
  2041-8213: American Astronomical Society
  2044-4052: Nature Publishing Group
  2055-0278: Nature Publishing Group
  2055-1010: Nature Publishing Group
  2055-5008: Nature Publishing Group
  2056-6387: Nature Publishing Group
  2056-676X: Nature Publishing Group
  2056-7189: Nature Publishing Group
  2056-7936: Nature Publishing Group
  2056-7944: Nature Publishing Group
  2057-3960: Nature Publishing Group
  2057-3995: Nature Publishing Group
  2058-5276: Nature Publishing Group
  2058-7546: Nature Publishing Group
  2058-8437: Nature Publishing Group
  2059-0105: Nature Publishing Group
  2059-7037: Nature Publishing Group
  2157-846X: Nature Publishing Group
  2162-2531: Cell Press
  2211-1247: Cell Press
  2213-6711: Cell Press
  2329-0501: Cell Press
  2373-8057: Nature Publishing Group
  2373-8065: Nature Publishing Group
  2374-4677: Nature Publishing Group
  2396-8370: Nature Publishing Group
  2397-2106: Nature Publishing Group
  2397-334X: Nature Publishing Group
  2397-3358: Nature Publishing Group
  2397-3366: Nature Publishing Group
  2397-3374: Nature Publishing Group
  2397-3722: Nature Publishing Group
  2397-4621: Nature Publishing Group
  2397-4648: Nature Publishing Group
  2397-7132: Nature Publishing Group
  2397-768X: Nature Publishing Group
  2398-6352: Nature Publishing Group
  2398-9629: Nature Publishing Group
  2405-4720: Cell Press
  2405-8025: Cell Press
  2405-8440: Cell Press
  2451-9294: Cell Press
  2451-9448: Cell Press
  2515-5172: American Astronomical Society
  2520-1131: Nature Publishing Group
  2520-1158: Nature Publishing Group
  2522-5812: Nature Publishing Group
  2522-5820: Nature Publishing Group
  2522-5839: Nature Publishing Group
  2542-4351: Cell Press
  2589-0042: Cell Press
  2589-5974: Cell Press
  2590-2385: Cell Press
  2590-3322: Cell Press
  2590-3462: Cell Press
  2632-3338: American Astronomical Society
  2661-8001: Nature Publishing Group
  2662-1347: Nature Publishing Group
  2662-1355: Nature Publishing Group
  2662-138X: Nature Publishing Group
  2662-8449: Nature Publishing Group
  2662-8457: Nature Publishing Group
  2662-8465: Nature Publishing Group
  2666-1667: Cell Press
  2666-2477: Cell Press
  2666-3791: Cell Press
  2666-3864: Cell Press
  2666-3899: Cell Press
  2666-6340: Cell Press
  2666-6758: Cell Press
  2666-979X: Cell Press
  2666-9986: Cell Press
  2667-0747: Cell Press
  2667-1093: Cell Press
  2667-2375: Cell Press
  2730-9878: Nature Publishing Group
  2731-0574: Nature Publishing Group
  2731-0582: Nature Publishing Group
  2731-0590: Nature Publishing Group
  2731-4243: Nature Publishing Group
  2731-4251: Nature Publishing Group
  2731-426X: Nature Publishing Group
  2731-4278: Nature Publishing Group
  2731-6068: Nature Publishing Group
  2731-6076: Nature Publishing Group
  2731-6084: Nature Publishing Group
  2731-6092: Nature Publishing Group
  2731-8745: Nature Publishing Group
  2731-8753: Nature Publishing Group
  2731-9202: Nature Publishing Group
  2731-9814: Nature Publishing Group
  2731-9997: Nature Publishing Group
  2752-8200: Royal Astronomical Society
  2948-1198: Nature Publishing Group
  2948-1201: Nature Publishing Group
  2948-1570: Nature Publishing Group
  2948-1716: Nature Publishing Group
  2948-1767: Nature Publishing Group
  2948-1775: Nature Publishing Group
  2948-197X: Nature Publishing Group
  2948-2100: Nature Publishing Group
  2948-2119: Nature Publishing Group
  2948-216X: Nature Publishing Group
  2948-281X: Nature Publishing Group
  2948-2828: Nature Publishing Group
  2948-2836: Nature Publishing Group
  2949-7906: Cell Press
  2950-1601: Cell Press
  2950-3299: Cell Press
  3004-8621: Nature Publishing Group
  3004-863X: Nature Publishing Group
  3004-8656: Nature Publishing Group
  3004-8664: Nature Publishing Group
  3004-8672: Nature Publishing Group
  3004-9806: Nature Publishing Group
  3005-0677: Nature Publishing Group
  3005-0685: Nature Publishing Group
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
	nWork, err := QueryCrossRefWork(app.Cfg, doi, options)
	if err != nil {
		return ENOENT, err
	}
	nRecord, err = CrosswalkCrossRefWork(app.Cfg, nWork, options)
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
	nWork, err := QueryDataCiteObject(app.Cfg, doi, options)
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
	if _, crErr := app.RunCrossRefToRdm(in, out, eout, optionFName, doi, diffFName); crErr != nil  {
		// Then try DataCiteToRdm
		if exitCode, dcErr := app.RunDataCiteToRdm(in, out, eout, optionFName, doi, diffFName); dcErr != nil {
			return exitCode, fmt.Errorf("crossref: %s, datacite: %s", crErr, dcErr)
		}
	}
	return EXIT_OK, nil
}
