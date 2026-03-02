package oracle

import "github.com/weaweawe01/freemarker-ast/internal/compat"

// TokenOracleFile stores token output for a single template/case.
type TokenOracleFile struct {
	CaseName string         `json:"case_name"`
	Tokens   []compat.Token `json:"tokens"`
}

// ASTOracleFile stores AST string output for a single template/case.
type ASTOracleFile struct {
	CaseName string `json:"case_name"`
	AST      string `json:"ast"`
}

// CanonicalOracleFile stores canonical form output for a single template/case.
type CanonicalOracleFile struct {
	CaseName  string `json:"case_name"`
	Canonical string `json:"canonical"`
}

// ErrorOracleFile stores parse/lex error output for a single template/case.
type ErrorOracleFile struct {
	CaseName string             `json:"case_name"`
	Error    *compat.ParseError `json:"error,omitempty"`
}

// OracleBundle can carry all expected artifacts for one case in a single payload.
type OracleBundle struct {
	CaseName  string             `json:"case_name"`
	Tokens    []compat.Token     `json:"tokens,omitempty"`
	AST       string             `json:"ast,omitempty"`
	Canonical string             `json:"canonical,omitempty"`
	Error     *compat.ParseError `json:"error,omitempty"`
}
