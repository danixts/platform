// Package timeutil centralises the time conventions used across XMart
// Cloud services: storage in UTC, presentation in the user's timezone
// (default America/La_Paz).
package timeutil

import "time"

// Now returns the current time in UTC. Use this instead of time.Now to
// keep all storage timestamps UTC.
func Now() time.Time { return time.Now().UTC() }

// NowRFC3339 is a shortcut for formatting Now() as RFC3339.
func NowRFC3339() string { return Now().Format(time.RFC3339) }

// NowUnix returns the current Unix timestamp in seconds.
func NowUnix() int64 { return Now().Unix() }

// Add returns Now() + d in UTC.
func Add(d time.Duration) time.Time { return Now().Add(d) }

// Unix is time.Unix(sec, 0) forced to UTC.
func Unix(sec int64) time.Time { return time.Unix(sec, 0).UTC() }

// Duration converts an integer number of seconds into a time.Duration.
func Duration(sec int) time.Duration { return time.Duration(sec) * time.Second }

// UTC is a no-op wrapper kept for call-site symmetry with the rest of
// the package.
func UTC(t time.Time) time.Time { return t.UTC() }

// ParseRFC3339 parses an RFC3339 string and normalises the result to UTC.
func ParseRFC3339(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

// DefaultLocation is the fallback location used by LoadLocation when the
// requested timezone is empty or invalid. It is America/La_Paz, matching
// the XMart Cloud product default.
var DefaultLocation = mustLoad("America/La_Paz")

// LoadLocation is a safe wrapper around time.LoadLocation that falls back
// to DefaultLocation on any error. Use this in handlers that read a
// per-user timezone from the Identity header and must never fail open.
func LoadLocation(tz string) *time.Location {
	if tz == "" {
		return DefaultLocation
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return DefaultLocation
	}
	return loc
}

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
