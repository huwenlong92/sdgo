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

func TestModulePathForTargetDefaultsToModule(t *testing.T) {
	if got := modulePathForTarget(DefaultInstallTarget); got != defaultModulePath {
		t.Fatalf("unexpected module path: %s", got)
	}
}

func TestModulePathForTargetTrimsCmdSdgo(t *testing.T) {
	if got := modulePathForTarget("example.com/acme/sdgo/cmd/sdgo"); got != "example.com/acme/sdgo" {
		t.Fatalf("unexpected module path: %s", got)
	}
}

func TestVersionsEqualIgnoresLeadingV(t *testing.T) {
	if !versionsEqual("v0.1.4", "0.1.4") {
		t.Fatalf("versions should match")
	}
}

func TestVersionsEqualRejectsEmpty(t *testing.T) {
	if versionsEqual("", "v0.1.4") {
		t.Fatalf("empty version should not match")
	}
}
