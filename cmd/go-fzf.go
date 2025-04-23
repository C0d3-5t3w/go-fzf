package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/C0d3-5t3w/go-fzf/int/config"
	"github.com/C0d3-5t3w/go-fzf/int/gui"
	"github.com/C0d3-5t3w/go-fzf/int/storage"
)

func main() {

	os.Setenv("FYNE_THEME", "dark")

	configDir := "pkg/config"
	configPath := filepath.Join(configDir, "config.yaml")
	_ = os.MkdirAll(configDir, 0755)

	storageDir := "pkg/storage"
	storagePath := filepath.Join(storageDir, "storage.json")
	_ = os.MkdirAll(storageDir, 0755)

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	store := storage.NewStorage(storagePath)

	ui := gui.NewGUI(cfg, store)
	ui.Run()
}
