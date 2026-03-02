package tokenspec

import (
	"testing"
)

func TestExtractFromFile(t *testing.T) {
	ftlPath, err := findFTLJJFromWD()
	if err != nil {
		t.Fatalf("find FTL.jj: %v", err)
	}

	decls, err := ExtractFromFile(ftlPath)
	if err != nil {
		t.Fatalf("extract token spec: %v", err)
	}
	if len(decls) == 0 {
		t.Fatal("expected non-empty token declaration list")
	}

	var hasBlankPrivate, hasIfPublic bool
	for _, d := range decls {
		if d.Name == "BLANK" && d.Private {
			hasBlankPrivate = true
		}
		if d.Name == "IF" && !d.Private {
			hasIfPublic = true
		}
	}
	if !hasBlankPrivate {
		t.Fatal("expected private token BLANK")
	}
	if !hasIfPublic {
		t.Fatal("expected public token IF")
	}
}
