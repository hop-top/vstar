// SPDX-License-Identifier: Apache-2.0

// Command fixtures-gen-supersession generates the four supersession
// edge-case .ics fixtures under testdata/supersession/. The
// `.canonical` and `.hash` regeneration is delegated to
// fixtures-verify; the `.notes.md` and README files are written by
// hand alongside the .ics in this directory and are NOT regenerated
// here.
//
// All times are fixed so re-running this generator produces
// byte-identical output. UIDs are spec/02-style human-readable
// stems for diff legibility.
//
// Outputs (each is a single VCALENDAR with UTF-8 / LF on disk):
//
//   - linear.ics          — original VTODO + one supersession VJOURNAL
//     flipping it to COMPLETED.
//   - multi_step.ics      — original VTODO + two supersession VJOURNALs
//     walking IN-PROCESS → COMPLETED.
//   - cross_component.ics — VTODO superseded by a VJOURNAL whose
//     RELATED-TO targets it; both live as
//     siblings in the same VCALENDAR (V*'s
//     "ledger is one container" model).
//   - corrupt_mutated.ics — original VTODO with X-VSTAR-HASH set,
//     then SUMMARY mutated post-stamp. The
//     hash no longer matches; consumers MUST
//     notice. NEGATIVE TEST FIXTURE.
package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	vstar "hop.top/vstar"
	"hop.top/vstar/codec/rfc5545"
	"hop.top/vstar/hashing"
	"hop.top/vstar/helpers"
	"hop.top/vstar/supersession"
)

func main() {
	root, err := repoRoot()
	if err != nil {
		fail(err)
	}
	out := filepath.Join(root, "testdata", "supersession")
	if err := os.MkdirAll(out, 0o755); err != nil {
		fail(fmt.Errorf("mkdir %s: %w", out, err))
	}

	// Fixed times — every run yields identical bytes.
	t0 := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 5, 4, 14, 30, 0, 0, time.UTC)
	t2 := time.Date(2026, 5, 4, 16, 45, 0, 0, time.UTC)

	if err := writeLinear(out, t0, t1); err != nil {
		fail(err)
	}
	if err := writeMultiStep(out, t0, t1, t2); err != nil {
		fail(err)
	}
	if err := writeCrossComponent(out, t0, t1); err != nil {
		fail(err)
	}
	if err := writeCorruptMutated(out, t0); err != nil {
		fail(err)
	}
	fmt.Println("fixtures-gen-supersession: wrote 4 .ics fixtures under", out)
}

// writeLinear: original VTODO at t0 + one supersession VJOURNAL at t1
// flipping it to COMPLETED.
func writeLinear(dir string, t0, t1 time.Time) error {
	target := newTodo("todo-1", "Buy milk", t0)
	hashing.SetXVSTAR(&target)
	sup, err := supersession.Supersedes(target, string(vstar.TodoCompleted), t1)
	if err != nil {
		return fmt.Errorf("linear: %w", err)
	}
	cal := vstar.Calendar{ProdID: "-//V*//Supersession-Linear//EN"}
	cal.Append(target)
	cal.Append(sup)
	return writeCalendar(dir, "linear", cal)
}

// writeMultiStep: original VTODO + two supersession VJOURNALs
// walking IN-PROCESS → COMPLETED.
func writeMultiStep(dir string, t0, t1, t2 time.Time) error {
	target := newTodo("todo-multi", "Ship Wave 5", t0)
	hashing.SetXVSTAR(&target)
	sup1, err := supersession.Supersedes(target, string(vstar.TodoInProcess), t1)
	if err != nil {
		return fmt.Errorf("multi_step sup1: %w", err)
	}
	sup2, err := supersession.Supersedes(target, string(vstar.TodoCompleted), t2)
	if err != nil {
		return fmt.Errorf("multi_step sup2: %w", err)
	}
	cal := vstar.Calendar{ProdID: "-//V*//Supersession-MultiStep//EN"}
	cal.Append(target)
	cal.Append(sup1)
	cal.Append(sup2)
	return writeCalendar(dir, "multi_step", cal)
}

// writeCrossComponent: VTODO superseded by a VJOURNAL whose
// RELATED-TO targets it; both as siblings in one VCALENDAR. The
// supersession journal carries an extra DESCRIPTION to flavor the
// "real" cross-component reference.
func writeCrossComponent(dir string, t0, t1 time.Time) error {
	target := newTodo("turn-42", "Resolve branch conflict", t0)
	hashing.SetXVSTAR(&target)

	sup, err := supersession.Supersedes(target, string(vstar.TodoCompleted), t1)
	if err != nil {
		return fmt.Errorf("cross_component sup: %w", err)
	}
	sup.Set(vstar.Property{Name: "DESCRIPTION", Value: "Resolved by Sami's review"})
	// Re-stamp the hash since we mutated the journal after Supersedes.
	hashing.SetXVSTAR(&sup)

	cal := vstar.Calendar{ProdID: "-//V*//Supersession-CrossComponent//EN"}
	cal.Append(target)
	cal.Append(sup)
	return writeCalendar(dir, "cross_component", cal)
}

// writeCorruptMutated: VTODO with X-VSTAR-HASH set, then SUMMARY
// mutated inline so the hash no longer matches. INTENTIONALLY
// CORRUPT — negative test fodder.
func writeCorruptMutated(dir string, t0 time.Time) error {
	target := newTodo("todo-corrupt", "Original summary", t0)
	hashing.SetXVSTAR(&target)
	// Hash now stamped. Mutate SUMMARY without re-stamping → the
	// stored hash no longer matches the canonical form.
	target.Set(vstar.Property{Name: "SUMMARY", Value: "MUTATED summary (hash NOT refreshed)"})
	cal := vstar.Calendar{ProdID: "-//V*//Supersession-Corrupt//EN"}
	cal.Append(target)
	return writeCalendar(dir, "corrupt_mutated", cal)
}

// newTodo builds a deterministic VTODO with UID/DTSTAMP/SUMMARY.
// Uses helpers.NewTodo for the constructor discipline (hash
// stamped at end), then strips DUE since this generator only
// wants UID + DTSTAMP + SUMMARY for the supersession story.
func newTodo(uid, summary string, t time.Time) vstar.Component {
	c, err := helpers.NewTodo(uid, t.Add(24*time.Hour))
	if err != nil {
		panic(fmt.Sprintf("newTodo(%s): %v", uid, err))
	}
	c.Remove("DUE")
	// Override DTSTAMP (helpers.NewTodo uses time.Now); we need
	// determinism.
	c.Set(vstar.Property{Name: "DTSTAMP", Value: vstar.FormatTime(t)})
	c.Set(vstar.Property{Name: "SUMMARY", Value: summary})
	return c
}

// writeCalendar encodes cal via the Go RFC 5545 codec, normalizes
// CRLF→LF for on-disk LF convention, then writes <stem>.ics under
// dir. The generator does NOT touch <stem>.notes.md or README.md;
// those are hand-authored to capture intent.
func writeCalendar(dir, stem string, cal vstar.Calendar) error {
	var buf bytes.Buffer
	if err := rfc5545.Encode(&buf, cal); err != nil {
		return fmt.Errorf("encode %s: %w", stem, err)
	}
	icsLF := crlfToLF(buf.Bytes())
	return os.WriteFile(filepath.Join(dir, stem+".ics"), icsLF, 0o644)
}

// crlfToLF strips \r before \n. Same convention as fixtures-verify.
func crlfToLF(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] == '\r' && i+1 < len(b) && b[i+1] == '\n' {
			continue
		}
		out = append(out, b[i])
	}
	return out
}

// repoRoot finds the repo root by walking up looking for testdata/.
func repoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		st, err := os.Stat(filepath.Join(dir, "testdata", "rfc5545"))
		if err == nil && st.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find testdata/ above %s", cwd)
		}
		dir = parent
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "fixtures-gen-supersession:", err)
	os.Exit(1)
}
