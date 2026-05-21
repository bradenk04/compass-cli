package main

import (
	"fmt"
	"os"

	"bradenkennedy.com/compass-cli/internal/config"
	"bradenkennedy.com/compass-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(tui.NewMainModel(config), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
