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
			return runAnalyzers(pass, analyzers)
		},
	}
}

type ignoreRange struct {
	start, end token.Pos
	analyzers  map[string]struct{}
}

func runAnalyzers(
	pass *analysis.Pass, analyzers []*analysis.Analyzer,
) (any, error) {
	ranges := ignoreRanges(pass)
	report := pass.Report
	defer func() { pass.Report = report }()
	for _, a := range analyzers {
		var diags []analysis.Diagnostic
		pass.Report = func(d analysis.Diagnostic) {
			diags = append(diags, d)
		}
		pass.Analyzer = a
		if _, err := a.Run(pass); err != nil {
			return nil, err
		}
		for d := range filter(diags, ranges, a.Name, pass.Fset) {
			report(d)
		}
	}
	return nil, nil
}

func ignoreRanges(pass *analysis.Pass) (ranges []ignoreRange) {
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
			analyzers, in := parseIgnore(c.Text)
			if analyzers == nil {
				continue
			}
			if in {
				ranges = append(ranges, ignoreCommentLine(c, fset, analyzers))
			} else {
				ranges = append(
					ranges,
					newIgnoreRange(c, analyzers, file, cmap, cg, fset),
				)
			}
		}
	}
	return
}

func newIgnoreRange(
	comment *ast.Comment, analyzers map[string]struct{},
	file *ast.File, cmap ast.CommentMap, group *ast.CommentGroup,
	fset *token.FileSet,
) ignoreRange {
	if comment.Pos() < file.Package {
		// Ignore analyzers on this entire file.
		return ignoreRange{file.Pos(), file.End(), analyzers}
	}
	node := findCommentNode(comment, cmap)
	if node != nil && group != nil {
		if fset.Position(node.Pos()).Line == fset.Position(group.Pos()).Line {
			// This is an inline comment. Ignore its associated line.
			return ignoreCommentLine(comment, fset, analyzers)
		} else {
			// Ignore analyzers on this comment group and its associated node.
			return ignoreRange{
				min(group.Pos(), node.Pos()),
				max(group.End(), node.End()),
				analyzers,
			}
		}
	} else if node != nil {
		// Ignore analyzers on this node.
		return ignoreRange{node.Pos(), node.End(), analyzers}
	} else if group != nil {
		// Ignore analyzers on this comment group.
		return ignoreRange{group.Pos(), group.End(), analyzers}
	} else {
		// Ignore analyzers on the current line.
		return ignoreCommentLine(comment, fset, analyzers)
	}
}

func ignoreCommentLine(
	comment *ast.Comment, fset *token.FileSet, analyzers map[string]struct{},
) ignoreRange {
	file := fset.File(comment.Pos())
	if file == nil {
		return ignoreRange{comment.Pos(), comment.End(), analyzers}
	}
	line := fset.Position(comment.Pos()).Line
	startPos := file.LineStart(line)

	var endPos token.Pos
	if line < file.LineCount() {
		endPos = file.LineStart(line+1) - 1
	} else {
		endPos = token.Pos(file.Base() + file.Size())
	}

	return ignoreRange{startPos, endPos, analyzers}
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

func filter(
	diags []analysis.Diagnostic, ranges []ignoreRange,
	analyzerName string, fset *token.FileSet,
) func(func(analysis.Diagnostic) bool) {
	return func(yield func(analysis.Diagnostic) bool) {
		for _, d := range diags {
			if !ignoreDiagnostic(&d, ranges, analyzerName, fset) {
				if !yield(d) {
					return
				}
			}
		}
	}
}

func ignoreDiagnostic(
	diag *analysis.Diagnostic, ranges []ignoreRange,
	analyzerName string, fset *token.FileSet,
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
		if analyzersContains(r.analyzers, analyzerName) {
			if diag.Pos >= r.start && diag.Pos <= r.end {
				return true
			}
		}
	}
	return false
}

func analyzersContains(
	analyzers map[string]struct{}, analyzerName string,
) bool {
	if _, ok := analyzers["all"]; ok {
		return true
	}
	_, exists := analyzers[analyzerName]
	return exists
}

func parseIgnore(text string) (analyzers map[string]struct{}, in bool) {
	re := regexp.MustCompile(`//ignore(?::([^/\s]+))?`)
	match := re.FindStringSubmatch(text)
	if match == nil {
		return
	}
	in = !strings.HasPrefix(text, match[0])
	analyzers = make(map[string]struct{})
	if match[1] == "" {
		analyzers["all"] = struct{}{}
		return
	}
	analyzerList := match[1]
	if analyzerList == "all" {
		analyzers["all"] = struct{}{}
		return
	}
	for name := range strings.SplitSeq(analyzerList, ",") {
		if name = strings.TrimSpace(name); name != "" {
			analyzers[name] = struct{}{}
		}
	}
	return
}
