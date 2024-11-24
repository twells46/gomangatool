package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/frontend"
)

func main() {
	f, _ := tea.LogToFile("debug.log", "debug")
	defer f.Close()

	p := tea.NewProgram(frontend.InitModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalln(err)
	}
}
