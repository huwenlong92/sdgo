package updater

import "testing"

func TestNormalizeVersionDefaultsToLatest(t *testing.T) {
	if got := NormalizeVersion(""); got != "latest" {
		t.Fatalf("unexpected version: %s", got)
	}
}

func TestNormalizeVersionTrimsInput(t *testing.T) {
	if got := NormalizeVersion(" v0.2.0 "); got != "v0.2.0" {
		t.Fatalf("unexpected version: %s", got)
	}
}
