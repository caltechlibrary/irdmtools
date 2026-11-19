package irdmtools

// Tests against a scratch MySQL 8 that holds the structure of the CaltechTHESIS
// EPrints database and only rows the tests write themselves. They skip unless
// EPRINTS_SCRATCH_DSN is set, the way the other live tests do. Start the
// scratch server with scripts/scratch_mysql.bash; it prints the DSN.

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// openScratchEPrints opens the scratch EPrints database named by
// EPRINTS_SCRATCH_DSN, or skips the test when the variable is not set.
//
// @param t {*testing.T} the running test
// @returns {*sql.DB} an open handle, closed when the test ends
// @example
//
//	db := openScratchEPrints(t)
//	_, err := db.Exec("DELETE FROM eprint")
func openScratchEPrints(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("EPRINTS_SCRATCH_DSN")
	if dsn == "" {
		t.Skip("EPRINTS_SCRATCH_DSN not set; run scripts/scratch_mysql.bash start")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open scratch database: %s", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("scratch database not reachable: %s", err)
	}
	return db
}

// TestScratchSchemaHasThesisTables checks that every table the thesis SQL
// crosswalk reads is present in the scratch database.
func TestScratchSchemaHasThesisTables(t *testing.T) {
	db := openScratchEPrints(t)
	for _, name := range []string{
		"eprint", "document", "file", "document_relation_type",
		"document_permission_group", "eprint_thesis_advisor_name",
		"eprint_thesis_advisor_id", "eprint_thesis_advisor_role",
		"eprint_thesis_committee_name", "eprint_thesis_committee_role",
		"eprint_divisions", "eprint_option_major", "eprint_option_minor",
		"eprint_local_group",
	} {
		var n int
		err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.tables
			WHERE table_schema = DATABASE() AND table_name = ?`, name).Scan(&n)
		if err != nil {
			t.Fatalf("%s: %s", name, err)
		}
		if n != 1 {
			t.Errorf("table %s is missing from the scratch schema", name)
		}
	}
}

// TestScratchSchemaStartsEmpty guards against production rows reaching the
// scratch database: right after a fresh load every table is empty.
func TestScratchSchemaStartsEmpty(t *testing.T) {
	if os.Getenv("EPRINTS_SCRATCH_FRESH") == "" {
		t.Skip("EPRINTS_SCRATCH_FRESH not set; only valid straight after a fresh load")
	}
	db := openScratchEPrints(t)
	rows, err := db.Query(`SELECT table_name FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		names = append(names, s)
	}
	if len(names) < 70 {
		t.Fatalf("expected at least 70 tables, got %d", len(names))
	}
	for _, name := range names {
		var n int
		// name comes from information_schema, not from input
		if err := db.QueryRow("SELECT COUNT(*) FROM `" + name + "`").Scan(&n); err != nil {
			t.Fatalf("%s: %s", name, err)
		}
		if n != 0 {
			t.Errorf("table %s holds %d rows; the scratch schema must be structure only", name, n)
		}
	}
}
