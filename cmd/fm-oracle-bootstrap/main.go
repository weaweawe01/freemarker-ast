package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/weaweawe01/freemarker-ast/internal/corpus"
	"github.com/weaweawe01/freemarker-ast/internal/oracle"
)

func main() {
	var coreRoot string
	var outDir string

	flag.StringVar(&coreRoot, "core-root", "", "path to test/resources/freemarker/core")
	flag.StringVar(&outDir, "out", "oracle/bootstrap", "output directory for oracle bundles")
	flag.Parse()

	if coreRoot == "" {
		autoRoot, err := corpus.FindCoreRootFromWD()
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve core root: %v\n", err)
			os.Exit(2)
		}
		coreRoot = autoRoot
	}

	bundles, err := oracle.BuildBootstrapBundles(coreRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build bundles: %v\n", err)
		os.Exit(1)
	}
	if err := oracle.WriteBundles(outDir, bundles); err != nil {
		fmt.Fprintf(os.Stderr, "write bundles: %v\n", err)
		os.Exit(1)
	}

	var astCount, canonicalCount int
	for _, b := range bundles {
		if b.AST != "" {
			astCount++
		}
		if b.Canonical != "" {
			canonicalCount++
		}
	}

	fmt.Printf("bootstrap oracle generated\n")
	fmt.Printf("core_root: %s\n", coreRoot)
	fmt.Printf("output_dir: %s\n", outDir)
	fmt.Printf("cases: %d\n", len(bundles))
	fmt.Printf("ast_payloads: %d\n", astCount)
	fmt.Printf("canonical_payloads: %d\n", canonicalCount)
}
