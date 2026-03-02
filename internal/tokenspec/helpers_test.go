package tokenspec

import (
	"fmt"
	"os"
	"path/filepath"
)

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
	return "", fmt.Errorf("FTL.jj not found from cwd")
}
