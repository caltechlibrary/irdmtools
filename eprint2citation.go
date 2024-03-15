package irdmtools

import (
	"fmt"
	"io"
	"log"
	"path"
	"strings"
	
	// Caltech Library packages
	"github.com/caltechlibrary/dataset/v2"
	"github.com/caltechlibrary/eprinttools"
	"github.com/caltechlibrary/simplified"
)

// This will talk to a EPrints3 database and retrieve an EPrint record
// and output a CiteProc record in JSON format.

// EPrintToCitation takes a single EPrint records and returns a single
// Citation struct
func EPrintToCitation(repoName string, key string, eprint *eprinttools.EPrint, repoHost string, resourceTypes map[string]string, contributorTypes map[string]string) (*Citation, error) {
	rec := &simplified.Record{}
    if err := CrosswalkEPrintToRecord(eprint, rec, resourceTypes, contributorTypes); err != nil {
    	return nil, err
    }

   	// This is the way an EPrint URL is actually formed.
   	eprintURL := fmt.Sprintf("http://%s/%d", repoHost, eprint.EPrintID)
    // NOTE: We're dealing with a squirly situation of URLs to use during our migration and
    // before the feeds v2.0 implementation.
    if strings.HasPrefix(eprint.ID, "http") {
    	eprintURL = eprint.ID
    } else if eprint.OfficialURL != "" {
    	eprintURL = eprint.OfficialURL
	}
    citation := new(Citation)
	if eprint.Collection == "" {
		eprint.Collection = repoName
	}
    err := citation.CrosswalkRecord(eprint.Collection, key, eprintURL, rec)
	log.Printf("DEBUG cite -> %q -> repository %q, id %q", citation.ID, citation.Collection, citation.CollectionID)
    return citation, err
}

// MigrateEPrintDatasetToCitationsDataset takes a dataset of EPrint objects and migrates the ones in the
// id list to a citation dataset collection.
func MigrateEPrintDatasetToCitationDataset(ep3CName string, ids []string, repoHost string, resourceTypes map[string]string, contributorTypes map[string]string, citeCName string) error {
	ep3, err := dataset.Open(ep3CName)
	if err != nil {
		return err
	}
	defer ep3.Close()
	cite, err := dataset.Open(citeCName)
	if err != nil {
		return err
	}
	defer cite.Close()
	for i, id := range ids {
		eprint := new(eprinttools.EPrint)
		if err := ep3.ReadObject(id, eprint); err != nil {
			log.Printf("failed to get %s (%d), %s", id, i, err)
			continue
		}
		repoName := eprint.Collection
		if repoName == "" {
			repoName = path.Base(strings.TrimSuffix(ep3CName, ".ds"))
		}
		citation, err := EPrintToCitation(repoName, id, eprint, repoHost, resourceTypes, contributorTypes)
		if err != nil {
			log.Printf("failed to convert %s (%d) to citation, %s", id, i, err)
		}
		if cite.HasKey(citation.ID) {
			err = cite.UpdateObject(citation.ID, citation)
		} else {
			err = cite.CreateObject(citation.ID, citation)
		}
		if err != nil {
			log.Printf("failed to save citation for %s (%d), %s", id, i, err)
		}
	}
	return nil
}

// RunEPrintDSToCitationDS migrates contents from an EPrint dataset collection to a citation dataset collection for
// a give list of ids and repostiory hostname.
func RunEPrintDSToCitationDS(in io.Reader, out io.Writer, eout io.Writer, args []string, repoHost string, ids []string, resourceTypes map[string]string, contributorTypes map[string]string) int {
	var (
		ep3CName string
		citeCName string
		keys []string
	)
	if len(args) < 1 {
		fmt.Fprintf(eout, "missing eprint collection name and citation collection name\n")
		return 1
	}
	if len(args) < 2 {
		fmt.Fprintf(eout, "missing or eprint or citation collection names\n")
		return 1
	}
	if len(args) >= 2 {
		ep3CName, citeCName = args[0], args[1]
	}
	if len(args) > 2 {
		keys = args[2:]
	}
	if len(ids) > 0 {
		keys = append(keys, ids...)
	}
	if len(keys) == 0 {
		fmt.Fprintf(eout, "no ids to process, aborting\n")
		return 1
	}
	if err := MigrateEPrintDatasetToCitationDataset(ep3CName, keys, repoHost, resourceTypes, contributorTypes, citeCName); err != nil  {
		fmt.Fprintf(eout,  "%s\n", err)
		return 1
	}
	return 0 // OK
}
