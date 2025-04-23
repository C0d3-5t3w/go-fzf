package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/C0d3-5t3w/go-fzf/int/config"
	"github.com/C0d3-5t3w/go-fzf/int/gui"
	"github.com/C0d3-5t3w/go-fzf/int/storage" // Import storage
)

func main() {
	// --- Configuration Path ---
	configDir := "pkg/config" // Adjust if needed
	configPath := filepath.Join(configDir, "config.yaml")
	_ = os.MkdirAll(configDir, 0755) // Ensure config dir exists

	// --- Storage Path ---
	storageDir := "pkg/storage" // Define storage directory
	storagePath := filepath.Join(storageDir, "storage.json")
	_ = os.MkdirAll(storageDir, 0755) // Ensure storage dir exists

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create storage manager
	store := storage.NewStorage(storagePath)

	// Create and run the GUI, passing config and storage
	ui := gui.NewGUI(cfg, store)
	ui.Run()
}
