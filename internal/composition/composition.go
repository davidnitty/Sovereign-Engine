package composition

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Composition struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
	Spec Spec `yaml:"spec"`
}

type Spec struct {
	Type        string            `yaml:"type"`
	Description string            `yaml:"description"`
	Parameters  map[string]string `yaml:"parameters"`
	Labels      map[string]string `yaml:"labels"`
}

func Load(dir, name string) (Composition, error) {
	path := filepath.Join(dir, name+".yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Composition{}, fmt.Errorf("read composition %q: %w", name, err)
	}
	var c Composition
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return Composition{}, fmt.Errorf("parse composition %q: %w", name, err)
	}
	if c.Metadata.Name == "" {
		return Composition{}, fmt.Errorf("composition %q is missing metadata.name", name)
	}
	if c.Spec.Type == "" {
		return Composition{}, fmt.Errorf("composition %q is missing spec.type", name)
	}
	if c.Spec.Parameters == nil {
		c.Spec.Parameters = map[string]string{}
	}
	if c.Spec.Labels == nil {
		c.Spec.Labels = map[string]string{}
	}
	return c, nil
}
