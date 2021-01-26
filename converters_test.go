package dxf

import (
	"fmt"
	"testing"
	"time"
)

func TestConvertJulianDaysToTime(t *testing.T) {
	// from AutoDesk spec: 2451544.91568287 = 31 December 1999, 9:58:35PM
	expected := time.Date(1999, 12, 31, 9+12, 58, 35, 0, time.UTC)
	actual := timeFromJulianDays(2451544.91568287)
	assert(t, expected == actual, fmt.Sprintf("Expected: %v\nActual: %v", expected, actual))
}

func TestConvertTimeToJulianDays(t *testing.T) {
	// from AutoDesk spec: 2451544.91568287 = 31 December 1999, 9:58:35PM
	expected := 2451544.91568287
	actual := julianDateFromTime(time.Date(1999, 12, 31, 9+12, 58, 35, 0, time.UTC))
	delta := expected - actual
	assert(t, delta < 1e-10, fmt.Sprintf("Expected: %.12f\nActual: %.12f\nDelta: %.12f", expected, actual, delta))
}

func TestConvertDaysToDuration(t *testing.T) {
	seconds := 4*60*60 + 13*60 // 4h 13m
	expected := time.Duration(seconds) * time.Second
	actual := durationFromDays(float64(seconds) / 86400.0)
	assert(t, expected == actual, fmt.Sprintf("Expected: %v\nActual: %v", expected, actual))
}

func TestConvertDurationToDays(t *testing.T) {
	seconds := 4*60*60 + 13*60 // 4h 13m
	expected := float64(seconds) / 86400.0
	actual := daysFromDuration(time.Duration(seconds) * time.Second)
	delta := expected - actual
	assert(t, delta < 1e-10, fmt.Sprintf("Expected: %f\nActual: %f", expected, actual))
}

func TestConvertStringToHandle(t *testing.T) {
	assertEqUInt64(t, uint64(0x01), uint64(handleFromString("1")))
	assertEqUInt64(t, uint64(0xABCD), uint64(handleFromString("ABCD")))
	assertEqUInt64(t, uint64(0xABCDABCDABCDABCD), uint64(handleFromString("ABCDABCDABCDABCD")))
}

func TestConvertHandleToString(t *testing.T) {
	assertEqString(t, "1", stringFromHandle(Handle(0x01)))
	assertEqString(t, "ABCD", stringFromHandle(Handle(0xABCD)))
	assertEqString(t, "ABCDABCDABCDABCD", stringFromHandle(Handle(uint64(0xABCDABCDABCDABCD))))
}
