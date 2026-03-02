package corpus

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Category is a coarse grouping used for migration and test planning.
type Category string

const (
	CategoryAST       Category = "ast"
	CategoryCanonical Category = "canonical"
	CategoryEncoding  Category = "encoding"
	CategoryUnknown   Category = "unknown"
)

// Case describes a discovered core test asset set keyed by base test name.
type Case struct {
	Name         string
	Category     Category
	FTLPath      string
	ASTPath      string
	CanonicalOut string
}

// Corpus is the discovered set of FreeMarker core resource cases.
type Corpus struct {
	Root  string
	Cases []Case
}

// Discover scans the given root directory and builds a corpus view.
func Discover(root string) (Corpus, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return Corpus{}, fmt.Errorf("read corpus root %q: %w", root, err)
	}

	byName := map[string]*Case{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		baseName, fileKind, ok := parseFileName(fileName)
		if !ok {
			continue
		}

		tc := byName[baseName]
		if tc == nil {
			tc = &Case{
				Name:     baseName,
				Category: detectCategory(baseName),
			}
			byName[baseName] = tc
		}

		fullPath := filepath.Join(root, fileName)
		switch fileKind {
		case "ftl":
			tc.FTLPath = fullPath
		case "ast":
			tc.ASTPath = fullPath
		case "out":
			tc.CanonicalOut = fullPath
		}
	}

	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)

	cases := make([]Case, 0, len(names))
	for _, name := range names {
		cases = append(cases, *byName[name])
	}

	return Corpus{
		Root:  root,
		Cases: cases,
	}, nil
}

func parseFileName(fileName string) (baseName string, fileKind string, ok bool) {
	switch {
	case strings.HasSuffix(fileName, ".ftl.out"):
		return strings.TrimSuffix(fileName, ".ftl.out"), "out", true
	case strings.HasSuffix(fileName, ".ftl"):
		return strings.TrimSuffix(fileName, ".ftl"), "ftl", true
	case strings.HasSuffix(fileName, ".ast"):
		return strings.TrimSuffix(fileName, ".ast"), "ast", true
	default:
		return "", "", false
	}
}

func detectCategory(baseName string) Category {
	switch {
	case strings.HasPrefix(baseName, "ast-"):
		return CategoryAST
	case strings.HasPrefix(baseName, "cano-"):
		return CategoryCanonical
	case strings.HasPrefix(baseName, "encodingOverride-"):
		return CategoryEncoding
	default:
		return CategoryUnknown
	}
}

// Validate checks required file pairings for known case categories.
func (c Corpus) Validate() error {
	var errs []error
	for _, tc := range c.Cases {
		switch tc.Category {
		case CategoryAST:
			if tc.FTLPath == "" {
				errs = append(errs, fmt.Errorf("%s: missing .ftl", tc.Name))
			}
			if tc.ASTPath == "" {
				errs = append(errs, fmt.Errorf("%s: missing .ast", tc.Name))
			}
		case CategoryCanonical:
			if tc.FTLPath == "" {
				errs = append(errs, fmt.Errorf("%s: missing .ftl", tc.Name))
			}
			if tc.CanonicalOut == "" {
				errs = append(errs, fmt.Errorf("%s: missing .ftl.out", tc.Name))
			}
		case CategoryEncoding:
			if tc.FTLPath == "" {
				errs = append(errs, fmt.Errorf("%s: missing .ftl", tc.Name))
			}
		}
	}
	return errors.Join(errs...)
}

// FindCoreRootFromWD locates test/resources/freemarker/core by walking parent directories.
func FindCoreRootFromWD() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	return FindCoreRoot(wd)
}

// FindCoreRoot locates test/resources/freemarker/core by walking parent directories from start.
func FindCoreRoot(start string) (string, error) {
	cur := start
	for {
		candidate := filepath.Join(cur, "test", "resources", "freemarker", "core")
		if st, err := os.Stat(candidate); err == nil && st.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return "", fmt.Errorf("unable to locate test/resources/freemarker/core from %q", start)
}

// ByName returns the case for a name, and false if not present.
func (c Corpus) ByName(name string) (Case, bool) {
	for _, tc := range c.Cases {
		if tc.Name == name {
			return tc, true
		}
	}
	return Case{}, false
}
