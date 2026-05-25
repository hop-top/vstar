// SPDX-License-Identifier: Apache-2.0

package validate

import (
	vstar "hop.top/vstar"
)

// Stable diagnostic codes for spec/05 §1 — required common
// properties. Per spec/02 every V* component MUST carry UID,
// DTSTAMP, X-VSTAR-HASH; absence of any one is an Error.
const (
	// CodeMissingUID — required UID property is absent (spec/02).
	CodeMissingUID = "VS001"
	// CodeMissingDTSTAMP — required DTSTAMP property is absent
	// (spec/02).
	CodeMissingDTSTAMP = "VS002"
	// CodeMissingXVSTARHash — required X-VSTAR-HASH property is
	// absent (spec/02). When the property is present but wrong,
	// see CodeBadXVSTARHash (VS010) instead.
	CodeMissingXVSTARHash = "VS003"
)

// xvstarHashProperty is the property name V* uses to carry the
// content hash. Mirrors hashing.XVSTARHashProperty without
// importing hashing here (validate has its own const so it can
// inspect raw property names without a transitive cycle risk).
const xvstarHashProperty = "X-VSTAR-HASH"

// checkRequiredCommon emits one Error diagnostic per missing
// required common property. UID, DTSTAMP, X-VSTAR-HASH are checked
// in that order so the diagnostic stream is deterministic.
//
// Path is the dotted component locator (e.g. "VCALENDAR.VTODO[uid=
// foo]"); per-property paths append the property name (e.g.
// "...].UID").
func checkRequiredCommon(c vstar.Component, path string) []Diagnostic {
	var out []Diagnostic
	if _, ok := c.Get("UID"); !ok {
		out = append(out, Diagnostic{
			Severity: SeverityError,
			Code:     CodeMissingUID,
			Message:  "required common property UID is missing (spec/02)",
			Path:     path + ".UID",
		})
	}
	if _, ok := c.Get("DTSTAMP"); !ok {
		out = append(out, Diagnostic{
			Severity: SeverityError,
			Code:     CodeMissingDTSTAMP,
			Message:  "required common property DTSTAMP is missing (spec/02)",
			Path:     path + ".DTSTAMP",
		})
	}
	if _, ok := c.Get(xvstarHashProperty); !ok {
		out = append(out, Diagnostic{
			Severity: SeverityError,
			Code:     CodeMissingXVSTARHash,
			Message:  "required common property X-VSTAR-HASH is missing (spec/02)",
			Path:     path + "." + xvstarHashProperty,
		})
	}
	return out
}
