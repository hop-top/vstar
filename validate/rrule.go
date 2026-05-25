// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"errors"
	"strings"

	vstar "hop.top/vstar"
	"hop.top/vstar/rrule"
)

// propRRULE is the RFC 5545 §3.8.5.3 property name. Pulled to a
// constant to keep the goconst linter happy and to centralize
// the case-insensitive name we match in checkRRule.
const propRRULE = "RRULE"

// CodeRRuleUnsupported — an RRULE property value parses
// syntactically but uses a feature ADR-0009 defers from the v0.2
// rrule scope. Post the BYSETPOS/BYWEEKNO/BYYEARDAY amendment, the
// remaining deferred surface is FREQ=SECONDLY, FREQ=MINUTELY, and
// RSCALE (RFC 7529 non-Gregorian calendars).
//
// This is a SeverityWarning, not an Error: the property still
// round-trips correctly through the codec layer; only its
// recurrence semantics are inaccessible to the v0.2 evaluator.
// Consumers using vstar to inspect recurrence MUST check for VS050
// before relying on rrule.NextOccurrence output.
const CodeRRuleUnsupported = "VS050"

// CodeRRuleMalformed — an RRULE property value is syntactically
// invalid (missing FREQ, INTERVAL=0/negative, both UNTIL+COUNT,
// BYMONTHDAY=0, etc. — every error rrule.ParseRRule wraps as
// vstar.ErrMalformed).
//
// SeverityError because a malformed RRULE means the producer
// cannot expect any consumer (vstar or otherwise) to evaluate
// recurrence correctly; the document violates RFC 5545 §3.3.10.
const CodeRRuleMalformed = "VS051"

// checkRRule emits one Diagnostic per RRULE property whose value
// fails rrule.ValidateRRule. The split is by sentinel:
//
//   - errors.Is(err, rrule.ErrUnsupportedRRule) → VS050 (Warning).
//   - errors.Is(err, vstar.ErrMalformed)        → VS051 (Error).
//
// When both classify (defensive — should not happen in practice)
// the Unsupported path wins because it is the more specific
// classification.
func checkRRule(c vstar.Component, path string) []Diagnostic {
	var out []Diagnostic
	for _, p := range c.Props {
		if !strings.EqualFold(p.Name, propRRULE) {
			continue
		}
		err := rrule.ValidateRRule(p.Value)
		if err == nil {
			continue
		}
		if errors.Is(err, rrule.ErrUnsupportedRRule) {
			out = append(out, Diagnostic{
				Severity: SeverityWarning,
				Code:     CodeRRuleUnsupported,
				Message:  "RRULE uses a feature outside the v0.2 rrule scope (ADR-0009): " + err.Error(),
				Path:     path + "." + propRRULE,
			})
			continue
		}
		if errors.Is(err, vstar.ErrMalformed) {
			out = append(out, Diagnostic{
				Severity: SeverityError,
				Code:     CodeRRuleMalformed,
				Message:  "RRULE is malformed (RFC 5545 §3.3.10): " + err.Error(),
				Path:     path + "." + propRRULE,
			})
			continue
		}
		// Unknown classification — emit as malformed conservatively
		// so the consumer at least sees the finding.
		out = append(out, Diagnostic{
			Severity: SeverityError,
			Code:     CodeRRuleMalformed,
			Message:  "RRULE failed validation: " + err.Error(),
			Path:     path + "." + propRRULE,
		})
	}
	return out
}
