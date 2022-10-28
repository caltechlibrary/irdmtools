package irdmtools

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Util struct {
	BaseURL string `json:"base_url"`
}

func MakeUtil(configFName string) (*Util, error) {
	util := new(Util)
	// By default we look for Invenio-RDM as installed with
	// docker on localhost:5000
	util.BaseURL = "http://localhost:5000"
	if configFName != "" {
		if _, err := os.Stat(configFName); os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to find %q, %s", configFName, err)
		}
		src, err := ioutil.ReadFile(configFName)
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}
		err = json.Unmarshal(src, &util)
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}
	}
	return util, nil
}

func (util *Util) Run(in io.Reader, out io.Writer, eout io.Writer, args []string) error {
	return fmt.Errorf("Run() not implemnted")
}
