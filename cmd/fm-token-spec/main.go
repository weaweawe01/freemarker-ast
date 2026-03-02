package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/weaweawe01/freemarker-ast/internal/oracle"
	"github.com/weaweawe01/freemarker-ast/internal/tokenspec"
)

func main() {
	var ftlPath string
	var outPath string

	flag.StringVar(&ftlPath, "ftl-jj", "", "path to main/javacc/freemarker/core/FTL.jj")
	flag.StringVar(&outPath, "out", "oracle/token-spec.json", "output json path")
	flag.Parse()

	if ftlPath == "" {
		ftlPath = filepath.Join("..", "main", "javacc", "freemarker", "core", "FTL.jj")
	}

	decls, err := tokenspec.ExtractFromFile(ftlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "extract token spec: %v\n", err)
		os.Exit(1)
	}

	if err := oracle.SaveJSON(outPath, decls); err != nil {
		fmt.Fprintf(os.Stderr, "save token spec: %v\n", err)
		os.Exit(1)
	}

	var privateCount, publicCount int
	for _, d := range decls {
		if d.Private {
			privateCount++
		} else {
			publicCount++
		}
	}
	fmt.Printf("token spec generated\n")
	fmt.Printf("input: %s\n", ftlPath)
	fmt.Printf("output: %s\n", outPath)
	fmt.Printf("total: %d\n", len(decls))
	fmt.Printf("public: %d\n", publicCount)
	fmt.Printf("private: %d\n", privateCount)
}
