package main

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/playmixer/short-link/pkg/analysis/exitchecker"
)

func main() {
	checks := []*analysis.Analyzer{
		sortslice.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		errcheck.Analyzer,
		ineffassign.Analyzer,
		exitchecker.ExitCheckAnalyzer,
	}
	// добавляем все анализаторы класса SA.
	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}
	// активные анализаторы.
	stChecks := map[string]bool{
		"ST1001": true,
		"QF1003": true,
		"S1001":  true,
	}
	// добавляем анализаторы стиля.
	for _, v := range stylecheck.Analyzers {
		if stChecks[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}
	// добавляем анализатор quickfix.
	for _, v := range quickfix.Analyzers {
		if stChecks[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}
	// добавляем анализатор упрощения кода.
	for _, v := range simple.Analyzers {
		if stChecks[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	multichecker.Main(
		checks...,
	)
}
