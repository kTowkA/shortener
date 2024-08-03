package linter

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(t *testing.T) {
	// функция analysistest.Run применяет тестируемый анализатор ErrCheckAnalyzer
	// к пакетам из папки testdata и проверяет ожидания
	// ./... — проверка всех поддиректорий в testdata
	analysistest.Run(t, analysistest.TestData(), NewAnalyzerNotRecommended().Analyzer(), "./pkg_01")
	analysistest.Run(t, analysistest.TestData(), NewAnalyzerNotRecommended(NotRecommended{Pkg: "", FromFunction: "yrun", Function: NotRecommendedFunc{Pkg: "", Name: "zrun"}}).Analyzer(), "./pkg_03")
	// analysistest.Run(t, analysistest.TestData(), NewAnalyzerNotRecommended({{Pkg: "", FromFunction: "yrun", Function: NotRecommendedFunc{Pkg: "", Name: "zrun"}}}).Analyzer(), "./pkg_03")
}
