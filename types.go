// SPDX-License-Identifier: Apache-2.0

// Package vstar is the Go reference implementation of the V*
// specification (a convention over RFC 5545 / RFC 6350 for agentic
// systems). This package exposes the in-memory data model —
// Property, Param, Component, Calendar, Card — plus the wire-string
// enums (CompType, Kind, TodoStatus) and error sentinels every later
// layer (codecs, canonicalization, validation, helpers) builds on.
//
// I/O, parsing, encoding, canonical form, time helpers, and
// validation live in dedicated packages; this file ships data types
// only.
//
// See README.md, ../spec/, and ../docs/adrs/.
package vstar

// CompType is the wire-string component type of a Component
// (e.g. "VEVENT"). Constants below mirror the RFC 5545 §3.6 component
// identifiers exactly. VCARD is intentionally absent — vCards are
// represented by Card, not Component.
type CompType string

// RFC 5545 §3.4 + §3.6 component identifiers, plus the VCALENDAR
// container identifier from §3.4.
const (
	CompCalendar CompType = "VCALENDAR"
	CompTodo     CompType = "VTODO"
	CompJournal  CompType = "VJOURNAL"
	CompEvent    CompType = "VEVENT"
	CompFreeBusy CompType = "VFREEBUSY"
	CompTimezone CompType = "VTIMEZONE"
	CompAlarm    CompType = "VALARM"
)

// Kind is the wire-string KIND value for VCARD components per
// RFC 6350 §6.1.4. Values are lowercase per the RFC's IANA registry.
type Kind string

// RFC 6350 §6.1.4 KIND values. The RFC also lists "location"; only
// the v0.1 enum values used by V* are defined here.
const (
	KindIndividual Kind = "individual"
	KindOrg        Kind = "org"
	KindGroup      Kind = "group"
)

// TodoStatus is the wire-string STATUS value for VTODO components
// per RFC 5545 §3.8.1.11. Values are uppercase per the RFC.
type TodoStatus string

// RFC 5545 §3.8.1.11 STATUS values applicable to VTODO.
const (
	TodoNeedsAction TodoStatus = "NEEDS-ACTION"
	TodoInProcess   TodoStatus = "IN-PROCESS"
	TodoCompleted   TodoStatus = "COMPLETED"
	TodoCancelled   TodoStatus = "CANCELLED" //nolint:misspell // RFC 5545 §3.8.1.11 wire string is "CANCELLED" (double L); not US "CANCELED".
)
