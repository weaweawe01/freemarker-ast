package tokenid

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/weaweawe01/freemarker-ast/internal/tokenspec"
)

func TestLookupHelpers(t *testing.T) {
	id, ok := ID("IF")
	if !ok {
		t.Fatal("missing token IF")
	}
	if id != TK_IF {
		t.Fatalf("IF id mismatch: got %d want %d", id, TK_IF)
	}

	name, ok := Name(TK_DIRECTIVE_END)
	if !ok {
		t.Fatal("missing token by id TK_DIRECTIVE_END")
	}
	if name != "DIRECTIVE_END" {
		t.Fatalf("DIRECTIVE_END name mismatch: got %q", name)
	}
}

func TestGeneratedMapMatchesFTLJJ(t *testing.T) {
	ftlPath, err := findFTLJJFromWD()
	if err != nil {
		t.Fatalf("find FTL.jj: %v", err)
	}
	decls, err := tokenspec.ExtractFromFile(ftlPath)
	if err != nil {
		t.Fatalf("extract token spec: %v", err)
	}
	ids := tokenspec.AssignIDs(decls)

	if len(NameToID) != len(ids) {
		t.Fatalf("NameToID size mismatch: got %d want %d", len(NameToID), len(ids))
	}
	if len(IDToName) != len(ids) {
		t.Fatalf("IDToName size mismatch: got %d want %d", len(IDToName), len(ids))
	}

	for _, tid := range ids {
		gotID, ok := NameToID[tid.Name]
		if !ok {
			t.Fatalf("missing token in NameToID: %s", tid.Name)
		}
		if gotID != tid.ID {
			t.Fatalf("id mismatch for %s: got %d want %d", tid.Name, gotID, tid.ID)
		}
		gotName, ok := IDToName[tid.ID]
		if !ok {
			t.Fatalf("missing id in IDToName: %d", tid.ID)
		}
		if gotName != tid.Name {
			t.Fatalf("name mismatch for id %d: got %s want %s", tid.ID, gotName, tid.Name)
		}
	}
}

func findFTLJJFromWD() (string, error) {
	root, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(root, "main", "javacc", "freemarker", "core", "FTL.jj")
		if st, statErr := os.Stat(candidate); statErr == nil && !st.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}
	return "", os.ErrNotExist
}
