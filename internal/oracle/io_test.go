package oracle

import (
	"path/filepath"
	"testing"

	"github.com/weaweawe01/freemarker-ast/internal/compat"
)

func TestSaveAndLoadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bundle.json")

	want := OracleBundle{
		CaseName: "ast-1",
		Tokens: []compat.Token{
			{
				Kind:  1,
				Image: "<#if",
				Begin: compat.Position{Line: 1, Column: 1},
				End:   compat.Position{Line: 1, Column: 4},
			},
		},
		AST: "IfBlock(...)",
	}

	if err := SaveJSON(path, want); err != nil {
		t.Fatalf("save json: %v", err)
	}

	var got OracleBundle
	if err := LoadJSON(path, &got); err != nil {
		t.Fatalf("load json: %v", err)
	}

	if got.CaseName != want.CaseName {
		t.Fatalf("case name mismatch: got %q want %q", got.CaseName, want.CaseName)
	}
	if got.AST != want.AST {
		t.Fatalf("ast mismatch: got %q want %q", got.AST, want.AST)
	}
	if len(got.Tokens) != len(want.Tokens) {
		t.Fatalf("token length mismatch: got %d want %d", len(got.Tokens), len(want.Tokens))
	}
	if got.Tokens[0].Image != want.Tokens[0].Image {
		t.Fatalf("token image mismatch: got %q want %q", got.Tokens[0].Image, want.Tokens[0].Image)
	}
}
