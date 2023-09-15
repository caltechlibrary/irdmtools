package irdmtools

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

type RorOrgAPIResponse struct {
	NumberOfResults int `json:"number_of_results,omitempty"`
	TimeTaken time.Duration `json:"time_taken,omitempty"`
	Items []map[string]interface{} `json:"items,omitempty"`
}

// lookupROR
func lookupROR(doiSuffix string) (string, bool) {
	// Call: https://api.ror.org/organizations?query={doiPrefix}
	orgAPI := "https://api.ror.org/organizations"
	client := &http.Client{}
	req, err := http.NewRequest("GET", orgAPI, nil)
	if err != nil {
		return "", false
	}
	// Add our query using the DOI prefix
	q := req.URL.Query()
	q.Set("query", doiSuffix)
	req.URL.RawQuery = q.Encode()
    resp, err := client.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		result := new(RorOrgAPIResponse)
		if err := json.Unmarshal(body, &result); err != nil {
			return "", false
		}
		if result.Items != nil {
			for _, item := range result.Items {
				if id, ok := item["id"].(string); ok {
					return id, true
				}
			}
		}
		return "", false
	}
	return "", false
}

