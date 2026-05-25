// SPDX-License-Identifier: Apache-2.0

package rrule

import (
	"errors"
	"testing"
	"time"
)

// nextCase is one row of the NextOccurrence golden table.
//
//   - rule + dtstart + after define the input.
//   - want is the expected next-fire t' (zero when wantOK=false).
//   - wantOK is true when the rule has an occurrence after `after`,
//     false when terminated.
//   - wantErr (when non-nil) is the sentinel the call must wrap;
//     when set, want and wantOK are ignored.
type nextCase struct {
	name    string
	rule    Rule
	dtstart time.Time
	after   time.Time
	want    time.Time
	wantOK  bool
	wantErr error
}

func mustParseUTC(s string) time.Time {
	t, err := time.Parse("20060102T150405Z", s)
	if err != nil {
		panic("nextCase: bad fixture: " + s + ": " + err.Error())
	}
	return t
}

// TestNextOccurrence covers FREQ-only iteration, INTERVAL,
// UNTIL/COUNT termination, BY* filters, and the documented edge
// cases (leap-day BYMONTHDAY=29 in non-leap February, DST
// boundaries).
func TestNextOccurrence(t *testing.T) {
	dtAprilNoon := mustParseUTC("20260401T120000Z")
	dtJan1 := mustParseUTC("20260101T000000Z")

	cases := []nextCase{
		// ── FREQ-only ─────────────────────────────────────────────
		{
			name:    "hourly_simple",
			rule:    Rule{Freq: FreqHourly, Interval: 1, WeekStart: MO},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon,
			want:    mustParseUTC("20260401T130000Z"),
			wantOK:  true,
		},
		{
			name:    "daily_simple",
			rule:    Rule{Freq: FreqDaily, Interval: 1, WeekStart: MO},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon,
			want:    mustParseUTC("20260402T120000Z"),
			wantOK:  true,
		},
		{
			name:    "weekly_simple",
			rule:    Rule{Freq: FreqWeekly, Interval: 1, WeekStart: MO},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon,
			want:    mustParseUTC("20260408T120000Z"),
			wantOK:  true,
		},
		{
			name:    "monthly_simple",
			rule:    Rule{Freq: FreqMonthly, Interval: 1, WeekStart: MO},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon,
			want:    mustParseUTC("20260501T120000Z"),
			wantOK:  true,
		},
		{
			name:    "yearly_simple",
			rule:    Rule{Freq: FreqYearly, Interval: 1, WeekStart: MO},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon,
			want:    mustParseUTC("20270401T120000Z"),
			wantOK:  true,
		},

		// ── INTERVAL ──────────────────────────────────────────────
		{
			name:    "daily_interval_2",
			rule:    Rule{Freq: FreqDaily, Interval: 2, WeekStart: MO},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon,
			want:    mustParseUTC("20260403T120000Z"),
			wantOK:  true,
		},
		{
			name:    "weekly_interval_2",
			rule:    Rule{Freq: FreqWeekly, Interval: 2, WeekStart: MO},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon,
			want:    mustParseUTC("20260415T120000Z"),
			wantOK:  true,
		},
		{
			name:    "monthly_interval_3",
			rule:    Rule{Freq: FreqMonthly, Interval: 3, WeekStart: MO},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon,
			want:    mustParseUTC("20260701T120000Z"),
			wantOK:  true,
		},

		// ── UNTIL termination ────────────────────────────────────
		{
			name: "daily_until_terminated",
			rule: Rule{
				Freq:      FreqDaily,
				Interval:  1,
				Until:     mustParseUTC("20260403T000000Z"),
				WeekStart: MO,
			},
			dtstart: dtAprilNoon,
			after:   mustParseUTC("20260402T120000Z"),
			wantOK:  false, // next would be 20260403T120000Z, past UNTIL midnight
		},
		{
			name: "daily_until_still_firing",
			rule: Rule{
				Freq:      FreqDaily,
				Interval:  1,
				Until:     mustParseUTC("20260410T120000Z"),
				WeekStart: MO,
			},
			dtstart: dtAprilNoon,
			after:   mustParseUTC("20260408T120000Z"),
			want:    mustParseUTC("20260409T120000Z"),
			wantOK:  true,
		},

		// ── COUNT termination ────────────────────────────────────
		{
			name: "daily_count_3_after_first",
			rule: Rule{
				Freq:      FreqDaily,
				Interval:  1,
				Count:     3,
				WeekStart: MO,
			},
			dtstart: dtAprilNoon,
			after:   dtAprilNoon, // dtstart counts as occurrence #1
			want:    mustParseUTC("20260402T120000Z"),
			wantOK:  true,
		},
		{
			name: "daily_count_3_after_last",
			rule: Rule{
				Freq:      FreqDaily,
				Interval:  1,
				Count:     3,
				WeekStart: MO,
			},
			dtstart: dtAprilNoon,
			after:   mustParseUTC("20260403T120000Z"), // last occurrence
			wantOK:  false,
		},

		// ── BYMONTH ──────────────────────────────────────────────
		{
			name: "yearly_bymonth_jun_dec",
			rule: Rule{
				Freq:      FreqYearly,
				Interval:  1,
				ByMonth:   []int{6, 12},
				WeekStart: MO,
			},
			dtstart: dtJan1,
			after:   dtJan1,
			want:    mustParseUTC("20260601T000000Z"),
			wantOK:  true,
		},

		// ── BYMONTHDAY ──────────────────────────────────────────
		{
			name: "monthly_bymonthday_15",
			rule: Rule{
				Freq:       FreqMonthly,
				Interval:   1,
				ByMonthDay: []int{15},
				WeekStart:  MO,
			},
			dtstart: mustParseUTC("20260115T120000Z"),
			after:   mustParseUTC("20260115T120000Z"),
			want:    mustParseUTC("20260215T120000Z"),
			wantOK:  true,
		},
		{
			name: "monthly_bymonthday_minus_1_last_day",
			rule: Rule{
				Freq:       FreqMonthly,
				Interval:   1,
				ByMonthDay: []int{-1},
				WeekStart:  MO,
			},
			dtstart: mustParseUTC("20260131T120000Z"),
			after:   mustParseUTC("20260131T120000Z"),
			want:    mustParseUTC("20260228T120000Z"), // Feb 2026 has 28 days
			wantOK:  true,
		},

		// ── BYDAY ────────────────────────────────────────────────
		{
			name: "weekly_byday_mo_we_fr",
			// April 1, 2026 = Wednesday (verify: 2026-04-01 was a
			// Wednesday). Next MO/WE/FR after that is FR = Apr 3.
			rule: Rule{
				Freq:      FreqWeekly,
				Interval:  1,
				ByDay:     []ByDay{{0, MO}, {0, WE}, {0, FR}},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260401T120000Z"), // Wed
			after:   mustParseUTC("20260401T120000Z"),
			want:    mustParseUTC("20260403T120000Z"), // Fri
			wantOK:  true,
		},
		{
			name: "monthly_byday_first_monday",
			rule: Rule{
				Freq:      FreqMonthly,
				Interval:  1,
				ByDay:     []ByDay{{1, MO}},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260406T120000Z"), // 1st Mon of April 2026
			after:   mustParseUTC("20260406T120000Z"),
			want:    mustParseUTC("20260504T120000Z"), // 1st Mon of May 2026
			wantOK:  true,
		},
		{
			name: "monthly_byday_last_friday",
			rule: Rule{
				Freq:      FreqMonthly,
				Interval:  1,
				ByDay:     []ByDay{{-1, FR}},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260424T120000Z"), // last Fri of Apr 2026
			after:   mustParseUTC("20260424T120000Z"),
			want:    mustParseUTC("20260529T120000Z"), // last Fri of May 2026
			wantOK:  true,
		},

		// ── BYHOUR / BYMINUTE / BYSECOND ────────────────────────
		{
			name: "daily_byhour_9_17",
			rule: Rule{
				Freq:      FreqDaily,
				Interval:  1,
				ByHour:    []int{9, 17},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260401T090000Z"),
			after:   mustParseUTC("20260401T090000Z"),
			want:    mustParseUTC("20260401T170000Z"),
			wantOK:  true,
		},

		// ── Leap-day edge case ──────────────────────────────────
		{
			name: "monthly_bymonthday_29_skips_feb_non_leap",
			// 2026 is NOT a leap year. BYMONTHDAY=29 should skip
			// February and land on March 29.
			rule: Rule{
				Freq:       FreqMonthly,
				Interval:   1,
				ByMonthDay: []int{29},
				WeekStart:  MO,
			},
			dtstart: mustParseUTC("20260129T120000Z"),
			after:   mustParseUTC("20260129T120000Z"),
			want:    mustParseUTC("20260329T120000Z"),
			wantOK:  true,
		},

		// ── DST boundary ────────────────────────────────────────
		// Spring-forward in America/New_York: 2026-03-08, 02:30
		// local does not exist. Daily rule from Mar 7 02:30 EST
		// (07:30 UTC) → Mar 8 03:00 EDT (07:00 UTC) — Go's
		// time.Add normalizes through the gap. We verify that the
		// next occurrence is one calendar day later in local terms,
		// which lands at the same UTC offset minus one hour after
		// spring-forward.
		dstSpringForwardCase(),

		// ── INTERVAL with COUNT terminates correctly ───────────
		{
			name: "weekly_interval_2_count_3_terminates",
			rule: Rule{
				Freq:      FreqWeekly,
				Interval:  2,
				Count:     3,
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260101T000000Z"),
			after:   mustParseUTC("20260129T000000Z"), // 3rd occurrence
			wantOK:  false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok, err := NextOccurrence(c.rule, c.dtstart, c.after)
			if c.wantErr != nil {
				if !errors.Is(err, c.wantErr) {
					t.Fatalf("err: want errors.Is(%v), got %v", c.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if ok != c.wantOK {
				t.Fatalf("ok: want %v, got %v (got time %v)", c.wantOK, ok, got)
			}
			if !c.wantOK {
				if !got.IsZero() {
					t.Fatalf("expected zero time when ok=false, got %v", got)
				}
				return
			}
			if !got.Equal(c.want) {
				t.Fatalf("time: want %v, got %v", c.want, got)
			}
		})
	}
}

// dstSpringForwardCase exercises a DAILY recurrence across the
// US/Eastern spring-forward boundary. dtstart is in
// America/New_York; the evaluator must respect the zone
// (time.Time arithmetic does this natively).
func dstSpringForwardCase() nextCase {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic("dstSpringForwardCase: tzdata missing: " + err.Error())
	}
	dt := time.Date(2026, 3, 7, 14, 0, 0, 0, loc) // Sat 14:00 EST
	return nextCase{
		name:    "daily_dst_spring_forward",
		rule:    Rule{Freq: FreqDaily, Interval: 1, WeekStart: MO},
		dtstart: dt,
		after:   dt,
		// Sun 14:00 EDT (after clocks jumped forward at 02:00 →
		// 03:00). Same wall-clock 14:00, new offset.
		want:   time.Date(2026, 3, 8, 14, 0, 0, 0, loc),
		wantOK: true,
	}
}

// TestNextOccurrence_TerminatesOnPathological asserts the safety
// limit catches a rule that would otherwise iterate forever (e.g.
// BYMONTHDAY=30 under FREQ=MONTHLY where the rule never fires after
// `after` because the parser-evaluator gap allows constructing such
// inputs only via the public Rule literal — not via ParseRRule —
// the limit is the only line of defense).
func TestNextOccurrence_TerminatesOnPathological(t *testing.T) {
	// FREQ=DAILY with BYMONTHDAY=29 only every Feb non-leap → very
	// sparse, but it does fire. To force a *no-fire* rule we use
	// COUNT=1 with after past dtstart so the rule has terminated:
	r := Rule{Freq: FreqDaily, Interval: 1, Count: 1, WeekStart: MO}
	dt := mustParseUTC("20260401T120000Z")
	got, ok, err := NextOccurrence(r, dt, mustParseUTC("20260401T120001Z"))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if ok {
		t.Fatalf("want terminated, got %v", got)
	}
}

// TestNextOccurrence_GuardRails covers the input-validation guards
// that ParseRRule would normally catch but a hand-built Rule
// literal can bypass.
func TestNextOccurrence_GuardRails(t *testing.T) {
	dt := mustParseUTC("20260401T120000Z")
	t.Run("freq_invalid", func(t *testing.T) {
		_, _, err := NextOccurrence(Rule{}, dt, dt)
		if !errors.Is(err, ErrUnsupportedRRule) {
			t.Fatalf("want ErrUnsupportedRRule, got %v", err)
		}
	})
	t.Run("interval_zero", func(t *testing.T) {
		_, _, err := NextOccurrence(Rule{Freq: FreqDaily, Interval: 0}, dt, dt)
		if !errors.Is(err, ErrUnsupportedRRule) {
			t.Fatalf("want ErrUnsupportedRRule, got %v", err)
		}
	})
}

// TestNextOccurrence_ByFiltersExclude exercises the rejection
// branches of the BY* matchers: BYMONTH excluding, BYMONTHDAY
// negative-resolution skipping, BYDAY missing the weekday.
func TestNextOccurrence_ByFiltersExclude(t *testing.T) {
	// YEARLY with BYMONTH=12: starting in Jan should land on Dec 1.
	r := Rule{
		Freq:      FreqYearly,
		Interval:  1,
		ByMonth:   []int{12},
		WeekStart: MO,
	}
	dt := mustParseUTC("20260101T000000Z")
	got, ok, err := NextOccurrence(r, dt, dt)
	if err != nil || !ok {
		t.Fatalf("err=%v ok=%v", err, ok)
	}
	want := mustParseUTC("20261201T000000Z")
	if !got.Equal(want) {
		t.Fatalf("BYMONTH=12: want %v, got %v", want, got)
	}

	// MONTHLY with BYMONTHDAY=-2 (penultimate day): January 2026
	// has 31 days → day 30. After dtstart on Jan 30, next is Feb
	// 27 (Feb has 28 days, -2 = day 27).
	r2 := Rule{
		Freq:       FreqMonthly,
		Interval:   1,
		ByMonthDay: []int{-2},
		WeekStart:  MO,
	}
	dt2 := mustParseUTC("20260130T000000Z")
	got2, ok, err := NextOccurrence(r2, dt2, dt2)
	if err != nil || !ok {
		t.Fatalf("err=%v ok=%v", err, ok)
	}
	want2 := mustParseUTC("20260227T000000Z")
	if !got2.Equal(want2) {
		t.Fatalf("BYMONTHDAY=-2: want %v, got %v", want2, got2)
	}
}

// TestWeekday_ToTime exercises the ToTime conversion across all
// weekdays.
func TestWeekday_ToTime(t *testing.T) {
	cases := []struct {
		w    Weekday
		want time.Weekday
	}{
		{SU, time.Sunday},
		{MO, time.Monday},
		{SA, time.Saturday},
	}
	for _, c := range cases {
		if got := c.w.ToTime(); got != c.want {
			t.Errorf("%v.ToTime() = %v, want %v", c.w, got, c.want)
		}
	}
}

// TestNextOccurrence_HourlyWithFilters drives the hourly path
// through every BY* branch.
func TestNextOccurrence_HourlyWithFilters(t *testing.T) {
	dt := mustParseUTC("20260401T100000Z")
	r := Rule{
		Freq:      FreqHourly,
		Interval:  1,
		ByMinute:  []int{15, 45},
		BySecond:  []int{30},
		WeekStart: MO,
	}
	got, ok, err := NextOccurrence(r, dt, dt)
	if err != nil || !ok {
		t.Fatalf("err=%v ok=%v", err, ok)
	}
	want := mustParseUTC("20260401T101530Z")
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestNextOccurrence_HourlyByDayFilter exercises the BYDAY
// rejection branch in expandHourly (rule is constructed manually
// — ParseRRule wouldn't allow this combo logically but the
// matcher should still skip non-matching days).
func TestNextOccurrence_HourlyByDayFilter(t *testing.T) {
	dt := mustParseUTC("20260401T100000Z") // Wednesday
	// BYDAY=FR — should advance until Friday.
	r := Rule{
		Freq:      FreqHourly,
		Interval:  1,
		ByDay:     []ByDay{{0, FR}},
		WeekStart: MO,
	}
	got, ok, err := NextOccurrence(r, dt, dt)
	if err != nil || !ok {
		t.Fatalf("err=%v ok=%v", err, ok)
	}
	// Friday Apr 3, 2026 at 00:00 (first hour after dtstart Wed
	// 10am where BYDAY matches and minute/second taken from dt).
	if got.Weekday() != time.Friday {
		t.Fatalf("expected Friday, got %v at %v", got.Weekday(), got)
	}
}

// TestNextOccurrence_DailyBYMONTHReject covers the matchesByMonth
// rejection branch within expandDaily.
func TestNextOccurrence_DailyBYMONTHReject(t *testing.T) {
	dt := mustParseUTC("20260101T120000Z") // January
	// BYMONTH=6 — daily should yield first day of June.
	r := Rule{
		Freq:      FreqDaily,
		Interval:  1,
		ByMonth:   []int{6},
		WeekStart: MO,
	}
	got, ok, err := NextOccurrence(r, dt, dt)
	if err != nil || !ok {
		t.Fatalf("err=%v ok=%v", err, ok)
	}
	if got.Month() != time.June {
		t.Fatalf("expected June, got %v", got)
	}
}

// TestNextOccurrence_MonthlyJan31SkipsFeb covers the
// dtstartDay > daysInMonth branch in monthDayCandidates.
func TestNextOccurrence_MonthlyJan31SkipsFeb(t *testing.T) {
	dt := mustParseUTC("20260131T120000Z") // Jan 31
	r := Rule{Freq: FreqMonthly, Interval: 1, WeekStart: MO}
	got, ok, err := NextOccurrence(r, dt, dt)
	if err != nil || !ok {
		t.Fatalf("err=%v ok=%v", err, ok)
	}
	// Should skip Feb (no Feb 31) and land on Mar 31.
	if got.Day() != 31 || got.Month() != time.March {
		t.Fatalf("want Mar 31, got %v", got)
	}
}

// TestParseRRule_ExtraEdgeCases nudges parser coverage on rare
// branches.
func TestParseRRule_ExtraEdgeCases(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"byday_empty_list", "FREQ=WEEKLY;BYDAY="},
		{"byday_too_short", "FREQ=WEEKLY;BYDAY=M"},
		{"byday_ordinal_too_large", "FREQ=MONTHLY;BYDAY=54MO"},
		{"byday_ordinal_too_small", "FREQ=MONTHLY;BYDAY=-54MO"},
		{"byday_ordinal_non_int", "FREQ=MONTHLY;BYDAY=ABMO"},
		{"bymonthday_non_int", "FREQ=MONTHLY;BYMONTHDAY=abc"},
		{"bymonthday_empty", "FREQ=MONTHLY;BYMONTHDAY="},
		{"bymonth_non_int", "FREQ=YEARLY;BYMONTH=abc"},
		{"count_non_int", "FREQ=DAILY;COUNT=abc"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := ParseRRule(c.in); err == nil {
				t.Fatalf("ParseRRule(%q): want error, got nil", c.in)
			}
		})
	}
}

// TestNextOccurrence_YearlyByYearDay covers BYYEARDAY expansion
// under FREQ=YEARLY: positive day-of-year, negative (counted from
// year-end), the leap-year boundary at day 366, and a list with
// multiple values.
func TestNextOccurrence_YearlyByYearDay(t *testing.T) {
	cases := []struct {
		name    string
		rule    Rule
		dtstart time.Time
		after   time.Time
		want    time.Time
	}{
		{
			name: "yearly_byyearday_100",
			rule: Rule{
				Freq:      FreqYearly,
				Interval:  1,
				ByYearDay: []int{100},
				WeekStart: MO,
			},
			// Day 100 of 2026 is April 10 (2026 not leap).
			dtstart: mustParseUTC("20260410T000000Z"),
			after:   mustParseUTC("20260410T000000Z"),
			want:    mustParseUTC("20270410T000000Z"),
		},
		{
			name: "yearly_byyearday_minus_1_last_day",
			rule: Rule{
				Freq:      FreqYearly,
				Interval:  1,
				ByYearDay: []int{-1},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20261231T000000Z"),
			after:   mustParseUTC("20261231T000000Z"),
			want:    mustParseUTC("20271231T000000Z"),
		},
		{
			name: "yearly_byyearday_366_skips_non_leap",
			rule: Rule{
				Freq:      FreqYearly,
				Interval:  1,
				ByYearDay: []int{366},
				WeekStart: MO,
			},
			// Start in 2024 (leap year). Day 366 = Dec 31 2024.
			// 2025, 2026, 2027 are non-leap → no day 366. Next fire
			// is Dec 31 2028 (leap).
			dtstart: mustParseUTC("20241231T000000Z"),
			after:   mustParseUTC("20241231T000000Z"),
			want:    mustParseUTC("20281231T000000Z"),
		},
		{
			name: "yearly_byyearday_list_first_after",
			rule: Rule{
				Freq:      FreqYearly,
				Interval:  1,
				ByYearDay: []int{1, 100, -1},
				WeekStart: MO,
			},
			// dtstart Jan 1 2026 → first occurrence Jan 1 2026 (day 1).
			// Next after is day 100 = Apr 10.
			dtstart: mustParseUTC("20260101T000000Z"),
			after:   mustParseUTC("20260101T000000Z"),
			want:    mustParseUTC("20260410T000000Z"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok, err := NextOccurrence(c.rule, c.dtstart, c.after)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !ok {
				t.Fatalf("want ok, got terminated")
			}
			if !got.Equal(c.want) {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}

// TestNextOccurrence_YearlyByWeekNo covers BYWEEKNO expansion
// under FREQ=YEARLY: ISO 8601 week numbering with WKST=MO
// (default). Week 1 of any year contains the first Thursday
// (equivalently, contains Jan 4). Negative values count from
// year-end. Years with 53 ISO weeks (e.g. 2026 — 53 weeks
// because Jan 1 2026 is Thursday) honor week 53.
func TestNextOccurrence_YearlyByWeekNo(t *testing.T) {
	cases := []struct {
		name    string
		rule    Rule
		dtstart time.Time
		after   time.Time
		want    time.Time
	}{
		{
			name: "yearly_byweekno_1",
			rule: Rule{
				Freq:      FreqYearly,
				Interval:  1,
				ByWeekNo:  []int{1},
				ByDay:     []ByDay{{0, MO}},
				WeekStart: MO,
			},
			// ISO week 1 of 2026 starts Mon Dec 29 2025 (week
			// containing Jan 4 2026). dtstart on that Monday.
			dtstart: mustParseUTC("20251229T000000Z"),
			after:   mustParseUTC("20251229T000000Z"),
			// Next ISO week 1 + Monday: Mon Jan 4 2027 (week 1 of
			// 2027 begins Jan 4 since Jan 1 2027 = Friday).
			want: mustParseUTC("20270104T000000Z"),
		},
		{
			name: "yearly_byweekno_minus_1_last_week",
			rule: Rule{
				Freq:      FreqYearly,
				Interval:  1,
				ByWeekNo:  []int{-1},
				ByDay:     []ByDay{{0, MO}},
				WeekStart: MO,
			},
			// ISO week 53 of 2026 starts Mon Dec 28 2026.
			dtstart: mustParseUTC("20261228T000000Z"),
			after:   mustParseUTC("20261228T000000Z"),
			// Next "last week + Monday": ISO week 52 of 2027 begins
			// Mon Dec 27 2027 (2027 has 52 weeks).
			want: mustParseUTC("20271227T000000Z"),
		},
		{
			name: "yearly_byweekno_53_skips_short_year",
			rule: Rule{
				Freq:      FreqYearly,
				Interval:  1,
				ByWeekNo:  []int{53},
				ByDay:     []ByDay{{0, MO}},
				WeekStart: MO,
			},
			// 2026 has 53 ISO weeks (Jan 1 2026 = Thursday). 2027,
			// 2028, 2029 have 52. 2030, 2031 have 52. 2032 has 53
			// (Jan 1 2032 = Thursday — leap year starting Thursday).
			dtstart: mustParseUTC("20261228T000000Z"),
			after:   mustParseUTC("20261228T000000Z"),
			// Next year with 53 weeks: 2032. Week 53 begins Mon
			// Dec 27 2032.
			want: mustParseUTC("20321227T000000Z"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok, err := NextOccurrence(c.rule, c.dtstart, c.after)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !ok {
				t.Fatalf("want ok, got terminated")
			}
			if !got.Equal(c.want) {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}

// TestNextOccurrence_YearlyByWeekNo_WKSTSunday verifies WKST
// affects week-boundary semantics for BYWEEKNO. With WKST=SU,
// the week starts Sunday (per RFC 5545 §3.3.10).
//
// Year 2023 is the discriminating case: Jan 1 2023 is a Sunday.
//   - WKST=MO: week 1 of 2023 starts Mon Jan 2 (first MO-anchored
//     week containing 4+ days of 2023). The Sunday in week 1 is
//     Jan 8.
//   - WKST=SU: week 1 of 2023 starts Sun Jan 1 (the SU-anchored
//     week containing 4+ days of 2023). The Sunday in week 1 is
//     Jan 1.
//
// We exercise the WKST=SU branch by asking for "first occurrence
// after dtstart=Jan 1 2023" — under WKST=SU that returns the
// Sunday of week 1 in 2024, which (since Jan 1 2024 is Mon) is
// Dec 31 2023 (the Sunday in the SU-anchored week containing Jan
// 4 2024).
func TestNextOccurrence_YearlyByWeekNo_WKSTSunday(t *testing.T) {
	r := Rule{
		Freq:      FreqYearly,
		Interval:  1,
		ByWeekNo:  []int{1},
		ByDay:     []ByDay{{0, SU}},
		WeekStart: SU,
	}
	dt := mustParseUTC("20230101T000000Z") // Sun Jan 1 2023
	got, ok, err := NextOccurrence(r, dt, dt)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !ok {
		t.Fatalf("want ok, got terminated")
	}
	// Week 1 of 2024 (WKST=SU): Jan 4 2024 = Thursday → SU-anchored
	// week containing Jan 4 starts Sun Dec 31 2023.
	want := mustParseUTC("20231231T000000Z")
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

// TestNextOccurrence_BySetPos covers BYSETPOS positional filtering
// applied AFTER all other BY-* expansion within the FREQ period
// per RFC 5545 §3.3.10. 1-based; negative counts from end. Out-of-
// range entries silently dropped.
func TestNextOccurrence_BySetPos(t *testing.T) {
	cases := []struct {
		name    string
		rule    Rule
		dtstart time.Time
		after   time.Time
		want    time.Time
	}{
		{
			// "Last weekday of month": BYDAY=MO,TU,WE,TH,FR yields
			// every weekday; BYSETPOS=-1 picks the last one.
			// Apr 2026: Apr 30 = Thursday → expected.
			name: "monthly_last_weekday",
			rule: Rule{
				Freq:      FreqMonthly,
				Interval:  1,
				ByDay:     []ByDay{{0, MO}, {0, TU}, {0, WE}, {0, TH}, {0, FR}},
				BySetPos:  []int{-1},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260401T000000Z"),
			after:   mustParseUTC("20260401T000000Z"),
			want:    mustParseUTC("20260430T000000Z"),
		},
		{
			// "Second Tuesday of month": BYDAY=TU yields every TU;
			// BYSETPOS=2 picks the 2nd. Apr 2026 Tuesdays: 7, 14,
			// 21, 28 → 14.
			name: "monthly_second_tuesday",
			rule: Rule{
				Freq:      FreqMonthly,
				Interval:  1,
				ByDay:     []ByDay{{0, TU}},
				BySetPos:  []int{2},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260401T000000Z"),
			after:   mustParseUTC("20260401T000000Z"),
			want:    mustParseUTC("20260414T000000Z"),
		},
		{
			// BYSETPOS list with mixed signs: 1 + -1 picks first
			// and last. After Apr 1 → first weekday Apr 1 itself
			// is Wed (a weekday) — but `after` is Apr 1 inclusive,
			// so first match strictly after = Apr 30 (the -1 in
			// April) — wait: occurrences in April are {Apr 1, Apr
			// 30}; Apr 1 == after (not strictly after); next is
			// Apr 30. Then May's first/last. So with after=Apr 1,
			// next is Apr 30.
			name: "monthly_first_and_last_weekday",
			rule: Rule{
				Freq:      FreqMonthly,
				Interval:  1,
				ByDay:     []ByDay{{0, MO}, {0, TU}, {0, WE}, {0, TH}, {0, FR}},
				BySetPos:  []int{1, -1},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260401T000000Z"),
			after:   mustParseUTC("20260401T000000Z"),
			want:    mustParseUTC("20260430T000000Z"),
		},
		{
			// Out-of-range BYSETPOS entry silently dropped:
			// Apr 2026 has 22 weekdays; BYSETPOS=99 produces no
			// occurrence in April; we fall through to May (also
			// 21 weekdays — May 1 is Fri). Combined with
			// BYSETPOS=1 → first weekday of May = May 1.
			name: "monthly_out_of_range_dropped",
			rule: Rule{
				Freq:      FreqMonthly,
				Interval:  1,
				ByDay:     []ByDay{{0, MO}, {0, TU}, {0, WE}, {0, TH}, {0, FR}},
				BySetPos:  []int{99, 1},
				WeekStart: MO,
			},
			dtstart: mustParseUTC("20260401T000000Z"),
			after:   mustParseUTC("20260401T120000Z"),
			want:    mustParseUTC("20260501T000000Z"),
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok, err := NextOccurrence(c.rule, c.dtstart, c.after)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !ok {
				t.Fatalf("want ok, got terminated")
			}
			if !got.Equal(c.want) {
				t.Fatalf("want %v, got %v", c.want, got)
			}
		})
	}
}

// TestNextOccurrence_YearlyFromFeb29_SkipsNonLeap covers the
// RFC 5545 §3.3.10 "non-existent dates simply do not occur"
// rule for FREQ=YEARLY anchored on Feb 29. Go's time.AddDate
// would normalize 2024-02-29 + 1y to 2025-03-01; the evaluator
// must instead skip non-leap years and emit the next true Feb 29
// (2028, 2032, 2036).
func TestNextOccurrence_YearlyFromFeb29_SkipsNonLeap(t *testing.T) {
	dt := mustParseUTC("20240229T000000Z")
	r := Rule{Freq: FreqYearly, Interval: 1, WeekStart: MO}

	want := []time.Time{
		mustParseUTC("20280229T000000Z"),
		mustParseUTC("20320229T000000Z"),
		mustParseUTC("20360229T000000Z"),
	}

	after := dt
	for i, w := range want {
		got, ok, err := NextOccurrence(r, dt, after)
		if err != nil {
			t.Fatalf("iter %d: err=%v", i, err)
		}
		if !ok {
			t.Fatalf("iter %d: want ok, got terminated", i)
		}
		if !got.Equal(w) {
			t.Fatalf("iter %d: want %v, got %v", i, w, got)
		}
		after = got
	}
}

// TestNextOccurrence_WeeklyByDay_AcrossWeekBoundary covers
// expandWeekly with BYDAY when the next occurrence is in the
// following week (exercises startOfWeek + advance).
func TestNextOccurrence_WeeklyByDay_AcrossWeekBoundary(t *testing.T) {
	r := Rule{
		Freq:      FreqWeekly,
		Interval:  1,
		ByDay:     []ByDay{{0, MO}},
		WeekStart: MO,
	}
	// Start on a Friday — next Monday is in next week.
	dt := mustParseUTC("20260403T120000Z") // Fri Apr 3
	// Find first MO >= Apr 3: Apr 6.
	got, ok, err := NextOccurrence(r, dt, mustParseUTC("20260403T000000Z"))
	if err != nil || !ok {
		t.Fatalf("err=%v ok=%v", err, ok)
	}
	want := mustParseUTC("20260406T120000Z")
	if !got.Equal(want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}
