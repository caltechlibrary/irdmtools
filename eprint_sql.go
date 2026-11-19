package irdmtools

// eprint_sql.go runs a per-collection SQL query against EPrints MySQL that
// emits RDM-shaped JSON (scripts/eprints_thesis_to_rdm.sql) and writes the
// rows as dataset {key, object} lines, optionally loading them into a
// collection with a count check (irdmtools DR-0010).

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	// Caltech Library Packages
	"github.com/caltechlibrary/dataset/v2"

	// 3rd Party Packages
	"github.com/go-sql-driver/mysql"
)

// objectStore is the part of a dataset collection the SQL mode writes to.
type objectStore interface {
	HasKey(key string) bool
	CreateJSON(key string, src []byte) error
	UpdateJSON(key string, src []byte) error
}

// sqlLine is one {key, object} line as `dataset load` reads it.
type sqlLine struct {
	Key    string          `json:"key"`
	Object json.RawMessage `json:"object"`
}

// RunEPrintSQL runs a crosswalk query and writes one {key, object} line per
// row, loading the objects into a store when one is given, then checks that
// every line's key is in the store. The query may read the session variable
// @eprint_ids (a comma separated list, NULL for no limit) to restrict itself.
// Every row is read and checked before anything is written or loaded, so a
// bad row leaves no partial output. The key is the row's eprintid identifier.
//
// Errors carry the DR-0014 classes: no row found is negative (1), a bad id is
// usage (2), a row that is not JSON, has no eprintid, or repeats one is data
// (65), a query MySQL rejects is config (78), a database that cannot be
// reached is unavailable (69), and a store that holds fewer objects than
// there were lines is io (74).
//
// @param db {*sql.DB} an open EPrints database
// @param query {string} the crosswalk query, one JSON object per row
// @param ids {[]int} eprint ids to limit the run to, or none for every record
// @param jsonl {io.Writer} where the {key, object} lines go, or nil
// @param store {objectStore} where the objects are loaded, or nil
// @returns {int} the number of lines produced, and an error
// @example
//
//	n, err := RunEPrintSQL(db, query, []int{1, 2}, os.Stdout, nil)
//	if err != nil {
//		os.Exit(ExitCode(err))
//	}
func RunEPrintSQL(db *sql.DB, query string, ids []int, jsonl io.Writer, store objectStore) (int, error) {
	idList := ""
	if len(ids) > 0 {
		parts := make([]string, 0, len(ids))
		for _, id := range ids {
			if id <= 0 {
				return 0, UsageError(fmt.Errorf("eprint id %d is not a positive number", id))
			}
			parts = append(parts, strconv.Itoa(id))
		}
		idList = strings.Join(parts, ",")
	}
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return 0, UnavailableError(fmt.Errorf("cannot reach the EPrints database: %w", err))
	}
	defer conn.Close()
	var limit interface{}
	if idList != "" {
		limit = idList
	}
	if _, err := conn.ExecContext(ctx, "SET @eprint_ids = ?", limit); err != nil {
		return 0, queryFailure(err)
	}
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return 0, queryFailure(err)
	}
	defer rows.Close()
	lines := []sqlLine{}
	seen := map[string]bool{}
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return 0, DataError(fmt.Errorf("a row cannot be read: %w", err))
		}
		key, err := eprintIDFromRow(raw)
		if err != nil {
			return 0, err
		}
		if seen[key] {
			return 0, DataError(fmt.Errorf("the query produced two rows for eprint %s", key))
		}
		seen[key] = true
		lines = append(lines, sqlLine{Key: key, Object: append(json.RawMessage(nil), raw...)})
	}
	if err := rows.Err(); err != nil {
		return 0, queryFailure(err)
	}
	if len(lines) == 0 {
		return 0, NegativeError(fmt.Errorf("the query found no record"))
	}
	for _, line := range lines {
		if jsonl != nil {
			src, err := json.Marshal(line)
			if err != nil {
				return 0, InternalError(err)
			}
			if _, err := jsonl.Write(append(src, '\n')); err != nil {
				return 0, IOError(fmt.Errorf("writing lines: %w", err))
			}
		}
		if store != nil {
			var err error
			if store.HasKey(line.Key) {
				err = store.UpdateJSON(line.Key, line.Object)
			} else {
				err = store.CreateJSON(line.Key, line.Object)
			}
			if err != nil {
				return len(lines), IOError(fmt.Errorf("loading eprint %s: %w", line.Key, err))
			}
		}
	}
	if store != nil {
		loaded := 0
		for _, line := range lines {
			if store.HasKey(line.Key) {
				loaded++
			}
		}
		if loaded != len(lines) {
			return len(lines), IOError(fmt.Errorf("loaded %d of %d lines; the load stopped part way", loaded, len(lines)))
		}
	}
	return len(lines), nil
}

// eprintIDFromRow reads the eprintid identifier out of a crosswalk row.
func eprintIDFromRow(raw []byte) (string, error) {
	var row struct {
		Metadata struct {
			Identifiers []struct {
				Scheme     string `json:"scheme"`
				Identifier string `json:"identifier"`
			} `json:"identifiers"`
		} `json:"metadata"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&row); err != nil {
		return "", DataError(fmt.Errorf("a row is not JSON: %w", err))
	}
	for _, ident := range row.Metadata.Identifiers {
		if ident.Scheme == "eprintid" && ident.Identifier != "" {
			return ident.Identifier, nil
		}
	}
	return "", DataError(fmt.Errorf("a row has no eprintid identifier"))
}

// queryFailure classifies an error from running the query: MySQL rejecting
// the statement means the query file is wrong (config); anything else is the
// database not answering (unavailable).
func queryFailure(err error) error {
	if _, ok := err.(*mysql.MySQLError); ok {
		return ConfigError(fmt.Errorf("the query was rejected: %w", err))
	}
	return UnavailableError(fmt.Errorf("the EPrints database did not answer: %w", err))
}

// ReadEPrintIDList reads a file of eprint ids, one per line. Blank lines and
// spaces around an id are ignored. A file that cannot be opened is no_input
// (66), and a line that is not a positive number, or a file with no id at all,
// is data (65).
//
// @param fName {string} a file with one eprint id per line
// @returns {[]int} the ids in file order, and an error
// @example
//
//	ids, err := ReadEPrintIDList("ids.txt")
//	if err != nil {
//		os.Exit(ExitCode(err))
//	}
func ReadEPrintIDList(fName string) ([]int, error) {
	src, err := os.ReadFile(fName)
	if err != nil {
		return nil, NoInputError(fmt.Errorf("cannot read the id list %q: %w", fName, err))
	}
	ids := []int{}
	for i, line := range strings.Split(string(src), "\n") {
		word := strings.TrimSpace(line)
		if word == "" {
			continue
		}
		id, err := strconv.Atoi(word)
		if err != nil || id <= 0 {
			return nil, DataError(fmt.Errorf("%s line %d: %q is not an eprint id", fName, i+1, word))
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil, DataError(fmt.Errorf("%s holds no eprint id", fName))
	}
	return ids, nil
}

// RunEPrintSQLFiles runs the SQL mode from files: the query from queryFName,
// the ids from idListFName when it is not empty, and the objects loaded into
// the dataset collection cName when it is not empty. With a collection the
// {key, object} lines are not written to out; without one they are.
//
// @param db {*sql.DB} an open EPrints database
// @param queryFName {string} the crosswalk query file
// @param idListFName {string} a file of eprint ids, or ""
// @param cName {string} a dataset collection to load, or ""
// @param out {io.Writer} where the lines go when no collection is given
// @returns {int} the number of lines produced, and an error
// @example
//
//	n, err := RunEPrintSQLFiles(db, "eprints_thesis_to_rdm.sql", "ids.txt", "thesis.ds", os.Stdout)
func RunEPrintSQLFiles(db *sql.DB, queryFName, idListFName, cName string, out io.Writer) (int, error) {
	src, err := os.ReadFile(queryFName)
	if err != nil {
		return 0, NoInputError(fmt.Errorf("cannot read the query file %q: %w", queryFName, err))
	}
	var ids []int
	if idListFName != "" {
		if ids, err = ReadEPrintIDList(idListFName); err != nil {
			return 0, err
		}
	}
	if cName == "" {
		return RunEPrintSQL(db, string(src), ids, out, nil)
	}
	c, err := dataset.Open(cName)
	if err != nil {
		return 0, NoInputError(fmt.Errorf("cannot open the collection %q: %w", cName, err))
	}
	defer c.Close()
	return RunEPrintSQL(db, string(src), ids, nil, c)
}

// EPrintDSN builds the MySQL data source name for the EPrints database from
// the EPRINT_DB_HOST, EPRINT_DB_USER, EPRINT_DB_PASSWORD and REPO_ID settings,
// the way GetEPrint does: a host of "localhost" is reached through the local
// socket. A missing setting is a config error (78) that names the variable
// and never shows the password.
//
// @param cfg {*Config} the configuration holding the EPRINT_DB_* settings
// @returns {string} the MySQL data source name, and an error
// @example
//
//	dsn, err := EPrintDSN(app.Cfg)
//	db, err := sql.Open("mysql", dsn)
func EPrintDSN(cfg *Config) (string, error) {
	missing := []string{}
	if cfg.EPrintDbHost == "" {
		missing = append(missing, "EPRINT_DB_HOST")
	}
	if cfg.EPrintDbUser == "" {
		missing = append(missing, "EPRINT_DB_USER")
	}
	if cfg.EPrintDbPassword == "" {
		missing = append(missing, "EPRINT_DB_PASSWORD")
	}
	if cfg.RepoID == "" {
		missing = append(missing, "REPO_ID")
	}
	if len(missing) > 0 {
		return "", ConfigError(fmt.Errorf("the SQL mode needs %s", strings.Join(missing, ", ")))
	}
	if cfg.EPrintDbHost == "localhost" {
		return fmt.Sprintf("%s:%s@/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.RepoID), nil
	}
	return fmt.Sprintf("%s:%s@%s/%s", cfg.EPrintDbUser, cfg.EPrintDbPassword, cfg.EPrintDbHost, cfg.RepoID), nil
}

// RunEPrintSQLMode is the entry point behind `eprint2rdm -sql`: it opens the
// EPrints database, runs RunEPrintSQLFiles, writes the {key, object} lines to
// out, and writes a one line summary or the error to eout. It returns the
// DR-0014 exit code for the caller to pass to os.Exit. The password in the
// data source name is never written.
//
// @param dsn {string} the MySQL data source name
// @param queryFName {string} the crosswalk query file
// @param idListFName {string} a file of eprint ids, or ""
// @param cName {string} a dataset collection to load, or ""
// @param out {io.Writer} where the lines go when no collection is given
// @param eout {io.Writer} where the summary or the error goes
// @returns {int} the exit code
// @example
//
//	dsn, _ := EPrintDSN(cfg)
//	os.Exit(RunEPrintSQLMode(dsn, "eprints_thesis_to_rdm.sql", "ids.txt", "thesis.ds", os.Stdout, os.Stderr))
func RunEPrintSQLMode(dsn, queryFName, idListFName, cName string, out, eout io.Writer) int {
	fail := func(err error) int {
		fmt.Fprintln(eout, scrubDSNPassword(err.Error(), dsn))
		return ExitCode(err)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fail(ConfigError(fmt.Errorf("the data source name is not valid: %w", err)))
	}
	defer db.Close()
	n, err := RunEPrintSQLFiles(db, queryFName, idListFName, cName, out)
	if err != nil {
		return fail(err)
	}
	fmt.Fprintf(eout, "%d records\n", n)
	return 0
}

// scrubDSNPassword replaces the password of a MySQL data source name
// ("user:password@host/db") wherever it appears in msg, so an error message
// can be shown without it.
//
// @param msg {string} the text to clean
// @param dsn {string} the data source name the password is taken from
// @returns {string} msg with the password replaced by "***"
// @example
//
//	fmt.Fprintln(os.Stderr, scrubDSNPassword(err.Error(), dsn))
func scrubDSNPassword(msg, dsn string) string {
	at := strings.LastIndex(dsn, "@")
	if at <= 0 {
		return msg
	}
	colon := strings.Index(dsn[:at], ":")
	if colon < 0 || colon+1 >= at {
		return msg
	}
	return strings.ReplaceAll(msg, dsn[colon+1:at], "***")
}
