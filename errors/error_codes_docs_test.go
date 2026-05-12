package errors

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestErrorCodes_DocumentationContainsEveryBuiltInCode(t *testing.T) {
	codes := parseDocumentedSourceCodes(t)
	codes = append(codes,
		"string.email.invalid",
		"string.ulid.invalid",
		"string.uuid.invalid",
	)
	sort.Strings(codes)

	doc, err := os.ReadFile("../docs/error-codes.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(doc)

	for _, code := range codes {
		if !strings.Contains(text, "`"+code+"`") {
			t.Fatalf("docs/error-codes.md is missing code %q", code)
		}
	}
}

func parseDocumentedSourceCodes(t *testing.T) []string {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "codes.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	var codes []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range valueSpec.Names {
				if !strings.HasPrefix(name.Name, "Code") || i >= len(valueSpec.Values) {
					continue
				}
				lit, ok := valueSpec.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				code, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Fatal(err)
				}
				codes = append(codes, code)
			}
		}
	}
	return codes
}
