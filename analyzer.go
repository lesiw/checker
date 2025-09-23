package checker

import (
	"go/ast"
	"go/token"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// NewAnalyzer creates a new analyzer that runs multiple analyzers and filters
// diagnostics based on //ignore directives.
func NewAnalyzer(analyzers ...*analysis.Analyzer) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "checker",
		Doc: "runs multiple analyzers and filters diagnostics based on " +
			"//ignore directives",
		Run: func(pass *analysis.Pass) (any, error) {
			return runWithAnalyzers(pass, analyzers)
		},
	}
}

func runWithAnalyzers(
	pass *analysis.Pass, analyzers []*analysis.Analyzer,
) (any, error) {
	var diags []analysis.Diagnostic
	capturePass := &analysis.Pass{
		Analyzer:     pass.Analyzer,
		Fset:         pass.Fset,
		Files:        pass.Files,
		OtherFiles:   pass.OtherFiles,
		IgnoredFiles: pass.IgnoredFiles,
		Pkg:          pass.Pkg,
		TypesInfo:    pass.TypesInfo,
		TypesSizes:   pass.TypesSizes,
		ResultOf:     pass.ResultOf,
		Report:       func(d analysis.Diagnostic) { diags = append(diags, d) },
	}
	for _, a := range analyzers {
		if _, err := a.Run(capturePass); err != nil {
			return nil, err
		}
	}
	for _, d := range filter(diags, pass) {
		pass.Report(d)
	}
	return nil, nil
}

func filter(
	diags []analysis.Diagnostic, pass *analysis.Pass,
) (filtered []analysis.Diagnostic) {
	ranges := passIgnores(pass)
	for _, d := range diags {
		if !ignoreDiagnostic(&d, ranges, pass.Fset) {
			filtered = append(filtered, d)
		}
	}
	return
}

func passIgnores(pass *analysis.Pass) (ranges []ignoreRange) {
	for _, file := range pass.Files {
		ranges = append(ranges, fileIgnores(file, pass.Fset)...)
	}
	slices.SortFunc(ranges, func(a, b ignoreRange) int {
		return int(a.start - b.start)
	})
	return
}

func fileIgnores(file *ast.File, fset *token.FileSet) (ranges []ignoreRange) {
	cmap := ast.NewCommentMap(fset, file, file.Comments)
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if analyzers := ignore(c.Text); analyzers != nil {
				ranges = append(ranges, ignores(c, analyzers, file, cmap))
			}
		}
	}
	return
}

func ignores(
	comment *ast.Comment, analyzers map[string]struct{},
	file *ast.File, cmap ast.CommentMap,
) ignoreRange {
	if comment.Pos() < file.Package {
		// Ignore analyzers on this entire file.
		return ignoreRange{file.Pos(), file.End(), analyzers}
	}
	group := findCommentGroup(comment, file)
	node := findCommentNode(comment, cmap)
	if node != nil && group != nil {
		// Ignore analyzers on this comment group and its associated node.
		return ignoreRange{
			min(group.Pos(), node.Pos()),
			max(group.End(), node.End()),
			analyzers,
		}
	} else if node != nil {
		// Ignore analyzers on this node.
		return ignoreRange{node.Pos(), node.End(), analyzers}
	} else if group != nil {
		// Ignore analyzers on this comment group.
		return ignoreRange{group.Pos(), group.End(), analyzers}
	} else {
		// Ignore analyzers on the comment itself.
		return ignoreRange{comment.Pos(), comment.End(), analyzers}
	}
}

func findCommentNode(comment *ast.Comment, cmap ast.CommentMap) ast.Node {
	for node, commentGroups := range cmap {
		for _, cg := range commentGroups {
			if slices.Contains(cg.List, comment) {
				return node
			}
		}
	}
	return nil
}

func findCommentGroup(comment *ast.Comment, file *ast.File) *ast.CommentGroup {
	for _, cg := range file.Comments {
		if slices.Contains(cg.List, comment) {
			return cg
		}
	}
	return nil
}

func ignoreDiagnostic(
	diag *analysis.Diagnostic, ranges []ignoreRange, fset *token.FileSet,
) bool {
	if !diag.Pos.IsValid() {
		return false
	}
	diagPos := fset.Position(diag.Pos)

	for _, r := range ranges {
		if !r.start.IsValid() {
			continue
		}
		if fset.Position(r.start).Filename != diagPos.Filename {
			continue
		}
		if analyzersContains(r.analyzers, diag) {
			if diag.Pos >= r.start && diag.Pos <= r.end {
				return true
			}
		}
	}
	return false
}

func analyzersContains(
	analyzers map[string]struct{}, diag *analysis.Diagnostic,
) bool {
	if _, ok := analyzers["all"]; ok {
		return true
	}
	name := diag.Category
	if name == "" && len(diag.Message) > 0 {
		if parts := strings.SplitN(diag.Message, ":", 2); len(parts) > 0 {
			name = strings.TrimSpace(parts[0])
		}
	}
	_, exists := analyzers[name]
	return exists
}

type ignoreRange struct {
	start, end token.Pos
	analyzers  map[string]struct{}
}

func ignore(text string) map[string]struct{} {
	re := regexp.MustCompile(`//ignore(?::([^/\s]+))?`)
	match := re.FindStringSubmatch(text)
	if match == nil {
		return nil
	}
	result := make(map[string]struct{})
	if match[1] == "" {
		result["all"] = struct{}{}
		return result
	}
	analyzerList := match[1]
	if analyzerList == "all" {
		result["all"] = struct{}{}
		return result
	}
	for _, name := range strings.Split(analyzerList, ",") {
		if name = strings.TrimSpace(name); name != "" {
			result[name] = struct{}{}
		}
	}
	return result
}
