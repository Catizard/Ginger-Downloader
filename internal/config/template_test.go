package config_test

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/Catizard/Ginger-Downloader/internal/config"
)

func TestCreateTemplateConfigFile(t *testing.T) {
	fp, cleanup := moveConfigFileToTempDir(t)
	defer cleanup()
	if err := config.CreateTemplateConfig(config.ENGLISH_LOCALE, false); err != nil {
		t.Fatalf("failed to create the config at the very beginning: %s", err)
	}

	if err := config.CreateTemplateConfig(config.ENGLISH_LOCALE, false); err == nil {
		t.Fatal("should return error when config file already existed")
	} else if !errors.Is(err, os.ErrExist) {
		t.Fatalf("expecting os.ErrExist, got: %s", err)
	}

	if err := config.CreateTemplateConfig(config.ENGLISH_LOCALE, true); err != nil {
		t.Fatalf("failed the override the config: %s", err)
	}

	f, err := os.Open(fp)
	if err != nil {
		t.Fatalf("open file: %s", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read file: %s", err)
	}
	log.Printf("Created template config:\n %s", string(b))

	if _, err := config.ReadConfig(); err != nil {
		t.Errorf("failed to read the empty config: %s", err)
	}
}

func moveConfigFileToTempDir(t *testing.T) (string, func()) {
	tempDir := t.TempDir()
	fp := filepath.Join(tempDir, "config.jsonc")
	config.MoveConfigFilePath(fp)
	return fp, func() {
		os.Remove(fp)
	}
}
