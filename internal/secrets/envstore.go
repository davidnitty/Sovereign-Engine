package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/yourusername/the-engine/internal/encryption"
)

type EnvStore struct {
	path string
	enc  *encryption.Service
}

func NewEnvStore(path string, enc *encryption.Service) *EnvStore {
	return &EnvStore{path: path, enc: enc}
}

func (s *EnvStore) Ensure() error {
	if _, err := os.Stat(s.path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat encrypted env file: %w", err)
	}
	return s.Save(map[string]string{})
}

func (s *EnvStore) Load() (map[string]string, error) {
	if err := s.Ensure(); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("read encrypted env file: %w", err)
	}
	if len(raw) == 0 {
		return map[string]string{}, nil
	}
	plaintext, err := s.enc.Decrypt(string(raw))
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	if err := json.Unmarshal([]byte(plaintext), &values); err != nil {
		return nil, fmt.Errorf("decode env payload: %w", err)
	}
	return values, nil
}

func (s *EnvStore) Save(values map[string]string) error {
	raw, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("encode env payload: %w", err)
	}
	ciphertext, err := s.enc.Encrypt(string(raw))
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.path, []byte(ciphertext), 0o600); err != nil {
		return fmt.Errorf("write encrypted env file: %w", err)
	}
	return nil
}
