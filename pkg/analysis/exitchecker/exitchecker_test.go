package exitchecker_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/playmixer/short-link/pkg/analysis/exitchecker"
)

func TestExitChecker(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exitchecker.ExitCheckAnalyzer, "./pkg1/...")
}
