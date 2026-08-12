// Package date is pure calendar arithmetic for Go and TinyGo: weekday of a
// date, days in a month, leap years, month names, and the "YYYY-MM"/
// "YYYY-MM-DD" key format used across the tinywasm ecosystem to identify a
// month or a day. No time zone, no wall clock, no JS interop — for that, use
// github.com/tinywasm/time. Every function here is a pure computation over
// (year, month, day) ints, so it runs identically in WASM and on the
// backend, and needs no build-tag split.
package date

import "github.com/tinywasm/fmt"

// IsLeapYear reports whether year is a leap year.
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in month (1-12) for year.
func DaysInMonth(year, month int) int {
	switch month {
	case 2:
		if IsLeapYear(year) {
			return 29
		}
		return 28
	case 4, 6, 9, 11:
		return 30
	default:
		return 31
	}
}

// Weekday returns the day of week for a date (0 = Sunday … 6 = Saturday)
// using the Sakamoto algorithm: pure arithmetic, no time zone, no parsing —
// deterministic in WASM and on the backend.
func Weekday(year, month, day int) int {
	t := [12]int{0, 3, 2, 5, 0, 3, 5, 1, 4, 6, 2, 4}
	y := year
	if month < 3 {
		y--
	}
	w := (y + y/4 - y/100 + y/400 + t[month-1] + day) % 7
	if w < 0 {
		w += 7
	}
	return w
}

// AddMonths adds delta months to (year, month), rolling the year over as
// needed — the only date arithmetic that depends on neither time zone nor
// the calendar (unlike day arithmetic, a month always has exactly 12 steps
// per year).
func AddMonths(year, month, delta int) (int, int) {
	t := year*12 + (month - 1) + delta
	y := t / 12
	m := t % 12
	if m < 0 {
		m += 12
		y--
	}
	return y, m + 1
}

// MonthName returns month's Spanish name ("Enero".."Diciembre"), or "" if
// month is out of range. Switch, not a map — TinyGo.
func MonthName(month int) string {
	switch month {
	case 1:
		return "Enero"
	case 2:
		return "Febrero"
	case 3:
		return "Marzo"
	case 4:
		return "Abril"
	case 5:
		return "Mayo"
	case 6:
		return "Junio"
	case 7:
		return "Julio"
	case 8:
		return "Agosto"
	case 9:
		return "Septiembre"
	case 10:
		return "Octubre"
	case 11:
		return "Noviembre"
	case 12:
		return "Diciembre"
	default:
		return ""
	}
}

// ParseMonthKey reads "YYYY-MM" (or the "YYYY-MM-DD" form, ignoring the
// day); returns (0, 0) if s is not a valid month key.
func ParseMonthKey(s string) (year, month int) {
	if len(s) < 7 || s[4] != '-' {
		return 0, 0
	}
	y, err := fmt.Convert(s[:4]).Int()
	if err != nil {
		return 0, 0
	}
	m, err := fmt.Convert(s[5:7]).Int()
	if err != nil || m < 1 || m > 12 {
		return 0, 0
	}
	return y, m
}

// MonthKey formats (year, month) as "YYYY-MM".
func MonthKey(year, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}

// DateKey formats (year, month, day) as "YYYY-MM-DD".
func DateKey(year, month, day int) string {
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}
