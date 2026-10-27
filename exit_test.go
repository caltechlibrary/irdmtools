package irdmtools

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
	"testing"
)

// TestExitClassNumbers pins the numbers of the workspace exit codes
// (DR-0014): scripts read the numbers, not the names.
func TestExitClassNumbers(t *testing.T) {
	tests := []struct {
		class ExitClass
		want  int
	}{
		{ExitOK, 0}, {ExitNegative, 1}, {ExitUsage, 2}, {ExitData, 65}, {ExitNoInput, 66},
		{ExitUnavailable, 69}, {ExitInternal, 70}, {ExitCantCreate, 73}, {ExitIO, 74},
		{ExitTempFail, 75}, {ExitNoPermission, 77}, {ExitConfig, 78},
	}
	for _, tt := range tests {
		if int(tt.class) != tt.want {
			t.Errorf("expected class %s to be %d, got %d", tt.class, tt.want, int(tt.class))
		}
	}
}

// TestExitClassNames checks each class has the name DR-0014 gives it.
func TestExitClassNames(t *testing.T) {
	tests := map[ExitClass]string{
		ExitOK: "ok", ExitNegative: "negative", ExitUsage: "usage", ExitData: "data",
		ExitNoInput: "no_input", ExitUnavailable: "unavailable", ExitInternal: "internal",
		ExitCantCreate: "cant_create", ExitIO: "io", ExitTempFail: "temp_fail",
		ExitNoPermission: "no_permission", ExitConfig: "config",
	}
	for class, want := range tests {
		if got := class.String(); got != want {
			t.Errorf("expected %d to be named %q, got %q", int(class), want, got)
		}
	}
}

// TestExitClassUnknownName checks a number that is not a class still prints.
func TestExitClassUnknownName(t *testing.T) {
	if got := ExitClass(3).String(); got != "exit 3" {
		t.Errorf("expected %q, got %q", "exit 3", got)
	}
}

// TestExitConstructors checks each constructor makes an error of its class
// that keeps the message and the cause.
func TestExitConstructors(t *testing.T) {
	cause := errors.New("the cause")
	tests := []struct {
		name string
		err  error
		code int
	}{
		{"NegativeError", NegativeError(cause), 1},
		{"UsageError", UsageError(cause), 2},
		{"DataError", DataError(cause), 65},
		{"NoInputError", NoInputError(cause), 66},
		{"UnavailableError", UnavailableError(cause), 69},
		{"InternalError", InternalError(cause), 70},
		{"IOError", IOError(cause), 74},
		{"TempFailError", TempFailError(cause), 75},
		{"ConfigError", ConfigError(cause), 78},
	}
	for _, tt := range tests {
		if got := ExitCode(tt.err); got != tt.code {
			t.Errorf("%s: expected exit code %d, got %d", tt.name, tt.code, got)
		}
		if tt.err.Error() != "the cause" {
			t.Errorf("%s: expected the message of the cause, got %q", tt.name, tt.err.Error())
		}
		if !errors.Is(tt.err, cause) {
			t.Errorf("%s: expected the error to wrap its cause", tt.name)
		}
	}
	if err := UsageError(nil); err == nil || err.Error() == "" {
		t.Errorf("expected an error with a message for a nil cause, got %v", err)
	}
}

// TestExitCode checks the one function that maps an error to its code.
func TestExitCode(t *testing.T) {
	var pathErr error = &fs.PathError{Op: "read", Path: "x", Err: errors.New("disk")}
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"classified", ConfigError(errors.New("bad")), 78},
		{"classified and wrapped", fmt.Errorf("while loading: %w", DataError(errors.New("bad"))), 65},
		{"wrapped twice, the inner class wins over the outer stdlib one", fmt.Errorf("a: %w", NoInputError(fmt.Errorf("b: %w", fs.ErrNotExist))), 66},
		{"missing file", &fs.PathError{Op: "open", Path: "x", Err: fs.ErrNotExist}, 66},
		{"missing file from os", func() error { _, err := os.ReadFile("/no/such/file/anywhere"); return err }(), 66},
		{"permission", &fs.PathError{Op: "open", Path: "x", Err: fs.ErrPermission}, 77},
		{"exists", &fs.PathError{Op: "mkdir", Path: "x", Err: fs.ErrExist}, 73},
		{"other path error", pathErr, 74},
		{"url error", &url.Error{Op: "Get", URL: "https://x", Err: errors.New("refused")}, 69},
		{"net error", &net.OpError{Op: "dial", Err: errors.New("refused")}, 69},
		{"unclassified", errors.New("something"), 70},
	}
	for _, tt := range tests {
		if got := ExitCode(tt.err); got != tt.want {
			t.Errorf("%s: expected exit code %d, got %d", tt.name, tt.want, got)
		}
	}
}
