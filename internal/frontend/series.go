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
	d := list.NewDefaultDelegate()
	//d.ShowDescription = false
	//d := NewSeriesDelegate()
	l := list.New([]list.Item{}, d, 80, 25)

	return Series{
		list: l,
	}
}

// Returns the model with a correctly set list.
func seriesRefreshList(m model) model {
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
	m.library.list.SetItem(m.library.list.Index(), m.series.manga)
	m.series.copied = false
	m.view = library
	m.series.list.SetItems([]list.Item{})
	return m
}

// Overall Series update function
func SeriesUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case ChapDlMsg:
		m.series.manga.Chapters[msg].Downloaded = true
		tmp := m.series.list.Items()[msg].(backend.Chapter)
		tmp.Downloaded = true
		cmds = append(cmds, m.series.list.SetItem(int(msg), tmp))
		m.series.list.StopSpinner()
	case ChapReadMsg:
		m.series.manga.Chapters[msg].IsRead = true
		tmp := m.series.list.Items()[msg].(backend.Chapter)
		tmp.IsRead = true
		cmds = append(cmds, m.series.list.SetItem(int(msg), tmp))
	case backend.Manga:
		m.series.manga = msg
		m.series.list.StopSpinner()
		return seriesRefreshList(m), nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return seriesExit(m), nil
		case "r":
			cmds = append(cmds, m.series.list.StartSpinner())
			cmds = append(cmds, refresh(m.series.manga, m.store))
		case "d":
			cmds = append(cmds, m.series.list.StartSpinner())
			cmds = append(cmds, dlChap(m.series.list.SelectedItem().(backend.Chapter), m.series.list.Index(), m.store))
			return m, tea.Batch(cmds...) // prevent 'd' from being handled by the list
		case "enter":
			cmds = append(cmds, readChap(m.series.list.SelectedItem().(backend.Chapter), m.series.list.Index(), m.store))
		}
	}

	var cmd tea.Cmd
	m.series.list, cmd = m.series.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func refresh(manga backend.Manga, store *backend.SQLite) tea.Cmd {
	return func() tea.Msg {
		return backend.RefreshFeed(manga, store)
	}
}

func dlChap(chapter backend.Chapter, idx int, store *backend.SQLite) tea.Cmd {
	return func() tea.Msg {
		backend.DownloadChapters(store, chapter)
		return ChapDlMsg(idx)
	}
}

func readChap(c backend.Chapter, idx int, store *backend.SQLite) tea.Cmd {
	return func() tea.Msg {
		readCmd := exec.Command("imv", "-f", "-d", "-r", c.ChapterPath)
		if err := readCmd.Run(); err != nil {
			log.Println(err)
		}
		store.UpdateChapterRead(c)
		return ChapReadMsg(idx)
	}
}

// Overall Series view function
func SeriesView(m model) string {
	info := fmt.Sprintf("%s\n%s%s",
		wrapStyle.Render(renderTags(m.series.manga.Tags)),
		boldStyle.Render("Description:\n"),
		wrapStyle.Render(m.series.manga.Descr))

	return lipgloss.JoinHorizontal(lipgloss.Left, m.series.list.View(), info)
}

func renderTags(tags []backend.Tag) string {
	tagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Padding(0, 1, 0, 0)

	var sb strings.Builder
	for _, t := range tags {
		sb.WriteString(tagStyle.Render(t.String()) + "\t")
	}

	return fmt.Sprintf("%s%s", boldStyle.Render("Tags:\n"), sb.String())
}
