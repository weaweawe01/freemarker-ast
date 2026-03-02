package tokenspec

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var tokenDeclPattern = regexp.MustCompile(`<(#?)([A-Za-z_][A-Za-z0-9_]*)\s*:`)

// TokenDecl is a token declaration found in FTL.jj.
type TokenDecl struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
	Line    int    `json:"line"`
	Order   int    `json:"order"`
}

// ExtractFromFile parses token declarations from an FTL.jj file.
func ExtractFromFile(path string) ([]TokenDecl, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer f.Close()

	var decls []TokenDecl
	sc := bufio.NewScanner(f)
	lineNo := 0
	order := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		m := tokenDeclPattern.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		order++
		decls = append(decls, TokenDecl{
			Name:    m[2],
			Private: m[1] == "#",
			Line:    lineNo,
			Order:   order,
		})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan %q: %w", path, err)
	}
	return decls, nil
}
