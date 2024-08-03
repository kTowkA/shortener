package linter

import (
	"golang.org/x/tools/go/analysis"
)

// Analyzers делаем переменные по аналогии с пакетами honnef.co
var Analyzers = map[string]*analysis.Analyzer{
	"CL001": func() *analysis.Analyzer {
		return NewAnalyzerNotRecommended(defaultsNotRecomendedFuncs...).Analyzer()
	}(),
}
