package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/weaweawe01/freemarker-ast/internal/corpus"
)

// BuildBootstrapBundles builds oracle bundles from existing core expected files.
// This is a bootstrap path before Java-side runtime exporters are wired in.
func BuildBootstrapBundles(coreRoot string) ([]OracleBundle, error) {
	cp, err := corpus.Discover(coreRoot)
	if err != nil {
		return nil, err
	}
	if err := cp.Validate(); err != nil {
		return nil, err
	}

	bundles := make([]OracleBundle, 0, len(cp.Cases))
	for _, tc := range cp.Cases {
		b := OracleBundle{
			CaseName: tc.Name,
		}
		if tc.ASTPath != "" {
			ast, err := readTextFile(tc.ASTPath)
			if err != nil {
				return nil, fmt.Errorf("read ast for %s: %w", tc.Name, err)
			}
			b.AST = ast
		}
		if tc.CanonicalOut != "" {
			out, err := readTextFile(tc.CanonicalOut)
			if err != nil {
				return nil, fmt.Errorf("read canonical for %s: %w", tc.Name, err)
			}
			b.Canonical = out
		}
		bundles = append(bundles, b)
	}

	sort.Slice(bundles, func(i, j int) bool {
		return bundles[i].CaseName < bundles[j].CaseName
	})
	return bundles, nil
}

func readTextFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return normalizeNewlines(string(raw)), nil
}

// WriteBundles writes one oracle JSON file per case in outDir.
func WriteBundles(outDir string, bundles []OracleBundle) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", outDir, err)
	}
	for _, b := range bundles {
		fileName := sanitizeCaseName(b.CaseName) + ".json"
		path := filepath.Join(outDir, fileName)
		if err := SaveJSON(path, b); err != nil {
			return err
		}
	}
	return nil
}

func sanitizeCaseName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return name
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}
