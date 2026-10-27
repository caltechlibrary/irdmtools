package irdmtools

import (
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/url"
)

// ExitClass is one of the workspace's exit-code classes (workspace DR-0014).
// The value is the exit code. A program that uses these never gives a number
// here a private meaning, so a script can tell what went wrong from the number
// alone.
type ExitClass int

const (
	// ExitOK is success.
	ExitOK ExitClass = 0
	// ExitNegative is a correct run whose answer is no: nothing found, no such item.
	ExitNegative ExitClass = 1
	// ExitUsage is a wrong command line; nothing was attempted.
	ExitUsage ExitClass = 2
	// ExitData is content the program reads that is wrong: a malformed document or record.
	ExitData ExitClass = 65
	// ExitNoInput is a named input that is missing or unreadable.
	ExitNoInput ExitClass = 66
	// ExitUnavailable is a service the run needs that cannot be reached.
	ExitUnavailable ExitClass = 69
	// ExitInternal is a bug, or an error nothing classified.
	ExitInternal ExitClass = 70
	// ExitCantCreate is an output that cannot be created.
	ExitCantCreate ExitClass = 73
	// ExitIO is a read or write that failed part way.
	ExitIO ExitClass = 74
	// ExitTempFail is a failure that retrying later may fix, such as a rate limit.
	ExitTempFail ExitClass = 75
	// ExitNoPermission is access the operating system refused.
	ExitNoPermission ExitClass = 77
	// ExitConfig is a configuration file that is present but wrong.
	ExitConfig ExitClass = 78
)

// exitClassNames are the class names DR-0014 gives each code.
var exitClassNames = map[ExitClass]string{
	ExitOK: "ok", ExitNegative: "negative", ExitUsage: "usage", ExitData: "data",
	ExitNoInput: "no_input", ExitUnavailable: "unavailable", ExitInternal: "internal",
	ExitCantCreate: "cant_create", ExitIO: "io", ExitTempFail: "temp_fail",
	ExitNoPermission: "no_permission", ExitConfig: "config",
}

// String returns the class's name from DR-0014, for example "no_input".
//
// @returns {string} the class name, or "exit NN" for a number that is not a class
// @example
//
//	fmt.Println(ExitNoInput) // no_input
func (c ExitClass) String() string {
	if name, ok := exitClassNames[c]; ok {
		return name
	}
	return fmt.Sprintf("exit %d", int(c))
}

// ExitError is an error that carries the exit class its failure belongs to,
// wrapping the cause. Make one with a constructor such as NoInputError and read
// its code with ExitCode.
type ExitError struct {
	// Class is the exit class of the failure.
	Class ExitClass
	// Err is the cause; its message is the message of the ExitError.
	Err error
}

// Error returns the message of the cause.
//
// @returns {string} the cause's message
func (e *ExitError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the cause, so errors.Is and errors.As see through the class.
//
// @returns {error} the cause
func (e *ExitError) Unwrap() error {
	return e.Err
}

// classified makes an ExitError of a class. A nil cause becomes an error that
// names the class, so the result always has a message.
func classified(class ExitClass, err error) error {
	if err == nil {
		err = fmt.Errorf("%s error", class)
	}
	return &ExitError{Class: class, Err: err}
}

// NegativeError classifies err as a correct run whose answer is no (exit 1).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitNegative
// @example
//
//	return NegativeError(fmt.Errorf("no such DOI %q", doi))
func NegativeError(err error) error { return classified(ExitNegative, err) }

// UsageError classifies err as a wrong command line (exit 2).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitUsage
// @example
//
//	return UsageError(errors.New("expected a DOI"))
func UsageError(err error) error { return classified(ExitUsage, err) }

// DataError classifies err as wrong content in something the program read (exit 65).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitData
// @example
//
//	return DataError(fmt.Errorf("%s: %w", fName, err))
func DataError(err error) error { return classified(ExitData, err) }

// NoInputError classifies err as a named input that is missing (exit 66).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitNoInput
// @example
//
//	return NoInputError(fmt.Errorf("%s: %w", fName, fs.ErrNotExist))
func NoInputError(err error) error { return classified(ExitNoInput, err) }

// UnavailableError classifies err as a service that cannot be reached or that
// failed (exit 69).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitUnavailable
// @example
//
//	return UnavailableError(fmt.Errorf("CrossRef answered %d", status))
func UnavailableError(err error) error { return classified(ExitUnavailable, err) }

// InternalError classifies err as a bug or an error nothing else classified (exit 70).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitInternal
// @example
//
//	return InternalError(err)
func InternalError(err error) error { return classified(ExitInternal, err) }

// IOError classifies err as a read or write that failed part way (exit 74).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitIO
// @example
//
//	return IOError(err)
func IOError(err error) error { return classified(ExitIO, err) }

// TempFailError classifies err as a failure that retrying later may fix (exit 75).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitTempFail
// @example
//
//	return TempFailError(errors.New("rate limited, try again later"))
func TempFailError(err error) error { return classified(ExitTempFail, err) }

// ConfigError classifies err as a configuration file that is present but wrong (exit 78).
//
// @param err {error} the cause
// @returns {error} an *ExitError of class ExitConfig
// @example
//
//	return ConfigError(fmt.Errorf("%s: %w", fName, err))
func ConfigError(err error) error { return classified(ExitConfig, err) }

// ExitClassOf returns the class of an error: the class of the first ExitError
// it wraps, else the class the standard library's error says (a missing file
// 66, a refused one 77, an existing one 73, another path error 74, a network
// error 69), else ExitInternal. A nil error is ExitOK.
//
// @param err {error} the error, which may be nil
// @returns {ExitClass} the class of the failure
// @example
//
//	if ExitClassOf(err) == ExitNegative {
//		fmt.Println("nothing found")
//	}
func ExitClassOf(err error) ExitClass {
	if err == nil {
		return ExitOK
	}
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Class
	}
	var netErr net.Error
	var urlErr *url.Error
	var pathErr *fs.PathError
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return ExitNoInput
	case errors.Is(err, fs.ErrPermission):
		return ExitNoPermission
	case errors.Is(err, fs.ErrExist):
		return ExitCantCreate
	case errors.As(err, &urlErr), errors.As(err, &netErr):
		return ExitUnavailable
	case errors.As(err, &pathErr):
		return ExitIO
	}
	return ExitInternal
}

// ExitCode returns the process exit code for an error, the single place the
// mapping from an error to a code lives.
//
// @param err {error} the error, which may be nil
// @returns {int} the exit code: 0 for nil, else the code of the error's class
// @example
//
//	if err != nil {
//		fmt.Fprintln(os.Stderr, err)
//		os.Exit(ExitCode(err))
//	}
func ExitCode(err error) int {
	return int(ExitClassOf(err))
}
