package mobile

import (
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	port, err := Start("/tmp/mobile-ipfs-lite", false)
	if err != nil {
		t.Fatalf("failed to Start: %v", err)
	}
	if port < 1 {
		t.Fatal("invalid port")
	}
}

func TestStop(t *testing.T) {
	time.Sleep(10 * time.Second)
	err := Stop()
	if err != nil {
		t.Fatalf("failed to StopServer: %v", err)
	}
}

func TestStartAgain(t *testing.T) {
	port, err := Start("/tmp/mobile-ipfs-lite", false)
	if err != nil {
		t.Fatalf("failed to Start again: %v", err)
	}
	if port < 1 {
		t.Fatal("invalid port again")
	}
}

func TestStopAgain(t *testing.T) {
	time.Sleep(10 * time.Second)
	err := Stop()
	if err != nil {
		t.Fatalf("failed to StopServer again: %v", err)
	}
}
