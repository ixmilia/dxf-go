package dxf

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func dublinOffset() (days float64) {
	return 2415020.0
}

func julianEpoch() time.Time {
	return time.Date(1899, 12, 31, 0, 0, 0, 0, time.UTC)
}

func boolFromShort(val int16) bool {
	return val != 0
}

func defaultIfEmpty(val, defaultValue string) string {
	if len(val) == 0 {
		return defaultValue
	}
	return val
}

func daysFromDuration(d time.Duration) (days float64) {
	return d.Hours() / 24.0
}

func julianDateFromTime(val time.Time) (julianDate float64) {
	sinceEpoch := val.Sub(julianEpoch())
	days := sinceEpoch.Hours() / 24.0
	return dublinOffset() + days
}

func durationFromDays(days float64) time.Duration {
	hours := days * 24.0
	minutes := hours * 60.0
	seconds := minutes * 60.0
	return time.Duration(int64(seconds)) * time.Second
}

func ensurePositiveOrDefault(val, defaultValue float64) float64 {
	if val < 0.0 {
		return defaultValue
	}
	return val
}

func handleFromString(val string) Handle {
	handle, err := strconv.ParseUint(val, 16, 64)
	if err != nil {
		return Handle(0)
	}
	return Handle(uint64(handle))
}

func shortFromBool(val bool) int16 {
	if val {
		return 1
	}
	return 0
}

func stringFromHandle(h Handle) string {
	return fmt.Sprintf("%X", uint64(h))
}

func timeFromJulianDays(juliandDays float64) time.Time {
	// manualy adjust for 1s difference to make the AutoDesk specified date match:
	//   2451544.91568287 = 31 December 1999, 9:58:35PM
	correction := 1
	asSeconds := (juliandDays-dublinOffset())*24*60*60 + float64(correction)
	offset := time.Duration(asSeconds) * time.Second
	return julianEpoch().Add(offset)
}

func uuidFromString(s string) uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.New()
	}

	return u
}
