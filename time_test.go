// SPDX-License-Identifier: Apache-2.0

package vstar

import (
	"testing"
	"time"
)

// TestFormatTime_KnownInstant verifies a well-known UTC instant is
// formatted using RFC 5545 §3.3.5 form #2 ("date with UTC time").
func TestFormatTime_KnownInstant(t *testing.T) {
	in := time.Date(2026, 5, 4, 18, 30, 45, 0, time.UTC)
	got := FormatTime(in)
	want := "20260504T183045Z"
	if got != want {
		t.Errorf("FormatTime(%v) = %q, want %q", in, got, want)
	}
}

// TestFormatTime_NonUTCConverted verifies non-UTC inputs are
// converted to UTC before formatting (callers don't have to pre-
// normalise; FormatTime is the chokepoint for the UTC invariant).
func TestFormatTime_NonUTCConverted(t *testing.T) {
	mtl := time.FixedZone("EST", -5*60*60) // arbitrary fixed -5h zone.
	// 13:30:45 -05:00 → 18:30:45 UTC.
	in := time.Date(2026, 5, 4, 13, 30, 45, 0, mtl)
	got := FormatTime(in)
	want := "20260504T183045Z"
	if got != want {
		t.Errorf("FormatTime(%v) = %q, want %q", in, got, want)
	}
}

// TestFormatTime_ZeroEmpty verifies the zero time renders as empty
// string. Callers (Set* writers) treat empty as "clear the property".
func TestFormatTime_ZeroEmpty(t *testing.T) {
	got := FormatTime(time.Time{})
	if got != "" {
		t.Errorf("FormatTime(zero) = %q, want %q", got, "")
	}
}

// TestFormatTime_NanosTruncated verifies sub-second precision is
// dropped — RFC 5545 form #2 has only second resolution.
func TestFormatTime_NanosTruncated(t *testing.T) {
	in := time.Date(2026, 5, 4, 18, 30, 45, 123456789, time.UTC)
	got := FormatTime(in)
	want := "20260504T183045Z"
	if got != want {
		t.Errorf("FormatTime(%v) = %q, want %q", in, got, want)
	}
}

// TestParseTime_FormTwoUTC verifies the canonical form #2
// (`YYYYMMDDTHHMMSSZ`) parses to the equivalent UTC instant.
func TestParseTime_FormTwoUTC(t *testing.T) {
	got, ok := ParseTime("20260504T183045Z")
	if !ok {
		t.Fatalf("ParseTime: ok=false, want true")
	}
	want := time.Date(2026, 5, 4, 18, 30, 45, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("ParseTime = %v, want %v", got, want)
	}
	if got.Location() != time.UTC {
		t.Errorf("ParseTime Location = %v, want UTC", got.Location())
	}
}

// TestParseTime_RejectsFormOneLocal verifies form #1 (no Z, local
// time) is rejected. Floating local times are not supported in v0.1
// and silent coercion would mask torn data.
func TestParseTime_RejectsFormOneLocal(t *testing.T) {
	got, ok := ParseTime("20260504T183045")
	if ok {
		t.Errorf("ParseTime(form #1 local) = (%v, true), want (zero, false)", got)
	}
	if !got.IsZero() {
		t.Errorf("ParseTime returned non-zero on failure: %v", got)
	}
}

// TestParseTime_RejectsFormThreeTZID verifies a form #3 input (with
// TZID parameter, handled by ParseTimeWithTZID) is rejected by
// ParseTime — that form must go through the TZID-aware entrypoint.
// Note: form #3 wire shape on the value side looks identical to
// form #1; the TZID parameter lives in the property params, not in
// the value. So we cover that via the form #1 rejection above. This
// test additionally rejects an input that looks like a date-only
// (form #1 of the DATE value type) which must NOT silently parse.
func TestParseTime_RejectsDateOnly(t *testing.T) {
	got, ok := ParseTime("20260504")
	if ok {
		t.Errorf("ParseTime(date-only) = (%v, true), want (zero, false)", got)
	}
}

// TestParseTime_RejectsRFC3339 verifies RFC 3339 layouts are
// rejected. Strict by design — readers should see torn data, not
// silently coerce.
func TestParseTime_RejectsRFC3339(t *testing.T) {
	cases := []string{
		"2026-05-04T18:30:45Z",
		"2026-05-04T18:30:45+00:00",
		"2026-05-04T18:30:45.123Z",
	}
	for _, in := range cases {
		got, ok := ParseTime(in)
		if ok {
			t.Errorf("ParseTime(%q) = (%v, true), want (zero, false)", in, got)
		}
	}
}

// TestParseTime_RejectsEmpty verifies the empty string fails.
func TestParseTime_RejectsEmpty(t *testing.T) {
	got, ok := ParseTime("")
	if ok {
		t.Errorf("ParseTime(\"\") = (%v, true), want (zero, false)", got)
	}
}

// TestParseTime_RejectsLowerZ verifies a lower-case `z` suffix is
// rejected. RFC 5545 §3.3.5 specifies upper-case Z; case-insensitive
// parsing here would mask invalid producer output.
func TestParseTime_RejectsLowerZ(t *testing.T) {
	got, ok := ParseTime("20260504T183045z")
	if ok {
		t.Errorf("ParseTime(lower z) = (%v, true), want (zero, false)", got)
	}
}

// TestParseTime_RejectsTrailingJunk verifies extra trailing or
// leading characters fail.
func TestParseTime_RejectsTrailingJunk(t *testing.T) {
	cases := []string{
		"20260504T183045Z ",
		" 20260504T183045Z",
		"20260504T183045ZZ",
		"X20260504T183045Z",
	}
	for _, in := range cases {
		got, ok := ParseTime(in)
		if ok {
			t.Errorf("ParseTime(%q) = (%v, true), want (zero, false)", in, got)
		}
	}
}

// TestParseTime_RejectsImpossibleDate verifies obviously invalid
// dates (wrong month/day) fail rather than rolling over.
func TestParseTime_RejectsImpossibleDate(t *testing.T) {
	cases := []string{
		"20261301T000000Z", // month 13
		"20260230T000000Z", // Feb 30
		"20260504T256000Z", // hour 25
	}
	for _, in := range cases {
		got, ok := ParseTime(in)
		if ok {
			t.Errorf("ParseTime(%q) = (%v, true), want (zero, false)", in, got)
		}
	}
}

// TestParseTime_RoundTrip verifies FormatTime → ParseTime returns
// the original instant.
func TestParseTime_RoundTrip(t *testing.T) {
	in := time.Date(2026, 5, 4, 18, 30, 45, 0, time.UTC)
	s := FormatTime(in)
	got, ok := ParseTime(s)
	if !ok {
		t.Fatalf("ParseTime(%q) ok=false", s)
	}
	if !got.Equal(in) {
		t.Errorf("round-trip = %v, want %v", got, in)
	}
}

// americaMontrealCalendar mirrors testdata/time/america_montreal.ics
// in model form. See that fixture for the wire shape; once the
// rfc5545 parser lands we'll switch this to read the fixture file.
func americaMontrealCalendar() Calendar {
	daylight := Component{
		Type: CompType("DAYLIGHT"),
		Props: []Property{
			{Name: "DTSTART", Value: "20070311T020000"},
			{Name: "TZOFFSETFROM", Value: "-0500"},
			{Name: "TZOFFSETTO", Value: "-0400"},
			{Name: "TZNAME", Value: "EDT"},
			{Name: "RRULE", Value: "FREQ=YEARLY;BYMONTH=3;BYDAY=2SU"},
		},
	}
	standard := Component{
		Type: CompType("STANDARD"),
		Props: []Property{
			{Name: "DTSTART", Value: "20071104T020000"},
			{Name: "TZOFFSETFROM", Value: "-0400"},
			{Name: "TZOFFSETTO", Value: "-0500"},
			{Name: "TZNAME", Value: "EST"},
			{Name: "RRULE", Value: "FREQ=YEARLY;BYMONTH=11;BYDAY=1SU"},
		},
	}
	tz := Component{
		Type:  CompTimezone,
		Props: []Property{{Name: "TZID", Value: "America/Montreal"}},
		Sub:   []Component{daylight, standard},
	}
	return Calendar{
		ProdID:     "-//V*//vstar-go time tests//EN",
		Components: []Component{tz},
	}
}

// utcOnlyCalendar mirrors testdata/time/utc_only.ics — a STANDARD-
// only VTIMEZONE for the trivial fixed-offset path.
func utcOnlyCalendar() Calendar {
	standard := Component{
		Type: CompType("STANDARD"),
		Props: []Property{
			{Name: "DTSTART", Value: "19700101T000000"},
			{Name: "TZOFFSETFROM", Value: "+0000"},
			{Name: "TZOFFSETTO", Value: "+0000"},
			{Name: "TZNAME", Value: "UTC"},
		},
	}
	tz := Component{
		Type:  CompTimezone,
		Props: []Property{{Name: "TZID", Value: "UTC"}},
		Sub:   []Component{standard},
	}
	return Calendar{
		ProdID:     "-//V*//vstar-go time tests//EN",
		Components: []Component{tz},
	}
}

// TestParseTimeWithTZID_MontrealStandard verifies that a winter
// (EST) instant in America/Montreal is interpreted with the -0500
// offset — so 13:30:45 local → 18:30:45 UTC.
func TestParseTimeWithTZID_MontrealStandard(t *testing.T) {
	cal := americaMontrealCalendar()
	got, ok := ParseTimeWithTZID("20260104T133045", "America/Montreal", cal)
	if !ok {
		t.Fatalf("ParseTimeWithTZID: ok=false, want true")
	}
	want := time.Date(2026, 1, 4, 18, 30, 45, 0, time.UTC)
	if !got.UTC().Equal(want) {
		t.Errorf("ParseTimeWithTZID winter = %v (UTC %v), want %v",
			got, got.UTC(), want)
	}
}

// TestParseTimeWithTZID_MontrealDaylight verifies that a summer
// (EDT) instant in America/Montreal is interpreted with -0400 —
// so 13:30:45 local → 17:30:45 UTC.
func TestParseTimeWithTZID_MontrealDaylight(t *testing.T) {
	cal := americaMontrealCalendar()
	got, ok := ParseTimeWithTZID("20260704T133045", "America/Montreal", cal)
	if !ok {
		t.Fatalf("ParseTimeWithTZID: ok=false, want true")
	}
	want := time.Date(2026, 7, 4, 17, 30, 45, 0, time.UTC)
	if !got.UTC().Equal(want) {
		t.Errorf("ParseTimeWithTZID summer = %v (UTC %v), want %v",
			got, got.UTC(), want)
	}
}

// TestParseTimeWithTZID_StandardOnly verifies the trivial single-
// STANDARD path (no DST) round-trips with a fixed offset.
func TestParseTimeWithTZID_StandardOnly(t *testing.T) {
	cal := utcOnlyCalendar()
	got, ok := ParseTimeWithTZID("20260504T183045", "UTC", cal)
	if !ok {
		t.Fatalf("ParseTimeWithTZID: ok=false, want true")
	}
	want := time.Date(2026, 5, 4, 18, 30, 45, 0, time.UTC)
	if !got.UTC().Equal(want) {
		t.Errorf("ParseTimeWithTZID = %v (UTC %v), want %v", got, got.UTC(), want)
	}
}

// TestParseTimeWithTZID_MissingVTIMEZONE verifies that an unknown
// TZID fails closed.
func TestParseTimeWithTZID_MissingVTIMEZONE(t *testing.T) {
	cal := americaMontrealCalendar()
	got, ok := ParseTimeWithTZID("20260104T133045", "America/Toronto", cal)
	if ok {
		t.Errorf("ParseTimeWithTZID(missing) = (%v, true), want (zero, false)", got)
	}
}

// TestParseTimeWithTZID_EmptyTZID verifies an empty tzid fails.
func TestParseTimeWithTZID_EmptyTZID(t *testing.T) {
	cal := americaMontrealCalendar()
	got, ok := ParseTimeWithTZID("20260104T133045", "", cal)
	if ok {
		t.Errorf("ParseTimeWithTZID(empty) = (%v, true), want (zero, false)", got)
	}
}

// TestParseTimeWithTZID_MalformedValue verifies a value that isn't
// form #1 fails (e.g. has a Z suffix — that's form #2 and belongs
// to ParseTime).
func TestParseTimeWithTZID_MalformedValue(t *testing.T) {
	cal := americaMontrealCalendar()
	cases := []string{
		"20260104T133045Z",
		"2026-01-04T13:30:45",
		"",
		"20260104",
	}
	for _, in := range cases {
		got, ok := ParseTimeWithTZID(in, "America/Montreal", cal)
		if ok {
			t.Errorf("ParseTimeWithTZID(%q) = (%v, true), want false", in, got)
		}
	}
}

// TestParseTimeWithTZID_MalformedVTIMEZONE verifies a VTIMEZONE
// with no STANDARD/DAYLIGHT children fails.
func TestParseTimeWithTZID_MalformedVTIMEZONE(t *testing.T) {
	cal := Calendar{
		Components: []Component{{
			Type:  CompTimezone,
			Props: []Property{{Name: "TZID", Value: "BrokenZone"}},
			// No Sub — no STANDARD/DAYLIGHT.
		}},
	}
	got, ok := ParseTimeWithTZID("20260104T133045", "BrokenZone", cal)
	if ok {
		t.Errorf("ParseTimeWithTZID(no rules) = (%v, true), want false", got)
	}
}

// TestParseTimeWithTZID_BadOffsetRejected verifies a malformed
// TZOFFSETTO/FROM (not ±HHMM) causes the lookup to fail.
func TestParseTimeWithTZID_BadOffsetRejected(t *testing.T) {
	cal := Calendar{
		Components: []Component{{
			Type:  CompTimezone,
			Props: []Property{{Name: "TZID", Value: "Bad"}},
			Sub: []Component{{
				Type: CompType("STANDARD"),
				Props: []Property{
					{Name: "DTSTART", Value: "19700101T000000"},
					{Name: "TZOFFSETFROM", Value: "garbage"},
					{Name: "TZOFFSETTO", Value: "garbage"},
				},
			}},
		}},
	}
	got, ok := ParseTimeWithTZID("20260104T133045", "Bad", cal)
	if ok {
		t.Errorf("ParseTimeWithTZID(bad offset) = (%v, true), want false", got)
	}
}

// TestParseTimeWithTZID_LastSundayBYDAY verifies the negative-
// ordinal BYDAY form (e.g. -1SU = "last Sunday of month") is
// honored. EU DST historically used "last Sunday of October" /
// "last Sunday of March".
func TestParseTimeWithTZID_LastSundayBYDAY(t *testing.T) {
	cal := Calendar{Components: []Component{{
		Type:  CompTimezone,
		Props: []Property{{Name: "TZID", Value: "Europe/Berlin"}},
		Sub: []Component{
			{
				Type: CompType("DAYLIGHT"),
				Props: []Property{
					{Name: "DTSTART", Value: "19960331T020000"},
					{Name: "TZOFFSETFROM", Value: "+0100"},
					{Name: "TZOFFSETTO", Value: "+0200"},
					{Name: "TZNAME", Value: "CEST"},
					{Name: "RRULE", Value: "FREQ=YEARLY;BYMONTH=3;BYDAY=-1SU"},
				},
			},
			{
				Type: CompType("STANDARD"),
				Props: []Property{
					{Name: "DTSTART", Value: "19961027T030000"},
					{Name: "TZOFFSETFROM", Value: "+0200"},
					{Name: "TZOFFSETTO", Value: "+0100"},
					{Name: "TZNAME", Value: "CET"},
					{Name: "RRULE", Value: "FREQ=YEARLY;BYMONTH=10;BYDAY=-1SU"},
				},
			},
		},
	}}}
	// July 4 2026 — DST active (CEST = +0200).
	got, ok := ParseTimeWithTZID("20260704T140000", "Europe/Berlin", cal)
	if !ok {
		t.Fatalf("ParseTimeWithTZID: ok=false")
	}
	want := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	if !got.UTC().Equal(want) {
		t.Errorf("ParseTimeWithTZID = %v (UTC %v), want %v", got, got.UTC(), want)
	}
	// January 4 2026 — STANDARD (CET = +0100).
	got2, ok := ParseTimeWithTZID("20260104T140000", "Europe/Berlin", cal)
	if !ok {
		t.Fatalf("ParseTimeWithTZID winter: ok=false")
	}
	want2 := time.Date(2026, 1, 4, 13, 0, 0, 0, time.UTC)
	if !got2.UTC().Equal(want2) {
		t.Errorf("ParseTimeWithTZID winter = %v (UTC %v), want %v", got2, got2.UTC(), want2)
	}
}

// TestParseTimeWithTZID_AllWeekdayCodes verifies every two-letter
// weekday code parses. Uses minimal STANDARD-only zones with the
// rule's BYDAY field exercised purely by the parser.
func TestParseTimeWithTZID_AllWeekdayCodes(t *testing.T) {
	codes := []string{"SU", "MO", "TU", "WE", "TH", "FR", "SA"}
	for _, code := range codes {
		cal := Calendar{Components: []Component{{
			Type:  CompTimezone,
			Props: []Property{{Name: "TZID", Value: "X"}},
			Sub: []Component{
				{
					Type: CompType("STANDARD"),
					Props: []Property{
						{Name: "DTSTART", Value: "19700101T000000"},
						{Name: "TZOFFSETFROM", Value: "+0000"},
						{Name: "TZOFFSETTO", Value: "+0000"},
						{Name: "RRULE", Value: "FREQ=YEARLY;BYMONTH=1;BYDAY=1" + code},
					},
				},
				{
					Type: CompType("DAYLIGHT"),
					Props: []Property{
						{Name: "DTSTART", Value: "19700601T000000"},
						{Name: "TZOFFSETFROM", Value: "+0000"},
						{Name: "TZOFFSETTO", Value: "+0100"},
						{Name: "RRULE", Value: "FREQ=YEARLY;BYMONTH=6;BYDAY=1" + code},
					},
				},
			},
		}}}
		_, ok := ParseTimeWithTZID("20260315T120000", "X", cal)
		if !ok {
			t.Errorf("BYDAY=1%s: ok=false, want true", code)
		}
	}
	// Bogus weekday code rejected.
	cal := Calendar{Components: []Component{{
		Type:  CompTimezone,
		Props: []Property{{Name: "TZID", Value: "Y"}},
		Sub: []Component{{
			Type: CompType("STANDARD"),
			Props: []Property{
				{Name: "DTSTART", Value: "19700101T000000"},
				{Name: "TZOFFSETFROM", Value: "+0000"},
				{Name: "TZOFFSETTO", Value: "+0000"},
				{Name: "RRULE", Value: "FREQ=YEARLY;BYMONTH=1;BYDAY=1XX"},
			},
		}},
	}}}
	if _, ok := ParseTimeWithTZID("20260315T120000", "Y", cal); ok {
		t.Errorf("bogus weekday code: ok=true, want false")
	}
}

// TestParseTimeWithTZID_UnsupportedRRULE verifies any RRULE that
// isn't FREQ=YEARLY is rejected — we don't silently apply rules we
// don't understand.
func TestParseTimeWithTZID_UnsupportedRRULE(t *testing.T) {
	cal := Calendar{
		Components: []Component{{
			Type:  CompTimezone,
			Props: []Property{{Name: "TZID", Value: "Weird"}},
			Sub: []Component{{
				Type: CompType("STANDARD"),
				Props: []Property{
					{Name: "DTSTART", Value: "20070101T000000"},
					{Name: "TZOFFSETFROM", Value: "-0500"},
					{Name: "TZOFFSETTO", Value: "-0500"},
					{Name: "RRULE", Value: "FREQ=MONTHLY;BYDAY=1SU"},
				},
			}},
		}},
	}
	got, ok := ParseTimeWithTZID("20260104T133045", "Weird", cal)
	if ok {
		t.Errorf("ParseTimeWithTZID(non-yearly rrule) = (%v, true), want false", got)
	}
}
