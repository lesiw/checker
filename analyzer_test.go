package checker

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

type identCheck func(pass *analysis.Pass, ident *ast.Ident)

func identAnalyze(checker identCheck) func(*analysis.Pass) (any, error) {
	return func(pass *analysis.Pass) (any, error) {
		for _, file := range pass.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.FuncDecl:
					checkFuncDecl(node, pass, checker)
				case *ast.GenDecl:
					checkGenDecl(node, pass, checker)
				case *ast.AssignStmt:
					checkAssignStmt(node, pass, checker)
				}
				return true
			})
		}
		return nil, nil
	}
}

func checkFuncDecl(
	node *ast.FuncDecl, pass *analysis.Pass, checker identCheck,
) {
	if node.Name != nil && node.Name.Name != "_" {
		checker(pass, node.Name)
	}
}

func checkGenDecl(node *ast.GenDecl, pass *analysis.Pass, checker identCheck) {
	for _, spec := range node.Specs {
		if tspec, ok := spec.(*ast.TypeSpec); ok && tspec.Name.Name != "_" {
			checker(pass, tspec.Name)
		} else if vspec, ok := spec.(*ast.ValueSpec); ok {
			for _, name := range vspec.Names {
				if name.Name != "_" {
					checker(pass, name)
				}
			}
		}
	}
}

func checkAssignStmt(
	node *ast.AssignStmt, pass *analysis.Pass, checker identCheck,
) {
	if node.Tok == token.DEFINE {
		for _, expr := range node.Lhs {
			if ident, ok := expr.(*ast.Ident); ok && ident.Name != "_" {
				checker(pass, ident)
			}
		}
	}
}

var publicNames = &analysis.Analyzer{
	Name: "publicnames",
	Doc:  "reports on public names",
	Run: identAnalyze(func(pass *analysis.Pass, ident *ast.Ident) {
		if ident.IsExported() {
			pass.Reportf(ident.Pos(), "%s is public", ident.Name)
		}
	}),
}

var numberedNames = &analysis.Analyzer{
	Name: "numberednames",
	Doc:  "reports on names with numbers in them",
	Run: identAnalyze(func(pass *analysis.Pass, ident *ast.Ident) {
		for _, r := range ident.Name {
			if r >= '0' && r <= '9' {
				pass.Reportf(ident.Pos(), "%s has numbers", ident.Name)
				break
			}
		}
	}),
}

var shortNames = &analysis.Analyzer{
	Name: "shortnames",
	Doc:  "reports on single-letter names",
	Run: identAnalyze(func(pass *analysis.Pass, ident *ast.Ident) {
		if len(ident.Name) == 1 {
			pass.Reportf(ident.Pos(), "%s is single letter", ident.Name)
		}
	}),
}

var underscoreNames = &analysis.Analyzer{
	Name: "underscorenames",
	Doc:  "reports on names containing underscores",
	Run: identAnalyze(func(pass *analysis.Pass, ident *ast.Ident) {
		if strings.Contains(ident.Name, "_") {
			pass.Reportf(ident.Pos(), "%s has underscore", ident.Name)
		}
	}),
}

var todoAnalyzer = &analysis.Analyzer{
	Name: "todocomments",
	Doc:  "reports on TODO comments",
	Run: func(pass *analysis.Pass) (any, error) {
		for _, file := range pass.Files {
			for _, cg := range file.Comments {
				for _, c := range cg.List {
					if strings.Contains(c.Text, "TODO") {
						pass.Reportf(c.Pos(), "comment contains TODO")
					}
				}
			}
		}
		return nil, nil
	},
}

func TestZeroAnalyzers(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(), "empty")
}

func TestOneAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(publicNames), "single")
}

func TestMultipleAnalyzers(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(publicNames, numberedNames), "basic")
}

func TestNolint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(publicNames, numberedNames), "ignore")
}

func TestMultipleNolint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(publicNames, numberedNames), "multiple")
}

func TestInlineNolint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(publicNames, numberedNames), "inline")
}

func TestBlockNolint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(publicNames, numberedNames), "block")
}

func TestFileLevelNolint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(publicNames, numberedNames), "filelevel")
}

func TestParseNolint(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]struct{}
	}{
		{"//ignore", map[string]struct{}{"all": {}}},
		{"//ignore:all", map[string]struct{}{"all": {}}},
		{"//ignore:test1", map[string]struct{}{"test1": {}}},
		{"//ignore:test1,test2", map[string]struct{}{
			"test1": {}, "test2": {},
		}},
		{"//ignore:all // This is a comment",
			map[string]struct{}{"all": {}}},
		{"//ignore:test1 // Explanatory comment",
			map[string]struct{}{"test1": {}}},
		{"//ignore:test1,test2 // Multiple analyzers",
			map[string]struct{}{"test1": {}, "test2": {}}},
		{"// not ignore", nil},
		{"//other", nil},
		{"// This is a comment. //ignore:all",
			map[string]struct{}{"all": {}}},
		{"// This is a comment. //ignore:all // And another comment.",
			map[string]struct{}{"all": {}}},
	}

	for _, tt := range tests {
		got := parseIgnore(tt.input)
		if tt.want == nil {
			if got != nil {
				t.Errorf("parseNolint(%q) = %v, want nil", tt.input, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("parseNolint(%q) = nil, want %v", tt.input, tt.want)
			continue
		}
		if len(got) != len(tt.want) {
			t.Errorf("parseNolint(%q) length = %d, want %d",
				tt.input, len(got), len(tt.want))
			continue
		}
		for k := range tt.want {
			if _, exists := got[k]; !exists {
				t.Errorf("parseNolint(%q)[%q] missing", tt.input, k)
			}
		}
	}
}

func TestBlockLevelNolint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(shortNames, underscoreNames), "blocklevel")
}

func TestMixedBlockNolint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), NewAnalyzer(
		publicNames, numberedNames, shortNames, underscoreNames),
		"mixedblock")
}

func TestCommentNolint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(todoAnalyzer), "commentignore")
}
