package config

import (
	"fmt"
	"testing"
)

func TestConfigLoad(t *testing.T) {
	cfg, err := MustLoad()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Config is nil")
	}
	fmt.Printf("Loaded config: %+v\n", cfg)
}
