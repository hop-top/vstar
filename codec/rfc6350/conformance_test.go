// SPDX-License-Identifier: Apache-2.0

package rfc6350

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	vstar "hop.top/vstar"
)

// fixtureDir resolves the conformance fixture directory at
// vstar/testdata/rfc6350. The package lives at codec/rfc6350,
// so we walk two levels up.
func fixtureDir(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "testdata", "rfc6350")
}

// TestConformance_RoundTrip exercises every .vcf fixture in
// vstar/testdata/rfc6350 by parsing, re-encoding, and re-parsing.
// The two parses MUST produce semantically equal Cards.
func TestConformance_RoundTrip(t *testing.T) {
	dir := fixtureDir(t)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("no fixtures found in %s", dir)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".vcf" {
			continue
		}
		t.Run(e.Name(), func(t *testing.T) {
			path := filepath.Join(dir, e.Name())
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			cards, err := NewParser().Parse(bytes.NewReader(raw))
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			if len(cards) == 0 {
				t.Fatalf("%s parsed to zero cards", path)
			}
			for i, c := range cards {
				var buf bytes.Buffer
				if err := NewEncoder().Encode(&buf, c); err != nil {
					t.Fatalf("encode %s[%d]: %v", path, i, err)
				}
				rt, err := NewParser().Parse(&buf)
				if err != nil {
					t.Fatalf("re-parse %s[%d]: %v\nwire=%s", path, i, err, buf.String())
				}
				if len(rt) != 1 {
					t.Fatalf("expected 1 card on round-trip; got %d", len(rt))
				}
				if !cardsConformanceEqual(c, rt[0]) {
					t.Errorf("%s[%d] round-trip mismatch\norig=%#v\nrt=%#v", path, i, c, rt[0])
				}
			}
		})
	}
}

// cardsConformanceEqual is a stricter version of cardsSemEqual that
// also enforces parameter equality via the package-root vstar.Equal.
func cardsConformanceEqual(a, b vstar.Card) bool {
	if a.UID != b.UID || a.Kind != b.Kind {
		return false
	}
	if len(a.Props) != len(b.Props) {
		return false
	}
	for i := range a.Props {
		if !vstar.Equal(a.Props[i], b.Props[i]) {
			return false
		}
	}
	return true
}
