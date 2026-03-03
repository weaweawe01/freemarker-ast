package risk

import "strings"

// Severity is the textual risk level computed from total risk score.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Finding is a single risk hit produced by one rule.
type Finding struct {
	Rule     string `json:"rule"`
	Score    int    `json:"score"`
	Message  string `json:"message"`
	Evidence string `json:"evidence,omitempty"`
}

// Report is the final risk assessment result.
type Report struct {
	TotalScore int       `json:"total_score"`
	Severity   Severity  `json:"severity"`
	Findings   []Finding `json:"findings"`
}

func scoreToSeverity(score int) Severity {
	switch {
	case score >= 300:
		return SeverityCritical
	case score >= 180:
		return SeverityHigh
	case score >= 80:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func normalizeClassName(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteByte(byte(r + ('a' - 'A')))
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			// Drop whitespace so "ObjectConst ructor" still matches.
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func unquoteLiteral(s string) string {
	if len(s) < 2 {
		return s
	}
	q := s[0]
	if (q == '"' || q == '\'') && s[len(s)-1] == q {
		return s[1 : len(s)-1]
	}
	return s
}

func looksLikeClassName(normalized string) bool {
	if strings.Count(normalized, ".") < 1 {
		return false
	}
	for _, r := range normalized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '$' || r == '_' {
			continue
		}
		return false
	}
	return true
}
