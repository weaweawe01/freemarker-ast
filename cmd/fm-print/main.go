package main

import (
	"fmt"
	"log"

	freemarker "github.com/weaweawe01/freemarker-ast"
	"github.com/weaweawe01/freemarker-ast/internal/ast"
)

func main() {
	root, err := freemarker.Parse(`<#assign x = "hello">${x?upper_case}`)
	if err != nil {
		log.Fatal(err)
	}
	for _, child := range root.Children {
		walk(child, 0)
	}
}

func walk(node ast.Node, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	switch n := node.(type) {
	case *ast.Text:
		fmt.Printf("%sText: %q\n", indent, n.Value)
	case *ast.Interpolation:
		fmt.Printf("%sInterpolation (opening=%s)\n", indent, n.Opening)
		walkExpr(n.Expr, depth+1)
	case *ast.If:
		fmt.Printf("%sIf (%d branches)\n", indent, len(n.Branches))
		for _, br := range n.Branches {
			walkExpr(br.Condition, depth+1)
			for _, c := range br.Children {
				walk(c, depth+1)
			}
		}
		for _, c := range n.Else {
			walk(c, depth+1)
		}
	case *ast.Assignment:
		fmt.Printf("%sAssignment (scope=%s)\n", indent, n.Scope)
		for _, item := range n.Items {
			fmt.Printf("%s  %s %s\n", indent, item.Target, item.Op)
			if item.Value != nil {
				walkExpr(item.Value, depth+2)
			}
		}
	case *ast.List:
		fmt.Printf("%sList (var=%s)\n", indent, n.LoopVar)
		walkExpr(n.Source, depth+1)
		for _, c := range n.Children {
			walk(c, depth+1)
		}
	case *ast.Macro:
		fmt.Printf("%sMacro: %s\n", indent, n.Name)
		for _, c := range n.Children {
			walk(c, depth+1)
		}
	case *ast.UnifiedCall:
		fmt.Printf("%sUnifiedCall\n", indent)
		walkExpr(n.Callee, depth+1)
	case *ast.Comment:
		fmt.Printf("%sComment: %q\n", indent, n.Content)
	default:
		fmt.Printf("%s%s\n", indent, node.Type())
	}
}

func walkExpr(expr ast.Expr, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		fmt.Printf("%sIdentifier: %s\n", indent, e.Name)
	case *ast.String:
		fmt.Printf("%sString: %q\n", indent, e.Literal)
	case *ast.Number:
		fmt.Printf("%sNumber: %s\n", indent, e.Literal)
	case *ast.Boolean:
		fmt.Printf("%sBoolean: %v\n", indent, e.Value)
	case *ast.Binary:
		fmt.Printf("%sBinary: %s\n", indent, e.Op)
		walkExpr(e.Left, depth+1)
		walkExpr(e.Right, depth+1)
	case *ast.Unary:
		fmt.Printf("%sUnary: %s\n", indent, e.Op)
		walkExpr(e.Expr, depth+1)
	case *ast.Dot:
		fmt.Printf("%sDot: .%s\n", indent, e.Name)
		walkExpr(e.Target, depth+1)
	case *ast.Builtin:
		fmt.Printf("%sBuiltin: ?%s\n", indent, e.Name)
		walkExpr(e.Target, depth+1)
	case *ast.Call:
		fmt.Printf("%sCall\n", indent)
		walkExpr(e.Target, depth+1)
		for _, a := range e.Args {
			walkExpr(a, depth+1)
		}
	case *ast.DynamicKey:
		fmt.Printf("%sDynamicKey\n", indent)
		walkExpr(e.Target, depth+1)
		walkExpr(e.Key, depth+1)
	case *ast.DefaultTo:
		fmt.Printf("%sDefaultTo\n", indent)
		walkExpr(e.Target, depth+1)
	case *ast.Exists:
		fmt.Printf("%sExists\n", indent)
		walkExpr(e.Target, depth+1)
	case *ast.Array:
		fmt.Printf("%sArray (%d items)\n", indent, len(e.Items))
	case *ast.Map:
		fmt.Printf("%sMap (%d entries)\n", indent, len(e.Items))
	default:
		fmt.Printf("%sExpr(%s)\n", indent, expr.Type())
	}
}
