package checker

import (
	"fmt"
	"strings"

	"golang.org/x/tools/go/analysis"
)

func detectCycles(analyzers []*analysis.Analyzer) error {
	visited := make(map[*analysis.Analyzer]struct{})
	recStack := make(map[*analysis.Analyzer]struct{})

	var visit func(*analysis.Analyzer, []string) error
	visit = func(a *analysis.Analyzer, path []string) error {
		if _, inStack := recStack[a]; inStack {
			cycle := append(path, a.Name)
			return fmt.Errorf(
				"circular dependency detected: %s",
				strings.Join(cycle, " -> "),
			)
		}
		if _, wasVisited := visited[a]; wasVisited {
			return nil
		}
		visited[a] = struct{}{}
		recStack[a] = struct{}{}
		newPath := append(path, a.Name)
		for _, req := range a.Requires {
			if err := visit(req, newPath); err != nil {
				return err
			}
		}
		delete(recStack, a)
		return nil
	}

	seen := make(map[*analysis.Analyzer]struct{})
	var collectAll func(*analysis.Analyzer)
	collectAll = func(a *analysis.Analyzer) {
		if _, wasSeen := seen[a]; wasSeen {
			return
		}
		seen[a] = struct{}{}
		for _, req := range a.Requires {
			collectAll(req)
		}
	}
	for _, a := range analyzers {
		collectAll(a)
	}
	for a := range seen {
		if err := visit(a, nil); err != nil {
			return err
		}
	}
	return nil
}

