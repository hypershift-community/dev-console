package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"hypershift-dev-console/pkg/tui"
)

func main() {
	p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
