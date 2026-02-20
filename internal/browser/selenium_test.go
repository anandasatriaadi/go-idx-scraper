package browser

import (
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("getFreePort failed: %v", err)
	}
	if port <= 0 {
		t.Errorf("expected positive port, got %d", port)
	}
}

func TestFindChromeDriver(t *testing.T) {
	path := findChromeDriver()
	if path == "" {
		t.Error("expected non-empty path for chromedriver")
	}
}
