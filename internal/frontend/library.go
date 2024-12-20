package frontend

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/backend"
)

// The components of the main, library view
type Library struct {
	list list.Model
}

// Initialize a new Library with the stored series
func initLibrary(store *backend.SQLite) Library {
	items := make([]list.Item, 0)
	for _, v := range store.GetAll() {
		items = append(items, list.Item(v))
	}
	d := list.NewDefaultDelegate()

	list := list.New(items, d, 80, 25)
	list.Title = "Library:"

	return Library{list}
}

// Overall Library update function
func LibraryUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.series.manga = m.library.list.SelectedItem().(backend.Manga)
			m.view = series
			// By calling newSeries here, the list will be loaded and rendered properly
			// instantly
			return seriesRefreshList(m), nil
		case "a":
			m.view = adder
			return m, nil
		case "r":
			new := backend.RefreshFeed(m.library.list.SelectedItem().(backend.Manga), m.store)
			m.library.list.SetItem(m.library.list.Index(), new)
		case "R":
			for i, manga := range m.library.list.Items() {
				new := backend.RefreshFeed(manga.(backend.Manga), m.store)
				m.library.list.SetItem(i, new)
			}
		}
	}

	var cmd tea.Cmd
	m.library.list, cmd = m.library.list.Update(msg)
	return m, cmd
}

// Overall Library view function
func LibraryView(m model) string {
	return m.library.list.View()
}
