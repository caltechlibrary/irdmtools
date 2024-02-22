package irdmtools

import (
	"fmt"
	"net/url"
	"strings"
)

// LinkToDoi removes a leading URL reference (DOI link) if found returning the remainder of the DOI string (prefix slash item identifier)
func LinkToDoi(s string) (string, error) {
	if strings.HasPrefix(s, "http:") || strings.HasPrefix(s, "https:") {
		u, err := url.Parse(s)
		if err != nil {
			return "", err
		}
		return strings.TrimPrefix(u.Path, "/"), nil
	}
	return s, nil
}

// DoiPrefix takes a DOI returns the publisher prefix
func DoiPrefix(s string) (string, error) {
	// Get the bare DOI without URL prefix
	doi, err := LinkToDoi(s)
	if err != nil {
		return "", err
	}
	// Split at the first "/"
	parts := strings.SplitN(doi, "/", 2)
	if len(parts) != 2 {
	    return "", fmt.Errorf("cannot determine prefix for %q", s)
	}
	return parts[0], nil
}
