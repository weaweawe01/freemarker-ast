package freemarker_test

import "testing"

func TestASTNestedIgnoredChildren(t *testing.T) {
	runASTCaseParity(t, "ast-nestedignoredchildren")
}
