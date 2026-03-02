package compat

import "testing"

func TestPositionIsZero(t *testing.T) {
	var zero Position
	if !zero.IsZero() {
		t.Fatal("expected zero position to report IsZero=true")
	}

	p := Position{Line: 1, Column: 1}
	if p.IsZero() {
		t.Fatal("expected non-zero position to report IsZero=false")
	}
}
