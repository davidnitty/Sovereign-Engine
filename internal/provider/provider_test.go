package provider

import "testing"

func TestRegistryGetNormalizesName(t *testing.T) {
	registry := DefaultRegistry()
	if _, err := registry.Get("AWS"); err != nil {
		t.Fatalf("Get(AWS) error = %v", err)
	}
	if _, err := registry.Get("missing"); err == nil {
		t.Fatal("Get(missing) error = nil, want error")
	}
}
