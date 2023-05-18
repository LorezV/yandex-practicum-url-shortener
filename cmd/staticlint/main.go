package staticlint

import (
	"strings"

	goc "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/gostaticanalysis/nilerr"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/LorezV/url-shorter.git/cmd/staticlint/analyzer"
)

var StaticChecks = []string{"SA"}
var StyleChecks = []string{"ST1000", "ST1005"}

func main() {
	analyzers := []*analysis.Analyzer{
		analyzer.OsExitAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		httpresponse.Analyzer,
		goc.Analyzer,
		nilerr.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		for _, sc := range StaticChecks {
			if strings.HasPrefix(v.Analyzer.Name, sc) {
				analyzers = append(analyzers, v.Analyzer)
			}
		}
	}
	for _, v := range stylecheck.Analyzers {
		for _, sc := range StyleChecks {
			if strings.HasPrefix(v.Analyzer.Name, sc) {
				analyzers = append(analyzers, v.Analyzer)
			}
		}
	}

	multichecker.Main(analyzers...)
}
