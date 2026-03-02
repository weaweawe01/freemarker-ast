package freemarker_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/weaweawe01/freemarker-ast/internal/astdump"
)

func findCoreRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	root := filepath.Join(filepath.Dir(filename), "..", "ast", "core")
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("ast/core directory not found at %s: %v", root, err)
	}
	return root
}

func readCaseFiles(t *testing.T, caseName string) (string, string) {
	t.Helper()

	coreRoot := findCoreRoot(t)
	ftlPath := filepath.Join(coreRoot, caseName+".ftl")
	astPath := filepath.Join(coreRoot, caseName+".ast")

	ftlRaw, err := os.ReadFile(ftlPath)
	if err != nil {
		t.Fatalf("read ftl %s: %v", ftlPath, err)
	}
	astRaw, err := os.ReadFile(astPath)
	if err != nil {
		t.Fatalf("read ast %s: %v", astPath, err)
	}

	return string(ftlRaw), string(astRaw)
}

func runASTCaseParity(t *testing.T, caseName string) {
	t.Helper()

	ftlRaw, astRaw := readCaseFiles(t, caseName)

	ftlInput := normalizeNewlines(ftlRaw)
	ftlInput = stripLeadingFTLComment(ftlInput)

	gotAST, err := astdump.ParseToJavaLikeAST(ftlInput)
	if err != nil {
		t.Fatalf("parse and dump %s: %v", caseName, err)
	}

	got := normalizeNewlines(gotAST)
	got = stripLeadingASTHeaderComment(got)
	got = finalizeComparableAST(got)

	want := normalizeNewlines(astRaw)
	want = stripLeadingASTHeaderComment(want)
	want = finalizeComparableAST(want)

	assertASTEqual(t, got, want, caseName)
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func stripLeadingFTLComment(s string) string {
	s = strings.TrimPrefix(s, "\uFEFF")
	return stripLeadingBlockComment(s, "<#--", "-->")
}

func stripLeadingASTHeaderComment(s string) string {
	s = strings.TrimPrefix(s, "\uFEFF")
	return stripLeadingBlockComment(s, "/*", "*/")
}

func stripLeadingBlockComment(s string, startToken string, endToken string) string {
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
	rest = strings.TrimLeft(rest, " \t\n")
	return rest
}

func finalizeComparableAST(s string) string {
	return strings.TrimRight(s, "\n")
}

func assertASTEqual(t *testing.T, got string, want string, caseName string) {
	t.Helper()
	if got == want {
		return
	}

	line, gotLine, wantLine := firstDiffLine(got, want)
	t.Fatalf(
		"%s AST mismatch at line %d\nwant: %q\ngot : %q",
		caseName,
		line,
		wantLine,
		gotLine,
	)
}

func firstDiffLine(got string, want string) (int, string, string) {
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")

	max := len(gotLines)
	if len(wantLines) > max {
		max = len(wantLines)
	}

	for i := 0; i < max; i++ {
		var g string
		if i < len(gotLines) {
			g = gotLines[i]
		}
		var w string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if g != w {
			return i + 1, g, w
		}
	}

	return max, "", ""
}

func TestStripLeadingFTLComment(t *testing.T) {
	in := "<#-- license -->\n1 <#assign x = 1>"
	got := stripLeadingFTLComment(in)
	want := "1 <#assign x = 1>"
	if got != want {
		t.Fatalf("strip ftl comment mismatch: got %q, want %q", got, want)
	}
}

func TestStripLeadingASTHeaderComment(t *testing.T) {
	in := "/*\n * license\n */\n#mixed_content  // f.c.MixedContent"
	got := stripLeadingASTHeaderComment(in)
	want := "#mixed_content  // f.c.MixedContent"
	if got != want {
		t.Fatalf("strip ast header mismatch: got %q, want %q", got, want)
	}
}
