package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/weaweawe01/freemarker-ast/internal/compat"
	"github.com/weaweawe01/freemarker-ast/internal/tokenid"
)

// State models JavaCC lexer states used by FreeMarker.
type State string

const (
	StateDefault     State = "DEFAULT"
	StateExpression  State = "FM_EXPRESSION"
	StateInParen     State = "IN_PAREN"
	StateNoParse     State = "NO_PARSE"
	StateNoDirective State = "NO_DIRECTIVE"
)

// Config controls lexer behavior.
type Config struct{}

// Lexer tokenizes a FreeMarker source string.
type Lexer struct {
	src                    string
	offset                 int
	line                   int
	column                 int
	state                  State
	postInterpolationState State
	inInterpolation        bool
	inDirective            bool
	directiveEnd           byte
	parenthesisNesting     int
	bracketNesting         int
	curlyBracketNesting    int
}

// New creates a lexer at DEFAULT state.
func New(src string, _ Config) *Lexer {
	return &Lexer{
		src:                    src,
		offset:                 0,
		line:                   1,
		column:                 1,
		state:                  StateDefault,
		postInterpolationState: StateDefault,
	}
}

// Next returns the next token. EOF is returned as TK_EOF with empty image.
func (l *Lexer) Next() (compat.Token, error) {
	if l.offset >= len(l.src) {
		pos := compat.Position{Line: l.line, Column: l.column}
		return compat.Token{
			Kind:     tokenid.TK_EOF,
			Image:    "",
			Begin:    pos,
			End:      pos,
			LexState: string(l.state),
		}, nil
	}

	switch l.state {
	case StateDefault, StateNoDirective:
		return l.nextDefault()
	case StateExpression, StateInParen:
		return l.nextExpression()
	case StateNoParse:
		return compat.Token{}, fmt.Errorf("lexer state %s not implemented yet", l.state)
	default:
		return compat.Token{}, fmt.Errorf("unknown lexer state: %s", l.state)
	}
}

func (l *Lexer) nextDefault() (compat.Token, error) {
	begin := compat.Position{Line: l.line, Column: l.column}

	if tok, ok := l.matchDirectiveToken(begin); ok {
		return tok, nil
	}

	if l.hasPrefix("${") {
		l.advanceN(2)
		l.startInterpolation()
		return compat.Token{
			Kind:     tokenid.TK_DOLLAR_INTERPOLATION_OPENING,
			Image:    "${",
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, nil
	}
	if l.hasPrefix("#{") {
		l.advanceN(2)
		l.startInterpolation()
		return compat.Token{
			Kind:     tokenid.TK_HASH_INTERPOLATION_OPENING,
			Image:    "#{",
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, nil
	}
	if l.hasPrefix("[=") {
		l.advanceN(2)
		l.startInterpolation()
		return compat.Token{
			Kind:     tokenid.TK_SQUARE_BRACKET_INTERPOLATION_OPENING,
			Image:    "[=",
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, nil
	}

	if isWS(l.peekByte()) {
		image := l.consumeWhile(func(b byte) bool { return isWS(b) })
		return compat.Token{
			Kind:     tokenid.TK_STATIC_TEXT_WS,
			Image:    image,
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, nil
	}

	if isFalseAlarmStart(l.peekByte()) {
		ch := l.src[l.offset : l.offset+1]
		l.advanceN(1)
		return compat.Token{
			Kind:     tokenid.TK_STATIC_TEXT_FALSE_ALARM,
			Image:    ch,
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, nil
	}

	image := l.consumeWhile(func(b byte) bool {
		return !isWS(b) && !isFalseAlarmStart(b)
	})
	return compat.Token{
		Kind:     tokenid.TK_STATIC_TEXT_NON_WS,
		Image:    image,
		Begin:    begin,
		End:      compat.Position{Line: l.line, Column: l.column - 1},
		LexState: string(l.state),
	}, nil
}

func (l *Lexer) nextExpression() (compat.Token, error) {
	l.skipExpressionWhitespace()
	if l.offset >= len(l.src) {
		return compat.Token{}, fmt.Errorf("unexpected EOF in %s state", l.state)
	}

	begin := compat.Position{Line: l.line, Column: l.column}

	if tok, ok := l.matchExpressionToken(begin); ok {
		return tok, nil
	}

	if isASCIIDigit(l.peekByte()) {
		return l.readNumber(begin), nil
	}
	if isIDStart(l.peekByte()) {
		if l.peekByte() == '\\' && (l.offset+1 >= len(l.src) || !isEscapableIDChar(l.src[l.offset+1])) {
			return compat.Token{}, fmt.Errorf("invalid escaped identifier start %q at %d:%d", l.peekByte(), l.line, l.column)
		}
		return l.readIdentifierOrKeyword(begin), nil
	}
	if l.peekByte() == '"' || l.peekByte() == '\'' {
		tok, err := l.readString(begin)
		if err != nil {
			return compat.Token{}, err
		}
		return tok, nil
	}

	return compat.Token{}, fmt.Errorf("unsupported expression character %q at %d:%d", l.peekByte(), l.line, l.column)
}

func (l *Lexer) matchExpressionToken(begin compat.Position) (compat.Token, bool) {
	if l.inDirective {
		if l.directiveEnd == '>' && l.hasPrefix("/>") {
			l.advanceN(2)
			tok := compat.Token{
				Kind:     tokenid.TK_EMPTY_DIRECTIVE_END,
				Image:    "/>",
				Begin:    begin,
				End:      compat.Position{Line: l.line, Column: l.column - 1},
				LexState: string(l.state),
			}
			l.applyStateSideEffects(&tok)
			return tok, true
		}
		if l.directiveEnd == ']' && l.hasPrefix("/]") {
			l.advanceN(2)
			tok := compat.Token{
				Kind:     tokenid.TK_EMPTY_DIRECTIVE_END,
				Image:    "/]",
				Begin:    begin,
				End:      compat.Position{Line: l.line, Column: l.column - 1},
				LexState: string(l.state),
			}
			l.applyStateSideEffects(&tok)
			return tok, true
		}
	}

	type pat struct {
		s    string
		kind int
	}
	patterns := []pat{
		{"...", tokenid.TK_ELLIPSIS},
		{"??", tokenid.TK_EXISTS},
		{"==", tokenid.TK_DOUBLE_EQUALS},
		{"!=", tokenid.TK_NOT_EQUALS},
		{">=", tokenid.TK_NATURAL_GTE},
		{"+=", tokenid.TK_PLUS_EQUALS},
		{"-=", tokenid.TK_MINUS_EQUALS},
		{"*=", tokenid.TK_TIMES_EQUALS},
		{"/=", tokenid.TK_DIV_EQUALS},
		{"%=", tokenid.TK_MOD_EQUALS},
		{"++", tokenid.TK_PLUS_PLUS},
		{"--", tokenid.TK_MINUS_MINUS},
		{"**", tokenid.TK_DOUBLE_STAR},
		{"..*", tokenid.TK_DOT_DOT_ASTERISK},
		{"..<", tokenid.TK_DOT_DOT_LESS},
		{"..!", tokenid.TK_DOT_DOT_LESS},
		{"..", tokenid.TK_DOT_DOT},
		{"<=", tokenid.TK_LESS_THAN_EQUALS},
		{"&&", tokenid.TK_AND},
		{"||", tokenid.TK_OR},
		{"->", tokenid.TK_LAMBDA_ARROW},
		{",", tokenid.TK_COMMA},
		{";", tokenid.TK_SEMICOLON},
		{":", tokenid.TK_COLON},
		{".", tokenid.TK_DOT},
		{"?", tokenid.TK_BUILT_IN},
		{"=", tokenid.TK_EQUALS},
		{"+", tokenid.TK_PLUS},
		{"-", tokenid.TK_MINUS},
		{"*", tokenid.TK_TIMES},
		{"/", tokenid.TK_DIVIDE},
		{"%", tokenid.TK_PERCENT},
		{"!", tokenid.TK_EXCLAM},
		{"<", tokenid.TK_LESS_THAN},
		{"|", tokenid.TK_OR},
		{"&", tokenid.TK_AND},
		{"{", tokenid.TK_OPENING_CURLY_BRACKET},
		{"}", tokenid.TK_CLOSING_CURLY_BRACKET},
		{"(", tokenid.TK_OPEN_PAREN},
		{")", tokenid.TK_CLOSE_PAREN},
		{"[", tokenid.TK_OPEN_BRACKET},
		{"]", tokenid.TK_CLOSE_BRACKET},
		{">", tokenid.TK_DIRECTIVE_END},
	}

	for _, p := range patterns {
		if !l.hasPrefix(p.s) {
			continue
		}
		l.advanceN(len(p.s))
		tok := compat.Token{
			Kind:     p.kind,
			Image:    p.s,
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}
		l.applyStateSideEffects(&tok)
		return tok, true
	}
	return compat.Token{}, false
}

func (l *Lexer) applyStateSideEffects(tok *compat.Token) {
	switch tok.Kind {
	case tokenid.TK_OPEN_BRACKET:
		l.bracketNesting++
	case tokenid.TK_CLOSE_BRACKET:
		if l.bracketNesting > 0 {
			l.bracketNesting--
		} else if l.inInterpolation {
			l.endInterpolation()
		} else if l.inDirective && l.directiveEnd == ']' {
			tok.Kind = tokenid.TK_DIRECTIVE_END
			l.endDirective()
		}
	case tokenid.TK_OPEN_PAREN:
		l.parenthesisNesting++
		l.state = StateInParen
	case tokenid.TK_CLOSE_PAREN:
		if l.parenthesisNesting > 0 {
			l.parenthesisNesting--
		}
		if l.parenthesisNesting == 0 {
			l.state = StateExpression
		}
	case tokenid.TK_OPENING_CURLY_BRACKET:
		l.curlyBracketNesting++
	case tokenid.TK_CLOSING_CURLY_BRACKET:
		if l.curlyBracketNesting > 0 {
			l.curlyBracketNesting--
		} else if l.inInterpolation {
			l.endInterpolation()
		}
	case tokenid.TK_DIRECTIVE_END:
		// In interpolation mode or inside (...), ">" is comparison.
		if l.inInterpolation || l.state == StateInParen {
			tok.Kind = tokenid.TK_NATURAL_GT
			break
		}
		if l.inDirective {
			if l.directiveEnd == ']' {
				tok.Kind = tokenid.TK_NATURAL_GT
				break
			}
			l.endDirective()
			break
		}
		l.state = StateDefault
	case tokenid.TK_EMPTY_DIRECTIVE_END:
		if l.inDirective {
			l.endDirective()
		}
	}
}

func (l *Lexer) startInterpolation() {
	l.postInterpolationState = l.state
	l.inInterpolation = true
	l.state = StateExpression
}

func (l *Lexer) endInterpolation() {
	l.inInterpolation = false
	l.state = l.postInterpolationState
}

func (l *Lexer) startDirective(end byte) {
	l.inDirective = true
	l.directiveEnd = end
	l.state = StateExpression
}

func (l *Lexer) endDirective() {
	l.inDirective = false
	l.directiveEnd = 0
	l.state = StateDefault
}

func (l *Lexer) skipExpressionWhitespace() {
	for l.offset < len(l.src) && isWS(l.src[l.offset]) {
		l.advanceOne()
	}
}

func (l *Lexer) readNumber(begin compat.Position) compat.Token {
	start := l.offset
	for l.offset < len(l.src) && isASCIIDigit(l.src[l.offset]) {
		l.advanceOne()
	}
	kind := tokenid.TK_INTEGER
	if l.offset < len(l.src) && l.src[l.offset] == '.' && l.offset+1 < len(l.src) && isASCIIDigit(l.src[l.offset+1]) {
		l.advanceOne() // dot
		for l.offset < len(l.src) && isASCIIDigit(l.src[l.offset]) {
			l.advanceOne()
		}
		kind = tokenid.TK_DECIMAL
	}
	return compat.Token{
		Kind:     kind,
		Image:    l.src[start:l.offset],
		Begin:    begin,
		End:      compat.Position{Line: l.line, Column: l.column - 1},
		LexState: string(l.state),
	}
}

func (l *Lexer) readIdentifierOrKeyword(begin compat.Position) compat.Token {
	start := l.offset

	for l.offset < len(l.src) {
		if l.src[l.offset] == '\\' {
			if l.offset+1 >= len(l.src) || !isEscapableIDChar(l.src[l.offset+1]) {
				break
			}
			l.advanceOne() // '\'
			l.advanceOne() // escaped char
			continue
		}
		if !isIDContinue(l.src[l.offset]) {
			break
		}
		l.advanceOne()
	}
	image := l.src[start:l.offset]
	kind := tokenid.TK_ID
	switch image {
	case "false":
		kind = tokenid.TK_FALSE
	case "true":
		kind = tokenid.TK_TRUE
	case "in":
		kind = tokenid.TK_IN
	case "as":
		kind = tokenid.TK_AS
	case "using":
		kind = tokenid.TK_USING
	case "lt":
		kind = tokenid.TK_LESS_THAN
	case "lte":
		kind = tokenid.TK_LESS_THAN_EQUALS
	case "gt":
		kind = tokenid.TK_ESCAPED_GT
	case "gte":
		kind = tokenid.TK_ESCAPED_GTE
	}
	return compat.Token{
		Kind:     kind,
		Image:    image,
		Begin:    begin,
		End:      compat.Position{Line: l.line, Column: l.column - 1},
		LexState: string(l.state),
	}
}

func (l *Lexer) readString(begin compat.Position) (compat.Token, error) {
	quote := l.peekByte()
	start := l.offset
	l.advanceOne() // open quote
	for l.offset < len(l.src) {
		b := l.peekByte()
		if b == '\\' {
			l.advanceOne()
			if l.offset < len(l.src) {
				l.advanceOne()
			}
			continue
		}
		l.advanceOne()
		if b == quote {
			return compat.Token{
				Kind:     tokenid.TK_STRING_LITERAL,
				Image:    l.src[start:l.offset],
				Begin:    begin,
				End:      compat.Position{Line: l.line, Column: l.column - 1},
				LexState: string(l.state),
			}, nil
		}
	}
	return compat.Token{}, fmt.Errorf("unterminated string literal at %d:%d", begin.Line, begin.Column)
}

func (l *Lexer) matchDirectiveToken(begin compat.Position) (compat.Token, bool) {
	if tok, ok := l.matchComment(begin); ok {
		return tok, true
	}
	if tok, end, ok := l.matchUnifiedCallStart(begin); ok {
		l.startDirective(end)
		return tok, true
	}
	if tok, ok := l.matchUnifiedCallEnd(begin); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveSimple(begin); ok {
		return tok, true
	}
	if tok, end, ok := l.matchDirectiveExprStart(begin); ok {
		l.startDirective(end)
		return tok, true
	}
	if tok, ok := l.matchDirectiveElse(begin); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveEnd(begin); ok {
		return tok, true
	}
	return compat.Token{}, false
}

func (l *Lexer) matchComment(begin compat.Position) (compat.Token, bool) {
	type form struct {
		open  string
		close string
		kind  int
	}
	forms := []form{
		{open: "<#--", close: "-->", kind: tokenid.TK_COMMENT},
		{open: "[#--", close: "--]", kind: tokenid.TK_COMMENT},
		{open: "<#-", close: "->", kind: tokenid.TK_TERSE_COMMENT},
		{open: "[#-", close: "-]", kind: tokenid.TK_TERSE_COMMENT},
	}

	for _, f := range forms {
		if !l.hasPrefix(f.open) {
			continue
		}
		searchFrom := l.offset + len(f.open)
		rel := strings.Index(l.src[searchFrom:], f.close)
		if rel < 0 {
			return compat.Token{}, false
		}
		endOffset := searchFrom + rel + len(f.close)
		image := l.src[l.offset:endOffset]
		l.advanceN(len(image))
		return compat.Token{
			Kind:     f.kind,
			Image:    image,
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, true
	}

	return compat.Token{}, false
}

func (l *Lexer) matchUnifiedCallStart(begin compat.Position) (compat.Token, byte, bool) {
	forms := []directivePrefixForm{
		{prefix: "<@", end: '>'},
		{prefix: "[@", end: ']'},
	}
	for _, form := range forms {
		if !l.hasPrefix(form.prefix) {
			continue
		}
		// Must be followed by identifier-like char to avoid false alarms.
		if l.offset+len(form.prefix) >= len(l.src) {
			continue
		}
		next := l.src[l.offset+len(form.prefix)]
		if !isIDStart(next) {
			continue
		}

		l.advanceN(len(form.prefix))
		return compat.Token{
			Kind:     tokenid.TK_UNIFIED_CALL,
			Image:    form.prefix,
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, form.end, true
	}
	return compat.Token{}, 0, false
}

func (l *Lexer) matchUnifiedCallEnd(begin compat.Position) (compat.Token, bool) {
	forms := []directivePrefixForm{
		{prefix: "</@", end: '>'},
		{prefix: "[/@", end: ']'},
	}
	for _, form := range forms {
		if !l.hasPrefix(form.prefix) {
			continue
		}
		// End tag can be </@foo>, </@foo.bar>, or </@>.
		endOffset, ok := l.parseUnifiedCallEndTag(l.offset+len(form.prefix), form.end)
		if !ok {
			continue
		}
		image := l.src[l.offset:endOffset]
		l.advanceN(len(image))
		return compat.Token{
			Kind:     tokenid.TK_UNIFIED_CALL_END,
			Image:    image,
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, true
	}
	return compat.Token{}, false
}

func (l *Lexer) parseUnifiedCallEndTag(at int, end byte) (int, bool) {
	i := at
	for i < len(l.src) && isWS(l.src[i]) {
		i++
	}
	// Optional callee name.
	for i < len(l.src) && isUnifiedCallNameChar(l.src[i]) {
		i++
	}
	for i < len(l.src) && isWS(l.src[i]) {
		i++
	}
	if i >= len(l.src) || l.src[i] != end {
		return 0, false
	}
	return i + 1, true
}

func isUnifiedCallNameChar(b byte) bool {
	return isIDContinue(b) || b == '.'
}

func (l *Lexer) matchDirectiveSimple(begin compat.Position) (compat.Token, bool) {
	if tok, ok := l.matchDirectiveKeywordCloseTag1(begin, []string{"sep"}, tokenid.TK_SEP); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveKeywordCloseTag1(begin, []string{"default"}, tokenid.TK_DEFAUL); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveKeywordCloseTag1(begin, []string{"attempt"}, tokenid.TK_ATTEMPT); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveKeywordCloseTag1(begin, []string{"recover"}, tokenid.TK_RECOVER); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveKeywordCloseTag1(begin, []string{"autoesc", "autoEsc"}, tokenid.TK_AUTOESC); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveKeywordCloseTag1(begin, []string{"noautoesc", "noAutoEsc"}, tokenid.TK_NOAUTOESC); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveKeywordCloseTag1(begin, []string{"compress"}, tokenid.TK_COMPRESS); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveKeywordCloseTag2(begin, []string{"return"}, tokenid.TK_SIMPLE_RETURN); ok {
		return tok, true
	}
	if tok, ok := l.matchDirectiveKeywordCloseTag2(begin, []string{"nested"}, tokenid.TK_SIMPLE_NESTED); ok {
		return tok, true
	}
	return compat.Token{}, false
}

func (l *Lexer) matchDirectiveExprStart(begin compat.Position) (compat.Token, byte, bool) {
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"if"}); ok {
		tok.Kind = tokenid.TK_IF
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"elseif", "elseIf"}); ok {
		tok.Kind = tokenid.TK_ELSE_IF
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"assign"}); ok {
		tok.Kind = tokenid.TK_ASSIGN
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"global"}); ok {
		tok.Kind = tokenid.TK_GLOBALASSIGN
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"local"}); ok {
		tok.Kind = tokenid.TK_LOCALASSIGN
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"macro"}); ok {
		tok.Kind = tokenid.TK_MACRO
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"function"}); ok {
		tok.Kind = tokenid.TK_FUNCTION
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"return"}); ok {
		tok.Kind = tokenid.TK_RETURN
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"outputformat", "outputFormat"}); ok {
		tok.Kind = tokenid.TK_OUTPUTFORMAT
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"list"}); ok {
		tok.Kind = tokenid.TK_LIST
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"items"}); ok {
		tok.Kind = tokenid.TK_ITEMS
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"switch"}); ok {
		tok.Kind = tokenid.TK_SWITCH
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"case"}); ok {
		tok.Kind = tokenid.TK_CASE
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"on"}); ok {
		tok.Kind = tokenid.TK_ON
		return tok, end, true
	}
	if tok, end, ok := l.matchDirectiveKeywordWithBlank(begin, []string{"nested"}); ok {
		tok.Kind = tokenid.TK_NESTED
		return tok, end, true
	}
	return compat.Token{}, 0, false
}

func (l *Lexer) matchDirectiveElse(begin compat.Position) (compat.Token, bool) {
	forms := directiveStartPrefixes()
	for _, form := range forms {
		full := form.prefix + "else"
		if !l.hasPrefix(full) {
			continue
		}
		endOffset, ok := l.parseCloseTag2(l.offset + len(full))
		if !ok {
			continue
		}
		image := l.src[l.offset:endOffset]
		l.advanceN(len(image))
		return compat.Token{
			Kind:     tokenid.TK_ELSE,
			Image:    image,
			Begin:    begin,
			End:      compat.Position{Line: l.line, Column: l.column - 1},
			LexState: string(l.state),
		}, true
	}
	return compat.Token{}, false
}

func (l *Lexer) matchDirectiveEnd(begin compat.Position) (compat.Token, bool) {
	endTags := []struct {
		name string
		kind int
	}{
		{name: "if", kind: tokenid.TK_END_IF},
		{name: "assign", kind: tokenid.TK_END_ASSIGN},
		{name: "global", kind: tokenid.TK_END_GLOBAL},
		{name: "local", kind: tokenid.TK_END_LOCAL},
		{name: "macro", kind: tokenid.TK_END_MACRO},
		{name: "function", kind: tokenid.TK_END_FUNCTION},
		{name: "list", kind: tokenid.TK_END_LIST},
		{name: "items", kind: tokenid.TK_END_ITEMS},
		{name: "sep", kind: tokenid.TK_END_SEP},
		{name: "switch", kind: tokenid.TK_END_SWITCH},
		{name: "outputformat", kind: tokenid.TK_END_OUTPUTFORMAT},
		{name: "outputFormat", kind: tokenid.TK_END_OUTPUTFORMAT},
		{name: "autoesc", kind: tokenid.TK_END_AUTOESC},
		{name: "autoEsc", kind: tokenid.TK_END_AUTOESC},
		{name: "noautoesc", kind: tokenid.TK_END_NOAUTOESC},
		{name: "noAutoEsc", kind: tokenid.TK_END_NOAUTOESC},
		{name: "compress", kind: tokenid.TK_END_COMPRESS},
		{name: "attempt", kind: tokenid.TK_END_ATTEMPT},
		{name: "recover", kind: tokenid.TK_END_RECOVER},
	}

	forms := directiveEndPrefixes()
	for _, form := range forms {
		for _, endTag := range endTags {
			full := form.prefix + endTag.name
			if !l.hasPrefix(full) {
				continue
			}
			endOffset, ok := l.parseCloseTag1(l.offset + len(full))
			if !ok {
				continue
			}
			image := l.src[l.offset:endOffset]
			l.advanceN(len(image))
			return compat.Token{
				Kind:     endTag.kind,
				Image:    image,
				Begin:    begin,
				End:      compat.Position{Line: l.line, Column: l.column - 1},
				LexState: string(l.state),
			}, true
		}
	}
	return compat.Token{}, false
}

type directivePrefixForm struct {
	prefix string
	end    byte
}

func directiveStartPrefixes() []directivePrefixForm {
	return []directivePrefixForm{
		{prefix: "<#", end: '>'},
		{prefix: "[#", end: ']'},
		{prefix: "<", end: '>'},
	}
}

func directiveEndPrefixes() []directivePrefixForm {
	return []directivePrefixForm{
		{prefix: "</#", end: '>'},
		{prefix: "[/#", end: ']'},
		{prefix: "</", end: '>'},
	}
}

func (l *Lexer) matchDirectiveKeywordWithBlank(
	begin compat.Position,
	keywords []string,
) (compat.Token, byte, bool) {
	for _, form := range directiveStartPrefixes() {
		for _, kw := range keywords {
			full := form.prefix + kw
			if !l.hasPrefix(full) {
				continue
			}
			if l.offset+len(full) >= len(l.src) || !isWS(l.src[l.offset+len(full)]) {
				continue
			}
			endOffset := l.offset + len(full)
			for endOffset < len(l.src) && isWS(l.src[endOffset]) {
				endOffset++
			}
			image := l.src[l.offset:endOffset]
			l.advanceN(len(image))
			return compat.Token{
				Image:    image,
				Begin:    begin,
				End:      compat.Position{Line: l.line, Column: l.column - 1},
				LexState: string(l.state),
			}, form.end, true
		}
	}
	return compat.Token{}, 0, false
}

func (l *Lexer) matchDirectiveKeywordCloseTag1(
	begin compat.Position,
	keywords []string,
	kind int,
) (compat.Token, bool) {
	for _, form := range directiveStartPrefixes() {
		for _, kw := range keywords {
			full := form.prefix + kw
			if !l.hasPrefix(full) {
				continue
			}
			endOffset, ok := l.parseCloseTag1(l.offset + len(full))
			if !ok {
				continue
			}
			image := l.src[l.offset:endOffset]
			l.advanceN(len(image))
			return compat.Token{
				Kind:     kind,
				Image:    image,
				Begin:    begin,
				End:      compat.Position{Line: l.line, Column: l.column - 1},
				LexState: string(l.state),
			}, true
		}
	}
	return compat.Token{}, false
}

func (l *Lexer) matchDirectiveKeywordCloseTag2(
	begin compat.Position,
	keywords []string,
	kind int,
) (compat.Token, bool) {
	for _, form := range directiveStartPrefixes() {
		for _, kw := range keywords {
			full := form.prefix + kw
			if !l.hasPrefix(full) {
				continue
			}
			endOffset, ok := l.parseCloseTag2(l.offset + len(full))
			if !ok {
				continue
			}
			image := l.src[l.offset:endOffset]
			l.advanceN(len(image))
			return compat.Token{
				Kind:     kind,
				Image:    image,
				Begin:    begin,
				End:      compat.Position{Line: l.line, Column: l.column - 1},
				LexState: string(l.state),
			}, true
		}
	}
	return compat.Token{}, false
}

func (l *Lexer) parseCloseTag1(at int) (int, bool) {
	i := at
	for i < len(l.src) && isWS(l.src[i]) {
		i++
	}
	if i >= len(l.src) {
		return 0, false
	}
	if l.src[i] != '>' && l.src[i] != ']' {
		return 0, false
	}
	return i + 1, true
}

func (l *Lexer) parseCloseTag2(at int) (int, bool) {
	i := at
	for i < len(l.src) && isWS(l.src[i]) {
		i++
	}
	if i < len(l.src) && l.src[i] == '/' {
		i++
	}
	if i >= len(l.src) {
		return 0, false
	}
	if l.src[i] != '>' && l.src[i] != ']' {
		return 0, false
	}
	return i + 1, true
}

func (l *Lexer) hasPrefix(s string) bool {
	return len(l.src)-l.offset >= len(s) && l.src[l.offset:l.offset+len(s)] == s
}

func (l *Lexer) peekByte() byte {
	return l.src[l.offset]
}

func (l *Lexer) consumeWhile(pred func(byte) bool) string {
	start := l.offset
	for l.offset < len(l.src) && pred(l.src[l.offset]) {
		l.advanceOne()
	}
	return l.src[start:l.offset]
}

func (l *Lexer) advanceN(n int) {
	for i := 0; i < n; i++ {
		l.advanceOne()
	}
}

func (l *Lexer) advanceOne() {
	if l.offset >= len(l.src) {
		return
	}

	r, size := utf8.DecodeRuneInString(l.src[l.offset:])
	if r == utf8.RuneError && size == 1 {
		// Keep progress even for malformed UTF-8.
		l.offset++
		l.column++
		return
	}
	l.offset += size
	if r == '\n' {
		l.line++
		l.column = 1
		return
	}
	l.column++
}

func isWS(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isFalseAlarmStart(b byte) bool {
	return b == '$' || b == '#' || b == '<' || b == '[' || b == '{'
}

func isASCIIDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isIDStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == '$' || b == '\\' || b == '@'
}

func isIDContinue(b byte) bool {
	return isIDStart(b) || isASCIIDigit(b)
}

func isEscapableIDChar(b byte) bool {
	return b == '-' || b == '.' || b == ':' || b == '#'
}
