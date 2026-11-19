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

// datesByType returns the date of each typed metadata.dates entry, keyed by
// the type id.
func datesByType(t *testing.T, rec *simplified.Record) map[string]string {
	t.Helper()
	got := map[string]string{}
	if rec.Metadata == nil {
		return got
	}
	for _, d := range rec.Metadata.Dates {
		if d.Type == nil {
			t.Errorf("date %q has no type", d.Date)
			continue
		}
		if _, dup := got[d.Type.ID]; dup {
			t.Errorf("date type %q appears twice", d.Type.ID)
		}
		got[d.Type.ID] = d.Date
	}
	return got
}

// TestThesisSQLTypedDates checks the thesis dates that are not part of
// thesis:thesis: approved_date, gradofc_approval_date and degree_date, with
// the ids the local date types use (the EPrints names minus "thesis_") and
// the precision EPrints holds. The ids pub_date and degree, which the Go
// crosswalk writes, are never used.
func TestThesisSQLTypedDates(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	// eprint 20: all three, each at a different precision
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type,
		thesis_approved_date_year, thesis_approved_date_month,
		gradofc_approval_date_year, gradofc_approval_date_month, gradofc_approval_date_day,
		thesis_degree_date_year)
		VALUES (20, 'archive', 'thesis', 2023, 6, 2023, 6, 20, 2023)`)
	// eprint 21: only a degree date
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, thesis_degree_date_year, thesis_degree_date_month)
		VALUES (21, 'archive', 'thesis', 1987, 6)`)
	// eprint 22: none, and the main date alone must not create one
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, date_year, date_type)
		VALUES (22, 'archive', 'thesis', 2001, 'degree')`)

	lines := runThesisSQL(t, db)
	want := map[string]map[string]string{
		"20": {"approved_date": "2023-06", "gradofc_approval_date": "2023-06-20", "degree_date": "2023"},
		"21": {"degree_date": "1987-06"},
		"22": {},
	}
	for id, w := range want {
		line, ok := lines[id]
		if !ok {
			t.Errorf("no output line for eprint %s", id)
			continue
		}
		got := datesByType(t, line.Record)
		if fmt.Sprint(got) != fmt.Sprint(w) {
			t.Errorf("eprint %s dates\n got  %v\n want %v", id, got, w)
		}
		if _, present := line.Raw["metadata"].(map[string]interface{})["dates"]; present && len(w) == 0 {
			t.Errorf("eprint %s has no dates but carries a dates key", id)
		}
	}
}

// TestThesisSQLDivision checks the division handling (brief ruling 16): the
// EPrints division ids go into caltech:division as a list of {id}, as
// caltech:groups does, and the first division's display name is copied into
// thesis:thesis.department for testing. A thesis with no division gets
// neither, and a division with no display name gets the id but no department.
func TestThesisSQLDivision(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	fixture(t, db, `INSERT INTO subject_name_name (subjectid, pos, name_name) VALUES
		('div_test_one', 0, 'Division of Test One'), ('div_test_two', 0, 'Division of Test Two')`)
	// eprint 30: two divisions; the department is the first, by list position
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (30, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_divisions (eprintid, pos, divisions) VALUES
		(30, 1, 'div_test_two'), (30, 0, 'div_test_one')`)
	// eprint 31: no division
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (31, 'archive', 'thesis')`)
	// eprint 32: a division with no display name
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (32, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_divisions (eprintid, pos, divisions) VALUES (32, 0, 'div_unnamed')`)
	// eprint 33: blank and NULL division rows count as no division
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (33, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_divisions (eprintid, pos, divisions) VALUES (33, 0, ''), (33, 1, NULL)`)
	// eprint 34: a division whose display name is blank has no department
	fixture(t, db, `INSERT INTO subject_name_name (subjectid, pos, name_name) VALUES ('div_blank', 0, '')`)
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (34, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_divisions (eprintid, pos, divisions) VALUES (34, 0, 'div_blank')`)

	lines := runThesisSQL(t, db)
	division := func(id string) []string {
		raw, _ := lines[id].Record.CustomFields["caltech:division"].([]interface{})
		ids := []string{}
		for _, item := range raw {
			m, _ := item.(map[string]interface{})
			entry := fmt.Sprint(m["id"])
			if title, ok := m["title"].(map[string]interface{}); ok {
				entry += "=" + fmt.Sprint(title["en"])
			}
			ids = append(ids, entry)
		}
		return ids
	}
	department := func(id string) interface{} {
		obj := thesisCustomField(t, lines[id])
		return obj["department"]
	}
	if got := division("30"); fmt.Sprint(got) != "[div_test_one=Division of Test One div_test_two=Division of Test Two]" {
		t.Errorf("eprint 30 caltech:division = %v, want the ids with their names", got)
	}
	if got := department("30"); got != "Division of Test One" {
		t.Errorf("eprint 30 department = %v, want the first division's name", got)
	}
	if _, present := lines["31"].Record.CustomFields["caltech:division"]; present {
		t.Errorf("eprint 31 has no division but carries caltech:division")
	}
	if got := department("31"); got != nil {
		t.Errorf("eprint 31 has no division but department = %v", got)
	}
	if got := division("32"); fmt.Sprint(got) != "[div_unnamed]" {
		t.Errorf("eprint 32 caltech:division = %v, want [div_unnamed]", got)
	}
	if got := department("32"); got != nil {
		t.Errorf("eprint 32 has no division name but department = %v", got)
	}
	if _, present := lines["33"].Record.CustomFields["caltech:division"]; present {
		t.Errorf("eprint 33 has only blank division rows but carries caltech:division")
	}
	if got := division("34"); fmt.Sprint(got) != "[div_blank]" {
		t.Errorf("eprint 34 caltech:division = %v, want [div_blank]", got)
	}
	if got := department("34"); got != nil {
		t.Errorf("eprint 34 has a blank division name but department = %v", got)
	}
}

// docFixture is one hand-written EPrints document with its files.
type docFixture struct {
	docid    int
	eprintid int
	pos      int
	security interface{} // string or nil
	main     string
	content  interface{}    // string or nil
	embargo  [3]interface{} // year, month, day
	files    []string       // file names; the first is the main file
}

// insertDoc writes a docFixture: the document row and one file row per name.
// Every file is 1000 bytes with the md5 "h<fileid>".
func insertDoc(t *testing.T, db *sql.DB, d docFixture) {
	t.Helper()
	fixture(t, db, `INSERT INTO document (docid, eprintid, pos, format, security, main, content,
		date_embargo_year, date_embargo_month, date_embargo_day, mime_type)
		VALUES (?, ?, ?, 'application/pdf', ?, ?, ?, ?, ?, ?, 'application/pdf')`,
		d.docid, d.eprintid, d.pos, d.security, d.main, d.content, d.embargo[0], d.embargo[1], d.embargo[2])
	for i, name := range d.files {
		fileid := d.docid*10 + i
		fixture(t, db, `INSERT INTO file (fileid, datasetid, objectid, filename, mime_type, hash, hash_type, filesize)
			VALUES (?, 'document', ?, ?, 'application/pdf', ?, 'MD5', 1000)`,
			fileid, d.docid, name, fmt.Sprintf("h%d", fileid))
	}
}

// sidecarDocs returns the sidecar's documents keyed by docid, as objects.
func sidecarDocs(t *testing.T, line *thesisLine) map[string]map[string]interface{} {
	t.Helper()
	out := map[string]map[string]interface{}{}
	raw, _ := line.Sidecar["documents"].([]interface{})
	for _, item := range raw {
		m, _ := item.(map[string]interface{})
		out[fmt.Sprint(m["docid"])] = m
	}
	return out
}

// publicFileKeys returns the sorted names of the files the record carries.
func publicFileKeys(line *thesisLine) []string {
	keys := []string{}
	if line.Record.Files != nil {
		for k := range line.Record.Files.Entries {
			keys = append(keys, k)
		}
	}
	sortStrings(keys)
	return keys
}

// sortStrings sorts a slice of strings in place.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// TestThesisSQLFileRule checks the file rule taken from ep3StaticSite's
// visibility.ts (brief ruling 17), one clause per document. A document is
// "public" when its eprint is in the archive, its security is public and it
// has no active embargo; "embargoed" when only an active embargo keeps a
// public document back (its files wait in an unpublished draft); "excluded"
// when its main file is an EPrints thumbnail or index name; and "non_public"
// otherwise. Only public documents' files appear in the record; the sidecar
// lists every document with its class and facts.
func TestThesisSQLFileRule(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	none := [3]interface{}{nil, nil, nil}
	// eprint 40: one document per clause, in the archive
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (40, 'archive', 'thesis')`)
	docs := []docFixture{
		{401, 40, 1, "public", "thesis.pdf", "final", none, []string{"thesis.pdf"}},
		{402, 40, 2, "public", "fig1.png", "supplemental", none, []string{"fig1.png", "data.csv"}},
		{403, 40, 3, "public", "second.pdf", "final", none, []string{"second.pdf"}},
		{404, 40, 4, "internal", "form.pdf", "approvalform", none, []string{"form.pdf"}},
		{405, 40, 5, "validuser", "campus.pdf", "final", [3]interface{}{2000, 1, 1}, []string{"campus.pdf"}},
		{406, 40, 6, nil, "nosec.pdf", "other", none, []string{"nosec.pdf"}},
		{407, 40, 7, "public", "later.pdf", "final", [3]interface{}{2100, 1, 1}, []string{"later.pdf"}},
		{408, 40, 8, "public", "Lightbox.JPG", nil, none, []string{"Lightbox.JPG"}},
		{409, 40, 9, "public", "preview.png", nil, none, []string{"preview.png"}},
		{410, 40, 10, "public", "MEDIUM.jpg", nil, none, []string{"MEDIUM.jpg"}},
		{411, 40, 11, "public", "small.gif", nil, none, []string{"small.gif"}},
		{412, 40, 12, "public", "indexcodes.txt", nil, none, []string{"indexcodes.txt"}},
		{413, 40, 13, "public", "released.pdf", "final", [3]interface{}{2000, 6, nil}, []string{"released.pdf"}},
		{414, 40, 14, "public", "lightbox_notes.pdf", "supplemental", none, []string{"lightbox_notes.pdf"}},
	}
	// eprint 41: not in the archive, so nothing is public
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (41, 'inbox', 'thesis')`)
	docs = append(docs, docFixture{415, 41, 1, "public", "draft.pdf", "final", none, []string{"draft.pdf"}})
	// eprint 42: only non-public documents
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (42, 'archive', 'thesis')`)
	docs = append(docs, docFixture{416, 42, 1, "internal", "arch.pdf", "archival", none, []string{"arch.pdf"}})
	// eprint 43: an embargo date that cannot be read counts as active
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (43, 'archive', 'thesis')`)
	docs = append(docs, docFixture{417, 43, 1, "public", "odd.pdf", "final", [3]interface{}{2000, 13, 1}, []string{"odd.pdf"}})
	for _, d := range docs {
		insertDoc(t, db, d)
	}

	lines := runThesisSQL(t, db)
	wantClass := map[string]map[string]string{
		"40": {"401": "public", "402": "public", "403": "public", "404": "non_public",
			"405": "non_public", "406": "non_public", "407": "embargoed",
			"408": "excluded", "409": "excluded", "410": "excluded", "411": "excluded", "412": "excluded",
			"413": "public", "414": "public"},
		"41": {"415": "non_public"},
		"42": {"416": "non_public"},
		"43": {"417": "embargoed"},
	}
	for id, w := range wantClass {
		got := map[string]string{}
		for docid, d := range sidecarDocs(t, lines[id]) {
			got[docid] = fmt.Sprint(d["class"])
		}
		if fmt.Sprint(got) != fmt.Sprint(w) {
			t.Errorf("eprint %s document classes\n got  %v\n want %v", id, got, w)
		}
	}
	// the record carries only the public documents' files, all of them
	wantKeys := map[string]string{
		"40": "[data.csv fig1.png lightbox_notes.pdf released.pdf second.pdf thesis.pdf]",
		"41": "[]", "42": "[]", "43": "[]",
	}
	for id, w := range wantKeys {
		if got := publicFileKeys(lines[id]); fmt.Sprint(got) != w {
			t.Errorf("eprint %s public files = %v, want %s", id, got, w)
		}
	}
	if f := lines["40"].Record.Files; f == nil || !f.Enabled {
		t.Errorf("eprint 40 has public files, so files.enabled must be true: %+v", f)
	}
	for _, id := range []string{"41", "42", "43"} {
		if f := lines[id].Record.Files; f != nil && f.Enabled {
			t.Errorf("eprint %s has no public files, so files.enabled must be false", id)
		}
	}
	// an entry carries the name, size and checksum
	if e := lines["40"].Record.Files.Entries["thesis.pdf"]; e == nil || e.Key != "thesis.pdf" || e.Size != 1000 || e.CheckSum != "md5:h4010" {
		t.Errorf("entry for thesis.pdf = %+v, want key thesis.pdf, size 1000, checksum md5:h4010", e)
	}
	// the sidecar keeps the facts for a document that is not public, files included
	d := sidecarDocs(t, lines["40"])["404"]
	if d["security"] != "internal" || d["main"] != "form.pdf" || d["content"] != "approvalform" {
		t.Errorf("sidecar document 404 = %v", d)
	}
	if files, _ := d["files"].([]interface{}); len(files) != 1 {
		t.Errorf("sidecar document 404 files = %v, want one", d["files"])
	}
}

// TestThesisSQLEmbargoBoundaries checks how an embargo date is compared with
// today, as ep3StaticSite does: an embargo that ends today is over, and a
// date with no month or day is compared as the first of the month or year.
// The dates come from the database server's own clock.
func TestThesisSQLEmbargoBoundaries(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	var y, m, d int
	if err := db.QueryRow(`SELECT YEAR(CURDATE()), MONTH(CURDATE()), DAY(CURDATE())`).Scan(&y, &m, &d); err != nil {
		t.Fatal(err)
	}
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (50, 'archive', 'thesis')`)
	docs := []docFixture{
		{501, 50, 1, "public", "today.pdf", "final", [3]interface{}{y, m, d}, []string{"today.pdf"}},
		{502, 50, 2, "public", "thisyear.pdf", "final", [3]interface{}{y, nil, nil}, []string{"thisyear.pdf"}},
		{503, 50, 3, "public", "thismonth.pdf", "final", [3]interface{}{y, m, nil}, []string{"thismonth.pdf"}},
		{504, 50, 4, "public", "nextyear.pdf", "final", [3]interface{}{y + 1, nil, nil}, []string{"nextyear.pdf"}},
		{505, 50, 5, "public", "tomorrow.pdf", "final", [3]interface{}{y + 1, 1, 1}, []string{"tomorrow.pdf"}},
	}
	for _, doc := range docs {
		insertDoc(t, db, doc)
	}
	lines := runThesisSQL(t, db)
	got := map[string]string{}
	for docid, doc := range sidecarDocs(t, lines["50"]) {
		got[docid] = fmt.Sprint(doc["class"])
	}
	want := map[string]string{"501": "public", "502": "public", "503": "public", "504": "embargoed", "505": "embargoed"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("document classes\n got  %v\n want %v", got, want)
	}
}

// TestThesisSQLEmbargoUntil checks caltech:embargo_until: the latest end date
// among the embargoed (public, not yet released) documents, at the precision
// EPrints holds. A permanent embargo (the sentinel years 2200 and 9999), an
// unreadable date, an expired embargo and a non-public document never set it;
// the sidecar marks a permanent embargo instead.
func TestThesisSQLEmbargoUntil(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	none := [3]interface{}{nil, nil, nil}
	for _, id := range []int{60, 61, 62, 63, 64, 65} {
		fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (?, 'archive', 'thesis')`, id)
	}
	docs := []docFixture{
		// 60: the latest of two embargoes wins, and keeps its precision
		{601, 60, 1, "public", "a.pdf", "final", [3]interface{}{2090, 6, nil}, []string{"a.pdf"}},
		{602, 60, 2, "public", "b.pdf", "final", [3]interface{}{2095, 3, 2}, []string{"b.pdf"}},
		{603, 60, 3, "public", "c.pdf", "final", none, []string{"c.pdf"}},
		// 61: a year-only embargo stays a year
		{611, 61, 1, "public", "a.pdf", "final", [3]interface{}{2100, nil, nil}, []string{"a.pdf"}},
		// 62: permanent embargoes are marked, not dated
		{621, 62, 1, "public", "a.pdf", "final", [3]interface{}{2200, 1, 1}, []string{"a.pdf"}},
		{622, 62, 2, "public", "b.pdf", "final", [3]interface{}{9999, nil, nil}, []string{"b.pdf"}},
		{623, 62, 3, "public", "c.pdf", "final", [3]interface{}{2100, 5, 5}, []string{"c.pdf"}},
		// 63: an expired embargo on a released document
		{631, 63, 1, "public", "a.pdf", "final", [3]interface{}{2000, 1, 1}, []string{"a.pdf"}},
		// 64: an unreadable date
		{641, 64, 1, "public", "a.pdf", "final", [3]interface{}{2000, 13, 1}, []string{"a.pdf"}},
		// 65: a campus-login document with a future embargo is not embargoed
		{651, 65, 1, "validuser", "a.pdf", "final", [3]interface{}{2100, 1, 1}, []string{"a.pdf"}},
	}
	for _, d := range docs {
		insertDoc(t, db, d)
	}
	lines := runThesisSQL(t, db)
	until := func(id string) interface{} {
		return lines[id].Record.CustomFields["caltech:embargo_until"]
	}
	for id, want := range map[string]interface{}{"60": "2095-03-02", "61": "2100", "62": "2100-05-05"} {
		if got := until(id); got != want {
			t.Errorf("eprint %s caltech:embargo_until = %v, want %v", id, got, want)
		}
	}
	for _, id := range []string{"63", "64", "65"} {
		if got := until(id); got != nil {
			t.Errorf("eprint %s must have no caltech:embargo_until, got %v", id, got)
		}
	}
	// the marker is on 621 and 622, and not on 623 or on an ordinary embargo
	side := sidecarDocs(t, lines["62"])
	for docid, want := range map[string]bool{"621": true, "622": true, "623": false} {
		if got := side[docid]["permanent_embargo"] == true; got != want {
			t.Errorf("sidecar document %s permanent_embargo = %v, want %v", docid, got, want)
		}
	}
	side60 := sidecarDocs(t, lines["60"])
	if side60["601"]["permanent_embargo"] == true {
		t.Errorf("an ordinary embargo must not be marked permanent")
	}
	// the sidecar keeps each document's own end date at its precision, and none where there is none
	for docid, want := range map[string]interface{}{"601": "2090-06", "602": "2095-03-02", "603": nil} {
		if got := side60[docid]["embargo_until"]; got != want {
			t.Errorf("sidecar document %s embargo_until = %v, want %v", docid, got, want)
		}
	}
}

// TestThesisSQLPublicationDate checks metadata.publication_date (brief ruling
// 15, accepted 2026-11-19): it is the main date of a degree-type eprint, at
// the precision EPrints holds, and never the datestamp, which records when
// the metadata was last corrected. A published or untyped date counts as
// well, as the Go crosswalk treats it; a submitted date does not, and an
// eprint with no usable date gets none, left for triage.
func TestThesisSQLPublicationDate(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	rows := []struct {
		id        int
		dateType  interface{}
		y, m, d   interface{}
		stampYear int
		want      interface{}
	}{
		{70, "degree", 2024, nil, nil, 2026, "2024"},
		{71, "degree", 2019, 6, 5, 2026, "2019-06-05"},
		{72, "published", 2010, 3, nil, 2026, "2010-03"},
		{73, nil, 2005, nil, nil, 2026, "2005"},
		{74, "submitted", 2003, nil, nil, 2026, nil},
		{75, "degree", nil, nil, nil, 2026, nil},
		{76, "", 1999, nil, nil, 2026, "1999"},
	}
	for _, r := range rows {
		fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, date_type, date_year, date_month, date_day,
			datestamp_year, datestamp_month, datestamp_day) VALUES (?, 'archive', 'thesis', ?, ?, ?, ?, ?, 11, 1)`,
			r.id, r.dateType, r.y, r.m, r.d, r.stampYear)
	}
	lines := runThesisSQL(t, db)
	for _, r := range rows {
		var got interface{}
		if md := lines[fmt.Sprint(r.id)].Record.Metadata; md != nil && md.PublicationDate != "" {
			got = md.PublicationDate
		}
		if got != r.want {
			t.Errorf("eprint %d (date_type %v) publication_date = %v, want %v", r.id, r.dateType, got, r.want)
		}
	}
}

// TestThesisSQLStateAndTombstone checks how an eprint's status reaches the
// loader. The sidecar carries eprint_status and rdm_state: archive is
// published, inbox and buffer (not yet accepted) are draft, deletion is
// deleted. A deleted eprint also carries a tombstone in the record, with the
// time its status changed when EPrints has it; COLD's query leaves deletions
// out, this one must not.
func TestThesisSQLStateAndTombstone(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	for _, st := range []struct {
		id     int
		status string
	}{{80, "inbox"}, {81, "buffer"}, {82, "deletion"}, {83, "archive"}, {84, "deletion"}} {
		fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (?, ?, 'thesis')`, st.id, st.status)
	}
	fixture(t, db, `UPDATE eprint SET status_changed_year = 2015, status_changed_month = 3, status_changed_day = 4,
		status_changed_hour = 5, status_changed_minute = 6, status_changed_second = 7 WHERE eprintid = 82`)
	// a thesis of another type is not part of this crosswalk
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (85, 'archive', 'article')`)

	lines := runThesisSQL(t, db)
	if _, ok := lines["85"]; ok {
		t.Errorf("an eprint that is not a thesis must not be emitted")
	}
	want := map[string]string{"80": "draft", "81": "draft", "82": "deleted", "83": "published", "84": "deleted"}
	status := map[string]string{"80": "inbox", "81": "buffer", "82": "deletion", "83": "archive", "84": "deletion"}
	for id, w := range want {
		line, ok := lines[id]
		if !ok {
			t.Errorf("no output line for eprint %s", id)
			continue
		}
		if got := line.Sidecar["rdm_state"]; got != w {
			t.Errorf("eprint %s rdm_state = %v, want %s", id, got, w)
		}
		if got := line.Sidecar["eprint_status"]; got != status[id] {
			t.Errorf("eprint %s eprint_status = %v, want %s", id, got, status[id])
		}
		hasTomb := line.Record.Tombstone != nil
		if hasTomb != (w == "deleted") {
			t.Errorf("eprint %s tombstone present = %v, want %v", id, hasTomb, w == "deleted")
		}
	}
	if tomb := lines["82"].Record.Tombstone; tomb != nil {
		if got := tomb.Timestamp.UTC().Format("2006-01-02T15:04:05Z"); got != "2015-03-04T05:06:07Z" {
			t.Errorf("eprint 82 tombstone timestamp = %s, want 2015-03-04T05:06:07Z", got)
		}
		if tomb.Reason == "" {
			t.Errorf("eprint 82 tombstone has no reason")
		}
	}
	if tomb := lines["84"].Record.Tombstone; tomb != nil && !tomb.Timestamp.IsZero() {
		t.Errorf("eprint 84 has no status_changed time but the tombstone has %v", tomb.Timestamp)
	}
}

// TestThesisSQLBaseRecord checks the base descriptive fields a thesis needs to
// load into RDM: resource type, title (with the quote escapes PHP magic quotes
// left in EPrints text removed, as COLD's query does), the author with their
// clpid and ORCID, the abstract, rights, the publisher and keywords.
func TestThesisSQLBaseRecord(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, title, abstract, keywords, rights,
		copyright_statement, publisher) VALUES (90, 'archive', 'thesis', ?, ?, ?, ?, ?, NULL)`,
		`On the \"Quiet\" Theory of Bob\'s Method`, `It doesn\'t say \"much\".`,
		`alpha; beta ;; cls; alpha; gamma\'s idea`, `Restricted for a time`, `Copyright 2020 Test`)
	fixture(t, db, `INSERT INTO eprint_creators_name (eprintid, pos, creators_name_family, creators_name_given)
		VALUES (90, 0, 'Authorone', 'Zed')`)
	fixture(t, db, `INSERT INTO eprint_creators_id (eprintid, pos, creators_id) VALUES (90, 0, 'Authorone-Z')`)
	fixture(t, db, `INSERT INTO eprint_creators_orcid (eprintid, pos, creators_orcid) VALUES (90, 0, '0000-0001-2345-6789')`)
	// eprint 91: nothing but the type; the publisher falls back to Caltech Library
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (91, 'archive', 'thesis')`)
	// eprint 92: a named publisher
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, publisher) VALUES (92, 'archive', 'thesis', 'California Institute of Technology')`)

	lines := runThesisSQL(t, db)
	md := lines["90"].Record.Metadata
	if got := md.ResourceType["id"]; got != "publication-thesis" {
		t.Errorf("resource_type id = %v, want publication-thesis", got)
	}
	if want := `On the "Quiet" Theory of Bob's Method`; md.Title != want {
		t.Errorf("title = %q, want %q", md.Title, want)
	}
	if want := `It doesn't say "much".`; md.Description != want {
		t.Errorf("description = %q, want %q", md.Description, want)
	}
	if len(md.Creators) != 1 {
		t.Fatalf("creators = %d, want 1", len(md.Creators))
	}
	p := md.Creators[0].PersonOrOrg
	if p.Type != "personal" || p.FamilyName != "Authorone" || p.GivenName != "Zed" || p.Name != "Authorone, Zed" {
		t.Errorf("creator = %+v", p)
	}
	ids := map[string]string{}
	for _, id := range p.Identifiers {
		ids[id.Scheme] = id.Identifier
	}
	if fmt.Sprint(ids) != "map[clpid:Authorone-Z orcid:0000-0001-2345-6789]" {
		t.Errorf("creator identifiers = %v", ids)
	}
	subjects := []string{}
	for _, s := range md.Subjects {
		subjects = append(subjects, s.Subject)
	}
	if fmt.Sprint(subjects) != "[alpha beta gamma's idea]" {
		t.Errorf("subjects = %v, want [alpha beta gamma's idea]", subjects)
	}
	rights := []string{}
	for _, r := range md.Rights {
		rights = append(rights, r.Description["en"])
	}
	if fmt.Sprint(rights) != "[Restricted for a time Copyright 2020 Test]" {
		t.Errorf("rights = %v", rights)
	}
	if md.Publisher != "Caltech Library" {
		t.Errorf("publisher = %q, want Caltech Library", md.Publisher)
	}
	if got := lines["92"].Record.Metadata.Publisher; got != "California Institute of Technology" {
		t.Errorf("named publisher = %q", got)
	}
	// a thesis with nothing else still has the resource type and publisher, and no creators or title key
	m91 := lines["91"].Raw["metadata"].(map[string]interface{})
	for _, k := range []string{"title", "creators", "subjects", "rights", "description"} {
		if _, present := m91[k]; present {
			t.Errorf("eprint 91 has nothing for %s but carries it", k)
		}
	}
}

// TestThesisSQLCreatorsAndBlanks checks a creator list whose members carry
// different identifiers (matched by position, not shared), and that blank
// strings are treated as absent.
func TestThesisSQLCreatorsAndBlanks(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	// eprint 93: blank title, abstract and publisher
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, title, abstract, publisher, keywords, rights)
		VALUES (93, 'archive', 'thesis', '', '', '', '', '')`)
	// eprint 94: two creators; the first has only a clpid, the second only an ORCID, the third neither
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (94, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_creators_name (eprintid, pos, creators_name_family, creators_name_given)
		VALUES (94, 0, 'First', 'Ann'), (94, 1, 'Second', 'Ben'), (94, 2, 'Third', 'Cat')`)
	fixture(t, db, `INSERT INTO eprint_creators_id (eprintid, pos, creators_id) VALUES (94, 0, 'First-A'), (94, 2, '')`)
	fixture(t, db, `INSERT INTO eprint_creators_orcid (eprintid, pos, creators_orcid) VALUES (94, 1, '0000-0002-0000-0001')`)

	lines := runThesisSQL(t, db)
	m93 := lines["93"].Raw["metadata"].(map[string]interface{})
	for _, k := range []string{"title", "description", "subjects", "rights"} {
		if _, present := m93[k]; present {
			t.Errorf("eprint 93 has a blank %s but carries it", k)
		}
	}
	if got := lines["93"].Record.Metadata.Publisher; got != "Caltech Library" {
		t.Errorf("a blank publisher must fall back to Caltech Library, got %q", got)
	}
	got := map[string]string{}
	for _, c := range lines["94"].Record.Metadata.Creators {
		ids := []string{}
		for _, id := range c.PersonOrOrg.Identifiers {
			ids = append(ids, id.Scheme+"="+id.Identifier)
		}
		got[c.PersonOrOrg.Name] = fmt.Sprint(ids)
	}
	want := "map[First, Ann:[clpid=First-A] Second, Ben:[orcid=0000-0002-0000-0001] Third, Cat:[]]"
	if fmt.Sprint(got) != want {
		t.Errorf("creator identifiers\n got  %v\n want %s", got, want)
	}
}

// TestThesisSQLTextCleaning checks that the backslash PHP magic quotes left
// before a quote is removed from every text field the query carries, and that
// a backslash that is not before a quote is kept.
func TestThesisSQLTextCleaning(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type, title, abstract, keywords, rights, copyright_statement)
		VALUES (95, 'archive', 'thesis', ?, ?, ?, ?, ?)`,
		`T \"a\" \'b\' c\d`, `A \"a\" \'b\'`, `k\'1; k\"2`, `R \"r\" \'s\'`, `C \"c\" \'d\'`)
	lines := runThesisSQL(t, db)
	md := lines["95"].Record.Metadata
	if want := `T "a" 'b' c\d`; md.Title != want {
		t.Errorf("title = %q, want %q", md.Title, want)
	}
	if want := `A "a" 'b'`; md.Description != want {
		t.Errorf("description = %q, want %q", md.Description, want)
	}
	subjects := []string{}
	for _, s := range md.Subjects {
		subjects = append(subjects, s.Subject)
	}
	if fmt.Sprint(subjects) != `[k'1 k"2]` {
		t.Errorf("subjects = %v", subjects)
	}
	rights := []string{}
	for _, r := range md.Rights {
		rights = append(rights, r.Description["en"])
	}
	if fmt.Sprint(rights) != `[R "r" 's' C "c" 'd']` {
		t.Errorf("rights = %v", rights)
	}
}

// TestThesisSQLOptions checks caltech:option_major and caltech:option_minor
// (the author's names, 2026-11-19): the EPrints option codes as lists of {id}
// in list order, as caltech:division does. Blank rows count as none.
func TestThesisSQLOptions(t *testing.T) {
	db := openScratchEPrints(t)
	resetScratch(t, db)
	fixture(t, db, `INSERT INTO subject_name_name (subjectid, pos, name_name) VALUES
		('chemistry', 0, 'Chemistry'), ('physics', 0, 'Physics'), ('art', 0, 'Art History')`)
	// eprint 100: two majors (inserted out of order), one minor
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (100, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_option_major (eprintid, pos, option_major) VALUES (100, 1, 'physics'), (100, 0, 'chemistry')`)
	fixture(t, db, `INSERT INTO eprint_option_minor (eprintid, pos, option_minor) VALUES (100, 0, 'history')`)
	// eprint 101: only a minor, and a blank major row
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (101, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_option_major (eprintid, pos, option_major) VALUES (101, 0, ''), (101, 1, NULL)`)
	fixture(t, db, `INSERT INTO eprint_option_minor (eprintid, pos, option_minor) VALUES (101, 0, 'art')`)
	// eprint 102: neither
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (102, 'archive', 'thesis')`)
	// eprint 104: the lowest-position name wins, and a code with only a later name still gets it
	fixture(t, db, `INSERT INTO subject_name_name (subjectid, pos, name_name) VALUES
		('multi', 1, 'Second'), ('multi', 0, 'First'), ('late', 1, 'Late')`)
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (104, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_option_major (eprintid, pos, option_major) VALUES (104, 0, 'multi')`)
	fixture(t, db, `INSERT INTO eprint_option_minor (eprintid, pos, option_minor) VALUES (104, 0, 'late')`)
	// eprint 103: blank minor rows
	fixture(t, db, `INSERT INTO eprint (eprintid, eprint_status, type) VALUES (103, 'archive', 'thesis')`)
	fixture(t, db, `INSERT INTO eprint_option_minor (eprintid, pos, option_minor) VALUES (103, 0, ''), (103, 1, NULL)`)

	lines := runThesisSQL(t, db)
	// each entry is "code" or "code=name"; a code with no name in EPrints has no title
	ids := func(id, key string) interface{} {
		raw, present := lines[id].Record.CustomFields[key]
		if !present {
			return nil
		}
		out := []string{}
		for _, item := range raw.([]interface{}) {
			m := item.(map[string]interface{})
			entry := fmt.Sprint(m["id"])
			if title, ok := m["title"].(map[string]interface{}); ok {
				entry += "=" + fmt.Sprint(title["en"])
			}
			out = append(out, entry)
		}
		return fmt.Sprint(out)
	}
	for _, c := range []struct {
		id, key string
		want    interface{}
	}{
		{"100", "caltech:option_major", "[chemistry=Chemistry physics=Physics]"},
		{"100", "caltech:option_minor", "[history]"},
		{"101", "caltech:option_major", nil},
		{"101", "caltech:option_minor", "[art=Art History]"},
		{"102", "caltech:option_major", nil},
		{"102", "caltech:option_minor", nil},
		{"103", "caltech:option_minor", nil},
		{"104", "caltech:option_major", "[multi=First]"},
		{"104", "caltech:option_minor", "[late=Late]"},
	} {
		if got := ids(c.id, c.key); got != c.want {
			t.Errorf("eprint %s %s = %v, want %v", c.id, c.key, got, c.want)
		}
	}
}
