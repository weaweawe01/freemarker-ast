package freemarker_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/weaweawe01/freemarker-ast"
)

func TestAstPrint(t *testing.T) {
	src := `<#assign value="freemarker.template.utility.Execute"?new()>${value("calc")}`
	out, err := freemarker.ParseToJavaLikeAST(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(out)
}
