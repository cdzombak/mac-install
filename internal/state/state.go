package state

import (
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	stateDir string
}

func NewStore() (*Store, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	stateDir := filepath.Join(homeDir, ".config", "dotfiles", "software")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, err
	}

	return &Store{stateDir: stateDir}, nil
}

func (s *Store) IsExcluded(softwareName string) bool {
	flagFile := filepath.Join(s.stateDir, "no-"+normalizeFilename(softwareName))
	_, err := os.Stat(flagFile)
	return err == nil
}

func (s *Store) SetExcluded(softwareName string) error {
	flagFile := filepath.Join(s.stateDir, "no-"+normalizeFilename(softwareName))
	file, err := os.Create(flagFile)
	if err != nil {
		return err
	}
	return file.Close()
}

func normalizeFilename(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}