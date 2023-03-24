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
}

// FromResponse takes an http.Response struct and extracts
// the header values realated to rate limits (e.g. X-RateLite-Limit)
//
// ```
// rl := new(RateLimit)
// rl.FromResponse(response)
// ```
func (rl *RateLimit) FromResponse(resp *http.Response) {
	if rl == nil {
		rl = new(RateLimit)
	}
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
// rl.FromHeader(header)
// ```
func (rl *RateLimit) FromHeader(header http.Header) {
	if rl == nil {
		rl = new(RateLimit)
	}
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
	fmt.Fprintf(out, "limit %d, ", rl.Limit)
	fmt.Fprintf(out, "remaining %d, ", rl.Remaining)
	if rl.Reset > 0 {
		resetTime := time.Unix(int64(rl.Reset), 0)
		fmt.Fprintf(os.Stderr, "reset in %s at %s", resetTime.Sub(time.Now()).Truncate(time.Second), resetTime.Format("03:04PM"))
	}
	fmt.Fprintln(out, "")
}

func (rl *RateLimit) String() string {
	s := []string{}
	s[0] = fmt.Sprintf("limit %d", rl.Limit)
	s[1] = fmt.Sprintf("remaining %d", rl.Remaining)
	if rl.Reset > 0 {
		resetTime := time.Unix(int64(rl.Reset), 0)
		s[2] = fmt.Sprintf("reset in %s at %s", resetTime.Sub(time.Now()).Truncate(time.Second), resetTime.Format("03:04PM"))
	}
	return strings.Join(s, ", ")
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
	return time.Duration(int(float64(rl.Limit)/60.0))
}

func (rl *RateLimit) TimeToReset() (time.Duration, time.Time) {
	resetTime := time.Unix(int64(rl.Reset), 0)
	return resetTime.Sub(time.Now()), resetTime
}

func (rl *RateLimit) Throttle(i int, tot int) {
	// Caltech the rate, rounding up
	// log wait to Stderr
	var speedBump time.Duration
	// NOTE: 5000 per hour rate from some RDM API
	// 500 per minutes for others. We need to throttle accordingly
	// An hout == 3600 seconds, a minute is 60 seconds
	if rl.Limit == 5000 {
		// Restart with Rate Limit is 500 per minute
		speedBump = time.Duration(int(rl.Limit/60)) * time.Second
	} else {
		// Slow down to Rate Limit is 5000 per hour
		speedBump = time.Duration(int(rl.Limit/3600)) * time.Second
	}
	//fmt.Fprintf(os.Stderr, "DEBUG should throttle for %s\n", speedLimit.Truncate(time.Second))
	callsRemaining := 0.0
	if rl.Limit > 0 {
		callsRemaining = float64(rl.Remaining)/float64(rl.Limit)
	}
	if callsRemaining <= 0.1 {
		timeUntilReset, resetAt := rl.TimeToReset()
		// We're throttled for which ever is further in the future
		fmt.Fprintf(os.Stderr, "waiting %s for reset (%s) before continuing (%d/%d)\n", timeUntilReset.Truncate(time.Second), resetAt.Format("3:04PM"), i, tot)
		time.Sleep(timeUntilReset)
	} else if callsRemaining <= 0.5 {
		fmt.Fprintf(os.Stderr, "waiting %s before continuing (%d/%d)\n", speedBump.Truncate(time.Second), i, tot)
		time.Sleep(speedBump)
	} else {
		time.Sleep(200 * time.Millisecond)
	}
}
