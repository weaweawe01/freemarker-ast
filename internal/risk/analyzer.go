package risk

import (
	"fmt"
	"strings"

	"github.com/weaweawe01/freemarker-ast/internal/ast"
)

type valueState struct {
	ClassName    string
	Instantiated bool
	Malicious    bool
}

type exprInfo struct {
	Value            valueState
	HasMalicious     bool
	HasInstantiation bool
	HasExecution     bool
	HasAPIAccess     bool
	HasFileRead      bool
}

func (i *exprInfo) merge(other exprInfo) {
	if other.Value.ClassName != "" {
		i.Value = other.Value
	}
	i.HasMalicious = i.HasMalicious || other.HasMalicious
	i.HasInstantiation = i.HasInstantiation || other.HasInstantiation
	i.HasExecution = i.HasExecution || other.HasExecution
	i.HasAPIAccess = i.HasAPIAccess || other.HasAPIAccess
	i.HasFileRead = i.HasFileRead || other.HasFileRead
}

// Analyze runs static risk analysis against a parsed FreeMarker AST.
func Analyze(root *ast.Root) *Report {
	a := analyzer{
		vars: make(map[string]valueState),
	}
	a.visitNodes(root.Children)
	return &Report{
		TotalScore: a.totalScore,
		Severity:   scoreToSeverity(a.totalScore),
		Findings:   a.findings,
	}
}

type analyzer struct {
	vars       map[string]valueState
	findings   []Finding
	totalScore int
}

func (a *analyzer) addFinding(rule string, score int, message string, evidence string) {
	if score <= 0 {
		return
	}
	a.findings = append(a.findings, Finding{
		Rule:     rule,
		Score:    score,
		Message:  message,
		Evidence: evidence,
	})
	a.totalScore += score
}

func (a *analyzer) visitNodes(nodes []ast.Node) {
	for _, n := range nodes {
		a.visitNode(n)
	}
}

func (a *analyzer) visitNode(node ast.Node) {
	switch n := node.(type) {
	case *ast.Root:
		a.visitNodes(n.Children)
	case *ast.Text:
		// Text blocks are not scored by default to reduce false positives.
	case *ast.Interpolation:
		a.evalTopExpr(n.Expr, "interpolation")
	case *ast.If:
		for idx, br := range n.Branches {
			a.evalTopExpr(br.Condition, fmt.Sprintf("if.branch[%d].condition", idx))
			a.visitNodes(br.Children)
		}
		a.visitNodes(n.Else)
	case *ast.Assignment:
		for _, item := range n.Items {
			if item.Value == nil {
				continue
			}
			info := a.evalTopExpr(item.Value, "assign."+item.Target)
			if item.Op == "=" {
				if info.Value.ClassName != "" {
					a.vars[item.Target] = info.Value
				} else {
					delete(a.vars, item.Target)
				}
			}
		}
		if n.Namespace != nil {
			a.evalTopExpr(n.Namespace, "assign.namespace")
		}
	case *ast.AssignBlock:
		if n.Namespace != nil {
			a.evalTopExpr(n.Namespace, "assign_block.namespace")
		}
		a.visitNodes(n.Children)
	case *ast.Macro:
		for _, p := range n.Params {
			if p.Default != nil {
				a.evalTopExpr(p.Default, "macro.param."+p.Name)
			}
		}
		a.visitNodes(n.Children)
	case *ast.Function:
		for _, p := range n.Params {
			if p.Default != nil {
				a.evalTopExpr(p.Default, "function.param."+p.Name)
			}
		}
		a.visitNodes(n.Children)
	case *ast.Return:
		if n.Value != nil {
			a.evalTopExpr(n.Value, "return")
		}
	case *ast.List:
		a.evalTopExpr(n.Source, "list.source")
		a.visitNodes(n.Children)
		a.visitNodes(n.Else)
	case *ast.Items:
		a.visitNodes(n.Children)
	case *ast.Sep:
		a.visitNodes(n.Children)
	case *ast.Switch:
		a.evalTopExpr(n.Value, "switch.value")
		for idx, br := range n.Branches {
			for cidx, cond := range br.Conditions {
				a.evalTopExpr(cond, fmt.Sprintf("switch.branch[%d].condition[%d]", idx, cidx))
			}
			a.visitNodes(br.Children)
		}
		a.visitNodes(n.Default)
	case *ast.OutputFormat:
		a.evalTopExpr(n.Value, "outputformat.value")
		a.visitNodes(n.Children)
	case *ast.AutoEsc:
		a.visitNodes(n.Children)
	case *ast.NoAutoEsc:
		a.visitNodes(n.Children)
	case *ast.Compress:
		a.visitNodes(n.Children)
	case *ast.Attempt:
		a.visitNodes(n.Attempt)
		a.visitNodes(n.Recover)
	case *ast.UnifiedCall:
		info := a.evalExpr(n.Callee, "unified_call.callee")
		for _, p := range n.Positional {
			info.merge(a.evalExpr(p, "unified_call.arg"))
		}
		for _, named := range n.Named {
			info.merge(a.evalExpr(named.Value, "unified_call.named."+named.Name))
		}
		if id, ok := n.Callee.(*ast.Identifier); ok {
			if st, exists := a.vars[id.Name]; exists && st.Instantiated && st.ClassName == "freemarker.template.utility.jythonruntime" {
				a.addFinding(
					"JYTHON_RUNTIME_CALL",
					90,
					"JythonRuntime object is invoked as directive body",
					id.Name,
				)
				info.HasExecution = true
				info.HasInstantiation = true
				info.HasMalicious = true
			}
		}
		a.applyChainBonus(info, "unified_call")
		a.visitNodes(n.Children)
	case *ast.Comment:
		// Ignored.
	case *ast.Nested:
		for _, v := range n.Values {
			a.evalTopExpr(v, "nested.value")
		}
	}
}

func (a *analyzer) evalTopExpr(expr ast.Expr, context string) exprInfo {
	info := a.evalExpr(expr, context)
	a.applyChainBonus(info, context)
	return info
}

func (a *analyzer) applyChainBonus(info exprInfo, context string) {
	if info.HasMalicious && info.HasInstantiation && info.HasExecution {
		a.addFinding(
			"FULL_EXPLOIT_CHAIN",
			fullExploitChainScore,
			"Detected class->instantiation->execution exploit chain",
			context,
		)
	}
	if info.HasAPIAccess && info.HasFileRead {
		a.addFinding(
			"API_FILE_READ_CHAIN",
			apiFileReadChainScore,
			"Detected ?api based file/resource read chain",
			context,
		)
	}
}

func (a *analyzer) evalExpr(expr ast.Expr, context string) exprInfo {
	if expr == nil {
		return exprInfo{}
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		if st, ok := a.vars[e.Name]; ok {
			return exprInfo{
				Value:            st,
				HasMalicious:     st.Malicious,
				HasInstantiation: st.Instantiated,
			}
		}
		return exprInfo{}
	case *ast.String:
		return a.evalStringLiteral(e.Literal, context)
	case *ast.Number, *ast.Boolean:
		return exprInfo{}
	case *ast.Binary:
		left := a.evalExpr(e.Left, context)
		right := a.evalExpr(e.Right, context)
		left.merge(right)
		return left
	case *ast.Unary:
		return a.evalExpr(e.Expr, context)
	case *ast.Parenthetical:
		return a.evalExpr(e.Expr, context)
	case *ast.Array:
		info := exprInfo{}
		for _, item := range e.Items {
			info.merge(a.evalExpr(item, context))
		}
		return info
	case *ast.Map:
		info := exprInfo{}
		for _, item := range e.Items {
			info.merge(a.evalExpr(item.Key, context))
			info.merge(a.evalExpr(item.Value, context))
		}
		return info
	case *ast.DynamicKey:
		info := a.evalExpr(e.Target, context)
		info.merge(a.evalExpr(e.Key, context))
		return info
	case *ast.DefaultTo:
		info := a.evalExpr(e.Target, context)
		if e.RHS != nil {
			info.merge(a.evalExpr(e.RHS, context))
		}
		return info
	case *ast.Exists:
		return a.evalExpr(e.Target, context)
	case *ast.Builtin:
		return a.evalBuiltinExpr(e, context)
	case *ast.Dot:
		return a.evalExpr(e.Target, context)
	case *ast.Call:
		return a.evalCallExpr(e, context)
	default:
		return exprInfo{}
	}
}

func (a *analyzer) evalStringLiteral(literal string, context string) exprInfo {
	raw := unquoteLiteral(literal)
	lowerRaw := strings.ToLower(raw)
	normalized := normalizeClassName(raw)

	info := exprInfo{}

	if looksLikeClassName(normalized) {
		info.Value.ClassName = normalized
		info.Value.Malicious = isMaliciousClass(normalized)
	}

	if isMaliciousClass(normalized) {
		info.HasMalicious = true
		info.Value.ClassName = normalized
		info.Value.Malicious = true
		a.addFinding(
			"MALICIOUS_CLASS",
			maliciousClassScore,
			"Detected high-risk class reference",
			raw,
		)
		trimmedNoCase := strings.ToLower(strings.TrimSpace(raw))
		if strings.ContainsAny(raw, " \t\n\r") && trimmedNoCase != normalized {
			a.addFinding(
				"OBFUSCATED_CLASS_NAME",
				obfuscatedClassNameScore,
				"Class name contains obfuscating whitespace",
				raw,
			)
		}
	}

	for _, indicator := range commandIndicators {
		if strings.Contains(lowerRaw, indicator) {
			a.addFinding(
				"COMMAND_EVIDENCE",
				commandEvidenceScore,
				"Suspicious command-like content in literal",
				raw,
			)
			break
		}
	}
	for _, indicator := range fileIndicators {
		if strings.Contains(lowerRaw, indicator) {
			a.addFinding(
				"FILE_PATH_EVIDENCE",
				filePathEvidenceScore,
				"Sensitive file/resource path evidence in literal",
				raw,
			)
			info.HasFileRead = true
			break
		}
	}

	_ = context
	return info
}

func (a *analyzer) evalBuiltinExpr(e *ast.Builtin, context string) exprInfo {
	info := a.evalExpr(e.Target, context)
	for _, arg := range e.Args {
		info.merge(a.evalExpr(arg, context))
	}

	if strings.EqualFold(e.Name, "new") {
		a.addFinding(
			"NEW_INSTANCE",
			newInstanceScore,
			"Detected ?new() instantiation",
			context,
		)
		info.HasInstantiation = true
		info.Value.Instantiated = true
		if info.Value.ClassName == "" {
			info.Value.ClassName = normalizeClassName(renderExprBrief(e.Target))
		}
		if info.Value.ClassName != "" && isMaliciousClass(info.Value.ClassName) {
			info.HasMalicious = true
			info.Value.Malicious = true
		}
	}
	if strings.EqualFold(e.Name, "api") {
		a.addFinding(
			"API_BUILTIN",
			apiBuiltinScore,
			"Detected ?api built-in access",
			context,
		)
		info.HasAPIAccess = true
	}

	return info
}

func (a *analyzer) evalCallExpr(e *ast.Call, context string) exprInfo {
	info := exprInfo{}

	methodName := ""
	receiverInfo := exprInfo{}
	switch tgt := e.Target.(type) {
	case *ast.Dot:
		methodName = strings.ToLower(tgt.Name)
		receiverInfo = a.evalExpr(tgt.Target, context)
		info.merge(receiverInfo)
	default:
		info.merge(a.evalExpr(e.Target, context))
	}

	firstArgClass := ""
	for idx, arg := range e.Args {
		argInfo := a.evalExpr(arg, context)
		info.merge(argInfo)
		if idx == 0 {
			if s, ok := arg.(*ast.String); ok {
				firstArgClass = normalizeClassName(unquoteLiteral(s.Literal))
			}
		}
	}

	if id, ok := e.Target.(*ast.Identifier); ok {
		if st, exists := a.vars[id.Name]; exists && st.Instantiated {
			switch st.ClassName {
			case "freemarker.template.utility.execute":
				a.addFinding(
					"EXECUTE_CALL",
					90,
					"Execute utility is invoked",
					id.Name,
				)
				info.HasExecution = true
				info.HasInstantiation = true
				info.HasMalicious = true
			case "freemarker.template.utility.objectconstructor":
				a.addFinding(
					"OBJECT_CONSTRUCTOR_CALL",
					objectConstructorCall,
					"ObjectConstructor is used to create an arbitrary class",
					id.Name,
				)
				info.HasInstantiation = true
				info.HasMalicious = true
				if firstArgClass != "" {
					info.Value = valueState{
						ClassName:    firstArgClass,
						Instantiated: true,
						Malicious:    isMaliciousClass(firstArgClass),
					}
					if info.Value.Malicious {
						info.HasMalicious = true
					}
				}
			case "freemarker.template.utility.jythonruntime":
				a.addFinding(
					"JYTHON_RUNTIME_CALL",
					90,
					"JythonRuntime is invoked as function call",
					id.Name,
				)
				info.HasExecution = true
				info.HasInstantiation = true
				info.HasMalicious = true
			}
		}
	}

	if methodName != "" {
		receiverClass := receiverInfo.Value.ClassName
		if receiverClass == "" {
			receiverClass = info.Value.ClassName
		}
		if receiverClass != "" && legacyUnsafeMethods.has(receiverClass, methodName) {
			a.addFinding(
				"UNSAFE_METHOD_BLACKLIST",
				unsafeMethodMatchScore,
				"Method is listed in freemarker unsafeMethods blacklist",
				receiverClass+"."+methodName+"(...)",
			)
		}
		if score, ok := sensitiveIOMethodScores[methodName]; ok {
			a.addFinding(
				"SENSITIVE_IO_METHOD",
				score,
				"Sensitive IO/resource access method detected",
				methodName,
			)
			info.HasFileRead = true
		}

		if score, ok := criticalMethodScores[methodName]; ok {
			a.addFinding(
				"CRITICAL_METHOD",
				score,
				"Critical method call detected",
				methodName,
			)
		}
		if isExecutionMethod(methodName) {
			info.HasExecution = true
		}
	}

	if info.Value.ClassName == "" {
		info.Value = receiverInfo.Value
	}
	return info
}

func renderExprBrief(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Name
	case *ast.String:
		return unquoteLiteral(e.Literal)
	case *ast.Dot:
		return renderExprBrief(e.Target) + "." + e.Name
	case *ast.Builtin:
		return renderExprBrief(e.Target) + "?" + e.Name
	default:
		return string(expr.Type())
	}
}
