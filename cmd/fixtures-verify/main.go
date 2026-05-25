// SPDX-License-Identifier: Apache-2.0

// Command fixtures-verify regenerates `<name>.canonical` and
// `<name>.hash` siblings for every parseable fixture in
// vstar/testdata/, plus syncs the canonical fuzz seed corpus into
// the go-fuzz convention dirs at go/codec/<rfc>/testdata/fuzz/.
//
// After regeneration, the caller (typically `task fixtures-verify`
// or CI) runs `git diff --exit-code testdata/ go/codec/*/testdata/`
// to detect drift. Any non-empty diff means either the
// implementation drifted (fix the implementation) or the fixture's
// canonical/hash changed deliberately (commit the regenerated
// siblings as part of the same PR).
//
// Walks:
//
//   - testdata/rfc5545/*.ics             → .canonical + .hash via Calendar
//   - testdata/rfc6350/*.vcf             → .canonical + .hash via Card
//   - testdata/supersession/*.ics        → .canonical + .hash via Calendar
//   - testdata/malformed/*               → no-op (these don't parse)
//   - testdata/fuzz-seed/rfc5545/*.bytes → copied into
//     go/codec/rfc5545/testdata/fuzz/
//     FuzzParse_RFC5545/seed_<stem>
//     (wrapped in go-fuzz format)
//   - testdata/fuzz-seed/rfc6350/*.bytes → same for rfc6350
//   - testdata/rrule/*/<name>.rrule      → asserts rrule.ParseRRule
//     classifies the value per the optional <name>.expect.json
//     sidecar (sentinel: "ErrMalformed" or "ErrUnsupportedRRule").
//     When no .expect.json is present, parsing must succeed.
//     For evaluator fixtures with <name>.next.json sidecars,
//     asserts rrule.NextOccurrence iteratively yields the listed
//     timestamps.
//
// Exits 0 on success; non-zero with a diagnostic on the first error.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	vstar "hop.top/vstar"
	"hop.top/vstar/canonical"
	"hop.top/vstar/codec/rfc5545"
	"hop.top/vstar/codec/rfc6350"
	"hop.top/vstar/hashing"
	"hop.top/vstar/rrule"
)

// repoRoot finds the repository root by walking up from the current
// working directory looking for a `testdata/` sibling. The intended
// invocation is `cd go && go run ./cmd/fixtures-verify`, so the
// search starts in `go/` and ascends.
func repoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	dir := cwd
	for {
		if st, err := os.Stat(filepath.Join(dir, "testdata", "rfc5545")); err == nil && st.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find testdata/rfc5545 walking up from %s", cwd)
		}
		dir = parent
	}
}

func main() {
	root, err := repoRoot()
	if err != nil {
		fail(err)
	}
	if err := regenerateCalendars(filepath.Join(root, "testdata", "rfc5545")); err != nil {
		fail(err)
	}
	if err := regenerateCalendars(filepath.Join(root, "testdata", "supersession")); err != nil {
		fail(err)
	}
	if err := regenerateCards(filepath.Join(root, "testdata", "rfc6350")); err != nil {
		fail(err)
	}
	if err := syncFuzzSeeds(root); err != nil {
		fail(err)
	}
	if err := verifyRRuleFixtures(filepath.Join(root, "testdata", "rrule")); err != nil {
		fail(err)
	}
	fmt.Println("fixtures-verify: regenerated all canonical/hash siblings + fuzz seeds; rrule fixtures verified")
}

// verifyRRuleFixtures walks every <name>.rrule under root and
// asserts the parser/evaluator output matches the optional
// sidecar contracts:
//
//   - <name>.expect.json — {"sentinel": "ErrMalformed" |
//     "ErrUnsupportedRRule"} — ValidateRRule must wrap that sentinel.
//     Absent → parsing must succeed.
//   - <name>.next.json — {"dtstart", "after", "expected": [...]}
//     — NextOccurrence iteratively must yield each `expected`
//     timestamp (using the previous result as `after`).
//
// All paths are resolved relative to `root`. Absent root is a
// silent no-op (lets the v0.1 verifier still run on a tree
// without the v0.2 rrule fixtures).
func verifyRRuleFixtures(root string) error {
	if _, err := os.Stat(root); errIsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".rrule") {
			return nil
		}
		return verifyOneRRuleFixture(path)
	})
}

func verifyOneRRuleFixture(rrulePath string) error {
	body, err := os.ReadFile(rrulePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", rrulePath, err)
	}
	value := strings.TrimRight(string(body), "\r\n")
	stem := strings.TrimSuffix(rrulePath, ".rrule")

	// Sentinel expectation.
	expectPath := stem + ".expect.json"
	if _, err := os.Stat(expectPath); err == nil {
		var spec struct {
			Sentinel string `json:"sentinel"`
		}
		raw, err := os.ReadFile(expectPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", expectPath, err)
		}
		if err := json.Unmarshal(raw, &spec); err != nil {
			return fmt.Errorf("parse %s: %w", expectPath, err)
		}
		err = rrule.ValidateRRule(value)
		if err == nil {
			return fmt.Errorf("%s: expected error %s, ValidateRRule succeeded", rrulePath, spec.Sentinel)
		}
		switch spec.Sentinel {
		case "ErrMalformed":
			if !errors.Is(err, vstar.ErrMalformed) {
				return fmt.Errorf("%s: expected ErrMalformed, got %w", rrulePath, err)
			}
		case "ErrUnsupportedRRule":
			if !errors.Is(err, rrule.ErrUnsupportedRRule) {
				return fmt.Errorf("%s: expected ErrUnsupportedRRule, got %w", rrulePath, err)
			}
		default:
			return fmt.Errorf("%s: unknown sentinel %q (want ErrMalformed or ErrUnsupportedRRule)", expectPath, spec.Sentinel)
		}
		return nil
	}

	// No sentinel sidecar → parse must succeed.
	rule, err := rrule.ParseRRule(value)
	if err != nil {
		return fmt.Errorf("%s: ParseRRule failed: %w", rrulePath, err)
	}

	// Evaluator sidecar.
	nextPath := stem + ".next.json"
	if _, err := os.Stat(nextPath); errIsNotExist(err) {
		return nil
	}
	var spec struct {
		DTStart  string   `json:"dtstart"`
		After    string   `json:"after"`
		Expected []string `json:"expected"`
	}
	raw, err := os.ReadFile(nextPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", nextPath, err)
	}
	if err := json.Unmarshal(raw, &spec); err != nil {
		return fmt.Errorf("parse %s: %w", nextPath, err)
	}
	dt, ok := vstar.ParseTime(spec.DTStart)
	if !ok {
		return fmt.Errorf("%s: bad dtstart %q", nextPath, spec.DTStart)
	}
	after, ok := vstar.ParseTime(spec.After)
	if !ok {
		return fmt.Errorf("%s: bad after %q", nextPath, spec.After)
	}
	for i, want := range spec.Expected {
		got, ok, err := rrule.NextOccurrence(rule, dt, after)
		if err != nil {
			return fmt.Errorf("%s: NextOccurrence step %d failed: %w", nextPath, i, err)
		}
		if !ok {
			return fmt.Errorf("%s: NextOccurrence step %d returned (zero, false), want %s", nextPath, i, want)
		}
		wantT, ok := vstar.ParseTime(want)
		if !ok {
			return fmt.Errorf("%s: bad expected[%d] %q", nextPath, i, want)
		}
		if !got.Equal(wantT) {
			return fmt.Errorf("%s: NextOccurrence step %d: got %s, want %s",
				nextPath, i, got.Format(time.RFC3339), wantT.Format(time.RFC3339))
		}
		after = got
	}
	return nil
}

// regenerateCalendars walks dir for *.ics, parses each, writes
// .canonical and .hash siblings.
func regenerateCalendars(dir string) error {
	if _, err := os.Stat(dir); errIsNotExist(err) {
		return nil // optional dir; skip silently.
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".ics") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		stem := strings.TrimSuffix(full, ".ics")
		input, err := os.ReadFile(full)
		if err != nil {
			return fmt.Errorf("read %s: %w", full, err)
		}
		cal, err := rfc5545.Parse(bytes.NewReader(input))
		if err != nil {
			return fmt.Errorf("parse %s: %w", full, err)
		}
		canonBytes := crlfToLF(canonical.Calendar(cal))
		if err := os.WriteFile(stem+".canonical", canonBytes, 0o644); err != nil {
			return fmt.Errorf("write %s.canonical: %w", stem, err)
		}
		hash := hashing.Calendar(cal) + "\n"
		if err := os.WriteFile(stem+".hash", []byte(hash), 0o644); err != nil {
			return fmt.Errorf("write %s.hash: %w", stem, err)
		}
	}
	return nil
}

// crlfToLF strips \r before \n. The on-disk convention is LF for
// diff-friendliness; the canonical bytes per spec/03 are CRLF and
// callers expand back at comparison time (TestGoldenFiles in
// canonical_test.go does this; TestHashGoldens reads bytes that
// hashing.Calendar already produced from the parsed Calendar so
// no expansion is needed there).
func crlfToLF(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for i := range len(b) {
		if b[i] == '\r' && i+1 < len(b) && b[i+1] == '\n' {
			continue
		}
		out = append(out, b[i])
	}
	return out
}

// regenerateCards walks dir for *.vcf, parses each, writes
// .canonical and .hash siblings.
func regenerateCards(dir string) error {
	if _, err := os.Stat(dir); errIsNotExist(err) {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read %s: %w", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".vcf") {
			continue
		}
		full := filepath.Join(dir, e.Name())
		stem := strings.TrimSuffix(full, ".vcf")
		input, err := os.ReadFile(full)
		if err != nil {
			return fmt.Errorf("read %s: %w", full, err)
		}
		cards, err := rfc6350.New().Parse(bytes.NewReader(input))
		if err != nil {
			return fmt.Errorf("parse %s: %w", full, err)
		}
		if len(cards) != 1 {
			return fmt.Errorf("parse %s: expected exactly one card, got %d", full, len(cards))
		}
		canonBytes := crlfToLF([]byte(canonical.Card(cards[0])))
		if err := os.WriteFile(stem+".canonical", canonBytes, 0o644); err != nil {
			return fmt.Errorf("write %s.canonical: %w", stem, err)
		}
		hash := hashing.Card(cards[0]) + "\n"
		if err := os.WriteFile(stem+".hash", []byte(hash), 0o644); err != nil {
			return fmt.Errorf("write %s.hash: %w", stem, err)
		}
	}
	return nil
}

// syncFuzzSeeds copies every testdata/fuzz-seed/<rfc>/seed_*.bytes
// into go/codec/<rfc>/testdata/fuzz/FuzzParse_RFC<RFC>/seed_<stem>
// using the go-fuzz wrapper format.
func syncFuzzSeeds(root string) error {
	pairs := []struct {
		src     string
		dst     string
		fuzzFn  string
		isBytes bool
	}{
		{
			filepath.Join(root, "testdata", "fuzz-seed", "rfc5545"),
			filepath.Join(root, "codec", "rfc5545", "testdata", "fuzz", "FuzzParse_RFC5545"),
			"FuzzParse_RFC5545",
			false, // .ics seeds wrap as string for nicer fuzz panics.
		},
		{
			filepath.Join(root, "testdata", "fuzz-seed", "rfc6350"),
			filepath.Join(root, "codec", "rfc6350", "testdata", "fuzz", "FuzzParse_RFC6350"),
			"FuzzParse_RFC6350",
			true, // .vcf seeds wrap as []byte (existing convention).
		},
	}
	for _, p := range pairs {
		if _, err := os.Stat(p.src); errIsNotExist(err) {
			continue
		}
		if err := os.MkdirAll(p.dst, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", p.dst, err)
		}
		entries, err := os.ReadDir(p.src)
		if err != nil {
			return fmt.Errorf("read %s: %w", p.src, err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".bytes") {
				continue
			}
			stem := strings.TrimSuffix(e.Name(), ".bytes")
			input, err := os.ReadFile(filepath.Join(p.src, e.Name()))
			if err != nil {
				return fmt.Errorf("read %s: %w", e.Name(), err)
			}
			wrapped := wrapFuzzSeed(input, p.isBytes)
			out := filepath.Join(p.dst, stem)
			if err := os.WriteFile(out, []byte(wrapped), 0o644); err != nil {
				return fmt.Errorf("write %s: %w", out, err)
			}
		}
	}
	return nil
}

// wrapFuzzSeed renders the go-fuzz "v1" wrapper around raw input.
// When isBytes is true the value is encoded as []byte("..."); when
// false it is encoded as string("..."). Both forms use Go-style
// escaping so any byte is representable.
func wrapFuzzSeed(input []byte, isBytes bool) string {
	var b strings.Builder
	b.WriteString("go test fuzz v1\n")
	if isBytes {
		b.WriteString("[]byte(")
	} else {
		b.WriteString("string(")
	}
	b.WriteString(quoteGoString(input))
	b.WriteString(")\n")
	return b.String()
}

// quoteGoString returns input as a double-quoted Go string literal.
// Uses fmt's %q which produces the canonical Go escape form (\r,
// \n, \t, \xNN, etc.) deterministically.
func quoteGoString(input []byte) string {
	return fmt.Sprintf("%q", string(input))
}

func errIsNotExist(err error) bool {
	return err != nil && (os.IsNotExist(err) || errors_Is_PathError(err))
}

// errors_Is_PathError treats fs.PathError wrapping ENOENT the same
// as os.IsNotExist (which it should already, but defensively).
// Uses errors.As so wrapped errors are still recognized.
func errors_Is_PathError(err error) bool {
	var p *fs.PathError
	if !errors.As(err, &p) {
		return false
	}
	return os.IsNotExist(p.Err)
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "fixtures-verify: %v\n", err)
	os.Exit(1)
}
