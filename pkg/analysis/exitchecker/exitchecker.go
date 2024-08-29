package exitchecker

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// ExitCheckAnalyzer - структура для анализатора проверки на выход из функций main.
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitchecker",
	Doc:  "checks for os.Exit points in code",
	Run:  run,
}

// run - проверка на выход из функций main.
func run(pass *analysis.Pass) (result interface{}, err error) {
	if pass.Pkg.Name() != "main" {
		return
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Проверка, является ли узел функцией main
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
				// Проходим по всем выражениям функции
				for _, stmt := range fn.Body.List {
					if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
						if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
							if fun, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
								if pkgIdent, ok := fun.X.(*ast.Ident); ok {
									if pkgIdent.Name == "os" && fun.Sel.Name == "Exit" {
										pass.Reportf(callExpr.Pos(), "found os.Exit in main function")
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}

	return
}
