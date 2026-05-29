package composition

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadComposition(t *testing.T) {
	dir := t.TempDir()
	raw := []byte(`apiVersion: sovereign.engine/v1alpha1
kind: Composition
metadata:
  name: database
spec:
  type: database
  parameters:
    region: us-east-1
    size: small
  labels:
    cleanup: "true"
`)
	if err := os.WriteFile(filepath.Join(dir, "database.yaml"), raw, 0o600); err != nil {
		t.Fatalf("write composition: %v", err)
	}
	c, err := Load(dir, "database")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if c.Metadata.Name != "database" {
		t.Fatalf("name = %q, want database", c.Metadata.Name)
	}
	if c.Spec.Parameters["region"] != "us-east-1" {
		t.Fatalf("region = %q, want us-east-1", c.Spec.Parameters["region"])
	}
}
