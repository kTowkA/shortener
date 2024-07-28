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
	mychecks := make([]*analysis.Analyzer, 0, 150)

	// SA
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	// QF

	qf := []string{"QF1009", "QF1004"}
	for _, v := range quickfix.Analyzers {
		for _, needA := range qf {
			if needA == v.Analyzer.Name {
				mychecks = append(mychecks, v.Analyzer)
				break
			}
		}
	}

	// S1
	s1 := []string{"S1001", "S1040"}
	for _, v := range simple.Analyzers {
		for _, needA := range s1 {
			if needA == v.Analyzer.Name {
				mychecks = append(mychecks, v.Analyzer)
				break
			}
		}
	}

	// ST
	st := []string{"ST1022", "ST1015"}
	for _, v := range stylecheck.Analyzers {
		for _, needA := range st {
			if needA == v.Analyzer.Name {
				mychecks = append(mychecks, v.Analyzer)
				break
			}
		}
	}

	// golang.org/x/tools/go/analysis/passes
	mychecks = append(mychecks, printf.Analyzer)
	mychecks = append(mychecks, slog.Analyzer)
	mychecks = append(mychecks, unreachable.Analyzer)
	mychecks = append(mychecks, loopclosure.Analyzer)

	// custom
	mychecks = append(mychecks, analyzer.New())
	mychecks = append(mychecks, bodyclose.Analyzer)
	// mychecks = append(mychecks, contextcheck.NewAnalyzer(contextcheck.Configuration{}))

	// ну и мою проверку добавляем
	mychecks = append(mychecks, linter.New(nil))

	multichecker.Main(
		mychecks...,
	)
}
