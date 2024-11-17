package frontend

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/twells46/gomangatool/internal/backend"
)

// The components to view an individual series
type Series struct {
	manga backend.Manga
	list  list.Model
	ready bool
}

// Create a new series view.
// Returns that model with a correctly set list
func newSeries(m model) model {
	items := make([]list.Item, 0)
	for _, chapter := range m.series.manga.Chapters {
		items = append(items, list.Item(chapter))
	}

	// TODO: Make a custom delegate to display more info
	d := list.NewDefaultDelegate()
	d.ShowDescription = false

	m.series.list = list.New(items, d, 80, 25)
	m.series.list.Title = m.series.manga.FullTitle
	m.series.ready = true

	return m
}

// Exit the series view and return to the Library
func seriesExit(m model) model {
	m.series.ready = false
	m.view = library
	return m
}

// Overall Series update function
func SeriesUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return seriesExit(m), nil
		}
		switch msg.String() {
		case "r":
			new := backend.RefreshFeed(m.library.list.SelectedItem().(backend.Manga), m.store)
			m.library.list.SetItem(m.library.list.Index(), new)
			m.series.manga = new
			return newSeries(m), nil
		}
	}

	if !m.series.ready {
		return newSeries(m), nil
	}

	var cmd tea.Cmd
	m.series.list, cmd = m.series.list.Update(msg)
	return m, cmd
}

// Overall Series view function
func SeriesView(m model) string {
	if !m.series.ready {
		return "Loading..."
	}

	info := fmt.Sprintf("%s\n%v\n%s", m.series.manga.FullTitle, m.series.manga.Tags, m.series.manga.Descr)
	return lipgloss.JoinVertical(lipgloss.Top, info, m.series.list.View())
}
