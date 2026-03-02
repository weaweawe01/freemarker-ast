package freemarker_test

import "testing"

func TestASTLocations(t *testing.T) {
	runASTCaseParity(t, "ast-locations")
}
