package freemarker_test

import "testing"

func TestASTStringLiteralInterpolation(t *testing.T) {
	runASTCaseParity(t, "ast-strlitinterpolation")
}
