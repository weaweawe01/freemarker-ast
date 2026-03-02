package parser

import (
	"fmt"
	"strings"

	"github.com/weaweawe01/freemarker-ast/internal/ast"
	"github.com/weaweawe01/freemarker-ast/internal/lexer"
	"github.com/weaweawe01/freemarker-ast/internal/tokenid"
)

// Parser parses template content into a minimal AST.
type Parser struct {
	lx      *lexer.Lexer
	lookBuf []token
}

type token struct {
	kind  int
	image string
}

type stopSet map[int]struct{}

const (
	scopeAssign = "assign"
	scopeGlobal = "global"
	scopeLocal  = "local"
)

type parseNodesOptions struct {
	allowItems bool
	allowSep   bool
}

// Parse parses full template source to root AST.
func Parse(src string) (*ast.Root, error) {
	p := &Parser{
		lx: lexer.New(src, lexer.Config{}),
	}
	return p.parseTemplate()
}

// ParseExpressionString parses an expression by wrapping it into an interpolation.
func ParseExpressionString(src string) (ast.Expr, error) {
	root, err := Parse("${" + src + "}")
	if err != nil {
		return nil, err
	}
	if len(root.Children) != 1 {
		return nil, fmt.Errorf("expected single interpolation root child, got %d", len(root.Children))
	}
	interp, ok := root.Children[0].(*ast.Interpolation)
	if !ok {
		return nil, fmt.Errorf("expected interpolation root child, got %T", root.Children[0])
	}
	return interp.Expr, nil
}

func (p *Parser) parseTemplate() (*ast.Root, error) {
	children, stopTok, hasStop, err := p.parseNodesUntil(nil)
	if err != nil {
		return nil, err
	}
	if hasStop && stopTok.kind != tokenid.TK_EOF {
		return nil, fmt.Errorf("unexpected trailing token: kind=%d image=%q", stopTok.kind, stopTok.image)
	}
	return &ast.Root{Children: children}, nil
}

func appendTextNode(nodes *[]ast.Node, text string) {
	if len(*nodes) == 0 {
		*nodes = append(*nodes, &ast.Text{Value: text})
		return
	}
	if last, ok := (*nodes)[len(*nodes)-1].(*ast.Text); ok {
		last.Value += text
		return
	}
	*nodes = append(*nodes, &ast.Text{Value: text})
}

func makeStopSet(kinds ...int) stopSet {
	if len(kinds) == 0 {
		return nil
	}
	s := make(stopSet, len(kinds))
	for _, k := range kinds {
		s[k] = struct{}{}
	}
	return s
}

func (p *Parser) parseNodesUntil(stops stopSet) ([]ast.Node, token, bool, error) {
	return p.parseNodesUntilWithOptions(stops, parseNodesOptions{})
}

func (p *Parser) parseNodesUntilWithOptions(stops stopSet, opts parseNodesOptions) ([]ast.Node, token, bool, error) {
	var nodes []ast.Node

	for {
		tok, err := p.peek()
		if err != nil {
			return nil, token{}, false, err
		}

		if stops != nil {
			if _, ok := stops[tok.kind]; ok {
				return nodes, tok, true, nil
			}
		}

		tok, err = p.next()
		if err != nil {
			return nil, token{}, false, err
		}

		switch tok.kind {
		case tokenid.TK_EOF:
			if stops != nil {
				return nil, token{}, false, fmt.Errorf("unexpected EOF before closing block")
			}
			return nodes, tok, true, nil
		case tokenid.TK_STATIC_TEXT_WS, tokenid.TK_STATIC_TEXT_NON_WS, tokenid.TK_STATIC_TEXT_FALSE_ALARM:
			appendTextNode(&nodes, tok.image)
		case tokenid.TK_DOLLAR_INTERPOLATION_OPENING, tokenid.TK_HASH_INTERPOLATION_OPENING, tokenid.TK_SQUARE_BRACKET_INTERPOLATION_OPENING:
			interp, err := p.parseInterpolation(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, interp)
		case tokenid.TK_IF:
			ifNode, err := p.parseIf(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, ifNode)
		case tokenid.TK_ASSIGN, tokenid.TK_GLOBALASSIGN, tokenid.TK_LOCALASSIGN:
			assignNode, err := p.parseAssignment(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, assignNode)
		case tokenid.TK_MACRO:
			macroNode, err := p.parseMacro(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, macroNode)
		case tokenid.TK_FUNCTION:
			functionNode, err := p.parseFunction(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, functionNode)
		case tokenid.TK_RETURN, tokenid.TK_SIMPLE_RETURN:
			returnNode, err := p.parseReturn(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, returnNode)
		case tokenid.TK_OUTPUTFORMAT:
			block, err := p.parseOutputFormat(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, block)
		case tokenid.TK_AUTOESC:
			block, err := p.parseAutoEsc(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, block)
		case tokenid.TK_NOAUTOESC:
			block, err := p.parseNoAutoEsc(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, block)
		case tokenid.TK_ATTEMPT:
			attempt, err := p.parseAttempt(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, attempt)
		case tokenid.TK_LIST:
			listNode, err := p.parseList(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, listNode)
		case tokenid.TK_SWITCH:
			switchNode, err := p.parseSwitch(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, switchNode)
		case tokenid.TK_UNIFIED_CALL:
			callNode, err := p.parseUnifiedCall(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, callNode)
		case tokenid.TK_COMMENT, tokenid.TK_TERSE_COMMENT:
			nodes = append(nodes, &ast.Comment{Content: extractCommentContent(tok.image)})
		case tokenid.TK_NESTED, tokenid.TK_SIMPLE_NESTED:
			nestedNode, err := p.parseNested(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, nestedNode)
		case tokenid.TK_ITEMS:
			if !opts.allowItems {
				return nil, token{}, false, fmt.Errorf("unexpected items token outside list context: kind=%d image=%q", tok.kind, tok.image)
			}
			itemsNode, err := p.parseItems(tok)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, itemsNode)
		case tokenid.TK_SEP:
			if !opts.allowSep {
				return nil, token{}, false, fmt.Errorf("unexpected sep token outside list/items context: kind=%d image=%q", tok.kind, tok.image)
			}
			sepNode, err := p.parseSep(stops)
			if err != nil {
				return nil, token{}, false, err
			}
			nodes = append(nodes, sepNode)
		case tokenid.TK_ELSE_IF, tokenid.TK_ELSE, tokenid.TK_END_IF:
			return nil, token{}, false, fmt.Errorf("unexpected if-control token: kind=%d image=%q", tok.kind, tok.image)
		case tokenid.TK_END_ASSIGN, tokenid.TK_END_GLOBAL, tokenid.TK_END_LOCAL:
			return nil, token{}, false, fmt.Errorf("unexpected assign-control token: kind=%d image=%q", tok.kind, tok.image)
		case tokenid.TK_END_MACRO:
			return nil, token{}, false, fmt.Errorf("unexpected macro-control token: kind=%d image=%q", tok.kind, tok.image)
		case tokenid.TK_END_FUNCTION:
			return nil, token{}, false, fmt.Errorf("unexpected function-control token: kind=%d image=%q", tok.kind, tok.image)
		case tokenid.TK_END_LIST, tokenid.TK_END_ITEMS, tokenid.TK_END_SEP:
			return nil, token{}, false, fmt.Errorf("unexpected list-control token: kind=%d image=%q", tok.kind, tok.image)
		case tokenid.TK_CASE, tokenid.TK_ON, tokenid.TK_DEFAUL, tokenid.TK_END_SWITCH:
			return nil, token{}, false, fmt.Errorf("unexpected switch-control token: kind=%d image=%q", tok.kind, tok.image)
		case tokenid.TK_UNIFIED_CALL_END:
			return nil, token{}, false, fmt.Errorf("unexpected unified-call end token: kind=%d image=%q", tok.kind, tok.image)
		case tokenid.TK_END_OUTPUTFORMAT, tokenid.TK_END_AUTOESC, tokenid.TK_END_NOAUTOESC:
			return nil, token{}, false, fmt.Errorf("unexpected escaping-control token: kind=%d image=%q", tok.kind, tok.image)
		case tokenid.TK_RECOVER, tokenid.TK_END_RECOVER, tokenid.TK_END_ATTEMPT:
			return nil, token{}, false, fmt.Errorf("unexpected attempt-control token: kind=%d image=%q", tok.kind, tok.image)
		default:
			return nil, token{}, false, fmt.Errorf("unexpected token at template level: kind=%d image=%q", tok.kind, tok.image)
		}
	}
}

func (p *Parser) parseUnifiedCall(start token) (*ast.UnifiedCall, error) {
	callee, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	var positional []ast.Expr
	var named []*ast.NamedArg
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind == tokenid.TK_DIRECTIVE_END || tok.kind == tokenid.TK_EMPTY_DIRECTIVE_END || tok.kind == tokenid.TK_SEMICOLON {
			break
		}
		if tok.kind == tokenid.TK_COMMA {
			if _, err := p.next(); err != nil {
				return nil, err
			}
			continue
		}

		if tok.kind == tokenid.TK_ID {
			look2, err := p.peekN(2)
			if err != nil {
				return nil, err
			}
			if look2.kind == tokenid.TK_EQUALS {
				nameTok, err := p.next()
				if err != nil {
					return nil, err
				}
				if _, err := p.next(); err != nil {
					return nil, err
				}
				value, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				named = append(named, &ast.NamedArg{Name: nameTok.image, Value: value})
				continue
			}
		}

		if !canStartExpression(tok.kind) {
			return nil, fmt.Errorf("expected unified call argument expression, got kind=%d image=%q", tok.kind, tok.image)
		}
		argExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		positional = append(positional, argExpr)
	}

	var loopVars []string
	peekTok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if peekTok.kind == tokenid.TK_SEMICOLON {
		if _, err := p.next(); err != nil {
			return nil, err
		}
		for {
			tok, err := p.peek()
			if err != nil {
				return nil, err
			}
			if tok.kind == tokenid.TK_DIRECTIVE_END || tok.kind == tokenid.TK_EMPTY_DIRECTIVE_END {
				break
			}
			if tok.kind == tokenid.TK_COMMA {
				if _, err := p.next(); err != nil {
					return nil, err
				}
				continue
			}
			idTok, err := p.next()
			if err != nil {
				return nil, err
			}
			if idTok.kind != tokenid.TK_ID {
				return nil, fmt.Errorf("expected loop variable ID after ';' in unified call, got kind=%d image=%q", idTok.kind, idTok.image)
			}
			loopVars = append(loopVars, idTok.image)
		}
	}

	headerEndTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if headerEndTok.kind != tokenid.TK_DIRECTIVE_END && headerEndTok.kind != tokenid.TK_EMPTY_DIRECTIVE_END {
		return nil, fmt.Errorf("expected directive end after unified call header, got kind=%d image=%q", headerEndTok.kind, headerEndTok.image)
	}
	selfClosing := headerEndTok.kind == tokenid.TK_EMPTY_DIRECTIVE_END
	if selfClosing {
		return &ast.UnifiedCall{
			Callee:     callee,
			Positional: positional,
			Named:      named,
			LoopVars:   loopVars,
		}, nil
	}

	children, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_UNIFIED_CALL_END))
	if err != nil {
		return nil, err
	}
	if !hasStop || stopTok.kind != tokenid.TK_UNIFIED_CALL_END {
		return nil, fmt.Errorf("unified call body not closed with unified-call end")
	}
	if _, err := p.next(); err != nil {
		return nil, err
	}

	return &ast.UnifiedCall{
		Callee:     callee,
		Positional: positional,
		Named:      named,
		LoopVars:   loopVars,
		Children:   children,
	}, nil
}

func (p *Parser) parseInterpolation(openTok token) (*ast.Interpolation, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.skipInterpolationTrailingOptions(openTok.kind); err != nil {
		return nil, err
	}
	if err := p.expectInterpolationClosing(openTok.kind); err != nil {
		return nil, err
	}
	return &ast.Interpolation{
		Opening: openTok.image,
		Expr:    expr,
	}, nil
}

func (p *Parser) skipInterpolationTrailingOptions(openingKind int) error {
	tok, err := p.peek()
	if err != nil {
		return err
	}
	if tok.kind != tokenid.TK_SEMICOLON {
		return nil
	}
	closingKind, err := interpolationClosingKind(openingKind)
	if err != nil {
		return err
	}
	for {
		tok, err := p.peek()
		if err != nil {
			return err
		}
		if tok.kind == closingKind {
			return nil
		}
		if tok.kind == tokenid.TK_EOF {
			return fmt.Errorf("unexpected EOF in interpolation options")
		}
		if _, err := p.next(); err != nil {
			return err
		}
	}
}

func (p *Parser) parseIf(start token) (*ast.If, error) {
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}

	ifNode := &ast.If{}
	first := &ast.IfBranch{Condition: cond}
	children, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_ELSE_IF, tokenid.TK_ELSE, tokenid.TK_END_IF))
	if err != nil {
		return nil, err
	}
	if !hasStop {
		return nil, fmt.Errorf("missing if terminator")
	}
	first.Children = children
	ifNode.Branches = append(ifNode.Branches, first)

	for {
		switch stopTok.kind {
		case tokenid.TK_ELSE_IF:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			branchCond, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expectDirectiveEnd(tokenid.TK_ELSE_IF); err != nil {
				return nil, err
			}
			branchChildren, nextStop, hasNextStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_ELSE_IF, tokenid.TK_ELSE, tokenid.TK_END_IF))
			if err != nil {
				return nil, err
			}
			if !hasNextStop {
				return nil, fmt.Errorf("missing elseif terminator")
			}
			ifNode.Branches = append(ifNode.Branches, &ast.IfBranch{
				Condition: branchCond,
				Children:  branchChildren,
			})
			stopTok = nextStop
		case tokenid.TK_ELSE:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			elseChildren, nextStop, hasNextStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_IF))
			if err != nil {
				return nil, err
			}
			if !hasNextStop || nextStop.kind != tokenid.TK_END_IF {
				return nil, fmt.Errorf("else block not closed with end_if")
			}
			ifNode.Else = elseChildren
			if _, err := p.next(); err != nil {
				return nil, err
			}
			return ifNode, nil
		case tokenid.TK_END_IF:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			return ifNode, nil
		default:
			return nil, fmt.Errorf("unexpected if block stop token: kind=%d image=%q", stopTok.kind, stopTok.image)
		}
	}
}

func (p *Parser) expectDirectiveEnd(openingKind int) error {
	tok, err := p.next()
	if err != nil {
		return err
	}
	if tok.kind != tokenid.TK_DIRECTIVE_END {
		return fmt.Errorf("expected directive end after kind=%d, got kind=%d image=%q", openingKind, tok.kind, tok.image)
	}
	return nil
}

func (p *Parser) parseAssignment(start token) (ast.Node, error) {
	scope, endKind, err := assignmentScopeFromToken(start.kind)
	if err != nil {
		return nil, err
	}

	targetTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if !isAssignmentTargetToken(targetTok.kind) {
		return nil, fmt.Errorf("expected assignment target after %s, got kind=%d image=%q", scope, targetTok.kind, targetTok.image)
	}

	nextTok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if nextTok.kind == tokenid.TK_DIRECTIVE_END {
		if _, err := p.next(); err != nil {
			return nil, err
		}
		children, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(endKind))
		if err != nil {
			return nil, err
		}
		if !hasStop || stopTok.kind != endKind {
			return nil, fmt.Errorf("capture assignment not closed with expected end token kind=%d", endKind)
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		return &ast.AssignBlock{
			Scope:    scope,
			Target:   targetTok.image,
			Children: children,
		}, nil
	}

	firstItem, err := p.parseAssignmentItemFromTarget(targetTok)
	if err != nil {
		return nil, err
	}
	items := []*ast.AssignmentItem{firstItem}

	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind == tokenid.TK_COMMA {
			if _, err := p.next(); err != nil {
				return nil, err
			}
		}

		nextTargetTok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if !isAssignmentTargetToken(nextTargetTok.kind) {
			break
		}

		target, err := p.next()
		if err != nil {
			return nil, err
		}
		item, err := p.parseAssignmentItemFromTarget(target)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	var namespace ast.Expr
	maybeIn, err := p.peek()
	if err != nil {
		return nil, err
	}
	if maybeIn.kind == tokenid.TK_IN {
		if _, err := p.next(); err != nil {
			return nil, err
		}
		nsExpr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		namespace = nsExpr
	}

	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}
	return &ast.Assignment{
		Scope:     scope,
		Items:     items,
		Namespace: namespace,
	}, nil
}

func (p *Parser) parseAssignmentItemFromTarget(target token) (*ast.AssignmentItem, error) {
	opTok, err := p.next()
	if err != nil {
		return nil, err
	}

	item := &ast.AssignmentItem{
		Target: normalizeAssignmentTarget(target),
		Op:     opTok.image,
	}

	switch opTok.kind {
	case tokenid.TK_EQUALS, tokenid.TK_PLUS_EQUALS, tokenid.TK_MINUS_EQUALS, tokenid.TK_TIMES_EQUALS, tokenid.TK_DIV_EQUALS, tokenid.TK_MOD_EQUALS:
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		item.Value = value
		return item, nil
	case tokenid.TK_PLUS_PLUS, tokenid.TK_MINUS_MINUS:
		return item, nil
	default:
		return nil, fmt.Errorf("unsupported assignment operator kind=%d image=%q", opTok.kind, opTok.image)
	}
}

func (p *Parser) parseParamDefsUntilDirectiveEnd() ([]*ast.ParamDef, string, error) {
	var params []*ast.ParamDef
	var catchAll string
	inParen := false

	for {
		tok, err := p.peek()
		if err != nil {
			return nil, "", err
		}
		if tok.kind == tokenid.TK_OPEN_PAREN {
			if _, err := p.next(); err != nil {
				return nil, "", err
			}
			inParen = true
			continue
		}
		if tok.kind == tokenid.TK_CLOSE_PAREN && inParen {
			if _, err := p.next(); err != nil {
				return nil, "", err
			}
			inParen = false
			continue
		}
		if tok.kind == tokenid.TK_DIRECTIVE_END {
			return params, catchAll, nil
		}
		if tok.kind == tokenid.TK_COMMA {
			if _, err := p.next(); err != nil {
				return nil, "", err
			}
			continue
		}

		nameTok, err := p.next()
		if err != nil {
			return nil, "", err
		}
		if nameTok.kind != tokenid.TK_ID && nameTok.kind != tokenid.TK_STRING_LITERAL {
			return nil, "", fmt.Errorf("expected parameter name, got kind=%d image=%q", nameTok.kind, nameTok.image)
		}
		paramName := normalizeNameToken(nameTok)

		nextTok, err := p.peek()
		if err != nil {
			return nil, "", err
		}
		if nextTok.kind == tokenid.TK_ELLIPSIS {
			if _, err := p.next(); err != nil {
				return nil, "", err
			}
			catchAll = paramName
			continue
		}

		var def ast.Expr
		if nextTok.kind == tokenid.TK_EQUALS {
			if _, err := p.next(); err != nil {
				return nil, "", err
			}
			expr, err := p.parseExpression()
			if err != nil {
				return nil, "", err
			}
			def = expr
		}

		params = append(params, &ast.ParamDef{
			Name:    paramName,
			Default: def,
		})
	}
}

func (p *Parser) parseMacro(start token) (*ast.Macro, error) {
	nameTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if nameTok.kind != tokenid.TK_ID && nameTok.kind != tokenid.TK_STRING_LITERAL {
		return nil, fmt.Errorf("expected macro name, got kind=%d image=%q", nameTok.kind, nameTok.image)
	}

	params, catchAll, err := p.parseParamDefsUntilDirectiveEnd()
	if err != nil {
		return nil, err
	}

	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}
	children, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_MACRO))
	if err != nil {
		return nil, err
	}
	if !hasStop || stopTok.kind != tokenid.TK_END_MACRO {
		return nil, fmt.Errorf("macro body not closed with end_macro")
	}
	if _, err := p.next(); err != nil {
		return nil, err
	}
	return &ast.Macro{
		Name:       normalizeNameToken(nameTok),
		Params:     params,
		CatchAll:   catchAll,
		IsFunction: false,
		Children:   children,
	}, nil
}

func (p *Parser) parseFunction(start token) (*ast.Function, error) {
	nameTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if nameTok.kind != tokenid.TK_ID && nameTok.kind != tokenid.TK_STRING_LITERAL {
		return nil, fmt.Errorf("expected function name, got kind=%d image=%q", nameTok.kind, nameTok.image)
	}

	params, catchAll, err := p.parseParamDefsUntilDirectiveEnd()
	if err != nil {
		return nil, err
	}

	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}
	children, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_FUNCTION))
	if err != nil {
		return nil, err
	}
	if !hasStop || stopTok.kind != tokenid.TK_END_FUNCTION {
		return nil, fmt.Errorf("function body not closed with end_function")
	}
	if _, err := p.next(); err != nil {
		return nil, err
	}
	return &ast.Function{
		Name:     normalizeNameToken(nameTok),
		Params:   params,
		CatchAll: catchAll,
		Children: children,
	}, nil
}

func (p *Parser) parseNested(start token) (*ast.Nested, error) {
	if start.kind == tokenid.TK_SIMPLE_NESTED {
		return &ast.Nested{}, nil
	}
	if start.kind != tokenid.TK_NESTED {
		return nil, fmt.Errorf("unexpected nested token kind=%d image=%q", start.kind, start.image)
	}

	var vals []ast.Expr
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind == tokenid.TK_DIRECTIVE_END || tok.kind == tokenid.TK_EMPTY_DIRECTIVE_END {
			break
		}
		if tok.kind == tokenid.TK_COMMA {
			if _, err := p.next(); err != nil {
				return nil, err
			}
			continue
		}
		if !canStartExpression(tok.kind) {
			return nil, fmt.Errorf("expected expression in nested arguments, got kind=%d image=%q", tok.kind, tok.image)
		}
		v, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}

	endTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if endTok.kind != tokenid.TK_DIRECTIVE_END && endTok.kind != tokenid.TK_EMPTY_DIRECTIVE_END {
		return nil, fmt.Errorf("expected directive end after nested arguments, got kind=%d image=%q", endTok.kind, endTok.image)
	}
	return &ast.Nested{Values: vals}, nil
}

func (p *Parser) parseReturn(start token) (*ast.Return, error) {
	if start.kind == tokenid.TK_SIMPLE_RETURN {
		return &ast.Return{}, nil
	}
	if start.kind != tokenid.TK_RETURN {
		return nil, fmt.Errorf("unsupported return token kind=%d", start.kind)
	}

	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.kind == tokenid.TK_DIRECTIVE_END {
		if err := p.expectDirectiveEnd(start.kind); err != nil {
			return nil, err
		}
		return &ast.Return{}, nil
	}
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}
	return &ast.Return{Value: value}, nil
}

func (p *Parser) parseOutputFormat(start token) (*ast.OutputFormat, error) {
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}
	children, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_OUTPUTFORMAT))
	if err != nil {
		return nil, err
	}
	if !hasStop || stopTok.kind != tokenid.TK_END_OUTPUTFORMAT {
		return nil, fmt.Errorf("outputFormat block not closed with end_outputformat")
	}
	if _, err := p.next(); err != nil {
		return nil, err
	}
	return &ast.OutputFormat{
		Value:    value,
		Children: children,
	}, nil
}

func (p *Parser) parseAutoEsc(start token) (*ast.AutoEsc, error) {
	children, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_AUTOESC))
	if err != nil {
		return nil, err
	}
	if !hasStop || stopTok.kind != tokenid.TK_END_AUTOESC {
		return nil, fmt.Errorf("autoEsc block not closed with end_autoesc")
	}
	if _, err := p.next(); err != nil {
		return nil, err
	}
	return &ast.AutoEsc{Children: children}, nil
}

func (p *Parser) parseNoAutoEsc(start token) (*ast.NoAutoEsc, error) {
	children, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_NOAUTOESC))
	if err != nil {
		return nil, err
	}
	if !hasStop || stopTok.kind != tokenid.TK_END_NOAUTOESC {
		return nil, fmt.Errorf("noAutoEsc block not closed with end_noautoesc")
	}
	if _, err := p.next(); err != nil {
		return nil, err
	}
	return &ast.NoAutoEsc{Children: children}, nil
}

func (p *Parser) parseAttempt(start token) (*ast.Attempt, error) {
	attemptChildren, stopTok, hasStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_RECOVER, tokenid.TK_END_ATTEMPT))
	if err != nil {
		return nil, err
	}
	if !hasStop {
		return nil, fmt.Errorf("attempt block missing terminator")
	}

	switch stopTok.kind {
	case tokenid.TK_END_ATTEMPT:
		if _, err := p.next(); err != nil {
			return nil, err
		}
		return &ast.Attempt{Attempt: attemptChildren}, nil
	case tokenid.TK_RECOVER:
		if _, err := p.next(); err != nil {
			return nil, err
		}
		recoverChildren, nextStop, hasNextStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_ATTEMPT, tokenid.TK_END_RECOVER))
		if err != nil {
			return nil, err
		}
		if !hasNextStop {
			return nil, fmt.Errorf("attempt recover block missing terminator")
		}
		if nextStop.kind == tokenid.TK_END_RECOVER {
			if _, err := p.next(); err != nil {
				return nil, err
			}
			extraChildren, extraStop, hasExtraStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_ATTEMPT))
			if err != nil {
				return nil, err
			}
			if !hasExtraStop || extraStop.kind != tokenid.TK_END_ATTEMPT {
				return nil, fmt.Errorf("attempt recover should be followed by end_attempt")
			}
			recoverChildren = append(recoverChildren, extraChildren...)
			nextStop = extraStop
		}
		if nextStop.kind != tokenid.TK_END_ATTEMPT {
			return nil, fmt.Errorf("attempt block not closed with end_attempt")
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		// Preserve explicit #recover presence even when body is empty.
		if recoverChildren == nil {
			recoverChildren = []ast.Node{}
		}
		return &ast.Attempt{
			Attempt: attemptChildren,
			Recover: recoverChildren,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected attempt stop token: kind=%d image=%q", stopTok.kind, stopTok.image)
	}
}

func (p *Parser) parseList(start token) (*ast.List, error) {
	source, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	var loopVar string
	nextTok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if nextTok.kind == tokenid.TK_AS {
		if _, err := p.next(); err != nil {
			return nil, err
		}
		loopVarTok, err := p.next()
		if err != nil {
			return nil, err
		}
		if loopVarTok.kind != tokenid.TK_ID {
			return nil, fmt.Errorf("expected loop variable ID after as in list, got kind=%d image=%q", loopVarTok.kind, loopVarTok.image)
		}
		loopVar = loopVarTok.image
	}

	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}

	children, stopTok, hasStop, err := p.parseNodesUntilWithOptions(
		makeStopSet(tokenid.TK_ELSE, tokenid.TK_END_LIST),
		parseNodesOptions{allowItems: true, allowSep: true},
	)
	if err != nil {
		return nil, err
	}
	if !hasStop {
		return nil, fmt.Errorf("missing list terminator")
	}

	listNode := &ast.List{
		Source:   source,
		LoopVar:  loopVar,
		Children: children,
	}

	switch stopTok.kind {
	case tokenid.TK_ELSE:
		if _, err := p.next(); err != nil {
			return nil, err
		}
		elseChildren, nextStop, hasNextStop, err := p.parseNodesUntil(makeStopSet(tokenid.TK_END_LIST))
		if err != nil {
			return nil, err
		}
		if !hasNextStop || nextStop.kind != tokenid.TK_END_LIST {
			return nil, fmt.Errorf("list else block not closed with end_list")
		}
		listNode.Else = elseChildren
		if _, err := p.next(); err != nil {
			return nil, err
		}
	case tokenid.TK_END_LIST:
		if _, err := p.next(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected list block stop token: kind=%d image=%q", stopTok.kind, stopTok.image)
	}
	return listNode, nil
}

func (p *Parser) parseItems(start token) (*ast.Items, error) {
	asTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if asTok.kind != tokenid.TK_AS {
		return nil, fmt.Errorf("expected as in items header, got kind=%d image=%q", asTok.kind, asTok.image)
	}
	loopVarTok, err := p.next()
	if err != nil {
		return nil, err
	}
	if loopVarTok.kind != tokenid.TK_ID {
		return nil, fmt.Errorf("expected loop variable ID in items header, got kind=%d image=%q", loopVarTok.kind, loopVarTok.image)
	}
	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}

	children, stopTok, hasStop, err := p.parseNodesUntilWithOptions(
		makeStopSet(tokenid.TK_END_ITEMS, tokenid.TK_END_LIST, tokenid.TK_ELSE),
		parseNodesOptions{allowSep: true},
	)
	if err != nil {
		return nil, err
	}
	if !hasStop {
		return nil, fmt.Errorf("missing items terminator")
	}

	// </#items> is optional in this minimal parser; if present, consume it.
	if stopTok.kind == tokenid.TK_END_ITEMS {
		if _, err := p.next(); err != nil {
			return nil, err
		}
	}
	return &ast.Items{
		LoopVar:  loopVarTok.image,
		Children: children,
	}, nil
}

func (p *Parser) parseSep(parentStops stopSet) (*ast.Sep, error) {
	stops := unionStopSets(parentStops, makeStopSet(tokenid.TK_END_SEP))
	children, stopTok, hasStop, err := p.parseNodesUntil(stops)
	if err != nil {
		return nil, err
	}
	if !hasStop {
		return nil, fmt.Errorf("missing sep terminator")
	}
	if stopTok.kind == tokenid.TK_END_SEP {
		if _, err := p.next(); err != nil {
			return nil, err
		}
	}
	return &ast.Sep{Children: children}, nil
}

func (p *Parser) parseSwitch(start token) (*ast.Switch, error) {
	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expectDirectiveEnd(start.kind); err != nil {
		return nil, err
	}

	switchNode := &ast.Switch{Value: value}
	var currentTarget *[]ast.Node
	stops := makeStopSet(tokenid.TK_CASE, tokenid.TK_ON, tokenid.TK_DEFAUL, tokenid.TK_END_SWITCH)

	for {
		segment, stopTok, hasStop, err := p.parseNodesUntil(stops)
		if err != nil {
			return nil, err
		}
		if !hasStop {
			return nil, fmt.Errorf("missing switch terminator")
		}

		if currentTarget == nil {
			if containsNonWhitespaceTextNodes(segment) {
				return nil, fmt.Errorf("unexpected content before first switch branch")
			}
		} else {
			*currentTarget = append(*currentTarget, segment...)
		}

		switch stopTok.kind {
		case tokenid.TK_END_SWITCH:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			return switchNode, nil
		case tokenid.TK_CASE:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			cond, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expectDirectiveEnd(tokenid.TK_CASE); err != nil {
				return nil, err
			}
			branch := &ast.SwitchBranch{
				Kind:       "case",
				Conditions: []ast.Expr{cond},
			}
			switchNode.Branches = append(switchNode.Branches, branch)
			currentTarget = &branch.Children
		case tokenid.TK_ON:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			conds, err := p.parsePositionalArgsNoParens()
			if err != nil {
				return nil, err
			}
			if err := p.expectDirectiveEnd(tokenid.TK_ON); err != nil {
				return nil, err
			}
			branch := &ast.SwitchBranch{
				Kind:       "on",
				Conditions: conds,
			}
			switchNode.Branches = append(switchNode.Branches, branch)
			currentTarget = &branch.Children
		case tokenid.TK_DEFAUL:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			// Preserve explicit #default presence even if its body is empty.
			if switchNode.Default == nil {
				switchNode.Default = []ast.Node{}
			}
			currentTarget = &switchNode.Default
		default:
			return nil, fmt.Errorf("unexpected switch stop token: kind=%d image=%q", stopTok.kind, stopTok.image)
		}
	}
}

func assignmentScopeFromToken(kind int) (string, int, error) {
	switch kind {
	case tokenid.TK_ASSIGN:
		return scopeAssign, tokenid.TK_END_ASSIGN, nil
	case tokenid.TK_GLOBALASSIGN:
		return scopeGlobal, tokenid.TK_END_GLOBAL, nil
	case tokenid.TK_LOCALASSIGN:
		return scopeLocal, tokenid.TK_END_LOCAL, nil
	default:
		return "", 0, fmt.Errorf("unsupported assignment scope token kind=%d", kind)
	}
}

func isAssignmentTargetToken(kind int) bool {
	return kind == tokenid.TK_ID || kind == tokenid.TK_STRING_LITERAL
}

func normalizeAssignmentTarget(tok token) string {
	return normalizeNameToken(tok)
}

func normalizeNameToken(tok token) string {
	if tok.kind != tokenid.TK_STRING_LITERAL {
		return tok.image
	}
	if len(tok.image) >= 2 {
		q := tok.image[0]
		if (q == '"' || q == '\'') && tok.image[len(tok.image)-1] == q {
			return tok.image[1 : len(tok.image)-1]
		}
	}
	return tok.image
}

func unionStopSets(a, b stopSet) stopSet {
	if a == nil && b == nil {
		return nil
	}
	out := make(stopSet, len(a)+len(b))
	for k := range a {
		out[k] = struct{}{}
	}
	for k := range b {
		out[k] = struct{}{}
	}
	return out
}

func containsNonWhitespaceTextNodes(nodes []ast.Node) bool {
	for _, n := range nodes {
		txt, ok := n.(*ast.Text)
		if !ok {
			return true
		}
		if trimWhitespace(txt.Value) != "" {
			return true
		}
	}
	return false
}

func trimWhitespace(s string) string {
	start := 0
	for start < len(s) && isByteWhitespace(s[start]) {
		start++
	}
	end := len(s)
	for end > start && isByteWhitespace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func isByteWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func (p *Parser) parsePositionalArgsNoParens() ([]ast.Expr, error) {
	var result []ast.Expr

	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.kind == tokenid.TK_DIRECTIVE_END {
		return result, nil
	}
	if !canStartExpression(tok.kind) {
		return nil, fmt.Errorf("expected expression or directive end in positional args, got kind=%d image=%q", tok.kind, tok.image)
	}

	first, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	result = append(result, first)

	for {
		nextTok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if nextTok.kind == tokenid.TK_DIRECTIVE_END {
			return result, nil
		}
		if nextTok.kind == tokenid.TK_COMMA {
			if _, err := p.next(); err != nil {
				return nil, err
			}
		}

		argStart, err := p.peek()
		if err != nil {
			return nil, err
		}
		if !canStartExpression(argStart.kind) {
			return nil, fmt.Errorf("expected positional arg expression, got kind=%d image=%q", argStart.kind, argStart.image)
		}
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		result = append(result, arg)
	}
}

func (p *Parser) parseExpression() (ast.Expr, error) {
	return p.parseLambda()
}

func (p *Parser) parseLambda() (ast.Expr, error) {
	left, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.kind != tokenid.TK_LAMBDA_ARROW {
		return left, nil
	}
	if _, err := p.next(); err != nil {
		return nil, err
	}
	right, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	return &ast.Binary{
		Op:    tok.image,
		Left:  left,
		Right: right,
	}, nil
}

func (p *Parser) parseOr() (ast.Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind != tokenid.TK_OR {
			return left, nil
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &ast.Binary{Op: tok.image, Left: left, Right: right}
	}
}

func (p *Parser) parseAnd() (ast.Expr, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind != tokenid.TK_AND {
			return left, nil
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &ast.Binary{
			Op:    tok.image,
			Left:  left,
			Right: right,
		}
	}
}

func (p *Parser) parseEquality() (ast.Expr, error) {
	left, err := p.parseRelational()
	if err != nil {
		return nil, err
	}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind != tokenid.TK_DOUBLE_EQUALS && tok.kind != tokenid.TK_EQUALS && tok.kind != tokenid.TK_NOT_EQUALS {
			return left, nil
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseRelational()
		if err != nil {
			return nil, err
		}
		left = &ast.Binary{
			Op:    tok.image,
			Left:  left,
			Right: right,
		}
	}
}

func (p *Parser) parseRelational() (ast.Expr, error) {
	left, err := p.parseRange()
	if err != nil {
		return nil, err
	}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if !(tok.kind == tokenid.TK_LESS_THAN ||
			tok.kind == tokenid.TK_LESS_THAN_EQUALS ||
			tok.kind == tokenid.TK_NATURAL_GT ||
			tok.kind == tokenid.TK_NATURAL_GTE ||
			tok.kind == tokenid.TK_ESCAPED_GT ||
			tok.kind == tokenid.TK_ESCAPED_GTE) {
			return left, nil
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseRange()
		if err != nil {
			return nil, err
		}
		left = &ast.Binary{
			Op:    tok.image,
			Left:  left,
			Right: right,
		}
	}
}

func (p *Parser) parseRange() (ast.Expr, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.kind != tokenid.TK_DOT_DOT && tok.kind != tokenid.TK_DOT_DOT_LESS && tok.kind != tokenid.TK_DOT_DOT_ASTERISK {
		return left, nil
	}
	if _, err := p.next(); err != nil {
		return nil, err
	}
	var rhs ast.Expr
	nextTok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if canStartExpression(nextTok.kind) {
		rhs, err = p.parseAdditive()
		if err != nil {
			return nil, err
		}
	}
	return &ast.Binary{
		Op:    tok.image,
		Left:  left,
		Right: rhs,
	}, nil
}

func (p *Parser) parseAdditive() (ast.Expr, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind != tokenid.TK_PLUS && tok.kind != tokenid.TK_MINUS {
			return left, nil
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &ast.Binary{
			Op:    tok.image,
			Left:  left,
			Right: right,
		}
	}
}

func (p *Parser) parseMultiplicative() (ast.Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind != tokenid.TK_TIMES && tok.kind != tokenid.TK_DIVIDE && tok.kind != tokenid.TK_PERCENT {
			return left, nil
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &ast.Binary{
			Op:    tok.image,
			Left:  left,
			Right: right,
		}
	}
}

func (p *Parser) parseUnary() (ast.Expr, error) {
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.kind == tokenid.TK_EXCLAM || tok.kind == tokenid.TK_PLUS || tok.kind == tokenid.TK_MINUS {
		if _, err := p.next(); err != nil {
			return nil, err
		}
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.Unary{
			Op:   tok.image,
			Expr: expr,
		}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	expr, err := p.parseAtomic()
	if err != nil {
		return nil, err
	}
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		switch tok.kind {
		case tokenid.TK_DOT:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			nameTok, err := p.next()
			if err != nil {
				return nil, err
			}
			if nameTok.kind != tokenid.TK_ID {
				return nil, fmt.Errorf("expected ID after dot, got kind=%d image=%q", nameTok.kind, nameTok.image)
			}
			expr = &ast.Dot{
				Target: expr,
				Name:   nameTok.image,
			}
		case tokenid.TK_OPEN_BRACKET:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			keyExpr, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			endTok, err := p.next()
			if err != nil {
				return nil, err
			}
			if endTok.kind != tokenid.TK_CLOSE_BRACKET {
				return nil, fmt.Errorf("expected closing bracket, got kind=%d image=%q", endTok.kind, endTok.image)
			}
			expr = &ast.DynamicKey{
				Target: expr,
				Key:    keyExpr,
			}
		case tokenid.TK_OPEN_PAREN:
			args, err := p.parseArgs()
			if err != nil {
				return nil, err
			}
			expr = &ast.Call{
				Target: expr,
				Args:   args,
			}
		case tokenid.TK_BUILT_IN:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			nameTok, err := p.next()
			if err != nil {
				return nil, err
			}
			if nameTok.kind != tokenid.TK_ID {
				return nil, fmt.Errorf("expected built-in name ID, got kind=%d image=%q", nameTok.kind, nameTok.image)
			}
			var args []ast.Expr
			nextTok, err := p.peek()
			if err != nil {
				return nil, err
			}
			if nextTok.kind == tokenid.TK_OPEN_PAREN {
				args, err = p.parseArgs()
				if err != nil {
					return nil, err
				}
			}
			expr = &ast.Builtin{
				Target: expr,
				Name:   nameTok.image,
				Args:   args,
			}
		case tokenid.TK_EXISTS:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			expr = &ast.Exists{Target: expr}
		case tokenid.TK_EXCLAM:
			if _, err := p.next(); err != nil {
				return nil, err
			}
			var rhs ast.Expr
			nextTok, err := p.peek()
			if err != nil {
				return nil, err
			}
			if canStartExpression(nextTok.kind) {
				rhs, err = p.parseExpression()
				if err != nil {
					return nil, err
				}
			}
			expr = &ast.DefaultTo{Target: expr, RHS: rhs}
		default:
			return expr, nil
		}
	}
}

func (p *Parser) parseAtomic() (ast.Expr, error) {
	tok, err := p.next()
	if err != nil {
		return nil, err
	}
	switch tok.kind {
	case tokenid.TK_ID:
		return &ast.Identifier{Name: tok.image}, nil
	case tokenid.TK_DOT:
		nameTok, err := p.next()
		if err != nil {
			return nil, err
		}
		if nameTok.kind != tokenid.TK_ID {
			return nil, fmt.Errorf("expected ID after builtin-variable dot, got kind=%d image=%q", nameTok.kind, nameTok.image)
		}
		return &ast.Identifier{Name: "." + nameTok.image}, nil
	case tokenid.TK_INTEGER, tokenid.TK_DECIMAL:
		return &ast.Number{Literal: tok.image}, nil
	case tokenid.TK_STRING_LITERAL:
		return &ast.String{Literal: tok.image}, nil
	case tokenid.TK_TRUE:
		return &ast.Boolean{Value: true}, nil
	case tokenid.TK_FALSE:
		return &ast.Boolean{Value: false}, nil
	case tokenid.TK_OPEN_BRACKET:
		return p.parseArrayLiteral()
	case tokenid.TK_OPENING_CURLY_BRACKET:
		return p.parseMapLiteral()
	case tokenid.TK_OPEN_PAREN:
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		endTok, err := p.next()
		if err != nil {
			return nil, err
		}
		if endTok.kind != tokenid.TK_CLOSE_PAREN {
			return nil, fmt.Errorf("expected closing parenthesis, got kind=%d image=%q", endTok.kind, endTok.image)
		}
		return &ast.Parenthetical{Expr: expr}, nil
	default:
		return nil, fmt.Errorf("unexpected token in expression: kind=%d image=%q", tok.kind, tok.image)
	}
}

func (p *Parser) parseArrayLiteral() (ast.Expr, error) {
	arr := &ast.Array{}
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.kind == tokenid.TK_CLOSE_BRACKET {
		if _, err := p.next(); err != nil {
			return nil, err
		}
		return arr, nil
	}
	for {
		item, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		arr.Items = append(arr.Items, item)
		nextTok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if nextTok.kind == tokenid.TK_COMMA {
			if _, err := p.next(); err != nil {
				return nil, err
			}
			continue
		}
		if nextTok.kind != tokenid.TK_CLOSE_BRACKET {
			return nil, fmt.Errorf("expected comma or closing bracket in array literal, got kind=%d image=%q", nextTok.kind, nextTok.image)
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		return arr, nil
	}
}

func (p *Parser) parseMapLiteral() (ast.Expr, error) {
	m := &ast.Map{}
	tok, err := p.peek()
	if err != nil {
		return nil, err
	}
	if tok.kind == tokenid.TK_CLOSING_CURLY_BRACKET {
		if _, err := p.next(); err != nil {
			return nil, err
		}
		return m, nil
	}
	for {
		key, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		colonTok, err := p.next()
		if err != nil {
			return nil, err
		}
		if colonTok.kind != tokenid.TK_COLON {
			return nil, fmt.Errorf("expected colon in map literal, got kind=%d image=%q", colonTok.kind, colonTok.image)
		}
		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		m.Items = append(m.Items, &ast.MapEntry{Key: key, Value: value})
		nextTok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if nextTok.kind == tokenid.TK_COMMA {
			if _, err := p.next(); err != nil {
				return nil, err
			}
			continue
		}
		if nextTok.kind != tokenid.TK_CLOSING_CURLY_BRACKET {
			return nil, fmt.Errorf("expected comma or closing curly in map literal, got kind=%d image=%q", nextTok.kind, nextTok.image)
		}
		if _, err := p.next(); err != nil {
			return nil, err
		}
		return m, nil
	}
}

func (p *Parser) parseArgs() ([]ast.Expr, error) {
	start, err := p.next()
	if err != nil {
		return nil, err
	}
	if start.kind != tokenid.TK_OPEN_PAREN {
		return nil, fmt.Errorf("expected open parenthesis, got kind=%d image=%q", start.kind, start.image)
	}

	var args []ast.Expr
	for {
		tok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if tok.kind == tokenid.TK_CLOSE_PAREN {
			if _, err := p.next(); err != nil {
				return nil, err
			}
			return args, nil
		}
		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		nextTok, err := p.peek()
		if err != nil {
			return nil, err
		}
		if nextTok.kind == tokenid.TK_COMMA {
			if _, err := p.next(); err != nil {
				return nil, err
			}
			continue
		}
		if canStartExpression(nextTok.kind) {
			// FreeMarker positional args allow optional commas in some contexts.
			continue
		}
		if nextTok.kind == tokenid.TK_CLOSE_PAREN {
			if _, err := p.next(); err != nil {
				return nil, err
			}
			return args, nil
		}
		return nil, fmt.Errorf("expected comma or close parenthesis in args, got kind=%d image=%q", nextTok.kind, nextTok.image)
	}
}

func (p *Parser) expectInterpolationClosing(openingKind int) error {
	wantKind, err := interpolationClosingKind(openingKind)
	if err != nil {
		return err
	}
	tok, err := p.next()
	if err != nil {
		return err
	}
	if tok.kind != wantKind {
		if wantKind == tokenid.TK_CLOSING_CURLY_BRACKET {
			return fmt.Errorf("expected closing curly, got kind=%d image=%q", tok.kind, tok.image)
		}
		return fmt.Errorf("expected closing bracket, got kind=%d image=%q", tok.kind, tok.image)
	}
	return nil
}

func interpolationClosingKind(openingKind int) (int, error) {
	switch openingKind {
	case tokenid.TK_DOLLAR_INTERPOLATION_OPENING, tokenid.TK_HASH_INTERPOLATION_OPENING:
		return tokenid.TK_CLOSING_CURLY_BRACKET, nil
	case tokenid.TK_SQUARE_BRACKET_INTERPOLATION_OPENING:
		return tokenid.TK_CLOSE_BRACKET, nil
	default:
		return 0, fmt.Errorf("unknown interpolation opening kind: %d", openingKind)
	}
}

func (p *Parser) peek() (token, error) {
	return p.peekN(1)
}

func (p *Parser) peekN(n int) (token, error) {
	if n <= 0 {
		return token{}, fmt.Errorf("peekN expects n >= 1, got %d", n)
	}
	for len(p.lookBuf) < n {
		tok, err := p.nextFromLexer()
		if err != nil {
			return token{}, err
		}
		p.lookBuf = append(p.lookBuf, tok)
	}
	return p.lookBuf[n-1], nil
}

func (p *Parser) next() (token, error) {
	if len(p.lookBuf) > 0 {
		tok := p.lookBuf[0]
		p.lookBuf = p.lookBuf[1:]
		return tok, nil
	}
	return p.nextFromLexer()
}

func (p *Parser) nextFromLexer() (token, error) {
	tok, err := p.lx.Next()
	if err != nil {
		return token{}, err
	}
	return token{kind: tok.Kind, image: tok.Image}, nil
}

func canStartExpression(kind int) bool {
	switch kind {
	case tokenid.TK_ID,
		tokenid.TK_DOT,
		tokenid.TK_INTEGER, tokenid.TK_DECIMAL,
		tokenid.TK_STRING_LITERAL,
		tokenid.TK_TRUE, tokenid.TK_FALSE,
		tokenid.TK_OPEN_PAREN, tokenid.TK_OPEN_BRACKET, tokenid.TK_OPENING_CURLY_BRACKET,
		tokenid.TK_EXCLAM, tokenid.TK_PLUS, tokenid.TK_MINUS:
		return true
	default:
		return false
	}
}

func extractCommentContent(image string) string {
	if strings.HasPrefix(image, "<#--") && strings.HasSuffix(image, "-->") {
		return image[4 : len(image)-3]
	}
	if strings.HasPrefix(image, "[#--") && strings.HasSuffix(image, "--]") {
		return image[4 : len(image)-3]
	}
	if strings.HasPrefix(image, "<#-") && strings.HasSuffix(image, "->") {
		return image[3 : len(image)-2]
	}
	if strings.HasPrefix(image, "[#-") && strings.HasSuffix(image, "-]") {
		return image[3 : len(image)-2]
	}
	return image
}
