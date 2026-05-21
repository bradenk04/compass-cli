package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const historyFileName = ".compass-cli-history.json"

func LoadHistory() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	historyPath := filepath.Join(home, historyFileName)
	bz, err := os.ReadFile(historyPath)
	if err != nil {
		return []string{}
	}

	var history []string
	err = json.Unmarshal(bz, &history)
	if err != nil {
		return []string{}
	}

	return history
}

func SaveHistory(history []string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	historyPath := filepath.Join(home, historyFileName)
	bz, err := json.Marshal(history)
	if err != nil {
		return
	}

	_ = os.WriteFile(historyPath, bz, 0644)
}

func AddToHistory(uri string) []string {
	if uri == "" {
		return LoadHistory()
	}

	history := LoadHistory()

	var newHistory []string
	newHistory = append(newHistory, uri)

	for _, item := range history {
		if item != uri {
			newHistory = append(newHistory, item)
		}
	}

	if len(newHistory) > 5 {
		newHistory = newHistory[:5]
	}

	SaveHistory(newHistory)
	return newHistory
}
