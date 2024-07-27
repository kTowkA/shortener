package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

const (
	// notRecomendedText текст для сообщении о нежелательности использования некоторых функций
	notRecomendedText = "not recommended function"
)

// notRecomended структура содержащая информацию о нерекомендованных к использованию функций
type NotRecomended struct {
	// pkg в каких пакетах нельзя использовать. если не указано ничего, ищется во всех пакетах
	Pkg string
	// fromFunction из какой функции нельзя вызывать нерекомендованную функцию. если не указано ничего, ищется во всех функциях
	FromFunction string
	// function функция не рекомендованная к использованию
	Function NotRecomendedFunc
}

// notRecomendedFunc функция не рекомендованная к использованию
type NotRecomendedFunc struct {
	// pkg имя пакета функции. Если ничего не указано, то только название функции будет использовано
	Pkg string
	// name имя функции
	Name string
}

// notRecomendedFuncs массив с нерекомендуемыми функциями
var notRecomendedFuncs = []NotRecomended{
	{
		// задание - в пакете main нельзя использовать os.Exit в функции main
		Pkg:          "main",
		FromFunction: "main",
		Function: NotRecomendedFunc{
			Pkg:  "os",
			Name: "Exit",
		},
	},
}

// New создание собственного анализатора. Если не переданы функции для анализа в config, то используется значение по умолчанию - в пакете main в функции main нельзя вызывать os.Exit
func New(config []NotRecomended) *analysis.Analyzer {
	if len(config) != 0 {
		notRecomendedFuncs = config
	}
	return &analysis.Analyzer{
		Name: "notrecomendedfuncs",
		Doc:  "Checking for undesirability of using functions",
		Run:  run,
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	// проверяем для каждой функции в notRecomendedFuncs
	for _, nrf := range notRecomendedFuncs {
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
