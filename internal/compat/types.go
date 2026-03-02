package compat

// Position is a 1-based source position.
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// IsZero reports whether the position is not set.
func (p Position) IsZero() bool {
	return p.Line == 0 && p.Column == 0
}

// Span represents an inclusive begin/end location.
type Span struct {
	Begin Position `json:"begin"`
	End   Position `json:"end"`
}

// Token is the canonical token payload used in parity checks.
type Token struct {
	Kind     int      `json:"kind"`
	Image    string   `json:"image"`
	Begin    Position `json:"begin"`
	End      Position `json:"end"`
	LexState string   `json:"lex_state,omitempty"`
}

// ParseErrorClass identifies lexical vs parser failures.
type ParseErrorClass string

const (
	ParseErrorClassLexical ParseErrorClass = "lexical"
	ParseErrorClassParser  ParseErrorClass = "parse"
)

// ParseError captures parser/lexer failures in parity runs.
type ParseError struct {
	Class   ParseErrorClass `json:"class"`
	Message string          `json:"message"`
	Begin   Position        `json:"begin"`
	End     Position        `json:"end"`
}
