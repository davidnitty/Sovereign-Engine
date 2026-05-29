package provider

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/yourusername/the-engine/internal/domain"
)

type LocalProvider struct {
	Name string
}

func (p LocalProvider) Deploy(resource domain.Resource) (domain.Resource, error) {
	if resource.Name == "" {
		return domain.Resource{}, fmt.Errorf("%s deploy: resource name is required", p.Name)
	}
	now := time.Now().UTC()
	if resource.ID == "" {
		id, err := resourceID(p.Name)
		if err != nil {
			return domain.Resource{}, err
		}
		resource.ID = id
	}
	resource.Provider = p.Name
	resource.Status = domain.StatusRunning
	resource.StatusMessage = "created in local control plane; provider SDK apply is not enabled"
	if resource.CreatedAt.IsZero() {
		resource.CreatedAt = now
	}
	resource.UpdatedAt = now
	resource.LastActivityAt = now
	return resource, nil
}

func (p LocalProvider) Destroy(id string) error {
	if id == "" {
		return fmt.Errorf("%s destroy: id is required", p.Name)
	}
	return nil
}

func (p LocalProvider) List(filter map[string]string) ([]domain.Resource, error) {
	return []domain.Resource{}, nil
}

func (p LocalProvider) Status(id string) (domain.ResourceStatus, error) {
	if id == "" {
		return "", fmt.Errorf("%s status: id is required", p.Name)
	}
	return domain.StatusRunning, nil
}

func DefaultRegistry() *Registry {
	return NewRegistry(map[string]Provider{
		"aws":     LocalProvider{Name: "aws"},
		"azure":   LocalProvider{Name: "azure"},
		"gcp":     LocalProvider{Name: "gcp"},
		"do":      LocalProvider{Name: "do"},
		"hetzner": LocalProvider{Name: "hetzner"},
		"ovh":     LocalProvider{Name: "ovh"},
	})
}

func resourceID(provider string) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate resource id: %w", err)
	}
	return fmt.Sprintf("%s-%s", provider, hex.EncodeToString(buf)), nil
}
