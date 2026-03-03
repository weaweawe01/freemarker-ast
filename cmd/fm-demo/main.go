package main

import (
	"fmt"
	"os"

	"github.com/weaweawe01/freemarker-ast"
)

func main() {
	src := `<#assign ex="freemarker.template.utility.Execute"?new()> ${ ex("id") }`
	out, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(out.TotalScore)

}
