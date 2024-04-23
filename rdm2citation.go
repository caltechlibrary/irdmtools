package irdmtools

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	// Caltech Library packages
	"github.com/caltechlibrary/simplified"
	"github.com/caltechlibrary/dataset/v2"
)

// This will talk to a RDM database and retrieve an RDM record
// and output a Citation record in JSON format.

// Convert an RDM record to a citation in a Citation struct
func RdmToCitation(repoName string, key string, record *simplified.Record, repoHost string) (*Citation, error) {
   	// This is the way an EPrint URL is actually formed.
	protocol := "https"
	if repoHost == "" {
		repoHost = "localhost:8000"
		protocol = "http"
	}
   	citeUsingURL := fmt.Sprintf("%s://%s/%s", protocol, repoHost, record.ID)
	repoHostURL := fmt.Sprintf("%s://%s", protocol, repoHost)
    // NOTE: We're dealing with a squirly situation of URLs to use during our migration and
    // before the feeds v2.0 implementation.
    citation := new(Citation)
    err := citation.CrosswalkRecord(repoName, key, citeUsingURL, repoHostURL, record)
    return citation, err
}

// MigrateRdmDatasetToCitationsDataset takes a dataset of RDM objects and migrates the ones in the
// id list to a citation dataset collection.
func MigrateRdmDatasetToCitationDataset(rdmCName string, ids []string, repoHost string, prefix string, citeCName string) error {
	rdm, err := dataset.Open(rdmCName)
	if err != nil {
		return err
	}
	defer rdm.Close()
	cite, err := dataset.Open(citeCName)
	if err != nil {
		return err
	}
	defer cite.Close()
	tot := len(ids)
	start := time.Now()
	iTime := time.Now()
	reportProgress := false
	i := 0
	log.Printf("%d/%d citations processed %s: %s", i, tot, time.Since(start).Truncate(time.Second).String(), ProgressETA(start, i, tot))
	for _, id := range ids {
		rdmRecord := new(simplified.Record)
		if err := rdm.ReadObject(id, rdmRecord); err != nil {
			log.Printf("failed to get %s (%d), %s", id, i, err)
			continue
		}
		if rdmRecord.Versions != nil && ! rdmRecord.Versions.IsLatest {
			log.Printf("skipping, not the latest version = %+v, %s (%d)", rdmRecord.Versions, id, i)
			continue
		}
		repoName := path.Base(strings.TrimSuffix(rdmCName, ".ds"))
		key := id
		if prefix != "" {
			key = fmt.Sprintf("%s:%s", repoName, id) // the key we will use as the suffix in citation.ds
		}
		citation, err := RdmToCitation(repoName, id, rdmRecord, repoHost)
		if err != nil {
			log.Printf("failed to convert (%d) id %s from %s to citation, %s", i, id, repoName, err)
			continue
		}
		if cite.HasKey(key) {
			err = cite.UpdateObject(key, citation)
		} else {
			err = cite.CreateObject(key, citation)
		}
		if err != nil {
			log.Printf("failed to save citation for %s (%d), %s", id, i, err)
		}
		i++
		if iTime, reportProgress = CheckWaitInterval(iTime, time.Minute); reportProgress || (i % 10000) == 0 {
			log.Printf("%d/%d citations processed %s: %s", i, tot, time.Since(start).Truncate(time.Second).String(), ProgressETA(start, i, tot))
		}
	}
	log.Printf("%d/%d citations processed %s: completed", i, tot, time.Since(start).Truncate(time.Second).String())
	return nil
}

// RunRdmDSToCitationDS migrates contents from an RDM dataset collection to a citation dataset collection for
// a give list of ids and repostiory hostname.
func RunRdmDSToCitationDS(in io.Reader, out io.Writer, eout io.Writer, args []string, repoHost string, prefix string, ids []string) int {
	var (
		rdmCName string
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
		rdmCName, citeCName = args[0], args[1]
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
	if repoHost == "" {
		if rdmUrl := os.Getenv("RDM_URL"); rdmUrl != "" {
			u, err := url.Parse(rdmUrl);
			if err == nil {
				repoHost = u.Host
			}
		}
	}
	if err := MigrateRdmDatasetToCitationDataset(rdmCName, keys, repoHost, prefix, citeCName); err != nil  {
		fmt.Fprintf(eout,  "%s\n", err)
		return 1
	}
	return 0 // OK
}
