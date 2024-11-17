package frontend

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/backend"
)

type Library struct {
	list    list.Model
	toAddID string
}

func initLibrary(store *backend.SQLite) Library {
	items := make([]list.Item, 0)
	for _, v := range store.GetAll() {
		items = append(items, list.Item(v))
	}
	d := list.NewDefaultDelegate()

	list := list.New(items, d, 80, 25)
	list.Title = "Library:"

	return Library{list, ""}
}

func LibraryUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			m.view = adder
		}
	}

	if m.library.toAddID != "" {
		return updateAfterAdd(m), nil
	}

	var cmd tea.Cmd
	m.library.list, cmd = m.library.list.Update(msg)
	return m, cmd
}

func LibraryView(m model) string {
	return m.library.list.View()
}

func updateAfterAdd(m model) model {
	new := m.store.GetByID(m.library.toAddID)
	m.library.list.InsertItem(2147483647, list.Item(new)) // Ghetto append using the max of an int
	m.library.toAddID = ""
	return m
}
