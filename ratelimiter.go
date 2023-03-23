package irdmtools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// RateLimit holds the values used to play nice with OAI-PMH or REST API.
// It normally is extracted from the response header.
type RateLimit struct {
	// Limit maps to X-RateLimit-Limit
	Limit int `json:"limit,omitempty"`
	// Remaining maps to X-RateLimit-Remaining
	Remaining int `json:"remaining,omitempty"`
	// Reset maps to X-RateLimit-Reset
	Reset int `json:"reset,omitempty"`
	// TimeUnit must be set explicitly.  Used by SecondsToWait
	// ```
	//    rateLimit.TimeUnit = time.Second
	// ```
	TimeUnit time.Duration `json:"time_unit,omitempty"`
}

// FromResponse takes an http.Response struct and extracts
// the header values realated to rate limits (e.g. X-RateLite-Limit)
//
// ```
// rl := new(RateLimit)
// rl.FromResponse(response, time.Minute)
// ```
func (rl *RateLimit) FromResponse(resp *http.Response, timeUnit time.Duration) {
	if rl == nil {
		rl = new(RateLimit)
	}
	rl.TimeUnit = timeUnit
	l := resp.Header.Values("X-RateLimit-Limit")
	if len(l) > 0 {
		if val, err := strconv.Atoi(l[0]); err == nil {
			rl.Limit = val
		} else {
			rl.Limit = 0
		}
	}
	l = resp.Header.Values("X-RateLimit-Remaining")
	if len(l) > 0 {
		if val, err := strconv.Atoi(l[0]); err == nil {
			rl.Remaining = val
		} else {
			rl.Remaining = 0
		}
	}
	l = resp.Header.Values("X-RateLimit-Reset")
	if len(l) > 0 {
		if val, err := strconv.Atoi(l[0]); err == nil {
			rl.Reset = val
		} else {
			rl.Reset = 0
		}
	}
}

// FromHeader takes an http.Header (e.g. http.Response.Header)
// and updates a rate limit struct.
//
// ```
// rl := new(RateLimit)
// rl.FromHeader(header, time.Hour)
// ```
func (rl *RateLimit) FromHeader(header http.Header, timeUnit time.Duration) {
	if rl == nil {
		rl = new(RateLimit)
	}
	rl.TimeUnit = timeUnit
	l := header.Values("X-RateLimit-Limit")
	if len(l) > 0 {
		if val, err := strconv.Atoi(l[0]); err == nil {
			rl.Limit = val
		} else {
			rl.Limit = 0
		}
	}
	l = header.Values("X-RateLimit-Remaining")
	if len(l) > 0 {
		if val, err := strconv.Atoi(l[0]); err == nil {
			rl.Remaining = val
		} else {
			rl.Remaining = 0
		}
	}
	l = header.Values("X-RateLimit-Reset")
	if len(l) > 0 {
		if val, err := strconv.Atoi(l[0]); err == nil {
			rl.Reset = val
		} else {
			rl.Reset = 0
		}
	}
}

func (rl *RateLimit) Fprintf(out io.Writer) {
	fmt.Fprintf(out, "limit %d\n", rl.Limit)
	fmt.Fprintf(out, "remaining %d\n", rl.Remaining)
	fmt.Fprintf(out, "time unit %s\n", rl.TimeUnit)
	if rl.Reset > 0 {
		resetTime := time.Unix(int64(rl.Reset), 0)
		fmt.Fprintf(os.Stderr, "reset in %s at %s\n", resetTime.Sub(time.Now()).Truncate(time.Second), resetTime.Format("03:04PM"))
	}
}

func (rl *RateLimit) String() string {
	s := []string{}
	s[0] = fmt.Sprintf("limit %d", rl.Limit)
	s[1] = fmt.Sprintf("remaining %d", rl.Remaining)
	s[2] = fmt.Sprintf("using time unit %q", rl.TimeUnit)
	if rl.Reset > 0 {
		resetTime := time.Unix(int64(rl.Reset), 0)
		s[3] = fmt.Sprintf("reset in %s at %s", resetTime.Sub(time.Now()).Truncate(time.Second), resetTime.Format("03:04PM"))
	}
	return strings.Join(s, "\n")
}

// SecondsToWait returns the number of seconds (as a time.Duratin) to wait to avoid
// a http status code 429 and a ratio (float64) of remaining per request limit.
//
// ```
//
//	rl := new(RateLimit)
//	rl.FromHeader(response.Header)
//	rl.TimeUnit = time.Minute
//	secondsToWait, remainingPerLimit := rl.TimeToWait()
//	if remainingPerLimit <= 0.5 {
//	    time.Sleep(secondsToWait)
//	}
//
// ```
func (rl *RateLimit) TimeToWait() time.Duration {
	return time.Duration(int(rl.Limit)) / rl.TimeUnit
}

func (rl *RateLimit) TimeToReset() (time.Duration, time.Time) {
	resetTime := time.Unix(int64(rl.Reset), 0)
	return resetTime.Sub(time.Now()), resetTime
}
