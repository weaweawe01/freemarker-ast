package lexer

import (
	"testing"

	"github.com/weaweawe01/freemarker-ast/internal/tokenid"
)

func TestDefaultTextTokenization(t *testing.T) {
	lx := New("abc  def", Config{})

	tok1, err := lx.Next()
	if err != nil {
		t.Fatalf("next1: %v", err)
	}
	if tok1.Kind != tokenid.TK_STATIC_TEXT_NON_WS || tok1.Image != "abc" {
		t.Fatalf("tok1 mismatch: %#v", tok1)
	}

	tok2, err := lx.Next()
	if err != nil {
		t.Fatalf("next2: %v", err)
	}
	if tok2.Kind != tokenid.TK_STATIC_TEXT_WS || tok2.Image != "  " {
		t.Fatalf("tok2 mismatch: %#v", tok2)
	}

	tok3, err := lx.Next()
	if err != nil {
		t.Fatalf("next3: %v", err)
	}
	if tok3.Kind != tokenid.TK_STATIC_TEXT_NON_WS || tok3.Image != "def" {
		t.Fatalf("tok3 mismatch: %#v", tok3)
	}

	eof, err := lx.Next()
	if err != nil {
		t.Fatalf("eof: %v", err)
	}
	if eof.Kind != tokenid.TK_EOF {
		t.Fatalf("expected EOF, got %#v", eof)
	}
}

func TestInterpolationOpeningTokens(t *testing.T) {
	cases := []struct {
		src      string
		wantKind int
		wantImg  string
	}{
		{src: "${x}", wantKind: tokenid.TK_DOLLAR_INTERPOLATION_OPENING, wantImg: "${"},
		{src: "#{x}", wantKind: tokenid.TK_HASH_INTERPOLATION_OPENING, wantImg: "#{"},
		{src: "[=x]", wantKind: tokenid.TK_SQUARE_BRACKET_INTERPOLATION_OPENING, wantImg: "[="},
	}

	for _, tc := range cases {
		lx := New(tc.src, Config{})
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("%s: %v", tc.src, err)
		}
		if tok.Kind != tc.wantKind || tok.Image != tc.wantImg {
			t.Fatalf("%s: token mismatch: %#v", tc.src, tok)
		}
	}
}

func TestSimpleInterpolationStream(t *testing.T) {
	lx := New("${x+1}", Config{})

	expectKinds := []int{
		tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_PLUS,
		tokenid.TK_INTEGER,
		tokenid.TK_CLOSING_CURLY_BRACKET,
		tokenid.TK_EOF,
	}
	expectImages := []string{"${", "x", "+", "1", "}", ""}

	for i := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != expectKinds[i] || tok.Image != expectImages[i] {
			t.Fatalf("token %d mismatch: got kind=%d image=%q", i, tok.Kind, tok.Image)
		}
	}
}

func TestInterpolationSurroundedByText(t *testing.T) {
	lx := New("a${x}b", Config{})
	expectKinds := []int{
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_CLOSING_CURLY_BRACKET,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d", i, tok.Kind, kind)
		}
	}
}

func TestSquareInterpolationSurroundedByText(t *testing.T) {
	lx := New("a[=x]b", Config{})
	expectKinds := []int{
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_SQUARE_BRACKET_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_CLOSE_BRACKET,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d", i, tok.Kind, kind)
		}
	}
}

func TestExpressionMultiCharOperators(t *testing.T) {
	lx := New("${a>=1 && b?? && c..*d && x**2 && p...}", Config{})
	expectKinds := []int{
		tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_NATURAL_GTE,
		tokenid.TK_INTEGER,
		tokenid.TK_AND,
		tokenid.TK_ID,
		tokenid.TK_EXISTS,
		tokenid.TK_AND,
		tokenid.TK_ID,
		tokenid.TK_DOT_DOT_ASTERISK,
		tokenid.TK_ID,
		tokenid.TK_AND,
		tokenid.TK_ID,
		tokenid.TK_DOUBLE_STAR,
		tokenid.TK_INTEGER,
		tokenid.TK_AND,
		tokenid.TK_ID,
		tokenid.TK_ELLIPSIS,
		tokenid.TK_CLOSING_CURLY_BRACKET,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestFalseAlarmCharacters(t *testing.T) {
	lx := New("$#<[{}", Config{})
	for i := 0; i < 5; i++ {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("next %d: %v", i, err)
		}
		if tok.Kind != tokenid.TK_STATIC_TEXT_FALSE_ALARM {
			t.Fatalf("token %d kind mismatch: %#v", i, tok)
		}
	}
}

func TestIfDirectiveTokenization(t *testing.T) {
	lx := New("<#if x>yes</#if>", Config{})
	expectKinds := []int{
		tokenid.TK_IF,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_IF,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestIfElseIfElseDirectiveTokenization(t *testing.T) {
	lx := New("<#if x>1<#elseif y>2<#else>3</#if>", Config{})
	expectKinds := []int{
		tokenid.TK_IF,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_ELSE_IF,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_ELSE,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_IF,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestSquareBracketIfDirectiveTokenization(t *testing.T) {
	lx := New("[#if x]y[/#if]", Config{})
	expectKinds := []int{
		tokenid.TK_IF,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_IF,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestElseVariantsTokenization(t *testing.T) {
	lx := New("<#if x>1<#else/>0</#if>", Config{})
	expectKinds := []int{
		tokenid.TK_IF,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_ELSE,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_IF,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestElseIfCamelCaseTokenization(t *testing.T) {
	lx := New("<#if x>1<#elseIf y>2</#if>", Config{})
	expectKinds := []int{
		tokenid.TK_IF,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_ELSE_IF,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_IF,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestSquareBracketDirectiveNaturalGt(t *testing.T) {
	lx := New("[#if x > y]ok[/#if]", Config{})
	expectKinds := []int{
		tokenid.TK_IF,
		tokenid.TK_ID,
		tokenid.TK_NATURAL_GT,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_IF,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestAssignmentDirectiveTokenization(t *testing.T) {
	lx := New("<#assign x = 1, y += 2 in ns>", Config{})
	expectKinds := []int{
		tokenid.TK_ASSIGN,
		tokenid.TK_ID,
		tokenid.TK_EQUALS,
		tokenid.TK_INTEGER,
		tokenid.TK_COMMA,
		tokenid.TK_ID,
		tokenid.TK_PLUS_EQUALS,
		tokenid.TK_INTEGER,
		tokenid.TK_IN,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestAssignmentAndMacroBoundaryTokenization(t *testing.T) {
	lx := New("<#macro m><#local x=1></#macro><#assign x>t</#assign>", Config{})
	expectKinds := []int{
		tokenid.TK_MACRO,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_LOCALASSIGN,
		tokenid.TK_ID,
		tokenid.TK_EQUALS,
		tokenid.TK_INTEGER,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_END_MACRO,
		tokenid.TK_ASSIGN,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_ASSIGN,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestListDirectiveTokenization(t *testing.T) {
	lx := New("<#list xs as x></#list>", Config{})
	expectKinds := []int{
		tokenid.TK_LIST,
		tokenid.TK_ID,
		tokenid.TK_AS,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_END_LIST,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestItemsAndSepDirectiveTokenization(t *testing.T) {
	lx := New("<#items as x>${x}<#sep>, </#items>", Config{})
	expectKinds := []int{
		tokenid.TK_ITEMS,
		tokenid.TK_AS,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_CLOSING_CURLY_BRACKET,
		tokenid.TK_SEP,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_STATIC_TEXT_WS,
		tokenid.TK_END_ITEMS,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestSwitchDirectiveTokenization(t *testing.T) {
	lx := New("<#switch x><#case 1>one<#default>other</#switch>", Config{})
	expectKinds := []int{
		tokenid.TK_SWITCH,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_CASE,
		tokenid.TK_INTEGER,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_DEFAUL,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_SWITCH,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestSwitchOnDirectiveTokenization(t *testing.T) {
	lx := New("<#switch x><#on 1,2>12</#switch>", Config{})
	expectKinds := []int{
		tokenid.TK_SWITCH,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_ON,
		tokenid.TK_INTEGER,
		tokenid.TK_COMMA,
		tokenid.TK_INTEGER,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_SWITCH,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestFunctionAndReturnDirectiveTokenization(t *testing.T) {
	lx := New("<#function foo x y><#local x = 1><#return x + y></#function>", Config{})
	expectKinds := []int{
		tokenid.TK_FUNCTION,
		tokenid.TK_ID,
		tokenid.TK_ID,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_LOCALASSIGN,
		tokenid.TK_ID,
		tokenid.TK_EQUALS,
		tokenid.TK_INTEGER,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_RETURN,
		tokenid.TK_ID,
		tokenid.TK_PLUS,
		tokenid.TK_ID,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_END_FUNCTION,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestSimpleReturnTokenization(t *testing.T) {
	lx := New("<#return>", Config{})
	tok, err := lx.Next()
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	if tok.Kind != tokenid.TK_SIMPLE_RETURN {
		t.Fatalf("expected TK_SIMPLE_RETURN, got kind=%d image=%q", tok.Kind, tok.Image)
	}
	eof, err := lx.Next()
	if err != nil {
		t.Fatalf("eof: %v", err)
	}
	if eof.Kind != tokenid.TK_EOF {
		t.Fatalf("expected EOF, got %#v", eof)
	}
}

func TestOutputFormatAndEscapingDirectiveTokenization(t *testing.T) {
	src := "<#outputFormat \"XML\"><#noAutoEsc>${a}<#autoEsc>${b}</#autoEsc>${c}</#noAutoEsc></#outputFormat>"
	lx := New(src, Config{})
	expectKinds := []int{
		tokenid.TK_OUTPUTFORMAT,
		tokenid.TK_STRING_LITERAL,
		tokenid.TK_DIRECTIVE_END,
		tokenid.TK_NOAUTOESC,
		tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_CLOSING_CURLY_BRACKET,
		tokenid.TK_AUTOESC,
		tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_CLOSING_CURLY_BRACKET,
		tokenid.TK_END_AUTOESC,
		tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_CLOSING_CURLY_BRACKET,
		tokenid.TK_END_NOAUTOESC,
		tokenid.TK_END_OUTPUTFORMAT,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestAttemptRecoverTokenization(t *testing.T) {
	lx := New("<#attempt>1<#recover>2</#attempt>", Config{})
	expectKinds := []int{
		tokenid.TK_ATTEMPT,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_RECOVER,
		tokenid.TK_STATIC_TEXT_NON_WS,
		tokenid.TK_END_ATTEMPT,
		tokenid.TK_EOF,
	}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
	}
}

func TestEscapedIdentifierTokenization(t *testing.T) {
	lx := New("${data\\-color}", Config{})
	expectKinds := []int{
		tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
		tokenid.TK_ID,
		tokenid.TK_CLOSING_CURLY_BRACKET,
		tokenid.TK_EOF,
	}
	expectImages := []string{"${", "data\\-color", "}", ""}
	for i, kind := range expectKinds {
		tok, err := lx.Next()
		if err != nil {
			t.Fatalf("token %d: %v", i, err)
		}
		if tok.Kind != kind {
			t.Fatalf("token %d kind mismatch: got %d want %d (%q)", i, tok.Kind, kind, tok.Image)
		}
		if tok.Image != expectImages[i] {
			t.Fatalf("token %d image mismatch: got %q want %q", i, tok.Image, expectImages[i])
		}
	}
}
