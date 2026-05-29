//go:build !windows

package runner

import (
	"os"
	"syscall"
	"testing"
)

func TestShutdownSignalsIncludeHangup(t *testing.T) {
	signals := shutdownSignals()

	for _, want := range []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGHUP} {
		found := false
		for _, got := range signals {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected shutdown signals to include %v, got %v", want, signals)
		}
	}
}
