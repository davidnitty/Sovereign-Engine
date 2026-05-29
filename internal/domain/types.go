package domain

import "time"

type ResourceStatus string

const (
	StatusPending    ResourceStatus = "pending"
	StatusRunning    ResourceStatus = "running"
	StatusStopped    ResourceStatus = "stopped"
	StatusDestroying ResourceStatus = "destroying"
	StatusDestroyed  ResourceStatus = "destroyed"
	StatusFailed     ResourceStatus = "failed"
)

type Resource struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Provider       string            `json:"provider"`
	Region         string            `json:"region"`
	Size           string            `json:"size"`
	Composition    string            `json:"composition"`
	Labels         map[string]string `json:"labels"`
	Status         ResourceStatus    `json:"status"`
	StatusMessage  string            `json:"status_message,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	LastActivityAt time.Time         `json:"last_activity_at"`
}

type CostRecord struct {
	ID         int64     `json:"id"`
	ResourceID string    `json:"resource_id"`
	Provider   string    `json:"provider"`
	AmountUSD  float64   `json:"amount_usd"`
	RecordedAt time.Time `json:"recorded_at"`
	Tags       string    `json:"tags"`
}

type CleanupMode string

const (
	CleanupDisabled CleanupMode = "disabled"
	CleanupEnabled  CleanupMode = "enabled"
)
