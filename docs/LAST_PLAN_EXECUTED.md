---
PLAN: "feat!: MonthName returns English (breaking), add WeekdayName and ParseDateKey"
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# Plan — English calendar-unit names + `ParseDateKey`

## Context — corrected after review

An earlier version of this plan added `WeekdayName` returning Spanish names,
matching `MonthName`'s existing behavior. That was wrong, and it exposed
that `MonthName` itself was already wrong: **this library ships to
`webtyp/fmt/lang`-aware consumers, and the rule for that ecosystem is the
library never hardcodes a human language — it returns the English canonical
name, and the CONSUMER decides what its users see.** This is not a new
policy invented for this plan; it is already load-bearing doctrine one
level up the stack — see
`https://github.com/webtyp/layout/blob/main/docs/DICTIONARY.md`:
*"The dictionary itself is consumer-owned: the library never registers
words... Without any dictionary, everything renders in English."*
`layout/crudview` already renders every one of its own strings through
`lang.Translate("Confirm")`-shaped calls for exactly this reason.

`date` itself does **not** import `webtyp/fmt/lang` and gains no new
dependency — it just returns the English string, same as it does today,
just in English instead of Spanish. Translation is entirely the
consumer's concern, downstream (`components/calendarslider`'s plans handle
that side — see `docs/PLAN_STAGE_1_ARROW_HOVER_LAYOUT.md` and
`docs/PLAN_STAGE_3_MOBILE_COLLAPSE.md` in `components`).

**This stage is a breaking change** to `MonthName`'s return values — the
only other consumer in the whole `webtyp` org today is
`components/calendarslider` (verified: `grep -rln "date.MonthName"` across
every local repo checkout returns exactly that one file), and its own plans
already account for the change. If you are executing this against a
checkout where some OTHER consumer of `date.MonthName` exists that this
plan does not know about, stop and flag it — do not silently ship a
behavior change for an unaccounted-for caller.

## Stage 1 — `MonthName` returns English (breaking)

**File: `date.go`** — replace the 12 Spanish return values with their
English names. Everything else about the function (signature, the
out-of-range `""`, the switch-not-map shape) is unchanged:

```go
// MonthName returns month's English name ("January".."December"), or "" if
// month is out of range. English is the canonical, untranslated form — see
// the package doc comment for why. Switch, not a map — TinyGo.
func MonthName(month int) string {
	switch month {
	case 1:
		return "January"
	case 2:
		return "February"
	case 3:
		return "March"
	case 4:
		return "April"
	case 5:
		return "May"
	case 6:
		return "June"
	case 7:
		return "July"
	case 8:
		return "August"
	case 9:
		return "September"
	case 10:
		return "October"
	case 11:
		return "November"
	case 12:
		return "December"
	default:
		return ""
	}
}
```

**File: `date_test.go`** — `TestMonthName` currently asserts the Spanish
values ("Enero", "Agosto", "Diciembre"). Update each expected string to its
English equivalent ("January", "August", "December"); the structure and the
out-of-range case (`MonthName(13) == ""`) are unchanged.

## Stage 2 — `WeekdayName` (new, English from the start)

**File: `date.go`** — add directly below `MonthName`:

```go
// WeekdayName returns w's English name ("Sunday".."Saturday"), or "" if w
// is out of range. w follows Weekday's convention: 0 = Sunday … 6 =
// Saturday. English is the canonical, untranslated form, same reasoning as
// MonthName. Switch, not a map — TinyGo.
func WeekdayName(w int) string {
	switch w {
	case 0:
		return "Sunday"
	case 1:
		return "Monday"
	case 2:
		return "Tuesday"
	case 3:
		return "Wednesday"
	case 4:
		return "Thursday"
	case 5:
		return "Friday"
	case 6:
		return "Saturday"
	default:
		return ""
	}
}
```

**File: `date_test.go`** — add directly below `TestMonthName`:

```go
func TestWeekdayName(t *testing.T) {
	if WeekdayName(0) != "Sunday" {
		t.Error("WeekdayName(0) should be Sunday")
	}
	if WeekdayName(1) != "Monday" {
		t.Error("WeekdayName(1) should be Monday")
	}
	if WeekdayName(6) != "Saturday" {
		t.Error("WeekdayName(6) should be Saturday")
	}
	if WeekdayName(7) != "" {
		t.Error("WeekdayName(7) should be empty")
	}
}
```

## Stage 3 — `ParseDateKey` (new, no language content — unaffected by any of the above)

**File: `date.go`** — add directly below `ParseMonthKey`, same shape and
error handling (`fmt.Convert(...).Int()`, `(0, 0, 0)` on any failure):

```go
// ParseDateKey reads "YYYY-MM-DD"; returns (0, 0, 0) if s is not a valid
// date key — including a day that does not exist in that month (e.g.
// "2026-02-30").
func ParseDateKey(s string) (year, month, day int) {
	y, m := ParseMonthKey(s)
	if y == 0 || len(s) < 10 || s[7] != '-' {
		return 0, 0, 0
	}
	d, err := fmt.Convert(s[8:10]).Int()
	if err != nil || d < 1 || d > DaysInMonth(y, m) {
		return 0, 0, 0
	}
	return y, m, d
}
```

**File: `date_test.go`** — add directly below the new `TestWeekdayName`:

```go
func TestParseDateKey(t *testing.T) {
	y, m, d := ParseDateKey("2026-08-18")
	if y != 2026 || m != 8 || d != 18 {
		t.Errorf("ParseDateKey(2026-08-18) = %d, %d, %d, want 2026, 8, 18", y, m, d)
	}
	if y, m, d := ParseDateKey("2026-02-30"); y != 0 || m != 0 || d != 0 {
		t.Errorf("ParseDateKey(2026-02-30) = %d, %d, %d, want 0, 0, 0 (Feb has no 30th)", y, m, d)
	}
	if y, m, d := ParseDateKey("not-a-date"); y != 0 || m != 0 || d != 0 {
		t.Errorf("ParseDateKey(not-a-date) = %d, %d, %d, want 0, 0, 0", y, m, d)
	}
}
```

(Verified against a real `go test` run before writing this plan — the
`Feb 30` and garbage-input cases both correctly return `(0, 0, 0)`.)

## Stage 4 — package doc comment

**File: `date.go`** — the package doc comment (top of file) currently ends
with "...for that, use github.com/webtyp/time." Add one sentence after
it:

```go
// Every calendar-unit name this package returns (MonthName, WeekdayName) is
// English — the canonical, untranslated form. A consumer that wants any
// other language translates it via github.com/webtyp/fmt/lang; this
// package does not import that dependency or make that decision itself.
```

## Acceptance criteria

- `grep -n "Enero\|Lunes" date.go` → no matches (the Spanish literals are
  gone, not merely supplemented).
- `grep -n "func WeekdayName\|func ParseDateKey" date.go` → one match each.
- `go build ./...` and `gotest` both green, including the updated
  `TestMonthName`.
- `date/go.mod` is untouched — no new dependency.

## Stages

| Stage | File(s) | Done when |
|---|---|---|
| 1 | `date.go`, `date_test.go` | `MonthName` returns English; `TestMonthName` asserts English |
| 2 | `date.go`, `date_test.go` | `WeekdayName` compiles, mirrors `MonthName`'s (now English) shape |
| 3 | `date.go`, `date_test.go` | `ParseDateKey` compiles, mirrors `ParseMonthKey`'s shape |
| 4 | `date.go` | package doc comment explains the English-canonical / consumer-translates split |
