// Package freemarker provides a Go implementation of the Apache FreeMarker
// template parser, producing an AST 100% compatible with the Java reference
// implementation (freemarker 2.3.34).
//
// Usage:
//
//	out, err := freemarker.ParseToJavaLikeAST(src)
package freemarker

import (
	"github.com/weaweawe01/freemarker-ast/internal/ast"
	"github.com/weaweawe01/freemarker-ast/internal/astdump"
	"github.com/weaweawe01/freemarker-ast/internal/parser"
	"github.com/weaweawe01/freemarker-ast/internal/risk"
)

// ParseToJavaLikeAST parses a FreeMarker template source string and returns
// the AST in the Java-like textual format used by the freemarker-core library.
//
// The output is byte-for-byte identical to the text produced by the Java
// FreeMarker 2.3.34 AST dump tool for the same template input.
//
// Example:
//
//	src := `<#assign x = "hello">${x?upper_case}`
//	out, err := freemarker.ParseToJavaLikeAST(src)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Print(out)
func ParseToJavaLikeAST(src string) (string, error) {
	return astdump.ParseToJavaLikeAST(src)
}

// 开放Parse 函数，供外部调用
func Parse(src string) (*ast.Root, error) {
	root, err := parser.Parse(src)
	if err != nil {
		return nil, err
	}
	return root, nil
}

// RiskSeverity is the textual risk level computed from risk score.
type RiskSeverity = risk.Severity

// RiskFinding is a single risk hit produced by one rule.
type RiskFinding = risk.Finding

// RiskReport is the aggregate risk analysis result.
type RiskReport = risk.Report

// AnalyzeRisk parses source and computes static risk score against the AST.
func AnalyzeRisk(src string) (*RiskReport, error) {
	root, err := Parse(src)
	if err != nil {
		return nil, err
	}
	return AnalyzeRiskAST(root), nil
}

// AnalyzeRiskAST computes static risk score from an already parsed AST.
func AnalyzeRiskAST(root *ast.Root) *RiskReport {
	return risk.Analyze(root)
}
