package composition

import (
	"fmt"
	"time"

	"github.com/yourusername/the-engine/internal/domain"
	"github.com/yourusername/the-engine/internal/provider"
	"github.com/yourusername/the-engine/internal/store"
)

type Engine struct {
	compositionDir string
	providers      *provider.Registry
	store          *store.Store
}

type DeployOptions struct {
	Composition string
	Provider    string
	Name        string
	Region      string
	Size        string
	Labels      map[string]string
}

func NewEngine(compositionDir string, registry *provider.Registry, store *store.Store) *Engine {
	return &Engine{compositionDir: compositionDir, providers: registry, store: store}
}

func (e *Engine) Deploy(opts DeployOptions) (domain.Resource, error) {
	if opts.Composition == "" {
		return domain.Resource{}, fmt.Errorf("composition is required")
	}
	if opts.Provider == "" {
		return domain.Resource{}, fmt.Errorf("provider is required")
	}
	if opts.Name == "" {
		opts.Name = opts.Composition + "-" + time.Now().UTC().Format("20060102150405")
	}
	c, err := Load(e.compositionDir, opts.Composition)
	if err != nil {
		return domain.Resource{}, err
	}
	p, err := e.providers.Get(opts.Provider)
	if err != nil {
		return domain.Resource{}, err
	}
	region := firstNonEmpty(opts.Region, c.Spec.Parameters["region"])
	size := firstNonEmpty(opts.Size, c.Spec.Parameters["size"])
	labels := map[string]string{}
	for k, v := range c.Spec.Labels {
		labels[k] = v
	}
	for k, v := range opts.Labels {
		labels[k] = v
	}
	now := time.Now().UTC()
	resource := domain.Resource{
		Name:           opts.Name,
		Type:           c.Spec.Type,
		Provider:       opts.Provider,
		Region:         region,
		Size:           size,
		Composition:    c.Metadata.Name,
		Labels:         labels,
		Status:         domain.StatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
		LastActivityAt: now,
	}
	resource, err = p.Deploy(resource)
	if err != nil {
		resource.Status = domain.StatusFailed
		resource.StatusMessage = err.Error()
		_ = e.store.SaveResource(resource)
		return domain.Resource{}, err
	}
	if err := e.store.SaveResource(resource); err != nil {
		return domain.Resource{}, err
	}
	return resource, nil
}

func (e *Engine) Destroy(id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	resource, err := e.store.GetResource(id)
	if err != nil {
		return err
	}
	p, err := e.providers.Get(resource.Provider)
	if err != nil {
		return err
	}
	resource.Status = domain.StatusDestroying
	resource.UpdatedAt = time.Now().UTC()
	if err := e.store.SaveResource(resource); err != nil {
		return err
	}
	if err := p.Destroy(id); err != nil {
		resource.Status = domain.StatusFailed
		resource.StatusMessage = err.Error()
		resource.UpdatedAt = time.Now().UTC()
		_ = e.store.SaveResource(resource)
		return err
	}
	return e.store.DeleteResource(id)
}

func (e *Engine) List(filter map[string]string) ([]domain.Resource, error) {
	return e.store.ListResources(filter)
}

func (e *Engine) Status(id string) (domain.Resource, error) {
	if id == "" {
		return domain.Resource{}, fmt.Errorf("id is required")
	}
	return e.store.GetResource(id)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
