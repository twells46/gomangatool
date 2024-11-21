package frontend

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/twells46/gomangatool/internal/backend"
)

type opDoneMsg byte

// The components to view an individual series
type Series struct {
	manga  backend.Manga
	list   list.Model
	copied bool
}

func blankSeries() Series {
	// TODO: Make a custom delegate to display more info
	d := list.NewDefaultDelegate()
	d.ShowDescription = false

	return Series{
		list: list.New([]list.Item{}, d, 80, 25),
	}
}

// Create a new series view.
// Returns that model with a correctly set list
func newSeries(m model) model {
	items := make([]list.Item, 0)
	for _, chapter := range m.series.manga.Chapters {
		items = append(items, list.Item(chapter))
	}

	m.series.list.SetItems(items)
	m.series.list.Title = m.series.manga.FullTitle
	m.series.copied = true

	return m
}

// Exit the series view and return to the Library
func seriesExit(m model) model {
	m.series.copied = false
	m.view = library
	m.series.list.SetItems([]list.Item{})
	return m
}

// Overall Series update function
func SeriesUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case list.Model:
		m.series.list = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return seriesExit(m), nil
		// TODO: Fix freezing when refreshing or downloading
		// Should probably be tea.cmds
		case "r":
			new := backend.RefreshFeed(m.library.list.SelectedItem().(backend.Manga), m.store)
			m.library.list.SetItem(m.library.list.Index(), new)
			m.series.manga = new
			return newSeries(m), nil
		case "d":
			return m, dlChap(&m.series.list, m.store)
		}
	}

	// TODO: Fix crappy loading
	if !m.series.copied {
		return newSeries(m), nil
	}

	var cmd tea.Cmd
	m.series.list, cmd = m.series.list.Update(msg)
	return m, cmd
}

// TODO: When I get styling set up, need to check that this function properly triggers
// the UI changes that it should
func dlChap(m *list.Model, store *backend.SQLite) tea.Cmd {
	// This function should download a chapter.
	// It must update the list with the chapter with the download flag triggered
	// so that I can style downloaded chapters differently.
	// It also needs to place the chapter correctly in the sorted list of chapters.
	// Currently it does not update the model from the parent Manga.
	// Maybe that should be an exit function thing?

	return func() tea.Msg {
		chapter := m.SelectedItem().(backend.Chapter)
		new := backend.DownloadChapters(store, chapter)
		m.SetItem(m.Index(), new[0])
		return m
	}
}

// Overall Series view function
func SeriesView(m model) string {
	info := fmt.Sprintf("%s\n%v\n%s", m.series.manga.FullTitle, m.series.manga.Tags, m.series.manga.Descr)
	return lipgloss.JoinVertical(lipgloss.Top, info, m.series.list.View())
}
