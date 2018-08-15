// copy from https://github.com/ChannelMeter/iso8601duration/blob/master/duration_test.go
package wsman

import (
	"reflect"
	"testing"
	"time"
)

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()

	switch expectedValue := expected.(type) {
	case error:
		if expected != actual {
			t.Error("except", expected, "got", actual)
		}
	case int:
		if expectedValue != actual.(int) {
			t.Error("except", expected, "got", actual)
		}
	default:
		if !reflect.DeepEqual(expected, actual) {
			t.Error("except", expected, "got", actual)
		}
	}
}

func TestDurationFromString(t *testing.T) {
	// test with bad format
	_, err := DurationFromString("asdf")
	assertEqual(t, err, ErrBadFormat)

	// test with month
	_, err = DurationFromString("P1M")
	assertEqual(t, err, ErrNoMonth)

	// test with good full string
	dur, err := DurationFromString("P1Y2DT3H4M5S")
	if err != nil {
		t.Error(err)
		return
	}
	assertEqual(t, 1, dur.Years)
	assertEqual(t, 2, dur.Days)
	assertEqual(t, 3, dur.Hours)
	assertEqual(t, 4, dur.Minutes)
	assertEqual(t, 5, dur.Seconds)

	// test with good week string
	dur, err = DurationFromString("P1W")
	if err != nil {
		t.Error(err)
		return
	}
	assertEqual(t, 1, dur.Weeks)
}

func TestDurationToString(t *testing.T) {
	// test empty
	d := Duration{}
	assertEqual(t, d.String(), "P")

	// test only larger-than-day
	d = Duration{Years: 1, Days: 2}
	assertEqual(t, d.String(), "P1Y2D")

	// test only smaller-than-day
	d = Duration{Hours: 1, Minutes: 2, Seconds: 3}
	assertEqual(t, d.String(), "PT1H2M3S")

	// test full format
	d = Duration{Years: 1, Days: 2, Hours: 3, Minutes: 4, Seconds: 5}
	assertEqual(t, d.String(), "P1Y2DT3H4M5S")

	// test week format
	d = Duration{Weeks: 1}
	assertEqual(t, d.String(), "P1W")
}

func TestToDuration(t *testing.T) {
	d := Duration{Years: 1}
	assertEqual(t, d.ToDuration(), time.Hour*24*365)

	d = Duration{Weeks: 1}
	assertEqual(t, d.ToDuration(), time.Hour*24*7)

	d = Duration{Days: 1}
	assertEqual(t, d.ToDuration(), time.Hour*24)

	d = Duration{Hours: 1}
	assertEqual(t, d.ToDuration(), time.Hour)

	d = Duration{Minutes: 1}
	assertEqual(t, d.ToDuration(), time.Minute)

	d = Duration{Seconds: 1}
	assertEqual(t, d.ToDuration(), time.Second)
}
