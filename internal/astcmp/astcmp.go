package astcmp

import (
	"fmt"
	"strings"
)

// Result stores the comparison result of two AST texts.
type Result struct {
	Equal    bool
	Line     int
	Oracle   string
	Actual   string
	DiffText string
}

// Normalize applies shared canonicalization rules before AST comparison.
func Normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	s = strings.TrimPrefix(s, "\uFEFF")
	s = stripLeadingBlockComment(s, "/*", "*/")
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	lines = trimTrailingEmptyLines(lines)
	lines = stripTrailingNewlineOnlyTextNodes(lines)
	return strings.Join(lines, "\n")
}

func stripTrailingNewlineOnlyTextNodes(lines []string) []string {
	for len(lines) >= 2 {
		n := len(lines)
		if lines[n-2] != "#text  // f.c.TextBlock" {
			return lines
		}
		if !isNewlineOnlyTextContentLine(lines[n-1]) {
			return lines
		}
		lines = lines[:n-2]
	}
	return lines
}

func trimTrailingEmptyLines(lines []string) []string {
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func isNewlineOnlyTextContentLine(line string) bool {
	const prefix = `- content: "`
	const suffix = `"  // String`
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
		return false
	}
	body := strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix)
	if len(body) < 2 || len(body)%2 != 0 {
		return false
	}
	for i := 0; i < len(body); i += 2 {
		if body[i] != '\\' || body[i+1] != 'n' {
			return false
		}
	}
	return true
}

// CompareNormalized compares two normalized AST payloads.
func CompareNormalized(oracle, actual string) Result {
	if oracle == actual {
		return Result{Equal: true}
	}

	line, o, a := firstDiffLine(oracle, actual)
	return Result{
		Equal:    false,
		Line:     line,
		Oracle:   o,
		Actual:   a,
		DiffText: makeDiff(oracle, actual, line),
	}
}

func stripLeadingBlockComment(s, startToken, endToken string) string {
	leading := len(s) - len(strings.TrimLeft(s, " \t\n"))
	if leading >= len(s) {
		return s
	}

	body := s[leading:]
	if !strings.HasPrefix(body, startToken) {
		return s
	}

	end := strings.Index(body[len(startToken):], endToken)
	if end < 0 {
		return s
	}
	end += len(startToken)

	rest := body[end+len(endToken):]
	return strings.TrimLeft(rest, " \t\n")
}

func firstDiffLine(oracle, actual string) (int, string, string) {
	oracleLines := strings.Split(oracle, "\n")
	actualLines := strings.Split(actual, "\n")

	maxLines := len(oracleLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := 0; i < maxLines; i++ {
		var o string
		if i < len(oracleLines) {
			o = oracleLines[i]
		}

		var a string
		if i < len(actualLines) {
			a = actualLines[i]
		}

		if o != a {
			return i + 1, o, a
		}
	}

	return maxLines, "", ""
}

func makeDiff(oracle, actual string, line int) string {
	oracleLines := strings.Split(oracle, "\n")
	actualLines := strings.Split(actual, "\n")

	start := line - 3
	if start < 1 {
		start = 1
	}
	end := line + 2

	var b strings.Builder
	b.WriteString("--- oracle\n")
	b.WriteString("+++ actual\n")
	b.WriteString(fmt.Sprintf("@@ first_diff_line=%d @@\n", line))

	for ln := start; ln <= end; ln++ {
		var o string
		if ln-1 < len(oracleLines) {
			o = oracleLines[ln-1]
		}

		var a string
		if ln-1 < len(actualLines) {
			a = actualLines[ln-1]
		}

		switch {
		case o == a:
			b.WriteString(fmt.Sprintf(" %4d | %s\n", ln, o))
		default:
			b.WriteString(fmt.Sprintf("-%4d | %s\n", ln, o))
			b.WriteString(fmt.Sprintf("+%4d | %s\n", ln, a))
		}
	}

	return b.String()
}
