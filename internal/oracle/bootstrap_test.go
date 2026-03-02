package oracle

import (
	"path/filepath"
	"testing"

	"github.com/weaweawe01/freemarker-ast/internal/corpus"
)

func TestBuildBootstrapBundles(t *testing.T) {
	coreRoot, err := corpus.FindCoreRootFromWD()
	if err != nil {
		t.Fatalf("find core root: %v", err)
	}

	bundles, err := BuildBootstrapBundles(coreRoot)
	if err != nil {
		t.Fatalf("build bootstrap bundles: %v", err)
	}
	if len(bundles) == 0 {
		t.Fatal("expected non-empty bundle list")
	}

	var hasAST, hasCanonical bool
	for _, b := range bundles {
		if b.AST != "" {
			hasAST = true
		}
		if b.Canonical != "" {
			hasCanonical = true
		}
	}
	if !hasAST {
		t.Fatal("expected at least one AST bundle payload")
	}
	if !hasCanonical {
		t.Fatal("expected at least one canonical bundle payload")
	}
}

func TestWriteBundles(t *testing.T) {
	coreRoot, err := corpus.FindCoreRootFromWD()
	if err != nil {
		t.Fatalf("find core root: %v", err)
	}
	bundles, err := BuildBootstrapBundles(coreRoot)
	if err != nil {
		t.Fatalf("build bundles: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "oracle")
	if err := WriteBundles(outDir, bundles); err != nil {
		t.Fatalf("write bundles: %v", err)
	}

	if len(bundles) > 0 {
		firstPath := filepath.Join(outDir, sanitizeCaseName(bundles[0].CaseName)+".json")
		var got OracleBundle
		if err := LoadJSON(firstPath, &got); err != nil {
			t.Fatalf("load bundle back: %v", err)
		}
		if got.CaseName != bundles[0].CaseName {
			t.Fatalf("bundle case name mismatch: got %q want %q", got.CaseName, bundles[0].CaseName)
		}
	}
}
