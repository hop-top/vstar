// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"time"

	vstar "hop.top/vstar"
	"hop.top/vstar/hashing"
)

// statusProp is the wire name for the VTODO STATUS property
// (RFC 5545 §3.8.1.11).
const statusProp = "STATUS"

// percentProp is the wire name for the VTODO PERCENT-COMPLETE
// property (RFC 5545 §3.8.1.8). Always carries an unsigned-int 0-100.
const percentProp = "PERCENT-COMPLETE"

// validTodoStatus reports whether s is one of the four wire values
// the spec enumerates for VTODO STATUS. Anything else is rejected by
// SetStatus and reported as ok=false by Status.
func validTodoStatus(s vstar.TodoStatus) bool {
	switch s {
	case vstar.TodoNeedsAction, vstar.TodoInProcess, vstar.TodoCompleted, vstar.TodoCancelled:
		return true
	default:
		return false
	}
}

// Status returns the parsed STATUS value of c when present and valid.
// Returns ok=false when STATUS is absent or carries an unrecognized
// value. Status does not enforce that c is a VTODO — callers wanting
// VTODO-specific semantics should gate on c.Type themselves.
func Status(c vstar.Component) (vstar.TodoStatus, bool) {
	p, ok := c.Get(statusProp)
	if !ok {
		return "", false
	}
	s := vstar.TodoStatus(p.Value)
	if !validTodoStatus(s) {
		return "", false
	}
	return s, true
}

// SetStatus writes the STATUS property on c and refreshes the
// X-VSTAR-HASH last. SetStatus is a no-op when:
//
//   - c is nil
//   - c.Type is not CompTodo (the four TodoStatus values are VTODO-
//     specific in v0.1; events/journals carry their own status enum
//     in RFC 5545 but we don't model those yet)
//   - s is not one of the four valid TodoStatus constants
//
// The no-op-on-mismatch shape is the Go convention for optional
// mutators on type-mixed inputs: callers don't get a false sense of
// success but also don't have to set up error-handling for a benign
// mistake.
func SetStatus(c *vstar.Component, s vstar.TodoStatus) {
	if c == nil || c.Type != vstar.CompTodo || !validTodoStatus(s) {
		return
	}
	c.Set(vstar.Property{Name: statusProp, Value: string(s)})
	hashing.SetXVSTAR(c)
}

// Complete finalizes a VTODO atomically: STATUS=COMPLETED,
// COMPLETED=t (UTC form), PERCENT-COMPLETE=100, X-VSTAR-HASH
// refreshed last.
//
// No-op when c is nil or c.Type is not CompTodo. The semantic mirrors
// the crm/internal/vcal precedent: if you call Complete you get the
// full set of "done" markers in one call and the stored hash matches.
func Complete(c *vstar.Component, t time.Time) {
	if c == nil || c.Type != vstar.CompTodo {
		return
	}
	c.Set(vstar.Property{Name: statusProp, Value: string(vstar.TodoCompleted)})
	c.SetCOMPLETED(t)
	c.Set(vstar.Property{Name: percentProp, Value: "100"})
	hashing.SetXVSTAR(c)
}
