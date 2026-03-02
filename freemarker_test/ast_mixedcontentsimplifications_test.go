package freemarker_test

import "testing"

func TestASTMixedContentSimplifications(t *testing.T) {
	runASTCaseParity(t, "ast-mixedcontentsimplifications")
}
