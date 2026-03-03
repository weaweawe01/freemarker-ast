package astcmp

import "testing"

func TestNormalize(t *testing.T) {
	raw := "\uFEFF/*\n * header\n */\r\n#line1\r\n#line2\r\n\r\n"
	got := Normalize(raw)
	want := "#line1\n#line2"
	if got != want {
		t.Fatalf("normalize mismatch:\nwant=%q\ngot =%q", want, got)
	}
}

func TestNormalizeTrimEachLineSpace(t *testing.T) {
	raw := "  #if  // f.c.ConditionalBlock  \n\t#if  // f.c.ConditionalBlock\t\n"
	got := Normalize(raw)
	want := "#if  // f.c.ConditionalBlock\n#if  // f.c.ConditionalBlock"
	if got != want {
		t.Fatalf("normalize line trim mismatch:\nwant=%q\ngot =%q", want, got)
	}
}

func TestNormalizeStripTrailingNewlineOnlyTextNode(t *testing.T) {
	raw := "#mixed_content  // f.c.MixedContent\n#assign  // f.c.Assignment\n#text  // f.c.TextBlock\n- content: \"\\n\"  // String\n"
	got := Normalize(raw)
	want := "#mixed_content  // f.c.MixedContent\n#assign  // f.c.Assignment"
	if got != want {
		t.Fatalf("normalize trailing text mismatch:\nwant=%q\ngot =%q", want, got)
	}
}

func TestCompareNormalizedEqual(t *testing.T) {
	res := CompareNormalized("a\nb", "a\nb")
	if !res.Equal {
		t.Fatalf("expected equal result, got %#v", res)
	}
}

func TestCompareNormalizedDifferent(t *testing.T) {
	res := CompareNormalized("a\nb\nc", "a\nx\nc")
	if res.Equal {
		t.Fatal("expected not equal result")
	}
	if res.Line != 2 {
		t.Fatalf("expected diff line 2, got %d", res.Line)
	}
	if res.Oracle != "b" || res.Actual != "x" {
		t.Fatalf("unexpected diff payload: oracle=%q actual=%q", res.Oracle, res.Actual)
	}
	if res.DiffText == "" {
		t.Fatal("expected diff text")
	}
}
