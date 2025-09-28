package checker

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

// dependentAnalyzer depends on publicNames and uses its results
var dependentAnalyzer = &analysis.Analyzer{
	Name:     "dependent",
	Doc:      "analyzer that depends on publicnames",
	Requires: []*analysis.Analyzer{publicNames},
	Run: func(pass *analysis.Pass) (any, error) {
		// This should have access to publicNames results via pass.ResultOf
		if result, ok := pass.ResultOf[publicNames]; ok {
			// Use the result somehow - report on the first file's package
			// declaration
			_ = result
			if len(pass.Files) > 0 {
				pass.Reportf(
					pass.Files[0].Package, "dependent analyzer ran",
				)
			}
		} else {
			if len(pass.Files) > 0 {
				pass.Reportf(
					pass.Files[0].Package,
					"dependent analyzer missing required result",
				)
			}
		}
		return "dependent result", nil
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
		input         string
		wantAnalyzers map[string]struct{}
		wantIn        bool
	}{
		{"//ignore", map[string]struct{}{"all": {}}, false},
		{"//ignore:all", map[string]struct{}{"all": {}}, false},
		{"//ignore:test1", map[string]struct{}{"test1": {}}, false},
		{"//ignore:test1,test2", map[string]struct{}{
			"test1": {}, "test2": {},
		}, false},
		{"//ignore:all // This is a comment",
			map[string]struct{}{"all": {}}, false},
		{"//ignore:test1 // Explanatory comment",
			map[string]struct{}{"test1": {}}, false},
		{"//ignore:test1,test2 // Multiple analyzers",
			map[string]struct{}{"test1": {}, "test2": {}}, false},
		{"// not ignore", nil, false},
		{"//other", nil, false},
		{"// This is a comment. //ignore:all",
			map[string]struct{}{"all": {}}, true},
		{"// This is a comment. //ignore:all // And another comment.",
			map[string]struct{}{"all": {}}, true},
	}

	for _, tt := range tests {
		gotAnalyzers, gotIn := parseIgnore(tt.input)
		if tt.wantAnalyzers == nil && gotAnalyzers != nil {
			t.Errorf("parseNolint(%q) = %v, want nil",
				tt.input, gotAnalyzers,
			)
		} else if gotAnalyzers == nil && tt.wantAnalyzers != nil {
			t.Errorf("parseNolint(%q) = nil, want %v",
				tt.input, tt.wantAnalyzers,
			)
		} else if !cmp.Equal(gotAnalyzers, tt.wantAnalyzers) {
			t.Errorf("analyzers -want +got\n%s",
				cmp.Diff(gotAnalyzers, tt.wantAnalyzers),
			)
		} else if gotIn != tt.wantIn {
			t.Errorf("parseNoLint(%q) in = %t, want %t",
				tt.input, gotIn, tt.wantIn,
			)
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

func TestAnalyzerDependencies(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(),
		NewAnalyzer(dependentAnalyzer), "dependency")
}

func TestCycleDetection(t *testing.T) {
	analyzerA := &analysis.Analyzer{
		Name:     "analyzerA",
		Doc:      "analyzer A",
		Requires: []*analysis.Analyzer{}, // Will be set to analyzerB
		Run: func(pass *analysis.Pass) (any, error) {
			return nil, nil
		},
	}
	analyzerB := &analysis.Analyzer{
		Name:     "analyzerB",
		Doc:      "analyzer B",
		Requires: []*analysis.Analyzer{analyzerA},
		Run: func(pass *analysis.Pass) (any, error) {
			return nil, nil
		},
	}
	analyzerA.Requires = []*analysis.Analyzer{analyzerB} // Create cycle.

	err := detectCycles([]*analysis.Analyzer{analyzerA})

	if err == nil {
		t.Error("Expected cycle detection error, but got none")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("Expected circular dependency error, got: %v", err)
	}
}
