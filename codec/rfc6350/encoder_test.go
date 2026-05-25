// SPDX-License-Identifier: Apache-2.0

package rfc6350

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	vstar "hop.top/vstar"
)

// TestEncode_MissingUID verifies that a Card with empty UID is rejected
// per the plan's T3 RED rule (no placeholder).
func TestEncode_MissingUID(t *testing.T) {
	enc := NewEncoder()
	var buf bytes.Buffer
	err := enc.Encode(&buf, vstar.Card{Props: []vstar.Property{{Name: "FN", Value: "Jad"}}})
	if err == nil {
		t.Fatalf("expected ErrMissingUID; got nil")
	}
	if !errors.Is(err, vstar.ErrMissingUID) {
		t.Fatalf("err mismatch: got %v want wraps %v", err, vstar.ErrMissingUID)
	}
}

// TestEncode_MinimalVCard checks the framing: BEGIN, VERSION:4.0, UID,
// other props, END — all CRLF-terminated, property names uppercased.
func TestEncode_MinimalVCard(t *testing.T) {
	enc := NewEncoder()
	var buf bytes.Buffer
	err := enc.Encode(&buf, vstar.Card{
		UID: "urn:uuid:1",
		Props: []vstar.Property{
			{Name: "fn", Value: "Jad Bitar"},
		},
	})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	want := "BEGIN:VCARD\r\n" +
		"VERSION:4.0\r\n" +
		"UID:urn:uuid:1\r\n" +
		"FN:Jad Bitar\r\n" +
		"END:VCARD\r\n"
	if got := buf.String(); got != want {
		t.Errorf("encoded mismatch\n got=%q\nwant=%q", got, want)
	}
}

// TestEncode_KindEmittedFirst verifies KIND, when set, is emitted
// after VERSION+UID and before other properties (RFC 6350 §6.1.4 is
// silent on order; we lock it for determinism).
func TestEncode_KindEmittedFirst(t *testing.T) {
	enc := NewEncoder()
	var buf bytes.Buffer
	err := enc.Encode(&buf, vstar.Card{
		UID:  "u",
		Kind: vstar.KindOrg,
		Props: []vstar.Property{
			{Name: "FN", Value: "ACME"},
		},
	})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	want := "BEGIN:VCARD\r\n" +
		"VERSION:4.0\r\n" +
		"UID:u\r\n" +
		"KIND:org\r\n" +
		"FN:ACME\r\n" +
		"END:VCARD\r\n"
	if got := buf.String(); got != want {
		t.Errorf("mismatch\n got=%q\nwant=%q", got, want)
	}
}

// TestEncode_EscapesText verifies TEXT escaping is applied on encode
// for ',' ';' '\n' '\\' inside property values.
func TestEncode_EscapesText(t *testing.T) {
	enc := NewEncoder()
	var buf bytes.Buffer
	err := enc.Encode(&buf, vstar.Card{
		UID: "u",
		Props: []vstar.Property{
			{Name: "N", Value: "Last,Comma;Jad;;;"},
			{Name: "NOTE", Value: "line\nbreak"},
		},
	})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got := buf.String()
	// Note: N's '\;' is part of structured-value semantics in RFC
	// 6350, but at the codec layer we treat all property values as
	// TEXT; consumers needing structured-value semantics live above.
	if !strings.Contains(got, `N:Last\,Comma\;Jad\;\;\;`+"\r\n") {
		t.Errorf("N escape: got=%q", got)
	}
	if !strings.Contains(got, `NOTE:line\nbreak`+"\r\n") {
		t.Errorf("NOTE escape: got=%q", got)
	}
}

// TestEncode_ParamSerialization verifies parameters are written as
// NAME;P1=V1;P2=V2:value with DQUOTE'ing when the value contains
// reserved chars (',' ':' ';').
func TestEncode_ParamSerialization(t *testing.T) {
	enc := NewEncoder()
	var buf bytes.Buffer
	err := enc.Encode(&buf, vstar.Card{
		UID: "u",
		Props: []vstar.Property{
			{
				Name:   "EMAIL",
				Value:  "jad@example.com",
				Params: []vstar.Param{{Name: "TYPE", Value: "work,home"}},
			},
		},
	})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !strings.Contains(buf.String(), `EMAIL;TYPE="work,home":jad@example.com`+"\r\n") {
		t.Errorf("EMAIL param: got=%q", buf.String())
	}
}

// TestEncode_PreservesGroupPrefix verifies group prefixes survive
// encode (lowercase preserved; only the *property name* segment is
// uppercased per RFC 6350 §3.3 conventions).
func TestEncode_PreservesGroupPrefix(t *testing.T) {
	enc := NewEncoder()
	var buf bytes.Buffer
	err := enc.Encode(&buf, vstar.Card{
		UID: "u",
		Props: []vstar.Property{
			{Name: "home.tel", Value: "tel:+15555550100"},
		},
	})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !strings.Contains(buf.String(), "home.TEL:tel:+15555550100\r\n") {
		t.Errorf("group preserve + name upper: got=%q", buf.String())
	}
}

// TestEncode_Folds verifies CRLF + SP folding at 75 octets per RFC
// 6350 §3.2 (which references RFC 5545 §3.1).
func TestEncode_Folds(t *testing.T) {
	long := strings.Repeat("a", 200)
	enc := NewEncoder()
	var buf bytes.Buffer
	err := enc.Encode(&buf, vstar.Card{
		UID: "u",
		Props: []vstar.Property{
			{Name: "ADR", Value: long},
		},
	})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got := buf.String()
	// Each non-first physical line in the fold begins with "\r\n " (SP).
	// Confirm at least two folds (200 + name overhead → ≥3 physical lines).
	if strings.Count(got, "\r\n ") < 2 {
		t.Errorf("expected fold markers; got=%q", got)
	}
	// No physical line >75 octets (count between CRLFs).
	for i, line := range strings.Split(got, "\r\n") {
		if len(line) > 75 {
			t.Errorf("line %d exceeds 75 octets (%d): %q", i, len(line), line)
		}
	}
}

// TestRoundTrip_FullCard parse → encode → parse should round-trip to
// a semantically equal Card.
func TestRoundTrip_FullCard(t *testing.T) {
	in := "BEGIN:VCARD\r\n" +
		"VERSION:4.0\r\n" +
		"UID:urn:uuid:abc\r\n" +
		"KIND:individual\r\n" +
		"FN:Jad Bitar\r\n" +
		"home.TEL;TYPE=voice:tel:+15555550100\r\n" +
		`N:Last\,Comma;Jad;;;` + "\r\n" +
		"END:VCARD\r\n"

	p := NewParser()
	cards, err := p.Parse(strings.NewReader(in))
	if err != nil {
		t.Fatalf("Parse(in): %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card; got %d", len(cards))
	}

	enc := NewEncoder()
	var buf bytes.Buffer
	if err := enc.Encode(&buf, cards[0]); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	cards2, err := p.Parse(&buf)
	if err != nil {
		t.Fatalf("Parse(out): %v\nwire=%q", err, buf.String())
	}
	if len(cards2) != 1 {
		t.Fatalf("expected 1 card on round-trip; got %d", len(cards2))
	}

	if !cardsSemEqual(cards[0], cards2[0]) {
		t.Errorf("round-trip mismatch\noriginal=%#v\nround-trip=%#v", cards[0], cards2[0])
	}
}

// cardsSemEqual returns true when two Cards carry the same UID, Kind,
// and property set (Property.Equal applied to each pair, in order-
// independent fashion via Get).
func cardsSemEqual(a, b vstar.Card) bool {
	if a.UID != b.UID || a.Kind != b.Kind {
		return false
	}
	if len(a.Props) != len(b.Props) {
		return false
	}
	for _, ap := range a.Props {
		bp, ok := b.Get(ap.Name)
		if !ok {
			return false
		}
		if !vstar.Equal(ap, bp) {
			return false
		}
	}
	return true
}

// TestEncode_DoesNotCloseWriter verifies Encode is composable with
// streaming writers — it must not call Close.
type noCloseWriter struct {
	bytes.Buffer
	closed bool
}

func (w *noCloseWriter) Close() error {
	w.closed = true
	return nil
}

func TestEncode_DoesNotCloseWriter(t *testing.T) {
	enc := NewEncoder()
	w := &noCloseWriter{}
	err := enc.Encode(w, vstar.Card{
		UID:   "u",
		Props: []vstar.Property{{Name: "FN", Value: "x"}},
	})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if w.closed {
		t.Errorf("Encode must not call Close")
	}
}
