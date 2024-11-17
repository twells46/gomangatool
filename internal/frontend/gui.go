package frontend

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/backend"
)

const (
	library int = iota
	series
	adder
	review
)

// The overall tea.Model, which contains the various sub-models
// for the different functionalities of the program.
type model struct {
	view     int
	adder    Adder
	library  Library
	series   Series
	err      error // NOTE: Currently unused
	store    *backend.SQLite
	quitting bool // NOTE: Currently unused
}

// Initialize a new model
func InitModel() model {
	store := backend.Opendb("manga.sqlite3")

	return model{
		view:    library,
		adder:   newAdder(),
		library: initLibrary(store),
		store:   store,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// Main update function, which handles universal quit keys
// then passes off to the appropriate sub-function
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	switch m.view {
	case library:
		return LibraryUpdate(msg, m)
	case adder:
		return AdderUpdate(msg, m)
	case series:
		return SeriesUpdate(msg, m)
	}

	return m, tea.Quit
}

// Main view function, which calls the correct sub-function
func (m model) View() string {
	switch m.view {
	case library:
		return LibraryView(m)
	case adder:
		return AdderView(m)
	case series:
		return SeriesView(m)
	}

	return "\n\nView got confused 🤮😭😨👿💔🔥💯💯💯\n\n"
}
