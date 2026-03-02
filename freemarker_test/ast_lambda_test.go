package freemarker_test

import "testing"

func TestASTLambda(t *testing.T) {
	runASTCaseParity(t, "ast-lambda")
}
