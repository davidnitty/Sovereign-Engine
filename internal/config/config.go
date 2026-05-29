package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/the-engine/internal/encryption"
)

const (
	EnvMasterKey = "ENGINE_MASTER_KEY"
	EnvDataDir   = "ENGINE_DATA_DIR"
)

type Config struct {
	DataDir          string
	DBPath           string
	EncryptedEnvPath string
	MasterKey        string
	Port             string
}

func Load() (*Config, error) {
	dataDir, err := dataDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	masterKey, err := masterKey(dataDir)
	if err != nil {
		return nil, err
	}
	port := os.Getenv("ENGINE_PORT")
	if port == "" {
		port = "8080"
	}
	return &Config{
		DataDir:          dataDir,
		DBPath:           filepath.Join(dataDir, "engine.db"),
		EncryptedEnvPath: filepath.Join(dataDir, ".env.enc"),
		MasterKey:        masterKey,
		Port:             port,
	}, nil
}

func dataDir() (string, error) {
	if dir := os.Getenv(EnvDataDir); dir != "" {
		return dir, nil
	}
	userConfig, err := os.UserConfigDir()
	if err == nil && userConfig != "" {
		return filepath.Join(userConfig, "sovereign-engine"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve data directory: %w", err)
	}
	return filepath.Join(home, ".sovereign-engine"), nil
}

func masterKey(dataDir string) (string, error) {
	if key := os.Getenv(EnvMasterKey); key != "" {
		return key, nil
	}
	path := filepath.Join(dataDir, "master.key")
	raw, err := os.ReadFile(path)
	if err == nil {
		if len(raw) == 0 {
			return "", errors.New("stored master key is empty")
		}
		return string(raw), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read master key: %w", err)
	}
	key, err := encryption.GenerateMasterKey()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(key), 0o600); err != nil {
		return "", fmt.Errorf("write generated master key: %w", err)
	}
	return key, nil
}
