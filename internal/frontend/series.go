package frontend

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/twells46/gomangatool/internal/backend"
)

type ChapDlMsg int
type ChapReadMsg int

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
			Padding(0, 0, 0, 2)
	wrapStyle = lipgloss.NewStyle().
			Width(100).
			Padding(0, 2).
			Foreground(lipgloss.AdaptiveColor{Light: "#5f5f5f", Dark: "#777777"})
	boldStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"}).
			Bold(true)
)

// The components to view an individual series
type Series struct {
	manga  backend.Manga
	list   list.Model
	copied bool
}

func blankSeries() Series {
	// TODO: Make item style different based on:
	// downloaded
	// isRead
	//d := list.NewDefaultDelegate()
	//d.ShowDescription = false
	d := NewSeriesDelegate()

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
	// TODO: dlChap and ReadChap both need to have their messages handled when the return
	// to the update function
	switch msg := msg.(type) {
	case list.Model:
		m.series.list = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return seriesExit(m), nil
		// Should probably be tea.cmd
		case "r":
			new := backend.RefreshFeed(m.library.list.SelectedItem().(backend.Manga), m.store)
			m.library.list.SetItem(m.library.list.Index(), new)
			m.series.manga = new
			return newSeries(m), nil
		case "d":
			return m, dlChap(m.series.list.SelectedItem().(backend.Chapter), m.series.list.Index(), m.store)
		case "enter":
			return m, ReadChap(m.series.list.SelectedItem().(backend.Chapter), m.series.list.Index(), m.store)
		}
	}

	var cmd tea.Cmd
	m.series.list, cmd = m.series.list.Update(msg)
	return m, cmd
}

func dlChap(chapter backend.Chapter, idx int, store *backend.SQLite) tea.Cmd {
	// This function should download a chapter.
	// It must update the list with the chapter with the download flag triggered
	// so that I can style downloaded chapters differently.
	// It also needs to place the chapter correctly in the sorted list of chapters.
	// Currently it does not update the model from the parent Manga.
	// Maybe that should be an exit function thing?

	return func() tea.Msg {
		backend.DownloadChapters(store, chapter)
		return ChapDlMsg(idx)
	}
}

// TODO: This should download the chapter if it isn't already
func ReadChap(c backend.Chapter, idx int, store *backend.SQLite) tea.Cmd {
	return func() tea.Msg {
		fullPath := fmt.Sprintf("/home/twells/media/manga/%s", c.DirName(store))
		readCmd := exec.Command("imv", "-f", "-d", "-r", fullPath)
		if err := readCmd.Run(); err != nil {
			log.Println(err)
		}
		return ChapReadMsg(idx)
	}
}

// Overall Series view function
func SeriesView(m model) string {
	info := fmt.Sprintf("%s\n%s%s",
		wrapStyle.Render(RenderTags(m.series.manga.Tags)),
		boldStyle.Render("Description:\n"),
		wrapStyle.Render(m.series.manga.Descr))

	return lipgloss.JoinHorizontal(lipgloss.Left, m.series.list.View(), info)
}

func RenderTags(tags []backend.Tag) string {
	tagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 1, 0, 0)

	var sb strings.Builder
	for _, t := range tags {
		sb.WriteString(tagStyle.Render(t.String()) + "\t")
	}

	return fmt.Sprintf("%s%s", boldStyle.Render("Tags:\n"), sb.String())
}
