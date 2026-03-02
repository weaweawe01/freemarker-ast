package freemarker_test

import "testing"

func TestASTMultipleIgnoredChildren(t *testing.T) {
	runASTCaseParity(t, "ast-multipleignoredchildren")
}
