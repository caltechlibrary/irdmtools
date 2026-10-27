package irdmtools

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestStructTagsParse checks that the tags `go vet` reported as malformed
// parse and carry the intended keys. A tag with a missing quote or a comma
// between its pairs is silently ignored by encoding/json and yaml.
func TestStructTagsParse(t *testing.T) {
	tests := []struct {
		typ   reflect.Type
		field string
		key   string
		want  string
	}{
		{reflect.TypeOf(Config{}), "Debug", "json", "-"},
		{reflect.TypeOf(Config{}), "Debug", "yaml", "-"},
		{reflect.TypeOf(EPrintRest{}), "DbHost", "json", "db_host,omitempty"},
		{reflect.TypeOf(EPrintRest{}), "DbUser", "json", "db_user,omitempty"},
		{reflect.TypeOf(EPrintRest{}), "DbPassword", "json", "db_password,omitempty"},
	}
	for _, tt := range tests {
		f, ok := tt.typ.FieldByName(tt.field)
		if !ok {
			t.Errorf("%s has no field %s", tt.typ.Name(), tt.field)
			continue
		}
		got, ok := f.Tag.Lookup(tt.key)
		if !ok || got != tt.want {
			t.Errorf("%s.%s %s tag, expected %q, got %q (found %t)", tt.typ.Name(), tt.field, tt.key, tt.want, got, ok)
		}
	}
}

// TestEPrintRestDecodesDbKeys checks that the database keys of an EPrintRest
// JSON document reach the struct.
func TestEPrintRestDecodesDbKeys(t *testing.T) {
	src := []byte(`{"db_host":"db.example.edu","db_user":"reader","db_password":"secret"}`)
	app := new(EPrintRest)
	if err := json.Unmarshal(src, app); err != nil {
		t.Fatal(err)
	}
	if app.DbHost != "db.example.edu" || app.DbUser != "reader" || app.DbPassword != "secret" {
		t.Errorf("expected host, user and password to be set, got %q %q %q", app.DbHost, app.DbUser, app.DbPassword)
	}
}

// TestConfigDebugIsNotConfigurable checks that Debug, which is a run-time flag
// set from the command line, is not read from a YAML or JSON configuration.
func TestConfigDebugIsNotConfigurable(t *testing.T) {
	cfg := new(Config)
	if err := yaml.Unmarshal([]byte("debug: true\n"), cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Debug {
		t.Error("a debug key in YAML set Config.Debug")
	}
	cfg = new(Config)
	if err := json.Unmarshal([]byte(`{"debug":true}`), cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Debug {
		t.Error("a debug key in JSON set Config.Debug")
	}
}
