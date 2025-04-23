package storage

import (
	"encoding/json"
	"os"
	"sync"
)

type Storage struct {
	filePath string
	history  []string
	mutex    sync.RWMutex
}

func NewStorage(filePath string) *Storage {
	return &Storage{
		filePath: filePath,
		history:  make([]string, 0),
	}
}

func (s *Storage) LoadHistory() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			s.history = make([]string, 0)
			return nil
		}
		return err 
	}

	if len(data) == 0 {
		s.history = make([]string, 0)
		return nil
	}

	err = json.Unmarshal(data, &s.history)
	if err != nil {
		s.history = make([]string, 0)
		return nil 	
	}

	return nil
}

func (s *Storage) SaveHistory() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := json.MarshalIndent(s.history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Storage) AddSearchTerm(term string) {
	if term == "" {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	newHistory := make([]string, 0, len(s.history)+1)
	newHistory = append(newHistory, term) 

	for _, existingTerm := range s.history {
		if existingTerm != term {
			newHistory = append(newHistory, existingTerm)
		}
	}
	s.history = newHistory
}

func (s *Storage) GetHistory() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	historyCopy := make([]string, len(s.history))
	copy(historyCopy, s.history)
	return historyCopy
}
