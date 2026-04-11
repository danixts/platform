package timeutil

import (
	"sync/atomic"
	"time"
)

func Now() time.Time { return time.Now().UTC() }

func NowRFC3339() string { return Now().Format(time.RFC3339) }

func NowUnix() int64 { return Now().Unix() }

func Add(d time.Duration) time.Time { return Now().Add(d) }

func Unix(sec int64) time.Time { return time.Unix(sec, 0).UTC() }

func Duration(sec int) time.Duration { return time.Duration(sec) * time.Second }

func UTC(t time.Time) time.Time { return t.UTC() }

func ParseRFC3339(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

var defaultLoc atomic.Pointer[time.Location]

func init() {
	defaultLoc.Store(time.UTC)
}

func DefaultLocation() *time.Location {
	return defaultLoc.Load()
}

func SetDefaultLocation(loc *time.Location) {
	if loc == nil {
		loc = time.UTC
	}
	defaultLoc.Store(loc)
}

func LoadLocation(tz string) *time.Location {
	if tz == "" {
		return DefaultLocation()
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return DefaultLocation()
	}
	return loc
}
