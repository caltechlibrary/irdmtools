package irdmtools

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	// Caltech Library packages
	"github.com/caltechlibrary/crossrefapi"
)

func QueryCrossRef(cfg *Config, doi string, mailTo string, dotInitials bool, downloadDocument bool, debug bool) (map[string]interface{}, error) {
	appName := path.Base(os.Args[0])
	client, err := crossrefapi.NewCrossRefClient(appName, mailTo)
	if err != nil {
		return nil, err
	}
	works, err := client.Works(doi)
	if err != nil {
		return nil, err
	}
	if debug {
		src, _ := json.MarshalIndent(works, "", "    ")
		fmt.Fprintf(os.Stderr, "works JSON:\n\n%s\n\n", src)
	}
	return works, nil
}
