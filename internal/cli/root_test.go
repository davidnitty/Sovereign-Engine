package cli

import "testing"

func TestParseLabels(t *testing.T) {
	labels, err := parseLabels([]string{"team=platform", "env=dev"})
	if err != nil {
		t.Fatalf("parseLabels() error = %v", err)
	}
	if labels["team"] != "platform" || labels["env"] != "dev" {
		t.Fatalf("labels = %#v", labels)
	}
}

func TestParseLabelsRejectsInvalidInput(t *testing.T) {
	if _, err := parseLabels([]string{"missing-separator"}); err == nil {
		t.Fatal("parseLabels() error = nil, want error")
	}
}
