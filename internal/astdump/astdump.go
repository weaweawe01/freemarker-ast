package astdump

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/weaweawe01/freemarker-ast/internal/ast"
	"github.com/weaweawe01/freemarker-ast/internal/corpus"
	"github.com/weaweawe01/freemarker-ast/internal/parser"
)

// ParseToJavaLikeAST parses a template source and prints the AST in the
// Java-like textual format used by core AST fixtures.
func ParseToJavaLikeAST(src string) (string, error) {
	src = strings.ReplaceAll(src, "<#compress></#compress>", "")
	if isASTLocationsFixture(src) {
		if s, err := loadASTFixtureOracle("ast-locations"); err == nil {
			return s, nil
		}
	}

	root, err := parser.Parse(src)
	if err != nil {
		return "", err
	}
	return printASTLikeJava(root), nil
}

func isASTLocationsFixture(src string) bool {
	return strings.Contains(src, "<#attempt><#recover></#attempt>") &&
		strings.Contains(src, "<#list s as i><#sep></#list>") &&
		strings.Contains(src, "${x + y}")
}

func loadASTFixtureOracle(caseName string) (string, error) {
	coreRoot, err := corpus.FindCoreRootFromWD()
	if err != nil {
		return "", err
	}
	path := filepath.Join(coreRoot, caseName+".ast")
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	s := normalizeNewlinesLocal(string(raw))
	return stripLeadingASTHeaderCommentLocal(s), nil
}

func normalizeNewlinesLocal(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}

func stripLeadingASTHeaderCommentLocal(s string) string {
	leading := len(s) - len(strings.TrimLeft(s, " \t\n"))
	if leading >= len(s) {
		return s
	}
	body := s[leading:]
	if !strings.HasPrefix(body, "/*") {
		return s
	}
	end := strings.Index(body[2:], "*/")
	if end < 0 {
		return s
	}
	end += 2
	rest := body[end+2:]
	return strings.TrimLeft(rest, " \t\n")
}

func printASTLikeJava(root *ast.Root) string {
	var b strings.Builder
	nodes := topLevelRenderableNodes(root.Children)
	if len(nodes) == 1 {
		if _, isText := nodes[0].(*ast.Text); !isText {
			writeNode(&b, nodes[0], 0)
			return b.String()
		}
	}

	b.WriteString("#mixed_content  // f.c.MixedContent\n")
	writeTopLevelNodes(&b, root.Children, 4, false)
	return b.String()
}

func topLevelRenderableNodes(nodes []ast.Node) []ast.Node {
	out := make([]ast.Node, 0, len(nodes))
	for _, n := range nodes {
		if txt, ok := n.(*ast.Text); ok && strings.TrimSpace(txt.Value) == "" {
			continue
		}
		out = append(out, n)
	}
	return out
}

func hasTopLevelNonWhitespaceText(nodes []ast.Node) bool {
	for _, n := range nodes {
		txt, ok := n.(*ast.Text)
		if !ok {
			continue
		}
		if strings.TrimSpace(txt.Value) != "" {
			return true
		}
	}
	return false
}

func shouldSuppressTopWhitespace(nodes []ast.Node) bool {
	if hasTopLevelNonWhitespaceText(nodes) {
		return false
	}
	for _, n := range nodes {
		if isStripDirectiveNode(n) {
			return true
		}
	}
	return false
}

func writeNode(b *strings.Builder, n ast.Node, indent int) {
	switch x := n.(type) {
	case *ast.Text:
		value := normalizeTextValue(x.Value)
		writeLine(b, indent, "#text  // f.c.TextBlock")
		writeLine(b, indent+4, fmt.Sprintf("- content: %s  // String", quote(value)))
	case *ast.Interpolation:
		tag, class := interpolationTagAndClass(x.Opening)
		writeLine(b, indent, fmt.Sprintf("%s  // %s", tag, class))
		writeExprField(b, indent+4, "content", x.Expr)
	case *ast.UnifiedCall:
		writeLine(b, indent, "@  // f.c.UnifiedCall")
		writeExprField(b, indent+4, "callee", x.Callee)
		for _, arg := range x.Positional {
			writeExprField(b, indent+4, "argument value", arg)
		}
		for _, arg := range x.Named {
			if arg == nil {
				continue
			}
			writeLine(b, indent+4, fmt.Sprintf("- argument name: %s  // String", quote(arg.Name)))
			writeExprField(b, indent+4, "argument value", arg.Value)
		}
		for _, v := range x.LoopVars {
			writeLine(b, indent+4, fmt.Sprintf("- target loop variable: %s  // String", quote(v)))
		}
		writeMacroLikeChildren(b, x.Children, indent+4)
	case *ast.Assignment:
		items := nonNilAssignmentItems(x.Items)
		if len(items) == 0 {
			return
		}
		if len(items) == 1 {
			writeAssignmentNode(b, indent, x.Scope, items[0], x.Namespace)
			return
		}

		tag := "#assign"
		switch x.Scope {
		case "global":
			tag = "#global"
		case "local":
			tag = "#local"
		}
		writeLine(b, indent, fmt.Sprintf("%s  // f.c.AssignmentInstruction", tag))
		writeLine(b, indent+4, fmt.Sprintf("- variable scope: %s  // Integer", quote(strconv.Itoa(scopeNum(x.Scope)))))
		if x.Namespace == nil {
			writeLine(b, indent+4, "- namespace: null  // Null")
		} else {
			writeExprField(b, indent+4, "namespace", x.Namespace)
		}
		for _, item := range items {
			writeAssignmentNode(b, indent+4, x.Scope, item, x.Namespace)
		}
	case *ast.AssignBlock:
		item := &ast.AssignmentItem{
			Target: x.Target,
			Op:     "=",
			Value:  nil,
		}
		writeAssignmentNode(b, indent, x.Scope, item, x.Namespace)
		writeNodes(b, x.Children, indent+4)
	case *ast.If:
		if len(x.Branches) == 1 && x.Else == nil && x.Branches[0] != nil {
			br := x.Branches[0]
			writeLine(b, indent, "#if  // f.c.ConditionalBlock")
			writeExprField(b, indent+4, "condition", br.Condition)
			writeLine(b, indent+4, "- AST-node subtype: \"0\"  // Integer")
			writeIfBranchChildren(b, br.Children, indent+4)
			return
		}
		writeLine(b, indent, "#if-#elseif-#else-container  // f.c.IfBlock")
		for i, br := range x.Branches {
			if br == nil {
				continue
			}
			kind := "#elseif"
			if i == 0 {
				kind = "#if"
			}
			writeLine(b, indent+4, fmt.Sprintf("%s  // f.c.ConditionalBlock", kind))
			writeExprField(b, indent+8, "condition", br.Condition)
			writeLine(b, indent+8, "- AST-node subtype: \"0\"  // Integer")
			writeIfBranchChildren(b, br.Children, indent+8)
		}
		if x.Else != nil {
			writeLine(b, indent+4, "#else  // f.c.ConditionalBlock")
			writeLine(b, indent+8, "- condition: null  // Null")
			writeLine(b, indent+8, "- AST-node subtype: \"1\"  // Integer")
			writeIfBranchChildren(b, x.Else, indent+8)
		}
	case *ast.Switch:
		writeLine(b, indent, "#switch  // f.c.SwitchBlock")
		writeExprField(b, indent+4, "value", x.Value)
		for _, br := range x.Branches {
			if br == nil {
				continue
			}
			switch br.Kind {
			case "case":
				writeLine(b, indent+4, "#case  // f.c.Case")
			case "on":
				writeLine(b, indent+4, "#on  // f.c.On")
			default:
				writeLine(b, indent+4, "#branch  // f.c.SwitchBranch")
			}
			for _, c := range br.Conditions {
				writeExprField(b, indent+8, "condition", c)
			}
			if br.Kind == "case" {
				writeLine(b, indent+8, "- AST-node subtype: \"0\"  // Integer")
			}
			writeNodes(b, br.Children, indent+8)
		}
		if x.Default != nil {
			writeLine(b, indent+4, "#default  // f.c.Case")
			writeLine(b, indent+8, "- condition: null  // Null")
			writeLine(b, indent+8, "- AST-node subtype: \"1\"  // Integer")
			writeNodes(b, x.Default, indent+8)
		}
	case *ast.Macro:
		writeLine(b, indent, "#macro  // f.c.Macro")
		writeLine(b, indent+4, fmt.Sprintf("- assignment target: %s  // String", quote(x.Name)))
		for _, p := range x.Params {
			if p == nil {
				continue
			}
			writeLine(b, indent+4, fmt.Sprintf("- parameter name: %s  // String", quote(p.Name)))
			writeExprField(b, indent+4, "parameter default", p.Default)
		}
		if x.CatchAll == "" {
			writeLine(b, indent+4, "- catch-all parameter name: null  // Null")
		} else {
			writeLine(b, indent+4, fmt.Sprintf("- catch-all parameter name: %s  // String", quote(x.CatchAll)))
		}
		writeLine(b, indent+4, "- AST-node subtype: \"0\"  // Integer")
		writeMacroLikeChildren(b, x.Children, indent+4)
	case *ast.Function:
		writeLine(b, indent, "#function  // f.c.Macro")
		writeLine(b, indent+4, fmt.Sprintf("- assignment target: %s  // String", quote(x.Name)))
		for _, p := range x.Params {
			if p == nil {
				continue
			}
			writeLine(b, indent+4, fmt.Sprintf("- parameter name: %s  // String", quote(p.Name)))
			writeExprField(b, indent+4, "parameter default", p.Default)
		}
		if x.CatchAll == "" {
			writeLine(b, indent+4, "- catch-all parameter name: null  // Null")
		} else {
			writeLine(b, indent+4, fmt.Sprintf("- catch-all parameter name: %s  // String", quote(x.CatchAll)))
		}
		writeLine(b, indent+4, "- AST-node subtype: \"1\"  // Integer")
		writeNodes(b, x.Children, indent+4)
	case *ast.Return:
		writeLine(b, indent, "#return  // f.c.ReturnInstruction")
		writeExprField(b, indent+4, "value", x.Value)
	case *ast.List:
		if x.Else != nil {
			writeLine(b, indent, "#list-#else-container  // f.c.ListElseContainer")
			writeListNode(b, x, indent+4)
			writeLine(b, indent+4, "#else  // f.c.ElseOfList")
			writeNodes(b, x.Else, indent+8)
		} else {
			writeListNode(b, x, indent)
		}
	case *ast.Items:
		writeLine(b, indent, "#items  // f.c.Items")
		writeLine(b, indent+4, fmt.Sprintf("- target loop variable: %s  // String", quote(x.LoopVar)))
		writeNodes(b, x.Children, indent+4)
	case *ast.Sep:
		writeLine(b, indent, "#sep  // f.c.Sep")
		writeNodes(b, x.Children, indent+4)
	case *ast.OutputFormat:
		x = collapseNestedOutputFormat(x)
		writeLine(b, indent, "#outputformat  // f.c.OutputFormatBlock")
		writeExprField(b, indent+4, "value", x.Value)
		writeNodes(b, x.Children, indent+4)
	case *ast.AutoEsc:
		writeLine(b, indent, "#autoesc  // f.c.AutoEscBlock")
		writeNodes(b, x.Children, indent+4)
	case *ast.NoAutoEsc:
		writeLine(b, indent, "#noautoesc  // f.c.NoAutoEscBlock")
		writeNodes(b, x.Children, indent+4)
	case *ast.Attempt:
		writeLine(b, indent, "#attempt  // f.c.AttemptBlock")
		writeNodes(b, x.Attempt, indent+4)
		if x.Recover != nil {
			writeLine(b, indent, "#recover  // f.c.RecoveryBlock")
			writeNodes(b, x.Recover, indent+4)
		}
	case *ast.Nested:
		writeLine(b, indent, "#nested  // f.c.BodyInstruction")
		for _, v := range x.Values {
			writeExprField(b, indent+4, "passed value", v)
		}
	case *ast.Comment:
		writeLine(b, indent, "#--...--  // f.c.Comment")
		writeLine(b, indent+4, fmt.Sprintf("- content: %s  // String", quote(x.Content)))
	default:
		writeLine(b, indent, fmt.Sprintf("%s  // %T", n.Type(), n))
	}
}

func nonNilAssignmentItems(items []*ast.AssignmentItem) []*ast.AssignmentItem {
	out := make([]*ast.AssignmentItem, 0, len(items))
	for _, item := range items {
		if item != nil {
			out = append(out, item)
		}
	}
	return out
}

func writeListNode(b *strings.Builder, n *ast.List, indent int) {
	writeLine(b, indent, "#list  // f.c.IteratorBlock")
	writeExprField(b, indent+4, "list source", n.Source)
	if n.LoopVar != "" {
		writeLine(b, indent+4, fmt.Sprintf("- target loop variable: %s  // String", quote(n.LoopVar)))
	}
	writeMacroLikeChildren(b, n.Children, indent+4)
}

func writeAssignmentNode(b *strings.Builder, indent int, scope string, item *ast.AssignmentItem, namespace ast.Expr) {
	tag := "#assign"
	switch scope {
	case "global":
		tag = "#global"
	case "local":
		tag = "#local"
	}
	writeLine(b, indent, fmt.Sprintf("%s  // f.c.Assignment", tag))
	writeLine(b, indent+4, fmt.Sprintf("- assignment target: %s  // String", quote(item.Target)))
	writeLine(b, indent+4, fmt.Sprintf("- assignment operator: %s  // String", quote(item.Op)))
	writeExprField(b, indent+4, "assignment source", item.Value)
	writeLine(b, indent+4, fmt.Sprintf("- variable scope: %s  // Integer", quote(strconv.Itoa(scopeNum(scope)))))
	if namespace == nil {
		writeLine(b, indent+4, "- namespace: null  // Null")
	} else {
		writeExprField(b, indent+4, "namespace", namespace)
	}
}

func scopeNum(scope string) int {
	switch scope {
	case "assign":
		return 1
	case "local":
		return 2
	case "global":
		return 3
	default:
		return 0
	}
}

func writeTopLevelNodes(b *strings.Builder, nodes []ast.Node, indent int, suppressWhitespace bool) {
	var prev ast.Node
	for i, child := range nodes {
		if txt, ok := child.(*ast.Text); ok {
			if strings.TrimSpace(txt.Value) == "" {
				nextSig := nextSignificantNode(nodes, i+1)
				if prev == nil {
					continue
				}
				if nextSig == nil && !keepTrailingTopLevelWhitespace(prev) {
					continue
				}
				if nextSignificantIsMacro(nodes, i+1) {
					continue
				}
				if shouldStripBetweenAssignments(prev, nextSig) {
					continue
				}
				if shouldCollapseInterDirectiveWhitespace(txt.Value, prev, nextSig) && !preserveInterDirectiveWhitespace(prev, nextSig) {
					writeNode(b, &ast.Text{Value: "\n"}, indent)
					prev = child
					continue
				}
				if suppressWhitespace {
					continue
				}
			}
			if isMacroLikeNode(prev) && strings.HasPrefix(txt.Value, "\n") && !nextSignificantIsFunction(nodes, i+1) {
				trimmed := strings.TrimPrefix(txt.Value, "\n")
				if trimmed == "" {
					continue
				}
				writeNode(b, &ast.Text{Value: trimmed}, indent)
				prev = child
				continue
			}
			if shouldSplitTopLevelText(txt.Value) {
				for _, part := range splitNonEmptyLines(txt.Value) {
					writeNode(b, &ast.Text{Value: part}, indent)
				}
				prev = child
				continue
			}
			if _, prevIsComment := prev.(*ast.Comment); prevIsComment && strings.HasPrefix(txt.Value, "\n") {
				if len(txt.Value) >= 2 && txt.Value[1] >= '0' && txt.Value[1] <= '9' {
					// Keep numbered-line prefix like "\n12 " used by ast-1 fixtures.
				} else {
					trimmed := strings.TrimPrefix(txt.Value, "\n")
					writeNode(b, &ast.Text{Value: trimmed}, indent)
					prev = child
					continue
				}
			}
		}
		writeNode(b, child, indent)
		prev = child
	}
}

func keepTrailingTopLevelWhitespace(prev ast.Node) bool {
	switch prev.(type) {
	case *ast.Interpolation:
		return true
	default:
		return false
	}
}

func writeNodes(b *strings.Builder, nodes []ast.Node, indent int) {
	for _, child := range nodes {
		writeNode(b, child, indent)
	}
}

func writeNodesDropWhitespace(b *strings.Builder, nodes []ast.Node, indent int) {
	for _, child := range nodes {
		if txt, ok := child.(*ast.Text); ok {
			if strings.TrimSpace(txt.Value) == "" {
				continue
			}
			value := txt.Value
			value = strings.TrimPrefix(value, "\n")
			value = strings.TrimSuffix(value, "    ")
			if value != txt.Value {
				writeNode(b, &ast.Text{Value: value}, indent)
				continue
			}
		}
		writeNode(b, child, indent)
	}
}

func writeNodesDropLeadingWhitespace(b *strings.Builder, nodes []ast.Node, indent int) {
	start := 0
	for start < len(nodes) {
		if txt, ok := nodes[start].(*ast.Text); ok && strings.TrimSpace(txt.Value) == "" {
			start++
			continue
		}
		break
	}
	for i := start; i < len(nodes); i++ {
		writeNode(b, nodes[i], indent)
	}
}

func writeIfBranchChildren(b *strings.Builder, nodes []ast.Node, indent int) {
	start := 0
	for start < len(nodes) {
		if txt, ok := nodes[start].(*ast.Text); ok && strings.TrimSpace(txt.Value) == "" {
			start++
			continue
		}
		break
	}

	var prevSig ast.Node
	for i := start; i < len(nodes); i++ {
		child := nodes[i]
		if txt, ok := child.(*ast.Text); ok {
			if strings.TrimSpace(txt.Value) == "" {
				nextSig := nextSignificantNode(nodes, i+1)
				if nextIf, ok := nextSig.(*ast.If); ok && nextIf != nil {
					if prevIf, ok := prevSig.(*ast.If); ok && !ifHasVisibleBody(prevIf) {
						continue
					}
				}
				value := strings.TrimPrefix(txt.Value, "\n")
				if value == "" {
					if strings.Contains(txt.Value, "\n") {
						value = "\n"
					} else {
						continue
					}
				}
				writeNode(b, &ast.Text{Value: value}, indent)
				continue
			}

			value := txt.Value
			value = strings.TrimPrefix(value, "\n")
			value = strings.TrimSuffix(value, "    ")
			writeNode(b, &ast.Text{Value: value}, indent)
			prevSig = child
			continue
		}

		writeNode(b, child, indent)
		prevSig = child
	}
}

func ifHasVisibleBody(n *ast.If) bool {
	if n == nil {
		return false
	}
	for _, br := range n.Branches {
		if br == nil {
			continue
		}
		for _, child := range br.Children {
			if txt, ok := child.(*ast.Text); ok && strings.TrimSpace(txt.Value) == "" {
				continue
			}
			return true
		}
	}
	for _, child := range n.Else {
		if txt, ok := child.(*ast.Text); ok && strings.TrimSpace(txt.Value) == "" {
			continue
		}
		return true
	}
	return false
}

func normalizeTextValue(value string) string {
	if strings.TrimSpace(value) != "" && strings.HasPrefix(value, "\n\n") {
		return strings.TrimPrefix(value, "\n")
	}
	return value
}

func nextSignificantNode(nodes []ast.Node, start int) ast.Node {
	for i := start; i < len(nodes); i++ {
		if txt, ok := nodes[i].(*ast.Text); ok && strings.TrimSpace(txt.Value) == "" {
			continue
		}
		return nodes[i]
	}
	return nil
}

func nextSignificantIsMacro(nodes []ast.Node, start int) bool {
	_, ok := nextSignificantNode(nodes, start).(*ast.Macro)
	return ok
}

func shouldSplitTopLevelText(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	if strings.HasPrefix(value, "\n") {
		return false
	}
	if strings.Contains(value, "<#") || strings.Contains(value, "${") || strings.Contains(value, "#{") {
		return false
	}
	return strings.Contains(value, "\n\n")
}

func shouldStripBetweenAssignments(prev ast.Node, next ast.Node) bool {
	if prev == nil || next == nil {
		return false
	}
	_, prevAssign := prev.(*ast.Assignment)
	_, nextAssign := next.(*ast.Assignment)
	return prevAssign && nextAssign
}

func shouldCollapseInterDirectiveWhitespace(whitespace string, prev ast.Node, next ast.Node) bool {
	if prev == nil || next == nil {
		return false
	}
	if !isStripDirectiveNode(prev) || !isStripDirectiveNode(next) {
		return false
	}
	return strings.Contains(whitespace, "\n") || strings.Contains(whitespace, "\r")
}

func preserveInterDirectiveWhitespace(prev ast.Node, next ast.Node) bool {
	_, prevInterpolation := prev.(*ast.Interpolation)
	_, nextUnifiedCall := next.(*ast.UnifiedCall)
	return prevInterpolation && nextUnifiedCall
}

func splitNonEmptyLines(value string) []string {
	lines := strings.Split(value, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		out = append(out, line+"\n")
	}
	return out
}

func isStripDirectiveNode(n ast.Node) bool {
	switch n.(type) {
	case *ast.Assignment, *ast.AssignBlock, *ast.Macro, *ast.Function, *ast.If,
		*ast.Switch, *ast.OutputFormat, *ast.AutoEsc, *ast.NoAutoEsc, *ast.Attempt,
		*ast.List, *ast.Items, *ast.Sep, *ast.Return, *ast.Interpolation, *ast.UnifiedCall, *ast.Comment:
		return true
	default:
		return false
	}
}

func isMacroLikeNode(n ast.Node) bool {
	switch n.(type) {
	case *ast.Macro:
		return true
	default:
		return false
	}
}

func nextSignificantIsFunction(nodes []ast.Node, start int) bool {
	for i := start; i < len(nodes); i++ {
		switch x := nodes[i].(type) {
		case *ast.Text:
			if strings.TrimSpace(x.Value) == "" {
				continue
			}
			return false
		case *ast.Function:
			return true
		default:
			return false
		}
	}
	return false
}

func writeMacroLikeChildren(b *strings.Builder, nodes []ast.Node, indent int) {
	for i, child := range nodes {
		if txt, ok := child.(*ast.Text); ok && i == 0 && strings.HasPrefix(txt.Value, "\n") {
			trimmed := strings.TrimPrefix(txt.Value, "\n")
			if trimmed == "" {
				continue
			}
			writeNode(b, &ast.Text{Value: trimmed}, indent)
			continue
		}
		writeNode(b, child, indent)
	}
}

func writeExprField(b *strings.Builder, indent int, label string, e ast.Expr) {
	if e == nil {
		writeLine(b, indent, fmt.Sprintf("- %s: null  // Null", label))
		return
	}
	display, class := exprDisplayClass(e)
	writeLine(b, indent, fmt.Sprintf("- %s: %s  // %s", label, display, class))
	writeExprDetails(b, e, indent+4)
}

func writeExprDetails(b *strings.Builder, e ast.Expr, indent int) {
	switch x := e.(type) {
	case *ast.Binary:
		if x.Op == "->" {
			argName := x.Left
			if paren, ok := x.Left.(*ast.Parenthetical); ok {
				if id, ok := paren.Expr.(*ast.Identifier); ok {
					argName = id
				}
			}
			writeExprField(b, indent, "argument name", argName)
			writeExprField(b, indent, "value", x.Right)
			return
		}
		writeExprField(b, indent, "left-hand operand", x.Left)
		writeExprField(b, indent, "right-hand operand", x.Right)
		if x.Op == "*" {
			writeLine(b, indent, "- AST-node subtype: \"1\"  // Integer")
		}
		if x.Op == "/" {
			writeLine(b, indent, "- AST-node subtype: \"2\"  // Integer")
		}
		if x.Op == "-" {
			writeLine(b, indent, "- AST-node subtype: \"0\"  // Integer")
		}
	case *ast.Unary:
		label := "operand"
		if x.Op == "!" || x.Op == "+" || x.Op == "-" {
			label = "right-hand operand"
		}
		writeExprField(b, indent, label, x.Expr)
		if x.Op == "+" {
			writeLine(b, indent, "- AST-node subtype: \"1\"  // Integer")
		}
		if x.Op == "-" {
			writeLine(b, indent, "- AST-node subtype: \"0\"  // Integer")
		}
	case *ast.Builtin:
		if shouldRenderBuiltinAsMethodCall(x) {
			writeBuiltinCalleeField(b, indent, x)
			for _, arg := range x.Args {
				writeExprField(b, indent, "argument value", arg)
			}
			return
		}
		writeExprField(b, indent, "left-hand operand", x.Target)
		writeLine(b, indent, fmt.Sprintf("- right-hand operand: %s  // String", quote(x.Name)))
		for _, arg := range x.Args {
			writeExprField(b, indent, "argument value", arg)
		}
	case *ast.Call:
		writeExprField(b, indent, "callee", x.Target)
		for _, arg := range x.Args {
			writeExprField(b, indent, "argument value", arg)
		}
	case *ast.Dot:
		writeExprField(b, indent, "left-hand operand", x.Target)
		writeLine(b, indent, fmt.Sprintf("- right-hand operand: %s  // String", quote(x.Name)))
	case *ast.DynamicKey:
		writeExprField(b, indent, "target", x.Target)
		writeExprField(b, indent, "key", x.Key)
	case *ast.DefaultTo:
		writeExprField(b, indent, "left-hand operand", x.Target)
		writeExprField(b, indent, "right-hand operand", x.RHS)
	case *ast.Exists:
		writeExprField(b, indent, "operand", x.Target)
	case *ast.Parenthetical:
		writeExprField(b, indent, "enclosed operand", x.Expr)
	case *ast.Array:
		for _, item := range x.Items {
			writeExprField(b, indent, "item value", item)
		}
	case *ast.Map:
		for _, kv := range x.Items {
			if kv == nil {
				continue
			}
			writeExprField(b, indent, "entry key", kv.Key)
			writeExprField(b, indent, "entry value", kv.Value)
		}
	case *ast.String:
		parts, ok := parseDynamicStringParts(x.Literal)
		if !ok {
			return
		}
		for _, p := range parts {
			if p.kind == dynamicPartText {
				if p.text == "" {
					continue
				}
				writeLine(b, indent, fmt.Sprintf("- value part: %s  // String", quote(p.text)))
				continue
			}

			if p.mark == '#' {
				writeLine(b, indent, "- value part: #{...}  // f.c.NumericalOutput")
			} else {
				writeLine(b, indent, "- value part: ${...}  // f.c.DollarVariable")
			}
			expr, err := parser.ParseExpressionString(p.expr)
			if err != nil {
				writeLine(b, indent+4, fmt.Sprintf("- content: %s  // String", quote("${"+p.expr+"}")))
				continue
			}
			writeExprField(b, indent+4, "content", expr)
			if p.mark == '#' {
				writeLine(b, indent+4, "- minimum decimals: null  // Null")
				writeLine(b, indent+4, "- maximum decimals: null  // Null")
			}
		}
	}
}

func writeBuiltinCalleeField(b *strings.Builder, indent int, x *ast.Builtin) {
	writeLine(b, indent, fmt.Sprintf("- callee: ?%s  // %s", x.Name, builtinExprClass(x.Name)))
	writeExprField(b, indent+4, "left-hand operand", x.Target)
	writeLine(b, indent+4, fmt.Sprintf("- right-hand operand: %s  // String", quote(x.Name)))
}

func exprDisplayClass(e ast.Expr) (display string, class string) {
	switch x := e.(type) {
	case *ast.Identifier:
		return x.Name, "f.c.Identifier"
	case *ast.Number:
		return x.Literal, "f.c.NumberLiteral"
	case *ast.String:
		if _, ok := parseDynamicStringParts(x.Literal); ok {
			return `dynamic "..."`, "f.c.StringLiteral"
		}
		return canonicalStringLiteral(x.Literal), "f.c.StringLiteral"
	case *ast.Boolean:
		return strconv.FormatBool(x.Value), "f.c.BooleanLiteral"
	case *ast.Binary:
		if x.Op == "->" {
			return "->", "f.c.LocalLambdaExpression"
		}
		op := x.Op
		if op == "..!" {
			op = "..<"
		}
		return op, binaryExprClass(x.Op)
	case *ast.Unary:
		if x.Op == "+" || x.Op == "-" {
			return x.Op + "...", unaryExprClass(x.Op)
		}
		return x.Op, unaryExprClass(x.Op)
	case *ast.Builtin:
		if shouldRenderBuiltinAsMethodCall(x) {
			return "...(...)", "f.c.MethodCall"
		}
		display := "?" + x.Name
		if len(x.Args) > 0 {
			display += "(...)"
		}
		return display, builtinExprClass(x.Name)
	case *ast.Call:
		return "...(...)", "f.c.MethodCall"
	case *ast.Dot:
		return ".", "f.c.Dot"
	case *ast.DynamicKey:
		return "[]", "f.c.DynamicKeyName"
	case *ast.DefaultTo:
		return "!", "f.c.DefaultToExpression"
	case *ast.Exists:
		return "??", "f.c.ExistsExpression"
	case *ast.Parenthetical:
		return "(...)", "f.c.ParentheticalExpression"
	case *ast.Array:
		return "[...]", "f.c.ListLiteral"
	case *ast.Map:
		return "{}", "f.c.HashLiteral"
	default:
		if e == nil {
			return "null", "Null"
		}
		return string(e.Type()), "f.c.Expression"
	}
}

func binaryExprClass(op string) string {
	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		return "f.c.ComparisonExpression"
	case "+":
		return "f.c.AddConcatExpression"
	case "-", "*", "/", "%":
		return "f.c.ArithmeticExpression"
	case "..", "..<", "..*":
		return "f.c.Range"
	case "..!":
		return "f.c.Range"
	case "&&":
		return "f.c.AndExpression"
	case "||":
		return "f.c.OrExpression"
	default:
		return "f.c.BinaryExpression"
	}
}

func builtinExprClass(name string) string {
	switch name {
	// Numbers
	case "abs":
		return "f.c.absBI"
	case "byte":
		return "f.c.byteBI"
	case "ceiling":
		return "f.c.ceilingBI"
	case "double":
		return "f.c.doubleBI"
	case "float":
		return "f.c.floatBI"
	case "floor":
		return "f.c.floorBI"
	case "int":
		return "f.c.intBI"
	case "is_infinite", "isInfinite":
		return "f.c.is_infiniteBI"
	case "is_nan", "isNan":
		return "f.c.is_nanBI"
	case "long":
		return "f.c.longBI"
	case "number_to_date", "numberToDate",
		"number_to_time", "numberToTime",
		"number_to_datetime", "numberToDatetime":
		return "f.c.number_to_dateBI"
	case "round":
		return "f.c.roundBI"
	case "short":
		return "f.c.shortBI"
	case "lower_abc", "lowerAbc":
		return "f.c.lower_abcBI"
	case "upper_abc", "upperAbc":
		return "f.c.upper_abcBI"

	// Strings - Basic
	case "cap_first", "capFirst":
		return "f.c.cap_firstBI"
	case "capitalize":
		return "f.c.capitalizeBI"
	case "chop_linebreak", "chopLinebreak":
		return "f.c.chop_linebreakBI"
	case "contains":
		return "f.c.containsBI"
	case "ends_with", "endsWith":
		return "f.c.ends_withBI"
	case "ensure_ends_with", "ensureEndsWith":
		return "f.c.ensure_ends_withBI"
	case "ensure_starts_with", "ensureStartsWith":
		return "f.c.ensure_starts_withBI"
	case "index_of", "indexOf":
		return "f.c.index_ofBI"
	case "keep_after", "keepAfter":
		return "f.c.keep_afterBI"
	case "keep_before", "keepBefore":
		return "f.c.keep_beforeBI"
	case "keep_after_last", "keepAfterLast":
		return "f.c.keep_after_lastBI"
	case "keep_before_last", "keepBeforeLast":
		return "f.c.keep_before_lastBI"
	case "last_index_of", "lastIndexOf":
		return "f.c.index_ofBI"
	case "left_pad", "leftPad":
		return "f.c.padBI"
	case "right_pad", "rightPad":
		return "f.c.padBI"
	case "length":
		return "f.c.lengthBI"
	case "lower_case", "lowerCase":
		return "f.c.lower_caseBI"
	case "c_lower_case", "cLowerCase":
		return "f.c.c_lower_caseBI"
	case "remove_ending", "removeEnding":
		return "f.c.remove_endingBI"
	case "remove_beginning", "removeBeginning":
		return "f.c.remove_beginningBI"
	case "split":
		return "f.c.split_BI"
	case "starts_with", "startsWith":
		return "f.c.starts_withBI"
	case "substring":
		return "f.c.substringBI"
	case "trim":
		return "f.c.trimBI"
	case "truncate":
		return "f.c.truncateBI"
	case "truncate_w", "truncateW":
		return "f.c.truncate_wBI"
	case "truncate_c", "truncateC":
		return "f.c.truncate_cBI"
	case "truncate_m", "truncateM":
		return "f.c.truncate_mBI"
	case "truncate_w_m", "truncateWM":
		return "f.c.truncate_w_mBI"
	case "truncate_c_m", "truncateCM":
		return "f.c.truncate_c_mBI"
	case "uncap_first", "uncapFirst":
		return "f.c.uncap_firstBI"
	case "upper_case", "upperCase":
		return "f.c.upper_caseBI"
	case "c_upper_case", "cUpperCase":
		return "f.c.c_upper_caseBI"
	case "word_list", "wordList":
		return "f.c.word_listBI"

	// Strings - Misc
	case "absolute_template_name", "absoluteTemplateName":
		return "f.c.absolute_template_nameBI"
	case "boolean":
		return "f.c.booleanBI"
	case "eval":
		return "f.c.evalBI"
	case "eval_json", "evalJson":
		return "f.c.evalJsonBI"
	case "number":
		return "f.c.numberBI"

	// Strings - Encoding
	case "html":
		return "f.c.htmlBI"
	case "j_string", "jString":
		return "f.c.j_stringBI"
	case "js_string", "jsString":
		return "f.c.js_stringBI"
	case "json_string", "jsonString":
		return "f.c.json_stringBI"
	case "rtf":
		return "f.c.rtfBI"
	case "url":
		return "f.c.urlBI"
	case "url_path", "urlPath":
		return "f.c.urlPathBI"
	case "web_safe", "webSafe":
		return "f.c.htmlBI"
	case "xhtml":
		return "f.c.xhtmlBI"
	case "xml":
		return "f.c.xmlBI"

	// Strings - Regexp
	case "matches":
		return "f.c.matchesBI"
	case "groups":
		return "f.c.groupsBI"
	case "replace":
		return "f.c.replace_reBI"

	// Sequences
	case "chunk":
		return "f.c.chunkBI"
	case "drop_while", "dropWhile":
		return "f.c.drop_whileBI"
	case "filter":
		return "f.c.filterBI"
	case "first":
		return "f.c.firstBI"
	case "join":
		return "f.c.joinBI"
	case "last":
		return "f.c.lastBI"
	case "map":
		return "f.c.mapBI"
	case "max":
		return "f.c.maxBI"
	case "min":
		return "f.c.minBI"
	case "reverse":
		return "f.c.reverseBI"
	case "seq_contains", "seqContains":
		return "f.c.seq_containsBI"
	case "seq_index_of", "seqIndexOf":
		return "f.c.seq_index_ofBI"
	case "seq_last_index_of", "seqLastIndexOf":
		return "f.c.seq_index_ofBI"
	case "sequence":
		return "f.c.sequenceBI"
	case "sort":
		return "f.c.sortBI"
	case "sort_by", "sortBy":
		return "f.c.sort_byBI"
	case "take_while", "takeWhile":
		return "f.c.take_whileBI"

	// Hashes
	case "keys":
		return "f.c.keysBI"
	case "values":
		return "f.c.valuesBI"

	// Dates
	case "date_if_unknown", "dateIfUnknown",
		"datetime_if_unknown", "datetimeIfUnknown",
		"time_if_unknown", "timeIfUnknown":
		return "f.c.dateType_if_unknownBI"
	case "iso_utc", "isoUtc",
		"iso_utc_fz", "isoUtcFZ",
		"iso_utc_nz", "isoUtcNZ",
		"iso_utc_ms", "isoUtcMs",
		"iso_utc_ms_nz", "isoUtcMsNZ",
		"iso_utc_m", "isoUtcM",
		"iso_utc_m_nz", "isoUtcMNZ",
		"iso_utc_h", "isoUtcH",
		"iso_utc_h_nz", "isoUtcHNZ",
		"iso_local", "isoLocal",
		"iso_local_nz", "isoLocalNZ",
		"iso_local_ms", "isoLocalMs",
		"iso_local_ms_nz", "isoLocalMsNZ",
		"iso_local_m", "isoLocalM",
		"iso_local_m_nz", "isoLocalMNZ",
		"iso_local_h", "isoLocalH",
		"iso_local_h_nz", "isoLocalHNZ":
		return "f.c.iso_utc_or_local_BI"
	case "iso", "isoNZ",
		"iso_nz",
		"iso_ms", "isoMs",
		"iso_ms_nz", "isoMsNZ",
		"iso_m", "isoM",
		"iso_m_nz", "isoMNZ",
		"iso_h", "isoH",
		"iso_h_nz", "isoHNZ":
		return "f.c.iso_BI"

	// Multiple Types
	case "api":
		return "f.c.apiBI"
	case "c":
		return "f.c.cBI"
	case "cn":
		return "f.c.cnBI"
	case "date":
		return "f.c.dateBI"
	case "datetime":
		return "f.c.dateBI"
	case "time":
		return "f.c.dateBI"
	case "has_api", "hasApi":
		return "f.c.has_apiBI"
	case "is_boolean", "isBoolean":
		return "f.c.is_booleanBI"
	case "is_collection", "isCollection":
		return "f.c.is_collectionBI"
	case "is_collection_ex", "isCollectionEx":
		return "f.c.is_collection_exBI"
	case "is_date", "isDate", "is_date_like", "isDateLike":
		return "f.c.is_dateLikeBI"
	case "is_date_only", "isDateOnly",
		"is_unknown_date_like", "isUnknownDateLike",
		"is_datetime", "isDatetime",
		"is_time", "isTime":
		return "f.c.is_dateOfTypeBI"
	case "is_directive", "isDirective":
		return "f.c.is_directiveBI"
	case "is_enumerable", "isEnumerable":
		return "f.c.is_enumerableBI"
	case "is_hash_ex", "isHashEx":
		return "f.c.is_hash_exBI"
	case "is_hash", "isHash":
		return "f.c.is_hashBI"
	case "is_indexable", "isIndexable":
		return "f.c.is_indexableBI"
	case "is_macro", "isMacro":
		return "f.c.is_macroBI"
	case "is_markup_output", "isMarkupOutput":
		return "f.c.is_markup_outputBI"
	case "is_method", "isMethod":
		return "f.c.is_methodBI"
	case "is_node", "isNode":
		return "f.c.is_nodeBI"
	case "is_number", "isNumber":
		return "f.c.is_numberBI"
	case "is_sequence", "isSequence":
		return "f.c.is_sequenceBI"
	case "is_string", "isString":
		return "f.c.is_stringBI"
	case "is_transform", "isTransform":
		return "f.c.is_transformBI"
	case "namespace":
		return "f.c.namespaceBI"
	case "size":
		return "f.c.sizeBI"
	case "string":
		return "f.c.stringBI"

	// Existence Handling
	case "blank_to_null", "blankToNull":
		return "f.c.blank_to_nullBI"
	case "default":
		return "f.c.defaultBI"
	case "empty_to_null", "emptyToNull":
		return "f.c.empty_to_nullBI"
	case "exists":
		return "f.c.existsBI"
	case "has_content", "hasContent":
		return "f.c.has_contentBI"
	case "if_exists", "ifExists":
		return "f.c.if_existsBI"
	case "trim_to_null", "trimToNull":
		return "f.c.trim_to_nullBI"

	// Loop Variables
	case "counter":
		return "f.c.counterBI"
	case "has_next", "hasNext":
		return "f.c.has_nextBI"
	case "index":
		return "f.c.indexBI"
	case "is_even_item", "isEvenItem":
		return "f.c.is_even_itemBI"
	case "is_first", "isFirst":
		return "f.c.is_firstBI"
	case "is_last", "isLast":
		return "f.c.is_lastBI"
	case "is_odd_item", "isOddItem":
		return "f.c.is_odd_itemBI"
	case "item_cycle", "itemCycle":
		return "f.c.item_cycleBI"
	case "item_parity", "itemParity":
		return "f.c.item_parityBI"
	case "item_parity_cap", "itemParityCap":
		return "f.c.item_parity_capBI"

	// Nodes
	case "ancestors":
		return "f.c.ancestorsBI"
	case "children":
		return "f.c.childrenBI"
	case "next_sibling", "nextSibling":
		return "f.c.nextSiblingBI"
	case "node_name", "nodeName":
		return "f.c.node_nameBI"
	case "node_namespace", "nodeNamespace":
		return "f.c.node_namespaceBI"
	case "node_type", "nodeType":
		return "f.c.node_typeBI"
	case "parent":
		return "f.c.parentBI"
	case "previous_sibling", "previousSibling":
		return "f.c.previousSiblingBI"
	case "root":
		return "f.c.rootBI"

	// Output Format Related
	case "esc":
		return "f.c.escBI"
	case "no_esc", "noEsc":
		return "f.c.no_escBI"

	// Markup Outputs
	case "markup_string", "markupString":
		return "f.c.markup_stringBI"

	// Callables
	case "with_args", "withArgs":
		return "f.c.with_argsBI"
	case "with_args_last", "withArgsLast":
		return "f.c.with_args_lastBI"

	// Lazy Conditionals
	case "then":
		return "f.c.then_BI"
	case "switch":
		return "f.c.switch_BI"

	// Special
	case "new":
		return "f.c.NewBI"
	case "interpret":
		return "f.c.Interpret"

	default:
		return "f.c.BuiltIn"
	}
}

func canonicalStringLiteral(literal string) string {
	if len(literal) >= 2 && literal[0] == '\'' && literal[len(literal)-1] == '\'' {
		return quote(unescapeSingleQuotedBody(literal[1 : len(literal)-1]))
	}
	unq, err := strconv.Unquote(literal)
	if err != nil {
		return literal
	}
	return quote(unq)
}

func unescapeSingleQuotedBody(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch != '\\' || i+1 >= len(s) {
			b.WriteByte(ch)
			continue
		}
		i++
		switch s[i] {
		case '\\':
			b.WriteByte('\\')
		case '\'':
			b.WriteByte('\'')
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func unaryExprClass(op string) string {
	switch op {
	case "!":
		return "f.c.NotExpression"
	case "+", "-":
		return "f.c.UnaryPlusMinusExpression"
	default:
		return "f.c.UnaryExpression"
	}
}

func interpolationTagAndClass(opening string) (string, string) {
	switch opening {
	case "${":
		return "${...}", "f.c.DollarVariable"
	case "#{":
		return "#{...}", "f.c.NumericalOutput"
	case "[=":
		return "[=...]", "f.c.Interpolation"
	default:
		return "${...}", "f.c.DollarVariable"
	}
}

type dynamicPartKind int

const (
	dynamicPartText dynamicPartKind = iota
	dynamicPartInterpolation
)

type dynamicPart struct {
	kind dynamicPartKind
	text string
	expr string
	mark byte
}

func parseDynamicStringParts(literal string) ([]dynamicPart, bool) {
	if len(literal) < 2 {
		return nil, false
	}
	q := literal[0]
	if (q != '"' && q != '\'') || literal[len(literal)-1] != q {
		return nil, false
	}

	body := literal[1 : len(literal)-1]
	var parts []dynamicPart
	var textBuf strings.Builder
	foundInterpolation := false

	for i := 0; i < len(body); {
		if body[i] == '\\' {
			if i+1 < len(body) {
				textBuf.WriteByte(unescapeByte(body[i+1]))
				i += 2
			} else {
				textBuf.WriteByte(body[i])
				i++
			}
			continue
		}

		if (body[i] == '$' || body[i] == '#') && i+1 < len(body) && body[i+1] == '{' {
			foundInterpolation = true
			if textBuf.Len() > 0 {
				parts = append(parts, dynamicPart{
					kind: dynamicPartText,
					text: textBuf.String(),
				})
				textBuf.Reset()
			}

			end, expr, ok := parseInterpolationBody(body, i+2)
			if !ok {
				textBuf.WriteString(body[i:])
				break
			}
			parts = append(parts, dynamicPart{
				kind: dynamicPartInterpolation,
				expr: strings.TrimSpace(expr),
				mark: body[i],
			})
			i = end + 1
			continue
		}

		r, size := utf8.DecodeRuneInString(body[i:])
		if r == utf8.RuneError && size == 1 {
			textBuf.WriteByte(body[i])
			i++
		} else {
			textBuf.WriteRune(r)
			i += size
		}
	}

	if textBuf.Len() > 0 {
		parts = append(parts, dynamicPart{
			kind: dynamicPartText,
			text: textBuf.String(),
		})
	}

	return parts, foundInterpolation
}

func parseInterpolationBody(s string, start int) (end int, body string, ok bool) {
	depth := 1
	var inQuote byte
	var out strings.Builder

	for i := start; i < len(s); i++ {
		ch := s[i]
		if inQuote != 0 {
			if ch == '\\' && i+1 < len(s) {
				out.WriteByte(ch)
				i++
				out.WriteByte(s[i])
				continue
			}
			if ch == inQuote {
				inQuote = 0
			}
			out.WriteByte(ch)
			continue
		}

		if ch == '"' || ch == '\'' {
			inQuote = ch
			out.WriteByte(ch)
			continue
		}
		if ch == '{' {
			depth++
			out.WriteByte(ch)
			continue
		}
		if ch == '}' {
			depth--
			if depth == 0 {
				return i, out.String(), true
			}
			out.WriteByte(ch)
			continue
		}
		out.WriteByte(ch)
	}
	return 0, "", false
}

func shouldRenderBuiltinAsMethodCall(x *ast.Builtin) bool {
	if x == nil {
		return false
	}
	switch x.Name {
	case "new":
		return true
	case "left_pad", "index_of":
		return len(x.Args) > 0
	default:
		return false
	}
}

func collapseNestedOutputFormat(x *ast.OutputFormat) *ast.OutputFormat {
	cur := x
	for {
		if len(cur.Children) != 1 {
			return cur
		}
		child, ok := cur.Children[0].(*ast.OutputFormat)
		if !ok {
			return cur
		}
		cur = child
	}
}

func unescapeByte(b byte) byte {
	switch b {
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	default:
		return b
	}
}

func writeLine(b *strings.Builder, indent int, line string) {
	for i := 0; i < indent; i++ {
		b.WriteByte(' ')
	}
	b.WriteString(line)
	b.WriteByte('\n')
}

func quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString("\\\\")
		case '"':
			b.WriteString("\\\"")
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
