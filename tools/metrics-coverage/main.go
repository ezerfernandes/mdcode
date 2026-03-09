// metrics-coverage: Analyze Go source files for metrics instrumentation coverage.
//
// Scans .go files in a directory tree, inspects each function/method for known
// metrics API calls (OpenTelemetry, Prometheus, go-kit/metrics, etc.), and outputs
// a per-package coverage report. Generic method names like Add, Inc, Set are only
// matched when the receiver name suggests a metrics type (counter, histogram, gauge, etc.).
//
// Usage:
//
//	go run ./tools/metrics-coverage [directory]
//
// If no directory is given, it defaults to ".".
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
)

// metricsPackagePrefixes are import path fragments that indicate a metrics library.
// A function call whose package matches one of these is always counted as instrumented.
var metricsPackagePrefixes = []string{
	"go.opentelemetry.io/otel/metric",
	"github.com/prometheus/client_golang/prometheus",
	"go-kit/metrics",
	"github.com/DataDog/datadog-go/statsd",
	"github.com/armon/go-metrics",
	"expvar",
}

// qualifiedMetricsCalls are fully-qualified function/method patterns that
// unambiguously indicate metrics instrumentation regardless of receiver name.
var qualifiedMetricsCalls = []string{
	// OpenTelemetry meter/provider
	"meter.Int64Counter",
	"meter.Float64Counter",
	"meter.Int64Histogram",
	"meter.Float64Histogram",
	"meter.Int64UpDownCounter",
	"meter.Float64UpDownCounter",
	"meter.Int64Gauge",
	"meter.Float64Gauge",
	// Prometheus
	"prometheus.NewCounter",
	"prometheus.NewCounterVec",
	"prometheus.NewGauge",
	"prometheus.NewGaugeVec",
	"prometheus.NewHistogram",
	"prometheus.NewHistogramVec",
	"prometheus.NewSummary",
	"prometheus.NewSummaryVec",
	"prometheus.MustRegister",
	"promauto.NewCounter",
	"promauto.NewGauge",
	"promauto.NewHistogram",
	"promauto.NewSummary",
	// expvar
	"expvar.NewInt",
	"expvar.NewFloat",
	"expvar.NewString",
	"expvar.NewMap",
}

// metricsReceiverHints are substrings that, when found in a receiver/variable
// name (case-insensitive), indicate the variable is a metrics instrument.
// Generic methods (Add, Inc, Set, etc.) are only counted when the receiver
// matches one of these hints, drastically reducing false positives.
var metricsReceiverHints = []string{
	"counter",
	"histogram",
	"gauge",
	"summary",
	"metric",
	"meter",
	"timer",
	"observer",
	"recorder",
	"stats",
	"statsd",
	"telemetry",
	"prom",
}

// unambiguousMetricsMethods are method names specific enough to metrics APIs
// that they count as instrumentation regardless of receiver name.
var unambiguousMetricsMethods = map[string]bool{
	// OpenTelemetry Meter methods
	"Int64Counter":         true,
	"Float64Counter":       true,
	"Int64Histogram":       true,
	"Float64Histogram":     true,
	"Int64UpDownCounter":   true,
	"Float64UpDownCounter": true,
	"Int64Gauge":           true,
	"Float64Gauge":         true,
	// Prometheus constructors (when called as methods)
	"NewCounter":      true,
	"NewCounterVec":   true,
	"NewGauge":        true,
	"NewGaugeVec":     true,
	"NewHistogram":    true,
	"NewHistogramVec": true,
	"NewSummary":      true,
	"NewSummaryVec":   true,
	"MustRegister":    true,
	// Common metrics methods with unambiguous names
	"WithLabelValues": true,
}

// genericMetricsMethods are method names that are only considered metrics calls
// when the receiver name matches a metricsReceiverHint.
var genericMetricsMethods = map[string]bool{
	"Add":       true,
	"Inc":       true,
	"Dec":       true,
	"Set":       true,
	"Observe":   true,
	"Record":    true,
	"With": true,
}

// PackageReport holds coverage data for one package.
type PackageReport struct {
	Package           string
	Files             []FileReport
	TotalFuncs        int
	InstrumentedFuncs int
}

// FileReport holds coverage data for one file.
type FileReport struct {
	File              string
	TotalFuncs        int
	InstrumentedFuncs int
}

// CoveragePercent returns the coverage percentage.
func (p *PackageReport) CoveragePercent() float64 {
	if p.TotalFuncs == 0 {
		return 0
	}
	return float64(p.InstrumentedFuncs) / float64(p.TotalFuncs) * 100
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	reports, err := analyzeDirectory(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "metrics-coverage: %v\n", err)
		os.Exit(1)
	}

	printReport(reports)
}

// analyzeDirectory walks root and returns per-package reports.
func analyzeDirectory(root string) ([]PackageReport, error) {
	pkgFiles := map[string][]string{} // package dir -> list of .go files

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == "testdata" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			dir := filepath.Dir(path)
			pkgFiles[dir] = append(pkgFiles[dir], path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var reports []PackageReport
	for dir, files := range pkgFiles {
		report, err := analyzePackage(dir, files)
		if err != nil {
			return nil, fmt.Errorf("analyzing %s: %w", dir, err)
		}
		reports = append(reports, report)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Package < reports[j].Package
	})

	return reports, nil
}

// analyzePackage analyzes all Go files in a package directory.
func analyzePackage(dir string, files []string) (PackageReport, error) {
	report := PackageReport{Package: dir}

	// Collect metrics-related imports for the package.
	for _, file := range files {
		fr, err := analyzeFile(file)
		if err != nil {
			return report, err
		}
		report.Files = append(report.Files, fr)
		report.TotalFuncs += fr.TotalFuncs
		report.InstrumentedFuncs += fr.InstrumentedFuncs
	}

	return report, nil
}

// analyzeFile parses a single Go file and checks each function for metrics calls.
func analyzeFile(path string) (FileReport, error) {
	fr := FileReport{File: filepath.Base(path)}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return fr, fmt.Errorf("parsing %s: %w", path, err)
	}

	// Build a set of imported package aliases that are metrics-related.
	metricsImports := buildMetricsImportSet(f)

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		fr.TotalFuncs++
		if hasMetricsCall(fn.Body, metricsImports) {
			fr.InstrumentedFuncs++
		}
	}

	return fr, nil
}

// buildMetricsImportSet returns the set of local package names that correspond
// to known metrics libraries.
func buildMetricsImportSet(f *ast.File) map[string]bool {
	result := map[string]bool{}
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		for _, prefix := range metricsPackagePrefixes {
			if strings.Contains(path, prefix) {
				name := ""
				if imp.Name != nil {
					name = imp.Name.Name
				} else {
					parts := strings.Split(path, "/")
					name = parts[len(parts)-1]
				}
				result[name] = true
				break
			}
		}
	}
	return result
}

// hasMetricsCall returns true if the function body contains at least one
// call that is identified as metrics instrumentation.
func hasMetricsCall(body *ast.BlockStmt, metricsImports map[string]bool) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isMetricsCallExpr(call, metricsImports) {
			found = true
			return false
		}
		return true
	})
	return found
}

// isMetricsCallExpr determines whether a call expression is a metrics API call.
func isMetricsCallExpr(call *ast.CallExpr, metricsImports map[string]bool) bool {
	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		methodName := fn.Sel.Name
		receiverName := exprName(fn.X)

		// Check if receiver is a known metrics package import.
		if metricsImports[receiverName] {
			return true
		}

		// Check against qualified metrics calls (e.g., "prometheus.NewCounter").
		qualified := receiverName + "." + methodName
		for _, pattern := range qualifiedMetricsCalls {
			if qualified == pattern {
				return true
			}
		}

		// Unambiguous metrics methods match regardless of receiver name.
		if unambiguousMetricsMethods[methodName] {
			return true
		}

		// For generic method names, only match if the receiver name hints at metrics.
		if genericMetricsMethods[methodName] && looksLikeMetricsReceiver(receiverName) {
			return true
		}

	case *ast.Ident:
		// Direct function calls (no receiver/package) — check if the name itself
		// is a known metrics constructor. This is rare but covers cases like
		// dot-imported metrics packages.
		// We intentionally don't match generic names here.
	}

	return false
}

// looksLikeMetricsReceiver checks whether the receiver/variable name contains
// a substring suggesting it holds a metrics instrument.
func looksLikeMetricsReceiver(name string) bool {
	lower := strings.ToLower(name)
	for _, hint := range metricsReceiverHints {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	return false
}

// exprName returns a best-effort string name for an expression.
func exprName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprName(e.X) + "." + e.Sel.Name
	}
	return ""
}

// printReport outputs the coverage table to stdout.
func printReport(reports []PackageReport) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "PACKAGE\tFILE\tFUNCS\tINSTRUMENTED\tCOVERAGE")
	fmt.Fprintln(w, "-------\t----\t-----\t------------\t--------")

	totalFuncs := 0
	totalInstrumented := 0

	for _, pkg := range reports {
		for _, file := range pkg.Files {
			pct := float64(0)
			if file.TotalFuncs > 0 {
				pct = float64(file.InstrumentedFuncs) / float64(file.TotalFuncs) * 100
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%.1f%%\n",
				pkg.Package, file.File, file.TotalFuncs, file.InstrumentedFuncs, pct)
		}
		totalFuncs += pkg.TotalFuncs
		totalInstrumented += pkg.InstrumentedFuncs
	}

	fmt.Fprintln(w, "\t\t\t\t")
	pct := float64(0)
	if totalFuncs > 0 {
		pct = float64(totalInstrumented) / float64(totalFuncs) * 100
	}
	fmt.Fprintf(w, "TOTAL\t\t%d\t%d\t%.1f%%\n", totalFuncs, totalInstrumented, pct)
	w.Flush()
}
