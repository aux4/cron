package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type CronEntry struct {
	Name  string `json:"name"`
	Every string `json:"every"`
	At    string `json:"at,omitempty"`
	Run   string `json:"run"`
	State string `json:"state"`
}

type HistoryEntry struct {
	Name      string `json:"name"`
	JobID     string `json:"jobId"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
}

type CronStore struct {
	mu      sync.RWMutex
	dir     string
	entries []CronEntry
	history []HistoryEntry
}

func NewCronStore(dir string) *CronStore {
	return &CronStore{
		dir:     dir,
		entries: []CronEntry{},
		history: []HistoryEntry{},
	}
}

func (s *CronStore) cronFilePath() string {
	return filepath.Join(s.dir, ".cron.json")
}

func (s *CronStore) historyFilePath() string {
	return filepath.Join(s.dir, ".cron-history.json")
}

func (s *CronStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.cronFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			s.entries = []CronEntry{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.entries)
}

func (s *CronStore) LoadHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.historyFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			s.history = []HistoryEntry{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.history)
}

func (s *CronStore) save() error {
	data, err := json.Marshal(s.entries)
	if err != nil {
		return err
	}
	return os.WriteFile(s.cronFilePath(), data, 0644)
}

func (s *CronStore) saveHistory() error {
	data, err := json.Marshal(s.history)
	if err != nil {
		return err
	}
	return os.WriteFile(s.historyFilePath(), data, 0644)
}

func (s *CronStore) Add(entry CronEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.entries {
		if e.Name == entry.Name {
			return errEntryExists(entry.Name)
		}
	}
	s.entries = append(s.entries, entry)
	return s.save()
}

func (s *CronStore) Remove(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, e := range s.entries {
		if e.Name == name {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errEntryNotFound(name)
	}
	s.entries = append(s.entries[:idx], s.entries[idx+1:]...)
	return s.save()
}

func (s *CronStore) SetState(name, state string) (*CronEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, e := range s.entries {
		if e.Name == name {
			s.entries[i].State = state
			if err := s.save(); err != nil {
				return nil, err
			}
			return &s.entries[i], nil
		}
	}
	return nil, errEntryNotFound(name)
}

func (s *CronStore) Get(name string) (*CronEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, e := range s.entries {
		if e.Name == name {
			return &e, nil
		}
	}
	return nil, errEntryNotFound(name)
}

func (s *CronStore) List() []CronEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]CronEntry, len(s.entries))
	copy(result, s.entries)
	return result
}

func (s *CronStore) AddHistory(entry HistoryEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = append(s.history, entry)

	// Keep last 1000 entries
	if len(s.history) > 1000 {
		s.history = s.history[len(s.history)-1000:]
	}

	return s.saveHistory()
}

func (s *CronStore) GetHistory(name string, limit int) []HistoryEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []HistoryEntry
	for _, h := range s.history {
		if h.Name == name {
			result = append(result, h)
		}
	}

	if result == nil {
		result = []HistoryEntry{}
	}

	if limit > 0 && len(result) > limit {
		result = result[len(result)-limit:]
	}

	return result
}

type cronError struct {
	message string
}

func (e *cronError) Error() string {
	return e.message
}

func errEntryExists(name string) error {
	return &cronError{message: "entry " + name + " already exists"}
}

func errEntryNotFound(name string) error {
	return &cronError{message: "entry " + name + " not found"}
}
