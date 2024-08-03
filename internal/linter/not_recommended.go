package linter

import (
	"go/ast"
	"log/slog"

	"golang.org/x/tools/go/analysis"
)

const (
	// notRecomendedText текст для сообщении о нежелательности использования некоторых функций
	notRecomendedText = "not recommended function"
)

// NotRecommended структура содержащая информацию о нерекомендованных к использованию функций
type NotRecommended struct {
	// pkg в каких пакетах нельзя использовать. если не указано ничего, ищется во всех пакетах
	Pkg string
	// fromFunction из какой функции нельзя вызывать нерекомендованную функцию. если не указано ничего, ищется во всех функциях
	FromFunction string
	// function функция не рекомендованная к использованию
	Function NotRecommendedFunc
}

// notRecomendedFunc функция не рекомендованная к использованию
type NotRecommendedFunc struct {
	// pkg имя пакета функции. Если ничего не указано, то только название функции будет использовано
	Pkg string
	// name имя функции
	Name string
}

// defaultsNotRecomendedFuncs массив с нерекомендуемыми функциями по умолчанию (задание инкремента)
var defaultsNotRecomendedFuncs = []NotRecommended{
	{
		// задание - в пакете main нельзя использовать os.Exit в функции main
		Pkg:          "main",
		FromFunction: "main",
		Function: NotRecommendedFunc{
			Pkg:  "os",
			Name: "Exit",
		},
	},
}

// AnalyzerNotRecommended анализатор нерекомендуемых к использованию функций
type AnalyzerNotRecommended struct {
	notRecomendedFuncs []NotRecommended
}

// NewAnalyzerNotRecommended создает новый анализатор типа AnalyzerNotRecommended и устанавливает ему правила cfg, при их отсутствии берет значения по умолчанию
func NewAnalyzerNotRecommended(cfg ...NotRecommended) *AnalyzerNotRecommended {
	anr := new(AnalyzerNotRecommended)
	if len(cfg) == 0 {
		slog.Info("в анализаторе установлено значение по умолчанию")
		anr.notRecomendedFuncs = defaultsNotRecomendedFuncs
	} else {
		anr.notRecomendedFuncs = cfg
	}
	return anr
}

// SetConfig устанавливает нерекомендуемые к использованию функции, если ничего не передано произойдет паника
func (anr *AnalyzerNotRecommended) SetConfig(cfg ...NotRecommended) *AnalyzerNotRecommended {
	if len(cfg) == 0 {
		panic("конфигурация не может быть пустой!")
	}
	anr.notRecomendedFuncs = cfg
	return anr
}

// Name возвращает имя анализатора AnalyzerNotRecommended
func (anr *AnalyzerNotRecommended) Name() string {
	return "Notrecomendedfuncs"
}

// Doc возвращает описание анализатора AnalyzerNotRecommended
func (anr *AnalyzerNotRecommended) Doc() string {
	return "Checking for undesirability of using functions"
}

// Run реализация функции запуска анализатора
func (anr *AnalyzerNotRecommended) Run(pass *analysis.Pass) (interface{}, error) {
	// проверяем для каждой функции в notRecomendedFuncs
	for _, nrf := range anr.notRecomendedFuncs {
		// если по каким-либо причинам не было передано имя функции, то пропускаем
		if nrf.Function.Name == "" {
			return nil, nil
		}
		// проверяем каждый файл
		for _, file := range pass.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.File:
					// если в конфиге указан пакет и в файле пакет не совпадает, то нечего и смотреть далее
					if nrf.Pkg != "" && x.Name.Name != nrf.Pkg {
						return false
					}
				case *ast.FuncDecl:
					// если в конфиге указана функция из которой запрещен вызов и вызывается не эта функция, то нечего и смотреть далее
					if nrf.FromFunction != "" && x.Name.Name != nrf.FromFunction {

						return false
					}
				case *ast.CallExpr:
					switch y := x.Fun.(type) {
					case *ast.Ident:
						// это для функции без пакета
						if nrf.Function.Pkg == "" && y.Name == nrf.Function.Name {
							pass.Reportf(x.Pos(), notRecomendedText)
						}
					case *ast.SelectorExpr:
						// тут ищем совпадение пакета и имя файла
						pkg := ""
						if pkgID, ok := y.X.(*ast.Ident); ok {
							pkg = pkgID.Name
						}
						if pkg == nrf.Function.Pkg && y.Sel.Name == nrf.Function.Name {
							pass.Reportf(x.Pos(), notRecomendedText)
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}

// Analyzer создание экземпляра *analysis.Analyzer
func (anr *AnalyzerNotRecommended) Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: anr.Name(),
		Doc:  anr.Doc(),
		Run:  anr.Run,
	}
}
