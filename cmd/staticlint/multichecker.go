package main

import (
	"strings"

	// "github.com/securego/gosec"
	"github.com/Antonboom/testifylint/analyzer"
	"github.com/kTowkA/shortener/internal/linter"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// ++ стандартных статических анализаторов пакета golang.org/x/tools/go/analysis/passes;
// ++ всех анализаторов класса SA пакета staticcheck.io;
// ++ не менее одного анализатора остальных классов пакета staticcheck.io;
// ++ двух или более любых публичных анализаторов на ваш выбор.
// ++ Напишите и добавьте в multichecker собственный анализатор, запрещающий использовать прямой вызов os.Exit в функции main пакета main. При необходимости перепишите код своего проекта так, чтобы он удовлетворял данному анализатору.

func main() {
	// берем с запасом
	analyzers := make([]*analysis.Analyzer, 0, 150)
	analyzers = addAnalyzersStaticcheck(analyzers, staticcheck.Analyzers, []string{"SA"})
	analyzers = addAnalyzersStaticcheck(analyzers, stylecheck.Analyzers, []string{"ST1022", "ST1015"})
	analyzers = addAnalyzersStaticcheck(analyzers, simple.Analyzers, []string{"S1001", "S1040"})
	analyzers = addAnalyzersStaticcheck(analyzers, quickfix.Analyzers, []string{"QF1009", "QF1004"})

	analyzers = addPasses(analyzers)
	analyzers = addCustomAnalyzers(analyzers)

	// ну и мою проверку добавляем (custom linter 001)
	analyzers = append(analyzers, linter.Analyzers["CL001"])

	multichecker.Main(
		analyzers...,
	)
}

func addAnalyzersStaticcheck(check []*analysis.Analyzer, analyzers []*lint.Analyzer, prefixs []string) []*analysis.Analyzer {
	for _, v := range analyzers {
		for _, prefix := range prefixs {
			if strings.HasPrefix(v.Analyzer.Name, prefix) {
				check = append(check, v.Analyzer)
				break
			}
		}
	}
	return check
}

// golang.org/x/tools/go/analysis/passes
func addPasses(check []*analysis.Analyzer) []*analysis.Analyzer {
	check = append(check, printf.Analyzer)
	check = append(check, slog.Analyzer)
	check = append(check, unreachable.Analyzer)
	check = append(check, loopclosure.Analyzer)
	return check
}

// custom
func addCustomAnalyzers(check []*analysis.Analyzer) []*analysis.Analyzer {
	check = append(check, analyzer.New())
	check = append(check, bodyclose.Analyzer)
	return check
}
