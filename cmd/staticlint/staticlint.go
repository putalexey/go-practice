package main

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"strings"
)

var honnefAnalizers = map[string]bool{
	"S1000":  true, // Use plain channel send or receive instead of single-case select
	"S1001":  true, // Replace for loop with call to copy
	"S1002":  true, // Omit comparison with boolean constant
	"S1003":  true, // Replace call to strings.Index with strings.Contains
	"S1004":  true, // Replace call to bytes.Compare with bytes.Equal
	"S1005":  true, // Drop unnecessary use of the blank identifier
	"S1007":  true, // Simplify regular expression by using raw string literal
	"S1008":  true, // Simplify returning boolean expression
	"S1009":  true, // Omit redundant nil check on slices
	"S1012":  true, // Replace time.Now().Sub(x) with time.Since(x)
	"S1018":  true, // Use copy for sliding elements
	"S1024":  true, // Replace x.Sub(time.Now()) with time.Until(x)
	"S1028":  true, // Simplify error construction with fmt.Errorf
	"S1030":  true, // Use bytes.Buffer.String or bytes.Buffer.Bytes
	"S1031":  true, // Omit redundant nil check around loop
	"S1036":  true, // Unnecessary guard around map access
	"ST1001": true, // Dot imports are discouraged
	"ST1005": true, // Incorrectly formatted error string
	"ST1008": true, // A function’s error value should be its last return value
	"ST1013": true, // Should use constants for HTTP error codes, not magic numbers
	"ST1015": true, // A switch’s default case should be the first or last case
	"ST1016": true, // Use consistent method receiver names
	"ST1019": true, // Importing the same package multiple times
	"QF1004": true, // Use strings.ReplaceAll instead of strings.Replace with n == -1
	"QF1005": true, // Expand call to math.Pow
	"QF1009": true, // Use time.Time.Equal instead of == operator
	"QF1010": true, // Convert slice of bytes to string when printing it
}

var OsExitCheckAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for os.Exit statements",
	Run:  run,
}

func main() {
	mychecks := getGoAnalysisPassesAnalyzers()
	mychecks = append(mychecks, OsExitCheckAnalyzer)

	for _, check := range staticcheck.Analyzers {
		if strings.HasPrefix(check.Analyzer.Name, "SA") {
			mychecks = append(mychecks, check.Analyzer)
		}
	}

	analyzerGroups := [][]*lint.Analyzer{
		simple.Analyzers,
		stylecheck.Analyzers,
		quickfix.Analyzers,
	}
	for _, ag := range analyzerGroups {
		for _, check := range ag {
			if honnefAnalizers[check.Analyzer.Name] {
				mychecks = append(mychecks, check.Analyzer)
			}
		}
	}

	multichecker.Main(mychecks...)
}

func getGoAnalysisPassesAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		//fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	}
}

func run(p *analysis.Pass) (interface{}, error) {
	for _, file := range p.Files {
		if file.Name.String() == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				if x, ok := node.(*ast.FuncDecl); ok {
					if x.Name.Name == "main" {
						inspectMainFunc(p, file)
						return false
					}
				}
				return true
			})
		}
	}

	return nil, nil
}

func inspectMainFunc(p *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		if isOsExitStmt(node) {
			p.Reportf(node.Pos(), "os.Exit called in main function of main package")
			return false
		}
		return true
	})
}

func isOsExitStmt(node ast.Node) bool {
	if x, ok := node.(*ast.ExprStmt); ok {
		if callexpr, ok := x.X.(*ast.CallExpr); ok {
			if x, ok := callexpr.Fun.(*ast.SelectorExpr); ok {
				if id, ok := x.X.(*ast.Ident); ok {
					if id.Name == "os" && x.Sel.Name == "Exit" {
						return true
					}
				}
			}
		}
	}
	return false
}
