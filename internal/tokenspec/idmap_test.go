package tokenspec

import (
	"testing"
)

func TestAssignIDs(t *testing.T) {
	ftlPath, err := findFTLJJFromWD()
	if err != nil {
		t.Fatalf("find FTL.jj: %v", err)
	}
	decls, err := ExtractFromFile(ftlPath)
	if err != nil {
		t.Fatalf("extract token spec: %v", err)
	}

	ids := AssignIDs(decls)
	if len(ids) != len(decls)+1 {
		t.Fatalf("id count mismatch: got %d want %d", len(ids), len(decls)+1)
	}

	nameToID, err := ToNameToID(ids)
	if err != nil {
		t.Fatalf("to name map: %v", err)
	}

	if got := nameToID["EOF"]; got != 0 {
		t.Fatalf("EOF id mismatch: got %d want 0", got)
	}
	if got := nameToID["IF"]; got != 8 {
		t.Fatalf("IF id mismatch: got %d want 8", got)
	}
	if got := nameToID["STATIC_TEXT_WS"]; got != 80 {
		t.Fatalf("STATIC_TEXT_WS id mismatch: got %d want 80", got)
	}
	if got := nameToID["DOLLAR_INTERPOLATION_OPENING"]; got != 83 {
		t.Fatalf("DOLLAR_INTERPOLATION_OPENING id mismatch: got %d want 83", got)
	}
	if got := nameToID["DIRECTIVE_END"]; got != 142 {
		t.Fatalf("DIRECTIVE_END id mismatch: got %d want 142", got)
	}
	if got := nameToID["NATURAL_GT"]; got != 144 {
		t.Fatalf("NATURAL_GT id mismatch: got %d want 144", got)
	}
}
