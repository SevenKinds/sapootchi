//go:build !js

package game

import (
	"os"
	"path/filepath"
)

// Desktop/mobile-native storage: a JSON file in the user config dir.

func savefilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "sapootchi_save.json"
	}
	return filepath.Join(dir, "sapootchi", "save.json")
}

func readSave() ([]byte, error) {
	return os.ReadFile(savefilePath())
}

func writeSave(data []byte) error {
	path := savefilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
