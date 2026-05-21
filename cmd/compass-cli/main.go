package main

import (
	"fmt"
	"os"

	"bradenkennedy.com/compass-cli/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(tui.NewMainModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
