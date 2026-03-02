package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/weaweawe01/freemarker-ast/internal/diff"
)

func main() {
	var oraclePath string
	var actualPath string

	flag.StringVar(&oraclePath, "oracle", "", "path to oracle JSON file")
	flag.StringVar(&actualPath, "actual", "", "path to actual JSON file")
	flag.Parse()

	if oraclePath == "" || actualPath == "" {
		fmt.Fprintln(os.Stderr, "usage: fm-oracle-diff --oracle <path> --actual <path>")
		os.Exit(2)
	}

	oraclePayload, err := loadJSONAny(oraclePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load oracle: %v\n", err)
		os.Exit(2)
	}
	actualPayload, err := loadJSONAny(actualPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load actual: %v\n", err)
		os.Exit(2)
	}

	diffs := diff.JSONLike(oraclePayload, actualPayload)
	if len(diffs) == 0 {
		fmt.Println("diff result: equal")
		return
	}

	fmt.Println("diff result: different")
	for i, d := range diffs {
		fmt.Printf("[%d] path=%s oracle=%v actual=%v\n", i+1, d.Path, d.Oracle, d.Actual)
	}
	os.Exit(1)
}

func loadJSONAny(path string) (any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}

	var out any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("decode %q: %w", path, err)
	}
	return out, nil
}
