// SPDX-License-Identifier: Apache-2.0

package validate

// standardProperties is the allow-list of property names known to
// be defined by RFC 5545 (iCalendar) §3.7-§3.8 and RFC 6350
// (vCard) §6. A property whose name is not on this list AND does
// not begin with "X-" is considered an unknown property and earns
// a SeverityWarning VS020 from spec/05 §3 (extension namespace
// compliance).
//
// This list is intentionally inline. The sister vstar-go-extensions
// track will publish ext.ScopeOf(name); when that lands, a follow-
// up PR replaces the local check with the shared helper. Until
// then, depending on ext would form a sister-package coupling that
// neither track is allowed to assume during Wave 4.
//
// The allow-list is conservative. If a future RFC errata or
// extension introduces a new property name we should accept, add
// it here — VS020 is a Warning, not an Error, so under-listing
// only produces noise, never a correctness failure.
//
// Keys are upper-cased; lookup is case-insensitive at the call
// site (see isStandardProperty).
var standardProperties = map[string]struct{}{
	// RFC 5545 §3.7 calendar properties.
	"CALSCALE": {},
	"METHOD":   {},
	"PRODID":   {},
	"VERSION":  {},

	// RFC 5545 §3.8.1 descriptive component properties.
	"ATTACH":           {},
	"CATEGORIES":       {},
	"CLASS":            {},
	"COMMENT":          {},
	"DESCRIPTION":      {},
	"GEO":              {},
	"LOCATION":         {},
	"PERCENT-COMPLETE": {},
	"PRIORITY":         {},
	"RESOURCES":        {},
	"STATUS":           {},
	"SUMMARY":          {},

	// RFC 5545 §3.8.2 date and time component properties.
	"COMPLETED": {},
	"DTEND":     {},
	"DUE":       {},
	"DTSTART":   {},
	"DURATION":  {},
	"FREEBUSY":  {},
	"TRANSP":    {},

	// RFC 5545 §3.8.3 time zone component properties.
	"TZID":         {},
	"TZNAME":       {},
	"TZOFFSETFROM": {},
	"TZOFFSETTO":   {},
	"TZURL":        {},

	// RFC 5545 §3.8.4 relationship component properties.
	"ATTENDEE":      {},
	"CONTACT":       {},
	"ORGANIZER":     {},
	"RECURRENCE-ID": {},
	"RELATED-TO":    {},
	"URL":           {},
	"UID":           {}, //nolint:goconst // map literal; propUID const lives in caller files (required.go, types.go) which use the bare string for symmetry with this allow-list.

	// RFC 5545 §3.8.5 recurrence component properties.
	"EXDATE": {},
	"EXRULE": {}, // deprecated by RFC 5545, still accepted by readers.
	"RDATE":  {},
	"RRULE":  {}, //nolint:goconst // map literal; propRRULE const in rrule.go is for diagnostic Path strings, not this allow-list.

	// RFC 5545 §3.8.6 alarm component properties.
	"ACTION":  {},
	"REPEAT":  {},
	"TRIGGER": {},

	// RFC 5545 §3.8.7 change management component properties.
	"CREATED":       {},
	"DTSTAMP":       {},
	"LAST-MODIFIED": {},
	"SEQUENCE":      {},

	// RFC 5545 §3.8.8 miscellaneous component properties.
	"REQUEST-STATUS": {},

	// RFC 6350 §6 vCard properties.
	"ADR":          {},
	"ANNIVERSARY":  {},
	"BDAY":         {},
	"CALADRURI":    {},
	"CALURI":       {},
	"CLIENTPIDMAP": {},
	"EMAIL":        {},
	"FBURL":        {},
	"FN":           {},
	"GENDER":       {},
	"IMPP":         {},
	"KEY":          {},
	"KIND":         {},
	"LANG":         {},
	"LOGO":         {},
	"MEMBER":       {},
	"N":            {},
	"NICKNAME":     {},
	"NOTE":         {},
	"ORG":          {},
	"PHOTO":        {},
	"RELATED":      {},
	"REV":          {},
	"ROLE":         {},
	"SOUND":        {},
	"SOURCE":       {},
	"TEL":          {},
	"TITLE":        {},
	"TZ":           {},
	"XML":          {},
}

// StandardPropertyCount returns the number of property names in
// the standard allow-list. Exported for diagnostic and test code
// (notably the docs-coverage CI test) — not part of the validator
// hot path.
func StandardPropertyCount() int {
	return len(standardProperties)
}
