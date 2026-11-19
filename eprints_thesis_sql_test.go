package irdmtools

// Tests for scripts/eprints_thesis_to_rdm.sql, the SQL crosswalk from the
// CaltechTHESIS EPrints tables to RDM-shaped JSON (irdmtools DR-0010). They run
// the query against the scratch MySQL (see eprints_scratch_mysql_test.go) with
// rows the tests write by hand, and skip when the scratch DSN is not set.

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	// Caltech Library Packages
	"github.com/caltechlibrary/simplified"
)

// thesisSidecarKey is the one top-level key a crosswalk line carries that is
// not part of an RDM record: per-document facts for the loader.
const thesisSidecarKey = "eprints_sidecar"

// thesisSQLPath is the crosswalk query file, relative to the package.
const thesisSQLPath = "scripts/eprints_thesis_to_rdm.sql"

// thesisLine is one decoded output line: the RDM record, checked strictly,
// and the sidecar kept raw.
type thesisLine struct {
	Record  *simplified.Record
	Sidecar map[string]interface{}
	Raw     map[string]interface{}
}

// resetScratch empties every table of the scratch database so a test starts
// from rows it wrote itself.
//
// @param t {*testing.T} the running test
// @param db {*sql.DB} the scratch database
// @example
//
//	resetScratch(t, db)
func resetScratch(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query(`SELECT table_name FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'`)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		names = append(names, s)
	}
	rows.Close()
	for _, name := range names {
		// name comes from information_schema, not from input
		if _, err := db.Exec("TRUNCATE TABLE `" + name + "`"); err != nil {
			t.Fatalf("truncate %s: %s", name, err)
		}
	}
}

// fixture executes one hand-written INSERT or UPDATE, failing the test on error.
//
// @param t {*testing.T} the running test
// @param db {*sql.DB} the scratch database
// @param stmt {string} the statement, with ? placeholders
// @param args {...interface{}} the placeholder values
// @example
//
//	fixture(t, db, "INSERT INTO eprint (eprintid, type) VALUES (?, ?)", 1, "thesis")
func fixture(t *testing.T, db *sql.DB, stmt string, args ...interface{}) {
	t.Helper()
	if _, err := db.Exec(stmt, args...); err != nil {
		t.Fatalf("fixture %q: %s", stmt, err)
	}
}

// runThesisSQL runs the crosswalk query and decodes every line strictly: the
// sidecar is split off, then the rest must fit simplified.Record with no
// unknown field, so a misspelled or misplaced key fails the test.
//
// @param t {*testing.T} the running test
// @param db {*sql.DB} the scratch database
// @returns {map[string]*thesisLine} the decoded lines keyed by the eprintid
// @example
//
//	lines := runThesisSQL(t, db)
//	rec := lines["1"].Record
func runThesisSQL(t *testing.T, db *sql.DB) map[string]*thesisLine {
	t.Helper()
	src, err := os.ReadFile(thesisSQLPath)
	if err != nil {
		t.Fatalf("read %s: %s", thesisSQLPath, err)
	}
	rows, err := db.Query(string(src))
	if err != nil {
		t.Fatalf("run %s: %s", thesisSQLPath, err)
	}
	defer rows.Close()
	out := map[string]*thesisLine{}
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			t.Fatal(err)
		}
		line := &thesisLine{Raw: map[string]interface{}{}}
		if err := json.Unmarshal(raw, &line.Raw); err != nil {
			t.Fatalf("output line is not JSON: %s\n%s", err, raw)
		}
		if sc, ok := line.Raw[thesisSidecarKey].(map[string]interface{}); ok {
			line.Sidecar = sc
		}
		delete(line.Raw, thesisSidecarKey)
		rest, _ := json.Marshal(line.Raw)
		dec := json.NewDecoder(bytes.NewReader(rest))
		dec.DisallowUnknownFields()
		line.Record = new(simplified.Record)
		if err := dec.Decode(line.Record); err != nil {
			t.Fatalf("line does not fit simplified.Record: %s\n%s", err, rest)
		}
		out[eprintIDOf(t, line)] = line
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// eprintIDOf returns the EPrints id a line carries in its identifiers.
func eprintIDOf(t *testing.T, line *thesisLine) string {
	t.Helper()
	if line.Record.Metadata != nil {
		for _, ident := range line.Record.Metadata.Identifiers {
			if ident.Scheme == "eprintid" {
				return ident.Identifier
			}
		}
	}
	t.Fatalf("line has no eprintid identifier: %+v", line.Raw)
	return ""
}

// contributorsByName returns the role id of each contributor, keyed by the
// "Family, Given" name, so a test does not depend on output order.
func contributorsByName(rec *simplified.Record) map[string]string {
	got := map[string]string{}
	if rec.Metadata == nil {
		return got
	}
	for _, c := range rec.Metadata.Contributors {
		role := ""
		if c.Role != nil {
			role = c.Role.ID
		}
		got[c.PersonOrOrg.Name] = role
	}
	return got
}

// TestThesisSQLRoles checks the advisor and committee role ids, and the
// default role for a row that has none (brief ruling 17): committee rows
// become committee_member, an advisor row becomes advisor, or co_advisor when
// the eprint has more than one advisor row.
func TestThesisSQLRoles(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	// eprint 1: every explicit role value EPrints holds
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (1, 'archive', 'thesis')`)
	advisors := []struct {
		pos           int
		family, given string
		role          interface{}
	}{
		{0, "Advisorone", "Ann", "advisor"},
		{1, "Coadvisor", "Bo", "co-advisor"},
	}
	for _, a := range advisors {
		fixture(t, db, `INSERT INTO eprint_thesis_advisor_name (eprintid, pos, thesis_advisor_name_family, thesis_advisor_name_given) VALUES (1, ?, ?, ?)`, a.pos, a.family, a.given)
		fixture(t, db, `INSERT INTO eprint_thesis_advisor_role (eprintid, pos, thesis_advisor_role) VALUES (1, ?, ?)`, a.pos, a.role)
	}
	committee := []struct {
		pos           int
		family, given string
		role          string
	}{
		{0, "Chairone", "Cy", "chair"},
		{1, "Cochair", "Di", "co-chair"},
		{2, "Memberone", "Ed", "member"},
	}
	for _, c := range committee {
		fixture(t, db, `INSERT INTO eprint_thesis_committee_name (eprintid, pos, thesis_committee_name_family, thesis_committee_name_given) VALUES (1, ?, ?, ?)`, c.pos, c.family, c.given)
		fixture(t, db, `INSERT INTO eprint_thesis_committee_role (eprintid, pos, thesis_committee_role) VALUES (1, ?, ?)`, c.pos, c.role)
	}
	// eprint 2: one advisor row and one committee row, both with no role row
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (2, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_thesis_advisor_name (eprintid, pos, thesis_advisor_name_family, thesis_advisor_name_given) VALUES (2, 0, 'Soleadvisor', 'Fay')`)
	fixture(t, db, `INSERT INTO eprint_thesis_committee_name (eprintid, pos, thesis_committee_name_family, thesis_committee_name_given) VALUES (2, 0, 'Plainmember', 'Gus')`)
	// eprint 3: two advisor rows, neither with a role: both are co-advisors
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (3, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_thesis_advisor_name (eprintid, pos, thesis_advisor_name_family, thesis_advisor_name_given) VALUES (3, 0, 'Pairone', 'Hal'), (3, 1, 'Pairtwo', 'Ida')`)

	lines := runThesisSQL(t, db)
	want := map[string]map[string]string{
		"1": {
			"Advisorone, Ann": "advisor", "Coadvisor, Bo": "co_advisor",
			"Chairone, Cy": "committee_chair", "Cochair, Di": "committee_co_chair",
			"Memberone, Ed": "committee_member",
		},
		"2": {"Soleadvisor, Fay": "advisor", "Plainmember, Gus": "committee_member"},
		"3": {"Pairone, Hal": "co_advisor", "Pairtwo, Ida": "co_advisor"},
	}
	for id, w := range want {
		line, ok := lines[id]
		if !ok {
			t.Errorf("no output line for eprint %s", id)
			continue
		}
		got := contributorsByName(line.Record)
		if fmt.Sprint(got) != fmt.Sprint(w) {
			t.Errorf("eprint %s contributors\n got  %v\n want %v", id, got, w)
		}
	}
}

// thesisCustomField returns the "thesis:thesis" object of a line, or nil.
func thesisCustomField(t *testing.T, line *thesisLine) map[string]interface{} {
	t.Helper()
	raw, ok := line.Record.CustomFields["thesis:thesis"]
	if !ok {
		return nil
	}
	obj, ok := raw.(map[string]interface{})
	if !ok {
		t.Fatalf("thesis:thesis is %T, expected an object", raw)
	}
	for k := range obj {
		switch k {
		case "university", "department", "type", "date_submitted", "date_defended":
		default:
			t.Errorf("thesis:thesis carries %q; the stock field has only university, department, type, date_submitted and date_defended", k)
		}
	}
	return obj
}

// TestThesisSQLThesisField checks the stock thesis:thesis object: the
// university, the thesis type, and the submitted and defended dates. A date
// keeps the precision EPrints holds: a year alone stays a year, a year and
// month stay a year and month, and a missing month or day is never padded
// with "01" (COLD's query pads).
func TestThesisSQLThesisField(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	// eprint 10: a full submitted date, a year-and-month defense date
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, institution, thesis_type,
		thesis_submitted_date_year, thesis_submitted_date_month, thesis_submitted_date_day,
		thesis_defense_date_year, thesis_defense_date_month)
		VALUES (10, 'archive', 'thesis', 'California Institute of Technology', 'phd', 2024, 5, 12, 2024, 3)`)
	// eprint 11: a year-only submitted date, no defense date, masters
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, institution, thesis_type,
		thesis_submitted_date_year)
		VALUES (11, 'archive', 'thesis', 'California Institute of Technology', 'masters', 1998)`)
	// eprint 12: a day with no month is a year-only date
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, institution, thesis_type,
		thesis_defense_date_year, thesis_defense_date_day)
		VALUES (12, 'archive', 'thesis', 'California Institute of Technology', 'engd', 2020, 15)`)
	// eprint 13: nothing to say, so no thesis:thesis object
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (13, 'archive', 'thesis')`)
	// eprint 14: a year-and-month submitted date; an empty institution and type are left out
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, institution, thesis_type,
		thesis_submitted_date_year, thesis_submitted_date_month)
		VALUES (14, 'archive', 'thesis', '', '', 2011, 9)`)
	// eprint 15: an empty institution alone leaves no thesis:thesis object
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, institution) VALUES (15, 'archive', 'thesis', '')`)

	lines := runThesisSQL(t, db)
	want := map[string]map[string]interface{}{
		"10": {"university": "California Institute of Technology", "type": "PhD",
			"date_submitted": "2024-05-12", "date_defended": "2024-03"},
		"11": {"university": "California Institute of Technology", "type": "Masters",
			"date_submitted": "1998"},
		"12": {"university": "California Institute of Technology", "type": "Engd",
			"date_defended": "2020"},
		"13": nil,
		"14": {"date_submitted": "2011-09"},
		"15": nil,
	}
	for id, w := range want {
		line, ok := lines[id]
		if !ok {
			t.Errorf("no output line for eprint %s", id)
			continue
		}
		got := thesisCustomField(t, line)
		if w == nil {
			if _, present := line.Record.CustomFields["thesis:thesis"]; present {
				t.Errorf("eprint %s has nothing to say but carries a thesis:thesis object: %v", id, got)
			}
			continue
		}
		if fmt.Sprint(got) != fmt.Sprint(w) && !(len(got) == 0 && len(w) == 0) {
			t.Errorf("eprint %s thesis:thesis\n got  %v\n want %v", id, got, w)
		}
	}
}
