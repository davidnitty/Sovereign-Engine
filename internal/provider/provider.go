package provider

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yourusername/the-engine/internal/domain"
)

type Provider interface {
	Deploy(resource domain.Resource) (domain.Resource, error)
	Destroy(id string) error
	List(filter map[string]string) ([]domain.Resource, error)
	Status(id string) (domain.ResourceStatus, error)
}

type Registry struct {
	providers map[string]Provider
}

func NewRegistry(providers map[string]Provider) *Registry {
	normalized := make(map[string]Provider, len(providers))
	for name, p := range providers {
		normalized[strings.ToLower(name)] = p
	}
	return &Registry{providers: normalized}
}

func (r *Registry) Get(name string) (Provider, error) {
	p, ok := r.providers[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q; supported providers: %s", name, strings.Join(r.Names(), ", "))
	}
	return p, nil
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
