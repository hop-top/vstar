// SPDX-License-Identifier: Apache-2.0

package validate_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDiagnosticCodes_AllDocumented walks every .go source file in
// the validate package and asserts that every exported "Code…"
// string constant appears verbatim in docs/validate-codes.md.
//
// This is the stability gate promised by the package: codes are
// part of the public surface. If a developer adds a new VSnnn
// without adding a doc entry, this test fails.
//
// The docs file lives at the repository root (../docs/), because
// the catalog is a user-facing artifact alongside ADRs.
func TestDiagnosticCodes_AllDocumented(t *testing.T) {
	srcDir := "."
	docsPath := filepath.Join("..", "docs", "validate-codes.md")

	docsBytes, err := os.ReadFile(docsPath)
	if err != nil {
		t.Fatalf("read %s: %v", docsPath, err)
	}
	docs := string(docsBytes)

	codes := collectCodeConstants(t, srcDir)
	if len(codes) == 0 {
		t.Fatalf("no Code* constants discovered in %s; test cannot meaningfully assert", srcDir)
	}

	for name, value := range codes {
		if !strings.Contains(docs, value) {
			t.Errorf("constant %s = %q is not documented in %s", name, value, docsPath)
		}
	}
}

// collectCodeConstants parses every .go file (excluding _test.go)
// in dir and returns the set of name→value pairs for top-level
// untyped string constants whose name starts with "Code".
func collectCodeConstants(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, n := range vs.Names {
					if !strings.HasPrefix(n.Name, "Code") {
						continue
					}
					if i >= len(vs.Values) {
						continue
					}
					lit, ok := vs.Values[i].(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					// Strip surrounding quotes from the literal.
					value := strings.Trim(lit.Value, "\"`")
					out[n.Name] = value
				}
			}
		}
	}
	return out
}
