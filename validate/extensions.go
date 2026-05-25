// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"strings"

	vstar "hop.top/vstar"
	"hop.top/vstar/ext"
)

// CodeUnknownProperty — a property's name is not on the standard
// allow-list and does not begin with the X- extension prefix
// (spec/05 §3, spec/04 — extension discipline).
//
// This is a SeverityWarning, not an Error: receivers MUST already
// ignore unknown X-* properties per RFC 5545, and a non-X- unknown
// property is more likely a typo or vendor-private than a hard
// spec violation. Emitting a Warning lets tooling surface the
// finding without blocking the document.
const CodeUnknownProperty = "VS020"

// checkExtensionNamespace emits one Warning per non-standard
// property whose name does not begin with the X- prefix. A
// trailing "." in the path locator marks the property name.
//
// The X- classification is delegated to ext.IsExtension to keep a
// single source of truth for "what counts as an extension". The
// standard-property allow-list (standard_properties.go) remains
// owned by validate because it answers the orthogonal question
// "is this a known RFC 5545/6350 property?", which the ext package
// has no business knowing.
func checkExtensionNamespace(c vstar.Component, path string) []Diagnostic {
	var out []Diagnostic
	for _, p := range c.Props {
		if isStandardProperty(p.Name) {
			continue
		}
		if ext.IsExtension(p.Name) {
			continue
		}
		out = append(out, Diagnostic{
			Severity: SeverityWarning,
			Code:     CodeUnknownProperty,
			Message:  "property " + p.Name + " is not a known RFC 5545/6350 property and does not use the X- extension prefix (spec/05 §3, spec/04)",
			Path:     path + "." + p.Name,
		})
	}
	return out
}

// isStandardProperty reports whether name (case-insensitive)
// matches a property in the RFC 5545/6350 allow-list.
func isStandardProperty(name string) bool {
	_, ok := standardProperties[strings.ToUpper(name)]
	return ok
}
