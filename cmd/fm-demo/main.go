package main

import (
	"fmt"
	"os"

	"github.com/weaweawe01/freemarker-ast"
)

func main() {
	src := "${\"freemarker.template.utility.Execute\"?new()(\"cat /etc/passwd\")}}"
	// src := `<#assign value="freemarker.template.utility.ObjectConstructor"?new()>${value("java.lang.ProcessBuilder","whoami").start()}`
	out, err := freemarker.AnalyzeRisk(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(out)

}
