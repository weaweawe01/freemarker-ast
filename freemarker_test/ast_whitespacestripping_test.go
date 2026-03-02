package freemarker_test

import "testing"

func TestASTWhitespaceStripping(t *testing.T) {
	runASTCaseParity(t, "ast-whitespacestripping")
}
