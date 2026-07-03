//go:build all || unit

package memberentitlementmanagement_test

import (
	"os"
	"testing"
)

func TestGapMatrixExists(t *testing.T) {
	_, err := os.Stat("../../../../docs/memberentitlementmanagement-gap-matrix.md")
	if os.IsNotExist(err) {
		t.Fatal("docs/memberentitlementmanagement-gap-matrix.md does not exist — run WI-1 to create it")
	}
	if err != nil {
		t.Fatalf("unexpected error checking gap matrix: %v", err)
	}
}
