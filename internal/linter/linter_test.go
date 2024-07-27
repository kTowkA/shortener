package linter

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(t *testing.T) {
	// функция analysistest.Run применяет тестируемый анализатор ErrCheckAnalyzer
	// к пакетам из папки testdata и проверяет ожидания
	// ./... — проверка всех поддиректорий в testdata
	analysistest.Run(t, analysistest.TestData(), New(nil), "./pkg_01")
	analysistest.Run(t, analysistest.TestData(), New([]NotRecomended{{Pkg: "", FromFunction: "yrun", Function: NotRecomendedFunc{Pkg: "", Name: "zrun"}}}), "./pkg_03")
}
