package freemarker_test

import "testing"

func TestAST1(t *testing.T) {
	runASTCaseParity(t, "ast-1")
}

func TestRCE(t *testing.T) {
	runASTCaseParity(t, "rce")

}

func TestRCE2(t *testing.T) {
	runASTCaseParity(t, "rce2")

}
func TestRCE3(t *testing.T) {
	runASTCaseParity(t, "rce3")

}
func TestRCE4(t *testing.T) {
	runASTCaseParity(t, "rce4")

}
func TestRCE5(t *testing.T) {
	runASTCaseParity(t, "rce5")

}
func TestRCE6(t *testing.T) {
	runASTCaseParity(t, "rce6")

}
func TestRCE7(t *testing.T) {
	runASTCaseParity(t, "rce7")

}
