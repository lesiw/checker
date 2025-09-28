package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"slices"
	"strings"
	"sync"

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
	// Most of this structure is borrowed from unitchecker.

	if err := detectCycles(analyzers); err != nil {
		return nil, err
	}
	ranges := ignoreRanges(pass)

	type action struct {
		once   sync.Once
		result any
		err    error
		diags  []analysis.Diagnostic
	}
	actions := make(map[*analysis.Analyzer]*action)

	// Initialize actions for all analyzers (including dependencies).
	var initActions func(a *analysis.Analyzer)
	initActions = func(a *analysis.Analyzer) {
		if _, ok := actions[a]; !ok {
			actions[a] = new(action)
			// Recursively initialize dependencies
			for _, req := range a.Requires {
				initActions(req)
			}
		}
	}
	for _, a := range analyzers {
		initActions(a)
	}

	// Execute analyzers on-demand with dependency resolution.
	var exec func(a *analysis.Analyzer) *action
	var execAll func(analyzers []*analysis.Analyzer)
	exec = func(a *analysis.Analyzer) *action {
		act := actions[a]
		act.once.Do(func() {
			execAll(a.Requires) // Prefetch dependencies concurrently.

			// The inputs to this analysis are the
			// results of its prerequisites.
			inputs := make(map[*analysis.Analyzer]any)
			var failed []string
			for _, req := range a.Requires {
				reqact := exec(req)
				if reqact.err != nil {
					failed = append(failed, req.String())
					continue
				}
				inputs[req] = reqact.result
			}
			if failed != nil {
				slices.Sort(failed)
				act.err = fmt.Errorf("failed prerequisites: %s",
					strings.Join(failed, ", "),
				)
				return
			}

			// Create a new pass for this analyzer.
			analyzerPass := *pass
			analyzerPass.Analyzer = a
			analyzerPass.ResultOf = inputs

			// Facts handling: preserve the original fact methods.
			// These will work correctly across analyzer boundaries
			// since they operate on the same underlying pass data.
			analyzerPass.ImportObjectFact = pass.ImportObjectFact
			analyzerPass.ExportObjectFact = pass.ExportObjectFact
			analyzerPass.AllObjectFacts = pass.AllObjectFacts
			analyzerPass.ImportPackageFact = pass.ImportPackageFact
			analyzerPass.ExportPackageFact = pass.ExportPackageFact
			analyzerPass.AllPackageFacts = pass.AllPackageFacts

			analyzerPass.Report = func(d analysis.Diagnostic) {
				act.diags = append(act.diags, d)
			}

			act.result, act.err = a.Run(&analyzerPass)
		})
		return act
	}
	execAll = func(analyzers []*analysis.Analyzer) {
		var wg sync.WaitGroup
		for _, a := range analyzers {
			wg.Add(1)
			go func(a *analysis.Analyzer) {
				_ = exec(a)
				wg.Done()
			}(a)
		}
		wg.Wait()
	}
	execAll(analyzers)

	// Check for errors from root analyzers.
	for _, a := range analyzers {
		act := actions[a]
		if act.err != nil {
			return nil, act.err
		}
	}
	for analyzer, act := range actions {
		if act.err != nil {
			continue
		}
		for d := range filter(act.diags, ranges, analyzer.Name, pass.Fset) {
			pass.Report(d)
		}
	}
	return nil, nil
}

func ignoreRanges(pass *analysis.Pass) (ranges []ignoreRange) {
	for _, file := range pass.Files {
		ranges = append(ranges, fileIgnores(file, pass)...)
	}
	slices.SortFunc(ranges, func(a, b ignoreRange) int {
		return int(a.start - b.start)
	})
	return
}

func fileIgnores(file *ast.File, pass *analysis.Pass) (ranges []ignoreRange) {
	cmap := ast.NewCommentMap(pass.Fset, file, file.Comments)
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			analyzers, in := parseIgnore(c.Text)
			if analyzers == nil {
				continue
			}
			if in {
				ranges = append(ranges, ignoreCommentLine(c, pass, analyzers))
			} else {
				ranges = append(
					ranges,
					newIgnoreRange(c, analyzers, file, cmap, cg, pass),
				)
			}
		}
	}
	return
}

func newIgnoreRange(
	comment *ast.Comment, analyzers map[string]struct{},
	file *ast.File, cmap ast.CommentMap, group *ast.CommentGroup,
	pass *analysis.Pass,
) ignoreRange {
	if comment.Pos() < file.Package {
		// Ignore analyzers on this entire file.
		return ignoreRange{file.Pos(), file.End(), analyzers}
	}
	node := findCommentNode(comment, cmap)
	if node != nil && group != nil {
		if isInlineComment(comment, pass) {
			// Ignore this comment's associated line.
			return ignoreCommentLine(comment, pass, analyzers)
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
		return ignoreCommentLine(comment, pass, analyzers)
	}
}

var commentRe = regexp.MustCompile(`^\s*\/$`)

func isInlineComment(comment *ast.Comment, pass *analysis.Pass) bool {
	file := pass.Fset.File(comment.Pos())
	buf, err := pass.ReadFile(file.Name())
	if err != nil {
		return false
	}
	pos := file.Offset(file.LineStart(pass.Fset.Position(comment.Pos()).Line))
	end := file.Offset(comment.Pos() + 1)
	return !commentRe.Match(buf[pos:end])
}

func ignoreCommentLine(
	comment *ast.Comment, pass *analysis.Pass, analyzers map[string]struct{},
) ignoreRange {
	start, end := linePos(pass, comment.Pos())
	if start == end {
		return ignoreRange{comment.Pos(), comment.End(), analyzers}
	}
	return ignoreRange{start, end, analyzers}
}

func linePos(pass *analysis.Pass, p token.Pos) (pos, end token.Pos) {
	file := pass.Fset.File(p)
	if file == nil {
		return p, p
	}
	line := pass.Fset.Position(p).Line
	pos = file.LineStart(line)
	if line < file.LineCount() {
		end = file.LineStart(line+1) - 1
	} else {
		end = token.Pos(file.Base() + file.Size())
	}
	return
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
