package storage

import (
	"encoding/json"
	"os"
	"sync"
)

// Storage manages the persistence of search history.
type Storage struct {
	filePath string
	history  []string
	mutex    sync.RWMutex
}

// NewStorage creates a new Storage instance.
func NewStorage(filePath string) *Storage {
	return &Storage{
		filePath: filePath,
		history:  make([]string, 0),
	}
}

// LoadHistory reads the search history from the JSON file.
func (s *Storage) LoadHistory() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with empty history
			s.history = make([]string, 0)
			return nil
		}
		return err // Other read error
	}

	if len(data) == 0 {
		// File is empty, start with empty history
		s.history = make([]string, 0)
		return nil
	}

	err = json.Unmarshal(data, &s.history)
	if err != nil {
		// If unmarshalling fails (e.g., corrupted file), start fresh
		s.history = make([]string, 0)
		// Optionally log the error
		// log.Printf("Warning: Could not unmarshal history file %s: %v. Starting fresh.", s.filePath, err)
		return nil // Treat as non-fatal for loading
	}

	return nil
}

// SaveHistory writes the current search history to the JSON file.
func (s *Storage) SaveHistory() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := json.MarshalIndent(s.history, "", "  ")
	if err != nil {
		return err
	}

	// Ensure the directory exists
	// dir := filepath.Dir(s.filePath)
	// if err := os.MkdirAll(dir, 0755); err != nil {
	// 	return err
	// }
	// Note: Directory creation is handled in main.go already

	return os.WriteFile(s.filePath, data, 0644)
}

// AddSearchTerm adds a new term to the history if it's not already present.
// It adds the new term to the beginning.
func (s *Storage) AddSearchTerm(term string) {
	if term == "" {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if term already exists and remove it
	newHistory := make([]string, 0, len(s.history)+1)
	newHistory = append(newHistory, term) // Add new term at the beginning

	for _, existingTerm := range s.history {
		if existingTerm != term {
			newHistory = append(newHistory, existingTerm)
		}
	}

	// Optional: Limit history size
	// const maxHistorySize = 100
	// if len(newHistory) > maxHistorySize {
	// 	newHistory = newHistory[:maxHistorySize]
	// }

	s.history = newHistory
}

// GetHistory returns a copy of the current search history.
func (s *Storage) GetHistory() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	// Return a copy to prevent external modification
	historyCopy := make([]string, len(s.history))
	copy(historyCopy, s.history)
	return historyCopy
}
