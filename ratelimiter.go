package irdmtools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// RateLimit holds the values used to play nice with OAI-PMH or REST API.
// It normally is extracted from the response header.
type RateLimit struct {
	// Limit maps to X-RateLimit-Limit
	Limit int `json:"limit,omitempty"`
	// OldLimit holds the last value of rate limit before change.
	OldLimit int `json:"-"`
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
	l := resp.Header.Values("X-RateLimit-Limit")
	if len(l) > 0 {
		if val, err := strconv.Atoi(l[0]); err == nil {
			if val != rl.OldLimit {
				rl.OldLimit = rl.Limit
			}
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
	l := header.Values("X-RateLimit-Limit")
	if len(l) > 0 {
		if val, err := strconv.Atoi(l[0]); err == nil {
			if val != rl.OldLimit {
				rl.OldLimit = rl.Limit
			}
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

func (rl *RateLimit) ResetString() string {
	var s string
	if rl.Reset > 0 {
		resetTime := time.Unix(int64(rl.Reset), 0)
		s = fmt.Sprintf("reset in %s at %s", resetTime.Sub(time.Now()).Truncate(time.Second), resetTime.Format("03:04PM"))
	}
	return s
}

func (rl *RateLimit) String() string {
	return fmt.Sprintf("limits %d/%d, %s", rl.Remaining, rl.Limit, rl.ResetString())
}

func (rl *RateLimit) Fprintf(out io.Writer) {
	fmt.Fprintln(out, rl.String())
}

// SecondsToWait returns the number of seconds (as a time.Duratin) to wait to avoid
// a http status code 429 and a ratio (float64) of remaining per request limit.
//
// ```
// rl := new(RateLimit)
// rl.FromHeader(response.Header)
// timeToWait := rl.TimeToWait()
// time.Sleep(timeToWait)
// ```
func (rl *RateLimit) TimeToWait(unit time.Duration) time.Duration {
	return time.Duration(int64(unit) / int64(rl.Limit))
}

func (rl *RateLimit) TimeToReset() (time.Duration, time.Time) {
	resetTime := time.Unix(int64(rl.Reset), 0)
	return resetTime.Sub(time.Now()), resetTime
}

// Throttle looks at the rate limit structure and implements
// an appropriate sleep time based on rate limits.
//
// ```
//
//	 i, tot := 0, 1000 // This ith' iteration and total number of records
//		rl := new(RateLimit)
//		// Set our rate limit from
//		rl.FromResponse(response)
//	 rl.Throttle(i, tot)
//
// ```
func (rl *RateLimit) Throttle(i int, tot int) {
	var speedBump time.Duration
	// NOTE: 5000 per hour rate from some RDM API
	// 500 per minutes for others. We need to throttle accordingly
	// An hout == 3600 seconds, a minute is 60 seconds.
	//
	// wait time = time unit / request limit
	//
	if tot == 1 || tot >= 5000 {
		// NOTE: Picking slower of the two rate limits, otherwise I stalling for an hour
		// at each 5000 records retrieved.
		speedBump = time.Duration(int64(time.Hour) / int64(5000))
	} else if rl.Limit == 5000 {
		// Slow down to Rate Limit is 5000 per hour
		speedBump = time.Duration(int64(time.Hour) / int64(rl.Limit))
	} else if rl.Limit > 0 {
		// Restart with Rate Limit is 500 per minute
		speedBump = time.Duration(int64(time.Minute) / int64(rl.Limit))
	} else {
		// Default rate limit of one per second
		speedBump = time.Second
	}
	if rl.OldLimit != rl.Limit {
		if rl.OldLimit > 0 {
			timeUntilReset, resetAt := rl.TimeToReset()
			// We're throttled for whichever is further in the future plus some padding
			timeUntilReset = timeUntilReset + (10 * time.Second)
			fmt.Fprintf(os.Stderr, "limits changed, waiting %s for reset (%s) before continuing (%d/%d)\n", timeUntilReset.Truncate(time.Second), resetAt.Format("3:04PM"), i, tot)
			time.Sleep(timeUntilReset)
		}
		// Update the old limit
		rl.OldLimit = rl.Limit
	} else {
		callsRemaining := 0.0
		if rl.Limit > 0 {
			callsRemaining = float64(rl.Remaining) / float64(rl.Limit)
		}
		if callsRemaining <= 0.25 {
			timeUntilReset, resetAt := rl.TimeToReset()
			// We're throttled for whichever is further in the future plus some padding
			timeUntilReset = timeUntilReset + (10 * time.Second)
			fmt.Fprintf(os.Stderr, "waiting %s for reset (%s) before continuing (%d/%d)\n", timeUntilReset.Truncate(time.Second), resetAt.Format("3:04PM"), i, tot)

			time.Sleep(timeUntilReset)
		} else {
			time.Sleep(speedBump)
		}
	}

}
