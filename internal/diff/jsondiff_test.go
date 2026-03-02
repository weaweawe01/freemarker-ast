package diff

import "testing"

func TestJSONLikeEqual(t *testing.T) {
	oracle := map[string]any{
		"case_name": "ast-1",
		"ast":       "IfBlock",
		"tokens": []any{
			map[string]any{"kind": 1, "image": "<#if"},
		},
	}
	actual := map[string]any{
		"case_name": "ast-1",
		"ast":       "IfBlock",
		"tokens": []any{
			map[string]any{"kind": 1, "image": "<#if"},
		},
	}

	diffs := JSONLike(oracle, actual)
	if len(diffs) != 0 {
		t.Fatalf("expected no diffs, got %+v", diffs)
	}
}

func TestJSONLikeDetectsDiffs(t *testing.T) {
	oracle := map[string]any{
		"case_name": "ast-1",
		"ast":       "IfBlock",
		"tokens": []any{
			map[string]any{"kind": 1, "image": "<#if"},
		},
	}
	actual := map[string]any{
		"case_name": "ast-1",
		"ast":       "IfBlockV2",
		"tokens": []any{
			map[string]any{"kind": 1, "image": "<#if"},
		},
	}

	diffs := JSONLike(oracle, actual)
	if len(diffs) == 0 {
		t.Fatal("expected diffs, got none")
	}
	if diffs[0].Path != "$.ast" {
		t.Fatalf("expected first diff path $.ast, got %s", diffs[0].Path)
	}
}
