package linter

import (
	"golang.org/x/tools/go/analysis"
)

var Analyzers = map[string]*analysis.Analyzer{
	"CL001": func() *analysis.Analyzer {
		return NewAnalyzerNotRecommended(defaultsNotRecomendedFuncs...).Analyzer()
	}(),
}
