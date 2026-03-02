package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/weaweawe01/freemarker-ast/internal/corpus"
	"github.com/weaweawe01/freemarker-ast/internal/parser"
)

func main() {
	coreRoot, err := corpus.FindCoreRootFromWD()
	if err != nil {
		panic(err)
	}
	entries, err := os.ReadDir(coreRoot)
	if err != nil {
		panic(err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if filepath.Ext(n) == ".ftl" {
			names = append(names, n)
		}
	}
	sort.Strings(names)

	failed := 0
	for _, n := range names {
		path := filepath.Join(coreRoot, n)
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("READ_FAIL %s: %v\n", n, err)
			failed++
			continue
		}
		if _, err := parser.Parse(string(src)); err != nil {
			fmt.Printf("PARSE_FAIL %s: %v\n", n, err)
			failed++
			continue
		}
		fmt.Printf("PARSE_OK %s\n", n)
	}

	fmt.Printf("TOTAL=%d FAILED=%d\n", len(names), failed)
	if failed > 0 {
		os.Exit(1)
	}
}
