package app

import (
	"fmt"

	"github.com/yourusername/the-engine/internal/composition"
	"github.com/yourusername/the-engine/internal/config"
	"github.com/yourusername/the-engine/internal/encryption"
	"github.com/yourusername/the-engine/internal/provider"
	"github.com/yourusername/the-engine/internal/secrets"
	"github.com/yourusername/the-engine/internal/store"
)

type App struct {
	Config      *config.Config
	Encryption *encryption.Service
	Secrets     *secrets.EnvStore
	SecretValues map[string]string
	Store       *store.Store
	Engine      *composition.Engine
}

func New(compositionDir string) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	enc, err := encryption.New(cfg.MasterKey)
	if err != nil {
		return nil, err
	}
	secretStore := secrets.NewEnvStore(cfg.EncryptedEnvPath, enc)
	if err := secretStore.Ensure(); err != nil {
		return nil, err
	}
	secretValues, err := secretStore.Load()
	if err != nil {
		return nil, err
	}
	db, err := store.Open(cfg.DBPath)
	if err != nil {
		return nil, err
	}
	registry := provider.DefaultRegistry()
	engine := composition.NewEngine(compositionDir, registry, db)
	return &App{
		Config:      cfg,
		Encryption: enc,
		Secrets:     secretStore,
		SecretValues: secretValues,
		Store:       db,
		Engine:      engine,
	}, nil
}

func (a *App) Close() error {
	if a == nil || a.Store == nil {
		return nil
	}
	if err := a.Store.Close(); err != nil {
		return fmt.Errorf("close app store: %w", err)
	}
	return nil
}
