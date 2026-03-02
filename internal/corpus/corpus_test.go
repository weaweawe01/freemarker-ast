package corpus

import "testing"

func TestDiscoverAndValidateCoreCorpus(t *testing.T) {
	root, err := FindCoreRootFromWD()
	if err != nil {
		t.Fatalf("find core root: %v", err)
	}

	corpus, err := Discover(root)
	if err != nil {
		t.Fatalf("discover corpus: %v", err)
	}
	if len(corpus.Cases) == 0 {
		t.Fatal("expected non-empty corpus")
	}

	if err := corpus.Validate(); err != nil {
		t.Fatalf("validate corpus: %v", err)
	}
}

func TestKnownPairings(t *testing.T) {
	root, err := FindCoreRootFromWD()
	if err != nil {
		t.Fatalf("find core root: %v", err)
	}

	corpus, err := Discover(root)
	if err != nil {
		t.Fatalf("discover corpus: %v", err)
	}

	astCase, ok := corpus.ByName("ast-1")
	if !ok {
		t.Fatal("expected ast-1 case")
	}
	if astCase.FTLPath == "" || astCase.ASTPath == "" {
		t.Fatalf("ast-1 must have .ftl and .ast, got %#v", astCase)
	}

	canoCase, ok := corpus.ByName("cano-builtins")
	if !ok {
		t.Fatal("expected cano-builtins case")
	}
	if canoCase.FTLPath == "" || canoCase.CanonicalOut == "" {
		t.Fatalf("cano-builtins must have .ftl and .ftl.out, got %#v", canoCase)
	}

	encCase, ok := corpus.ByName("encodingOverride-UTF-8")
	if !ok {
		t.Fatal("expected encodingOverride-UTF-8 case")
	}
	if encCase.FTLPath == "" {
		t.Fatalf("encodingOverride-UTF-8 must have .ftl, got %#v", encCase)
	}
}
