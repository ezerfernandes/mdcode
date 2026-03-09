package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func parseSource(t *testing.T, src string) (*ast.File, map[string]bool) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	imports := buildMetricsImportSet(f)
	return f, imports
}

func firstFuncBody(t *testing.T, f *ast.File) *ast.BlockStmt {
	t.Helper()
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			return fn.Body
		}
	}
	t.Fatal("no function found")
	return nil
}

func TestHasMetricsCall_PrometheusImport(t *testing.T) {
	src := `package foo

import "github.com/prometheus/client_golang/prometheus"

func setup() {
	c := prometheus.NewCounter(prometheus.CounterOpts{Name: "requests"})
	_ = c
}
`
	f, imports := parseSource(t, src)
	body := firstFuncBody(t, f)

	if !hasMetricsCall(body, imports) {
		t.Error("expected metrics call to be detected for prometheus.NewCounter")
	}
}

func TestHasMetricsCall_OpenTelemetry(t *testing.T) {
	src := `package foo

import "go.opentelemetry.io/otel/metric"

func setup(m metric.Meter) {
	counter, _ := m.Int64Counter("requests")
	_ = counter
}
`
	f, imports := parseSource(t, src)
	body := firstFuncBody(t, f)

	// The import is matched, so any call on 'metric' package counts.
	if !hasMetricsCall(body, imports) {
		t.Error("expected metrics call to be detected for otel metric")
	}
}

func TestHasMetricsCall_GenericMethodOnMetricsReceiver(t *testing.T) {
	src := `package foo

func doWork(counter Counter) {
	counter.Add(1)
}
`
	f, imports := parseSource(t, src)
	body := firstFuncBody(t, f)

	if !hasMetricsCall(body, imports) {
		t.Error("expected counter.Add to be detected as metrics call")
	}
}

func TestHasMetricsCall_GenericMethodOnNonMetricsReceiver(t *testing.T) {
	src := `package foo

import "sync"

func doWork(wg *sync.WaitGroup) {
	wg.Add(1)
}
`
	f, imports := parseSource(t, src)
	body := firstFuncBody(t, f)

	if hasMetricsCall(body, imports) {
		t.Error("wg.Add should NOT be detected as metrics call")
	}
}

func TestHasMetricsCall_NoMetrics(t *testing.T) {
	src := `package foo

func hello() string {
	return "hello"
}
`
	f, imports := parseSource(t, src)
	body := firstFuncBody(t, f)

	if hasMetricsCall(body, imports) {
		t.Error("plain function should not be detected as having metrics")
	}
}

func TestHasMetricsCall_MapSet(t *testing.T) {
	// A map "Set" call should not be flagged.
	src := `package foo

func doWork(m map[string]int) {
	m["key"] = 1
}

func doWork2(data Data) {
	data.Set("x", 1)
}
`
	f, imports := parseSource(t, src)
	// Check the second function (data.Set) — "data" does not hint at metrics.
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Name.Name != "doWork2" {
			continue
		}
		if hasMetricsCall(fn.Body, imports) {
			t.Error("data.Set should NOT be detected as metrics call")
		}
	}
}

func TestHasMetricsCall_HistogramObserve(t *testing.T) {
	src := `package foo

func record(histogram Histogram) {
	histogram.Observe(1.5)
}
`
	f, imports := parseSource(t, src)
	body := firstFuncBody(t, f)

	if !hasMetricsCall(body, imports) {
		t.Error("histogram.Observe should be detected as metrics call")
	}
}

func TestLooksLikeMetricsReceiver(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"counter", true},
		{"requestCounter", true},
		{"histogram", true},
		{"latencyHistogram", true},
		{"gauge", true},
		{"myGauge", true},
		{"wg", false},
		{"data", false},
		{"client", false},
		{"db", false},
		{"metricsCollector", true},
		{"promRegistry", true},
		{"telemetryRecorder", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := looksLikeMetricsReceiver(tt.name)
			if got != tt.want {
				t.Errorf("looksLikeMetricsReceiver(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestBuildMetricsImportSet(t *testing.T) {
	src := `package foo

import (
	"fmt"
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/metric"
)
`
	f, _ := parseSource(t, src)
	imports := buildMetricsImportSet(f)

	if !imports["prom"] {
		t.Error("expected 'prom' alias to be in metrics imports")
	}
	if !imports["metric"] {
		t.Error("expected 'metric' to be in metrics imports")
	}
	if imports["fmt"] {
		t.Error("'fmt' should not be in metrics imports")
	}
}

func TestAnalyzeFile_Integration(t *testing.T) {
	// Write a temp Go file and analyze it.
	dir := t.TempDir()
	src := `package foo

import "github.com/prometheus/client_golang/prometheus"

func instrumented() {
	c := prometheus.NewCounter(prometheus.CounterOpts{})
	_ = c
}

func notInstrumented() {
	x := 1 + 2
	_ = x
}

func alsoNotInstrumented() {
	println("hello")
}
`
	path := filepath.Join(dir, "foo.go")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	fr, err := analyzeFile(path)
	if err != nil {
		t.Fatalf("analyzeFile error: %v", err)
	}

	if fr.TotalFuncs != 3 {
		t.Errorf("expected 3 total funcs, got %d", fr.TotalFuncs)
	}
	if fr.InstrumentedFuncs != 1 {
		t.Errorf("expected 1 instrumented func, got %d", fr.InstrumentedFuncs)
	}
}

func TestAnalyzeDirectory_Integration(t *testing.T) {
	dir := t.TempDir()

	// Create a sub-package.
	pkgDir := filepath.Join(dir, "mypkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}

	src := `package mypkg

func plain() {
	_ = 42
}
`
	if err := os.WriteFile(filepath.Join(pkgDir, "a.go"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	reports, err := analyzeDirectory(dir)
	if err != nil {
		t.Fatalf("analyzeDirectory error: %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 package report, got %d", len(reports))
	}

	r := reports[0]
	if r.TotalFuncs != 1 {
		t.Errorf("expected 1 total func, got %d", r.TotalFuncs)
	}
	if r.InstrumentedFuncs != 0 {
		t.Errorf("expected 0 instrumented funcs, got %d", r.InstrumentedFuncs)
	}
	if !strings.HasSuffix(r.Package, "mypkg") {
		t.Errorf("expected package to end with 'mypkg', got %q", r.Package)
	}
}
