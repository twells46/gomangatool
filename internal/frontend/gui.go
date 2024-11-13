package frontend

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/backend"
)

const (
	all int = iota
	series
	adder
)

// TODO: Rework stages to incorporate more than just adder
type model struct {
	view     int
	adder    Adder
	err      error // NOTE: Currently unused
	store    *backend.SQLite
	quitting bool // NOTE: Currently unused
}

// Initialize a new model
func InitModel() model {
	store := backend.Opendb("manga.sqlite3")

	return model{
		view:  adder,
		adder: initAdder(),
		store: store,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// Main update function, which handles universal quit keys
// then passes off to the appropriate sub-function
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.adder.list.SetWidth(msg.Width)
		m.adder.list.SetHeight(msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	switch m.view {
	case adder:
		switch m.adder.stage {
		case 0:
			return AdderUpdateIDInput(msg, m)
		case 1:
			return AdderUpdateChooser(msg, m)
		case 2:
			return AdderUpdateAbbrevInput(msg, m)
		}
	}

	return m, tea.Quit
}

// Main view function, which calls the correct sub-function
func (m model) View() string {
	switch m.view {
	case adder:
		switch m.adder.stage {
		case 0:
			return AdderViewIDInput(m)
		case 1:
			return AdderViewChooser(m)
		case 2:
			return AdderViewAbbrevInput(m)
		}
	}

	return "\n\nHow did we get here?\n\n"
}
