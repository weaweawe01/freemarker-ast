package parser

import (
	"testing"

	"github.com/weaweawe01/freemarker-ast/internal/ast"
)

func TestParseTextOnly(t *testing.T) {
	root, err := Parse("hello world")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	txt, ok := root.Children[0].(*ast.Text)
	if !ok {
		t.Fatalf("expected Text child, got %T", root.Children[0])
	}
	if txt.Value != "hello world" {
		t.Fatalf("text mismatch: got %q", txt.Value)
	}
}

func TestParseInterpolationIdentifier(t *testing.T) {
	root, err := Parse("a${x}b")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 3 {
		t.Fatalf("children length mismatch: got %d want 3", len(root.Children))
	}
	interp, ok := root.Children[1].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected Interpolation child, got %T", root.Children[1])
	}
	id, ok := interp.Expr.(*ast.Identifier)
	if !ok {
		t.Fatalf("expected Identifier expr, got %T", interp.Expr)
	}
	if id.Name != "x" {
		t.Fatalf("identifier mismatch: got %q", id.Name)
	}
}

func TestParseInterpolationAdditiveExpression(t *testing.T) {
	root, err := Parse("${x+1}")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	interp, ok := root.Children[0].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected Interpolation child, got %T", root.Children[0])
	}
	bin, ok := interp.Expr.(*ast.Binary)
	if !ok {
		t.Fatalf("expected Binary expr, got %T", interp.Expr)
	}
	if bin.Op != "+" {
		t.Fatalf("binary op mismatch: got %q", bin.Op)
	}
}

func TestParseSquareInterpolation(t *testing.T) {
	root, err := Parse("a[=x]b")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 3 {
		t.Fatalf("children length mismatch: got %d want 3", len(root.Children))
	}
	interp, ok := root.Children[1].(*ast.Interpolation)
	if !ok {
		t.Fatalf("expected Interpolation child, got %T", root.Children[1])
	}
	if interp.Opening != "[=" {
		t.Fatalf("opening mismatch: got %q", interp.Opening)
	}
}

func TestParseExpressionBuiltIn(t *testing.T) {
	expr, err := ParseExpressionString("x?trim")
	if err != nil {
		t.Fatalf("parse expression: %v", err)
	}
	bi, ok := expr.(*ast.Builtin)
	if !ok {
		t.Fatalf("expected Builtin, got %T", expr)
	}
	if bi.Name != "trim" {
		t.Fatalf("builtin name mismatch: got %q", bi.Name)
	}
}

func TestParseExpressionDefaultAndExists(t *testing.T) {
	expr, err := ParseExpressionString("x?? || y!'d'")
	if err != nil {
		t.Fatalf("parse expression: %v", err)
	}
	bin, ok := expr.(*ast.Binary)
	if !ok {
		t.Fatalf("expected Binary, got %T", expr)
	}
	if _, ok := bin.Left.(*ast.Exists); !ok {
		t.Fatalf("left should be Exists, got %T", bin.Left)
	}
	def, ok := bin.Right.(*ast.DefaultTo)
	if !ok {
		t.Fatalf("right should be DefaultTo, got %T", bin.Right)
	}
	if def.RHS == nil {
		t.Fatal("default RHS should not be nil")
	}
}

func TestParseExpressionRangeUnbound(t *testing.T) {
	expr, err := ParseExpressionString("0..")
	if err != nil {
		t.Fatalf("parse expression: %v", err)
	}
	rng, ok := expr.(*ast.Binary)
	if !ok {
		t.Fatalf("expected Binary range, got %T", expr)
	}
	if rng.Op != ".." {
		t.Fatalf("range op mismatch: got %q", rng.Op)
	}
	if rng.Right != nil {
		t.Fatalf("unbound range should have nil right, got %T", rng.Right)
	}
}

func TestParseExpressionCallAndDot(t *testing.T) {
	expr, err := ParseExpressionString("ns.f(1, 2).g")
	if err != nil {
		t.Fatalf("parse expression: %v", err)
	}
	dot, ok := expr.(*ast.Dot)
	if !ok {
		t.Fatalf("expected final Dot, got %T", expr)
	}
	if dot.Name != "g" {
		t.Fatalf("dot name mismatch: got %q", dot.Name)
	}
}

func TestParseIfSimple(t *testing.T) {
	root, err := Parse("<#if x>yes</#if>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	ifNode, ok := root.Children[0].(*ast.If)
	if !ok {
		t.Fatalf("expected If node, got %T", root.Children[0])
	}
	if len(ifNode.Branches) != 1 {
		t.Fatalf("if branches mismatch: got %d want 1", len(ifNode.Branches))
	}
	cond, ok := ifNode.Branches[0].Condition.(*ast.Identifier)
	if !ok || cond.Name != "x" {
		t.Fatalf("if condition mismatch: %#v", ifNode.Branches[0].Condition)
	}
	if len(ifNode.Branches[0].Children) != 1 {
		t.Fatalf("if children mismatch: got %d want 1", len(ifNode.Branches[0].Children))
	}
	txt, ok := ifNode.Branches[0].Children[0].(*ast.Text)
	if !ok || txt.Value != "yes" {
		t.Fatalf("if child text mismatch: %#v", ifNode.Branches[0].Children[0])
	}
}

func TestParseIfElse(t *testing.T) {
	root, err := Parse("<#if x>yes<#else>no</#if>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	ifNode, ok := root.Children[0].(*ast.If)
	if !ok {
		t.Fatalf("expected If node, got %T", root.Children[0])
	}
	if len(ifNode.Else) != 1 {
		t.Fatalf("else children mismatch: got %d want 1", len(ifNode.Else))
	}
	txt, ok := ifNode.Else[0].(*ast.Text)
	if !ok || txt.Value != "no" {
		t.Fatalf("else child text mismatch: %#v", ifNode.Else[0])
	}
}

func TestParseIfElseIfElse(t *testing.T) {
	root, err := Parse("<#if x>1<#elseif y>2<#else>3</#if>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	ifNode, ok := root.Children[0].(*ast.If)
	if !ok {
		t.Fatalf("expected If node, got %T", root.Children[0])
	}
	if len(ifNode.Branches) != 2 {
		t.Fatalf("if branches mismatch: got %d want 2", len(ifNode.Branches))
	}
	cond0, ok := ifNode.Branches[0].Condition.(*ast.Identifier)
	if !ok || cond0.Name != "x" {
		t.Fatalf("branch0 condition mismatch: %#v", ifNode.Branches[0].Condition)
	}
	cond1, ok := ifNode.Branches[1].Condition.(*ast.Identifier)
	if !ok || cond1.Name != "y" {
		t.Fatalf("branch1 condition mismatch: %#v", ifNode.Branches[1].Condition)
	}
	if len(ifNode.Else) != 1 {
		t.Fatalf("else children mismatch: got %d want 1", len(ifNode.Else))
	}
}

func TestParseIfNested(t *testing.T) {
	root, err := Parse("<#if x>a<#if y>b</#if>c</#if>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	top, ok := root.Children[0].(*ast.If)
	if !ok {
		t.Fatalf("expected top If, got %T", root.Children[0])
	}
	if len(top.Branches) != 1 {
		t.Fatalf("top branch count mismatch: got %d want 1", len(top.Branches))
	}
	if len(top.Branches[0].Children) != 3 {
		t.Fatalf("top branch children mismatch: got %d want 3", len(top.Branches[0].Children))
	}
	nested, ok := top.Branches[0].Children[1].(*ast.If)
	if !ok {
		t.Fatalf("expected nested If at index 1, got %T", top.Branches[0].Children[1])
	}
	if len(nested.Branches) != 1 {
		t.Fatalf("nested branch count mismatch: got %d want 1", len(nested.Branches))
	}
}

func TestParseIfWithElseIfCamelCase(t *testing.T) {
	root, err := Parse("<#if x>1<#elseIf y>2</#if>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	ifNode, ok := root.Children[0].(*ast.If)
	if !ok {
		t.Fatalf("expected If node, got %T", root.Children[0])
	}
	if len(ifNode.Branches) != 2 {
		t.Fatalf("if branches mismatch: got %d want 2", len(ifNode.Branches))
	}
}

func TestParseSquareBracketIf(t *testing.T) {
	root, err := Parse("[#if x]yes[/#if]")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	ifNode, ok := root.Children[0].(*ast.If)
	if !ok {
		t.Fatalf("expected If node, got %T", root.Children[0])
	}
	if len(ifNode.Branches) != 1 {
		t.Fatalf("if branches mismatch: got %d want 1", len(ifNode.Branches))
	}
}

func TestParseAssignInstruction(t *testing.T) {
	root, err := Parse("<#assign x = 1, y += 2 in ns>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	assignNode, ok := root.Children[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment node, got %T", root.Children[0])
	}
	if assignNode.Scope != "assign" {
		t.Fatalf("scope mismatch: got %q", assignNode.Scope)
	}
	if len(assignNode.Items) != 2 {
		t.Fatalf("assignment item count mismatch: got %d want 2", len(assignNode.Items))
	}
	if assignNode.Items[0].Target != "x" || assignNode.Items[0].Op != "=" {
		t.Fatalf("first assignment mismatch: %+v", assignNode.Items[0])
	}
	if assignNode.Items[1].Target != "y" || assignNode.Items[1].Op != "+=" {
		t.Fatalf("second assignment mismatch: %+v", assignNode.Items[1])
	}
	if assignNode.Namespace == nil {
		t.Fatal("namespace should not be nil")
	}
}

func TestParseAssignCaptureBlock(t *testing.T) {
	root, err := Parse("<#assign x>foo ${bar}</#assign>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	block, ok := root.Children[0].(*ast.AssignBlock)
	if !ok {
		t.Fatalf("expected AssignBlock node, got %T", root.Children[0])
	}
	if block.Scope != "assign" {
		t.Fatalf("scope mismatch: got %q", block.Scope)
	}
	if block.Target != "x" {
		t.Fatalf("target mismatch: got %q", block.Target)
	}
	if len(block.Children) != 2 {
		t.Fatalf("assign block children mismatch: got %d want 2", len(block.Children))
	}
	if _, ok := block.Children[1].(*ast.Interpolation); !ok {
		t.Fatalf("expected interpolation child at index 1, got %T", block.Children[1])
	}
}

func TestParseMacroWithLocalAssignment(t *testing.T) {
	root, err := Parse("<#macro m><#local x = 1></#macro>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	macroNode, ok := root.Children[0].(*ast.Macro)
	if !ok {
		t.Fatalf("expected Macro node, got %T", root.Children[0])
	}
	if macroNode.Name != "m" {
		t.Fatalf("macro name mismatch: got %q", macroNode.Name)
	}
	if len(macroNode.Children) != 1 {
		t.Fatalf("macro child count mismatch: got %d want 1", len(macroNode.Children))
	}
	localAssign, ok := macroNode.Children[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment child, got %T", macroNode.Children[0])
	}
	if localAssign.Scope != "local" {
		t.Fatalf("local scope mismatch: got %q", localAssign.Scope)
	}
}

func TestParseListSimple(t *testing.T) {
	root, err := Parse("<#list xs as x>${x}</#list>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	listNode, ok := root.Children[0].(*ast.List)
	if !ok {
		t.Fatalf("expected List node, got %T", root.Children[0])
	}
	if listNode.LoopVar != "x" {
		t.Fatalf("list loop var mismatch: got %q", listNode.LoopVar)
	}
	if len(listNode.Children) != 1 {
		t.Fatalf("list children mismatch: got %d want 1", len(listNode.Children))
	}
	if _, ok := listNode.Children[0].(*ast.Interpolation); !ok {
		t.Fatalf("expected interpolation child, got %T", listNode.Children[0])
	}
}

func TestParseListWithItemsAndSepAndElse(t *testing.T) {
	src := "<#list xs>[<#items as x>${x}<#sep>, </#items>]<#else>None</#list>"
	root, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	listNode, ok := root.Children[0].(*ast.List)
	if !ok {
		t.Fatalf("expected List node, got %T", root.Children[0])
	}
	if listNode.LoopVar != "" {
		t.Fatalf("list loop var should be empty, got %q", listNode.LoopVar)
	}
	if len(listNode.Children) != 3 {
		t.Fatalf("list children mismatch: got %d want 3", len(listNode.Children))
	}
	itemsNode, ok := listNode.Children[1].(*ast.Items)
	if !ok {
		t.Fatalf("expected Items node at index 1, got %T", listNode.Children[1])
	}
	if itemsNode.LoopVar != "x" {
		t.Fatalf("items loop var mismatch: got %q", itemsNode.LoopVar)
	}
	if len(itemsNode.Children) != 2 {
		t.Fatalf("items children mismatch: got %d want 2", len(itemsNode.Children))
	}
	sepNode, ok := itemsNode.Children[1].(*ast.Sep)
	if !ok {
		t.Fatalf("expected Sep node at index 1, got %T", itemsNode.Children[1])
	}
	if len(sepNode.Children) != 1 {
		t.Fatalf("sep children mismatch: got %d want 1", len(sepNode.Children))
	}
	if len(listNode.Else) != 1 {
		t.Fatalf("list else children mismatch: got %d want 1", len(listNode.Else))
	}
}

func TestParseSwitchCaseDefault(t *testing.T) {
	src := "<#switch x><#case 1>one<#case 2>two<#default>more</#switch>"
	root, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	sw, ok := root.Children[0].(*ast.Switch)
	if !ok {
		t.Fatalf("expected Switch node, got %T", root.Children[0])
	}
	if len(sw.Branches) != 2 {
		t.Fatalf("switch branch count mismatch: got %d want 2", len(sw.Branches))
	}
	if sw.Branches[0].Kind != "case" || sw.Branches[1].Kind != "case" {
		t.Fatalf("switch branch kinds mismatch: %+v", sw.Branches)
	}
	if len(sw.Default) != 1 {
		t.Fatalf("switch default children mismatch: got %d want 1", len(sw.Default))
	}
}

func TestParseSwitchOnDefault(t *testing.T) {
	src := "<#switch x><#on 1, 2>one or two<#on 3>three<#default>more</#switch>"
	root, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	sw, ok := root.Children[0].(*ast.Switch)
	if !ok {
		t.Fatalf("expected Switch node, got %T", root.Children[0])
	}
	if len(sw.Branches) != 2 {
		t.Fatalf("switch branch count mismatch: got %d want 2", len(sw.Branches))
	}
	if sw.Branches[0].Kind != "on" || sw.Branches[1].Kind != "on" {
		t.Fatalf("switch on branch kinds mismatch: %+v", sw.Branches)
	}
	if len(sw.Branches[0].Conditions) != 2 {
		t.Fatalf("first on branch conditions mismatch: got %d want 2", len(sw.Branches[0].Conditions))
	}
	if len(sw.Default) != 1 {
		t.Fatalf("switch default children mismatch: got %d want 1", len(sw.Default))
	}
}

func TestParseFunctionWithReturn(t *testing.T) {
	src := "<#function foo x y><#local x = 123><#return x + y></#function>"
	root, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	fn, ok := root.Children[0].(*ast.Function)
	if !ok {
		t.Fatalf("expected Function node, got %T", root.Children[0])
	}
	if fn.Name != "foo" {
		t.Fatalf("function name mismatch: got %q", fn.Name)
	}
	if len(fn.Children) != 2 {
		t.Fatalf("function children mismatch: got %d want 2", len(fn.Children))
	}
	if _, ok := fn.Children[0].(*ast.Assignment); !ok {
		t.Fatalf("expected Assignment as first function child, got %T", fn.Children[0])
	}
	ret, ok := fn.Children[1].(*ast.Return)
	if !ok {
		t.Fatalf("expected Return as second function child, got %T", fn.Children[1])
	}
	if ret.Value == nil {
		t.Fatal("return value should not be nil")
	}
}

func TestParseSimpleReturn(t *testing.T) {
	root, err := Parse("<#return>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	ret, ok := root.Children[0].(*ast.Return)
	if !ok {
		t.Fatalf("expected Return node, got %T", root.Children[0])
	}
	if ret.Value != nil {
		t.Fatalf("simple return should have nil value, got %T", ret.Value)
	}
}

func TestParseOutputFormatNestedEscaping(t *testing.T) {
	src := "<#outputFormat \"XML\"><#noAutoEsc>${a}<#autoEsc>${b}</#autoEsc>${c}</#noAutoEsc></#outputFormat>"
	root, err := Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	out, ok := root.Children[0].(*ast.OutputFormat)
	if !ok {
		t.Fatalf("expected OutputFormat node, got %T", root.Children[0])
	}
	if len(out.Children) != 1 {
		t.Fatalf("outputFormat children mismatch: got %d want 1", len(out.Children))
	}
	noAuto, ok := out.Children[0].(*ast.NoAutoEsc)
	if !ok {
		t.Fatalf("expected NoAutoEsc child, got %T", out.Children[0])
	}
	if len(noAuto.Children) != 3 {
		t.Fatalf("noAutoEsc children mismatch: got %d want 3", len(noAuto.Children))
	}
	if _, ok := noAuto.Children[1].(*ast.AutoEsc); !ok {
		t.Fatalf("expected AutoEsc in middle child, got %T", noAuto.Children[1])
	}
}

func TestParseAttemptRecover(t *testing.T) {
	root, err := Parse("<#attempt>1<#recover>2</#attempt>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	attempt, ok := root.Children[0].(*ast.Attempt)
	if !ok {
		t.Fatalf("expected Attempt node, got %T", root.Children[0])
	}
	if len(attempt.Attempt) != 1 {
		t.Fatalf("attempt body children mismatch: got %d want 1", len(attempt.Attempt))
	}
	if len(attempt.Recover) != 1 {
		t.Fatalf("recover body children mismatch: got %d want 1", len(attempt.Recover))
	}
}

func TestParseExpressionLambda(t *testing.T) {
	expr, err := ParseExpressionString("x -> !x")
	if err != nil {
		t.Fatalf("parse expression: %v", err)
	}
	bin, ok := expr.(*ast.Binary)
	if !ok {
		t.Fatalf("expected Binary lambda expression, got %T", expr)
	}
	if bin.Op != "->" {
		t.Fatalf("lambda op mismatch: got %q", bin.Op)
	}
}

func TestParseAssignmentArrayLiteral(t *testing.T) {
	root, err := Parse("<#assign xs = []>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	assignNode, ok := root.Children[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("expected Assignment node, got %T", root.Children[0])
	}
	arr, ok := assignNode.Items[0].Value.(*ast.Array)
	if !ok {
		t.Fatalf("expected Array literal value, got %T", assignNode.Items[0].Value)
	}
	if len(arr.Items) != 0 {
		t.Fatalf("expected empty array literal, got %d items", len(arr.Items))
	}
}

func TestParseInterpolationWithSemicolonOptions(t *testing.T) {
	root, err := Parse("#{x; M2}")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	if _, ok := root.Children[0].(*ast.Interpolation); !ok {
		t.Fatalf("expected interpolation node, got %T", root.Children[0])
	}
}

func TestParseMacroQuotedName(t *testing.T) {
	root, err := Parse("<#macro \"m-b2\"></#macro>")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children length mismatch: got %d want 1", len(root.Children))
	}
	m, ok := root.Children[0].(*ast.Macro)
	if !ok {
		t.Fatalf("expected Macro node, got %T", root.Children[0])
	}
	if m.Name != "m-b2" {
		t.Fatalf("macro name mismatch: got %q", m.Name)
	}
}
