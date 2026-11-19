package irdmtools

// Tests for RunEPrintSQL (eprint_sql.go) against the scratch MySQL with the
// thesis crosswalk query, and against a real temporary dataset collection.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	// Caltech Library Packages
	"github.com/caltechlibrary/dataset/v2"
)

// thesisQuery reads the crosswalk query file.
func thesisQuery(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile(thesisSQLPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(src)
}

// threeTheses writes three thesis eprints and one article.
func threeTheses(t *testing.T) {
	t.Helper()
	db := openScratchEPrints(t)
	resetScratch(t, db)
	for _, id := range []int{1, 2, 3} {
		fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, title) VALUES (?, 'archive', 'thesis', ?)`,
			id, fmt.Sprintf("Thesis %d", id))
		fixture(t, db, `INSERT INTO document (docid, eprintid, pos, security, main, content) VALUES (?, ?, 1, 'public', 'a.pdf', 'final')`, id*10, id)
		fixture(t, db, `INSERT INTO file (fileid, datasetid, objectid, filename, hash, hash_type, filesize) VALUES (?, 'document', ?, 'a.pdf', 'h', 'MD5', 10)`, id*100, id*10)
	}
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (9, 'archive', 'article')`)
}

// keysOf returns the sorted keys of the {key, object} lines in out.
func keysOf(t *testing.T, out []byte) []string {
	t.Helper()
	keys := []string{}
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 1024*1024), 64*1024*1024)
	for sc.Scan() {
		var line struct {
			Key    string                 `json:"key"`
			Object map[string]interface{} `json:"object"`
		}
		if err := json.Unmarshal(sc.Bytes(), &line); err != nil {
			t.Fatalf("not a {key, object} line: %s\n%s", err, sc.Bytes())
		}
		if line.Object == nil {
			t.Fatalf("line has no object: %s", sc.Bytes())
		}
		keys = append(keys, line.Key)
	}
	sort.Strings(keys)
	return keys
}

// TestRunEPrintSQLAllAndIDs checks that every thesis is written when no id
// list is given, and only the listed ones when it is.
func TestRunEPrintSQLAllAndIDs(t *testing.T) {
	threeTheses(t)
	db := openScratchEPrints(t)
	var all bytes.Buffer
	n, err := RunEPrintSQL(db, thesisQuery(t), nil, &all, nil)
	if err != nil || n != 3 {
		t.Fatalf("all: n = %d, err = %v, want 3 lines", n, err)
	}
	if got := fmt.Sprint(keysOf(t, all.Bytes())); got != "[1 2 3]" {
		t.Errorf("all keys = %s", got)
	}
	var some bytes.Buffer
	n, err = RunEPrintSQL(db, thesisQuery(t), []int{3, 1}, &some, nil)
	if err != nil || n != 2 {
		t.Fatalf("ids: n = %d, err = %v, want 2 lines", n, err)
	}
	if got := fmt.Sprint(keysOf(t, some.Bytes())); got != "[1 3]" {
		t.Errorf("id-limited keys = %s, want [1 3]", got)
	}
}

// TestRunEPrintSQLErrors checks each failure gets its DR-0014 exit class.
func TestRunEPrintSQLErrors(t *testing.T) {
	threeTheses(t)
	db := openScratchEPrints(t)
	q := thesisQuery(t)
	cases := []struct {
		name  string
		query string
		ids   []int
		want  ExitClass
	}{
		{"no record found", q, []int{999}, ExitNegative},
		{"bad id", q, []int{1, 0}, ExitUsage},
		{"negative id", q, []int{-4}, ExitUsage},
		{"bad sql", "SELECT FROM nowhere", nil, ExitConfig},
		{"row is not json", "SELECT 'plain text' FROM eprint LIMIT 1", nil, ExitData},
		{"row has no eprintid", "SELECT JSON_OBJECT('metadata', JSON_OBJECT()) FROM eprint LIMIT 1", nil, ExitData},
		{"identifier of another scheme", `SELECT JSON_OBJECT('metadata', JSON_OBJECT('identifiers', JSON_ARRAY(JSON_OBJECT('scheme','doi','identifier','10.1/x')))) FROM eprint LIMIT 1`, nil, ExitData},
		{"duplicate key", `SELECT JSON_OBJECT('metadata', JSON_OBJECT('identifiers', JSON_ARRAY(JSON_OBJECT('scheme','eprintid','identifier','7'))))
			UNION ALL SELECT JSON_OBJECT('metadata', JSON_OBJECT('identifiers', JSON_ARRAY(JSON_OBJECT('scheme','eprintid','identifier','7'))))`, nil, ExitData},
	}
	for _, c := range cases {
		var out bytes.Buffer
		store := &fakeStore{keep: 100}
		n, err := RunEPrintSQL(db, c.query, c.ids, &out, store)
		if err == nil {
			t.Errorf("%s: no error", c.name)
			continue
		}
		if got := ExitClassOf(err); got != c.want {
			t.Errorf("%s: exit class %d, want %d (%v)", c.name, got, c.want, err)
		}
		if n != 0 || len(store.loaded) != 0 {
			t.Errorf("%s: n = %d, loaded = %d, want nothing written or loaded", c.name, n, len(store.loaded))
		}
	}
}

// fakeStore is an objectStore that silently drops every object after keep of
// them, as an aborted load would, so the count check has something to find.
type fakeStore struct {
	keep   int
	loaded map[string][]byte
}

func (s *fakeStore) HasKey(key string) bool { _, ok := s.loaded[key]; return ok }
func (s *fakeStore) CreateJSON(key string, src []byte) error {
	if s.loaded == nil {
		s.loaded = map[string][]byte{}
	}
	if len(s.loaded) < s.keep {
		s.loaded[key] = src
	}
	return nil
}
func (s *fakeStore) UpdateJSON(key string, src []byte) error { return s.CreateJSON(key, src) }

// TestRunEPrintSQLCountCheck checks that a load that stops part way is an
// error naming how many lines were produced and how many are in the store.
func TestRunEPrintSQLCountCheck(t *testing.T) {
	threeTheses(t)
	db := openScratchEPrints(t)
	store := &fakeStore{keep: 2}
	_, err := RunEPrintSQL(db, thesisQuery(t), nil, nil, store)
	if err == nil {
		t.Fatal("a store that drops a record must be an error")
	}
	if ExitClassOf(err) != ExitIO {
		t.Errorf("exit class = %d, want %d", ExitClassOf(err), ExitIO)
	}
	if !strings.Contains(err.Error(), "loaded 2 of 3") {
		t.Errorf("the error must name the counts, got: %s", err)
	}
}

// TestRunEPrintSQLLoadsCollection loads into a real dataset collection and
// reads an object back; running again with the same ids replaces, not fails.
func TestRunEPrintSQLLoadsCollection(t *testing.T) {
	threeTheses(t)
	db := openScratchEPrints(t)
	cName := path.Join(t.TempDir(), "thesis.ds")
	c, err := dataset.Init(cName, "")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for round := 0; round < 2; round++ {
		n, err := RunEPrintSQL(db, thesisQuery(t), []int{1, 2}, nil, c)
		if err != nil || n != 2 {
			t.Fatalf("round %d: n = %d, err = %v", round, n, err)
		}
	}
	if c.Length() != 2 {
		t.Errorf("collection holds %d objects, want 2", c.Length())
	}
	obj := map[string]interface{}{}
	if err := c.Read("1", obj); err != nil {
		t.Fatal(err)
	}
	md, _ := obj["metadata"].(map[string]interface{})
	if md == nil || md["title"] != "Thesis 1" {
		t.Errorf("object 1 = %v", obj)
	}
}

// TestRunEPrintSQLWriteFailure checks a failing output is an io error.
func TestRunEPrintSQLWriteFailure(t *testing.T) {
	threeTheses(t)
	db := openScratchEPrints(t)
	_, err := RunEPrintSQL(db, thesisQuery(t), nil, failingWriter{}, nil)
	if err == nil || ExitClassOf(err) != ExitIO {
		t.Errorf("a failing writer must give an io error, got %v (class %d)", err, ExitClassOf(err))
	}
}

// TestReadEPrintIDList checks the id list file: one id per line, blank lines
// and surrounding spaces ignored, a missing file is no_input, a bad line is data.
func TestReadEPrintIDList(t *testing.T) {
	dir := t.TempDir()
	good := path.Join(dir, "ids.txt")
	if err := os.WriteFile(good, []byte("12\n\n  34 \n56\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	ids, err := ReadEPrintIDList(good)
	if err != nil || fmt.Sprint(ids) != "[12 34 56]" {
		t.Errorf("ids = %v, err = %v, want [12 34 56]", ids, err)
	}
	if _, err := ReadEPrintIDList(path.Join(dir, "missing.txt")); ExitClassOf(err) != ExitNoInput {
		t.Errorf("missing file: class %d, want %d (%v)", ExitClassOf(err), ExitNoInput, err)
	}
	bad := path.Join(dir, "bad.txt")
	if err := os.WriteFile(bad, []byte("12\nabc\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadEPrintIDList(bad); ExitClassOf(err) != ExitData {
		t.Errorf("bad line: class %d, want %d (%v)", ExitClassOf(err), ExitData, err)
	}
	for _, word := range []string{"0", "-5"} {
		nonPositive := path.Join(dir, "nonpositive.txt")
		if err := os.WriteFile(nonPositive, []byte("12\n"+word+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadEPrintIDList(nonPositive); ExitClassOf(err) != ExitData {
			t.Errorf("id %s: class %d, want %d (%v)", word, ExitClassOf(err), ExitData, err)
		}
	}
	empty := path.Join(dir, "empty.txt")
	if err := os.WriteFile(empty, []byte("\n\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadEPrintIDList(empty); ExitClassOf(err) != ExitData {
		t.Errorf("an id list with no id: class %d, want %d (%v)", ExitClassOf(err), ExitData, err)
	}
}

// TestRunEPrintSQLFiles checks the file-driven entry point: query and id list
// from files, lines to the writer without a collection, a load with one, and
// the exit classes of a missing query file, id list and collection.
func TestRunEPrintSQLFiles(t *testing.T) {
	threeTheses(t)
	db := openScratchEPrints(t)
	dir := t.TempDir()
	queryFName := path.Join(dir, "q.sql")
	if err := os.WriteFile(queryFName, []byte(thesisQuery(t)), 0o600); err != nil {
		t.Fatal(err)
	}
	idsFName := path.Join(dir, "ids.txt")
	if err := os.WriteFile(idsFName, []byte("2\n3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	n, err := RunEPrintSQLFiles(db, queryFName, idsFName, "", &out)
	if err != nil || n != 2 || fmt.Sprint(keysOf(t, out.Bytes())) != "[2 3]" {
		t.Fatalf("lines: n = %d, err = %v, keys = %v", n, err, keysOf(t, out.Bytes()))
	}
	// with a collection the lines are loaded and not written
	cName := path.Join(dir, "thesis.ds")
	c, err := dataset.Init(cName, "")
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	out.Reset()
	if n, err := RunEPrintSQLFiles(db, queryFName, "", cName, &out); err != nil || n != 3 {
		t.Fatalf("load: n = %d, err = %v", n, err)
	}
	if out.Len() != 0 {
		t.Errorf("with a collection no lines are written, got %d bytes", out.Len())
	}
	c, err = dataset.Open(cName)
	if err != nil {
		t.Fatal(err)
	}
	if c.Length() != 3 {
		t.Errorf("collection holds %d, want 3", c.Length())
	}
	c.Close()
	for _, tc := range []struct {
		name             string
		query, ids, coll string
		want             ExitClass
	}{
		{"missing query file", path.Join(dir, "none.sql"), "", "", ExitNoInput},
		{"missing id list", queryFName, path.Join(dir, "none.txt"), "", ExitNoInput},
		{"missing collection", queryFName, "", path.Join(dir, "none.ds"), ExitNoInput},
	} {
		if _, err := RunEPrintSQLFiles(db, tc.query, tc.ids, tc.coll, &out); ExitClassOf(err) != tc.want {
			t.Errorf("%s: class %d, want %d (%v)", tc.name, ExitClassOf(err), tc.want, err)
		}
	}
}

// TestEPrintDSN checks the data source name built from the EPRINT_DB_*
// settings the same way GetEPrint builds it, and that a missing setting is a
// config error that names the variable and never shows the password.
func TestEPrintDSN(t *testing.T) {
	cfg := &Config{EPrintDbHost: "tcp(db.example.edu:3306)", EPrintDbUser: "reader", EPrintDbPassword: "s3cret", RepoID: "caltechthesis"}
	if got, err := EPrintDSN(cfg); err != nil || got != "reader:s3cret@tcp(db.example.edu:3306)/caltechthesis" {
		t.Errorf("dsn = %q, err = %v", got, err)
	}
	cfg.EPrintDbHost = "localhost"
	if got, err := EPrintDSN(cfg); err != nil || got != "reader:s3cret@/caltechthesis" {
		t.Errorf("localhost dsn = %q, err = %v", got, err)
	}
	for _, c := range []struct {
		name string
		cfg  Config
		want string
	}{
		{"no host", Config{EPrintDbUser: "u", EPrintDbPassword: "s3cret", RepoID: "r"}, "EPRINT_DB_HOST"},
		{"no user", Config{EPrintDbHost: "h", EPrintDbPassword: "s3cret", RepoID: "r"}, "EPRINT_DB_USER"},
		{"no password", Config{EPrintDbHost: "h", EPrintDbUser: "u", RepoID: "r"}, "EPRINT_DB_PASSWORD"},
		{"no repo id", Config{EPrintDbHost: "h", EPrintDbUser: "u", EPrintDbPassword: "s3cret"}, "REPO_ID"},
	} {
		_, err := EPrintDSN(&c.cfg)
		if ExitClassOf(err) != ExitConfig {
			t.Errorf("%s: class %d, want %d (%v)", c.name, ExitClassOf(err), ExitConfig, err)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error %q does not name %s", c.name, err, c.want)
		}
		if strings.Contains(err.Error(), "s3cret") {
			t.Errorf("%s: the error shows the password", c.name)
		}
	}
}

// TestRunEPrintSQLMode checks the entry point the command calls: it returns
// the DR-0014 exit code, writes lines to out, and writes the summary or the
// error to eout, never the other way round.
func TestRunEPrintSQLMode(t *testing.T) {
	threeTheses(t)
	openScratchEPrints(t)
	dsn := os.Getenv("EPRINTS_SCRATCH_DSN")
	dir := t.TempDir()
	queryFName := path.Join(dir, "q.sql")
	if err := os.WriteFile(queryFName, []byte(thesisQuery(t)), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, eout bytes.Buffer
	if code := RunEPrintSQLMode(dsn, queryFName, "", "", &out, &eout); code != 0 {
		t.Fatalf("exit %d, want 0: %s", code, eout.String())
	}
	if got := fmt.Sprint(keysOf(t, out.Bytes())); got != "[1 2 3]" {
		t.Errorf("keys = %s", got)
	}
	if !strings.Contains(eout.String(), "3 records") {
		t.Errorf("the summary must say how many records, got %q", eout.String())
	}
	// nothing found: exit 1, the message on eout, nothing on out
	idsFName := path.Join(dir, "ids.txt")
	if err := os.WriteFile(idsFName, []byte("999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	eout.Reset()
	if code := RunEPrintSQLMode(dsn, queryFName, idsFName, "", &out, &eout); code != int(ExitNegative) {
		t.Errorf("no record: exit %d, want %d", code, ExitNegative)
	}
	if out.Len() != 0 || eout.Len() == 0 {
		t.Errorf("no record: out %d bytes, eout %d bytes", out.Len(), eout.Len())
	}
	// missing query file: 66
	eout.Reset()
	if code := RunEPrintSQLMode(dsn, path.Join(dir, "none.sql"), "", "", &out, &eout); code != int(ExitNoInput) {
		t.Errorf("missing query: exit %d, want %d", code, ExitNoInput)
	}
	// a database that cannot be reached: 69, and the password is not echoed
	eout.Reset()
	if code := RunEPrintSQLMode("u:s3cret@tcp(127.0.0.1:1)/x", queryFName, "", "", &out, &eout); code != int(ExitUnavailable) {
		t.Errorf("unreachable: exit %d, want %d (%s)", code, ExitUnavailable, eout.String())
	}
	if strings.Contains(eout.String(), "s3cret") {
		t.Errorf("the error shows the password: %s", eout.String())
	}
}

// TestScrubDSNPassword checks the password is replaced everywhere it appears,
// and that a data source name with no password leaves the message alone.
func TestScrubDSNPassword(t *testing.T) {
	dsn := "reader:p@ss:w0rd@tcp(db.example.edu:3306)/caltechthesis"
	got := scrubDSNPassword("access denied for reader using p@ss:w0rd; p@ss:w0rd again", dsn)
	if strings.Contains(got, "w0rd") || strings.Count(got, "***") != 2 {
		t.Errorf("password not scrubbed everywhere: %q", got)
	}
	for _, d := range []string{"root@tcp(h)/db", "reader:@tcp(h)/db", "nothing", "@x"} {
		if got := scrubDSNPassword("an error", d); got != "an error" {
			t.Errorf("dsn %q changed the message to %q", d, got)
		}
	}
}
