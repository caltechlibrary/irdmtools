// irdmtools is a package for working with institutional repositories and
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
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	// 3rd Party packages
	_ "github.com/go-sql-driver/mysql"
)

// EPrintRest the "app" structure for the service.
type EPrintRest struct {
	RepoID             string `json:"repo_id,required"`
	EPrintArchivesPath string `json:"eprint_archives_path,required"`
	Port               string `json:"rest_port,omitempty"`
	DbHost             string `json:"db_host,omitempty`
	DbUser             string `json:"db_user,omitempty`
	DbPassword         string `json:"db_password,omitempty`
	in                 io.Reader
	out                io.Writer
	eout               io.Writer
}

var (
	restPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>{repo_id} REST: Datasets</title>
</head>
<body>
  <h1>{repo_id} REST: Datasets</h1>
<ul>
<li><a href='/rest/eprint/'>EPrints</a></li>
<li><a href='/rest/user/'>Users</a></li>
<li><a href='/rest/subject/'>Subjects</a></li>
</ul>
</body>
</html>`

	datasetPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <title>{repo_id} REST: {dataset} DataSet</title>
</head>
<body>
  <h1>{repo_id} REST: {dataset} DataSet</h1>
  <ul>
  	{list_of_li}
  </ul>
</body>
</html>`

	idFields = map[string]string{
		"eprint":  "eprintid",
		"user":    "userid",
		"subject": "subjectid",
	}
)

// LoadEnv settings from the enviroment to run a local host clone
// of the EPrints REST API using the archives content and MySQL database.
func (app *EPrintRest) LoadEnv() {
	app.RepoID = os.Getenv("REPO_ID")
	app.EPrintArchivesPath = os.Getenv("EPRINT_ARCHIVES_PATH")
	app.Port = os.Getenv("REST_PORT")
	if app.Port == "" {
		app.Port = ":8003"
	} else if ! strings.HasPrefix(app.Port, ":") {
		app.Port = ":" + app.Port
	}
	app.DbUser = os.Getenv("DB_USER")
	app.DbPassword = os.Getenv("DB_PASSWORD")
}

func transformTxt(s string, target string, dest string) string {
	return strings.ReplaceAll(s, target, dest)
}

// MkDatasetPage takes the simple page template and maps the dataset and label
// rendering the HTML page as a string
func (app *EPrintRest) MkDatasetPage(db *sql.DB, tmpl string, dataset string, label string) (string, error) {
	src := transformTxt(transformTxt(tmpl, "{repo_id}", app.RepoID), "{dataset}", label)
	// Figure out which dataset we're working with and retrieve the ids building a list of
	// LI elements in the form `<li><a href="/rest/{dataset}/{id}.xml">{id}.xml</a></li>`
	items := []string{}
	idField, ok := idFields[dataset]
	if !ok {
		return "", fmt.Errorf("%q is not a supported dataset", dataset)
	}
	queryTxt := transformTxt(transformTxt(`SELECT {idField} AS id FROM {dataset} ORDER BY {idField}`, `{idField}`, idField), `{dataset}`, dataset)
	rows, err := db.Query(queryTxt)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	liTmpl := transformTxt(`<li><a href="/rest/{dataset}/{id}.xml">{id}.xml</a></li>`, `{dataset}`, dataset)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		items = append(items, transformTxt(liTmpl, `{id}`, id))

	}
	return transformTxt(src, "{list_of_li}", strings.Join(items, "\n")), nil
}

// RequestLogger logs http request to service
func RequestLogger(targetMux http.Handler) http.Handler {
         return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                 start := time.Now()
                 targetMux.ServeHTTP(w, r)

                 // log request by who(IP address)
                 requesterIP := r.RemoteAddr

                 log.Printf(
                         "%s\t\t%s\t\t%s\t\t%v",
                         r.Method,
                         r.RequestURI,
                         requesterIP,
                         time.Since(start),
                 )
         })
 }

// EPrintXMLPath takes the app setup and generates the path do the EPrintXML document from
// an id.
func (app *EPrintRest) EPrintXMLPath(db *sql.DB, id string) (string, error) {
	queryTxt := `SELECT IFNULL(dir, "") AS dir, IFNULL(rev_number, "") AS rev_number FROM eprint WHERE eprintid = ?`
	rows, err := db.Query(queryTxt, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			dir string
			revNumber string
		)
		if err := rows.Scan(&dir, &revNumber); err != nil {
			return "", err
		}
		return fmt.Sprintf("%s/%s/documents/%s/revisions/%s.xml", app.EPrintArchivesPath, app.RepoID, dir, revNumber), nil
	}
	return "", fmt.Errorf("failed to get next row for eprintid %q", id)
}

func getIds(db *sql.DB, dataset string) ([]string, error) {
	idField, ok := idFields[dataset]
	if ! ok {
		return nil, fmt.Errorf("unsupported dataset %q", dataset)
	}
	whereClause := ""
	if dataset == "eprint" {
		whereClause = `WHERE eprint_status = "archive"`
	}
	queryTxt := fmt.Sprintf(`SELECT %s FROM %s %s ORDER BY %s`, idField, dataset, whereClause, idField)
	ids := []string{}
	rows, err := db.Query(queryTxt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Serve runs the web service minimally replicating the EPrints 3.x
// REST API.
func (app *EPrintRest) ListenAndServe() error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s", app.DbUser, app.DbPassword, app.RepoID))
	if err != nil {
		return err
	}
	defer db.Close()

	// Preload set pages, content and cache the id list pages
	eprintSrc, err := app.MkDatasetPage(db, datasetPage, "eprint", "EPrints")
	if err != nil {
		return err
	}
	userSrc, err := app.MkDatasetPage(db, datasetPage, "user", "Users")
	if err != nil {
		return err
	}
	subjectSrc, err := app.MkDatasetPage(db, datasetPage, "subject", "Subjects")
	if err != nil {
		return err
	}
	restPageSrc := transformTxt(restPage, "{repo_id}", app.RepoID)


	// Set up our server Mux
	mux := http.NewServeMux()
	
	// Map the on URLs to the disk path to EPrint XML files with the latest version.
	eprintXML := map[string]string{}
	ids, err := getIds(db, "eprint")
	tot := len(ids)
	cnt := 0
	for i, id := range ids {
		var (
			fName string
			src []byte
			err error
		)
		apiPath := fmt.Sprintf("/rest/eprint/%s.xml", id)
		fName, err = app.EPrintXMLPath(db, id)
		if err != nil {
			return err
		} 
		src, err = os.ReadFile(fName)
		if err != nil {
			log.Printf("skipping, failed to read record %s from %s", id, fName)
			continue
		}
		eprintXML[apiPath] = fmt.Sprintf("%s", src)
		cnt += 1
		if (i % 1000) == 0 {
			log.Printf("scanned %d/%d records, read %d", i, tot, cnt)
		}
	}
	log.Printf("found %d/%d records", cnt, tot)
	// Handle `/rest/eprint/` and the individual EPrint XML responses
	mux.HandleFunc("/rest/eprint/", func(w http.ResponseWriter, req *http.Request) {
		// FIXME: Set content type to HTML
		if req.URL.Path == "/rest/eprint/" {
			io.WriteString(w, eprintSrc)
			return
		}
		if txt, ok := eprintXML[req.URL.Path]; ok {
			io.WriteString(w, txt)
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	// Handle `/rest/user/`
	mux.HandleFunc("/rest/user/", func(w http.ResponseWriter, req *http.Request) {
		// FIXME: Set content type to HTML
		if req.URL.Path == "/rest/user/" {
			io.WriteString(w, userSrc)
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	// Handle `/rest/subject/`
	mux.HandleFunc("/rest/subject/", func(w http.ResponseWriter, req *http.Request) {
		// FIXME: Set content type to HTML
		if req.URL.Path == "/rest/subject/" {
			io.WriteString(w, subjectSrc)
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})

	// Handle of REST API page `/rest/` and '/' with restPage
	mux.HandleFunc("/rest/", func(w http.ResponseWriter, req *http.Request) {
		// FIXME: Set content type to HTML
		io.WriteString(w, restPageSrc)
	})
	/*
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// FIXME: Set content type to HTML
		io.WriteString(w, restPageSrc)
	})
	*/
	// FIXME need to setup a 404 error page response.
	log.Printf("Starting REST API for %q up listening on http://localhost%s", app.RepoID, app.Port)
	return http.ListenAndServe(app.Port, RequestLogger(mux))
}

// Run loads a configuration from the environment and does a sanity check of the service setup
// maps standard in, out and error to the service then invokes app.ListenAndServe().
func (app *EPrintRest) Run(in io.Reader, out io.Writer, eout io.Writer) error {
	app.LoadEnv()
	// Sanity check the application's settings.
	if app.RepoID == "" || app.EPrintArchivesPath == "" || app.Port == "" {
		return fmt.Errorf("repoID, eprint archives path or rest port number missing")
	}
	if app.DbUser == "" || app.DbPassword == "" {
		return fmt.Errorf("MySQL EPrint database access not configured")
	}
	app.in = in
	app.out = out
	app.eout = eout
	return app.ListenAndServe()
}
