package frontend

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"
	"github.com/twells46/gomangatool/internal/backend"
)

var meta backend.MangaMeta

type (
	errMsg error
	tOpt   string
)

func (t tOpt) FilterValue() string { return string(t) }
func (t tOpt) Title() string       { return string(t) }
func (t tOpt) Description() string { return "" }

type model struct {
	textInput        textinput.Model
	list             list.Model
	seriesID         string
	fullTitle        string
	abbrevTitle      string
	err              error // NOTE: Currently unused
	stage            int
	fetched          bool // For stage 1: have the titles been fetched?
	textInputUpdated bool // For stage 2: has textInput been cleared and updated?
	store            *backend.SQLite
	quitting         bool // NOTE: Currently unused
}

// Initialize a new model
func InitModel() model {
	ti := textinput.New()
	ti.Placeholder = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 64

	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	l := list.New([]list.Item{}, d, 80, 20)
	l.Title = "Choose a title:"

	return model{
		textInput: ti,
		list:      l,
		err:       nil,
		stage:     0,
		fetched:   false,
		store:     backend.Opendb("manga.sqlite3"),
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
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	if m.stage == 0 {
		return UpdateIDInput(msg, m)
	} else if m.stage == 1 {
		return UpdateChooser(msg, m)
	} else if m.stage == 2 {
		return UpdateAbbrevInput(msg, m)
	}

	return m, tea.Quit
}

// Update function for the inputting the manga ID (stage 0)
func UpdateIDInput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.seriesID = m.textInput.Value()
			m.stage = 1 // Move to title choices list
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// Update function for choosing the title (stage 1)
func UpdateChooser(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if val, ok := m.list.SelectedItem().(tOpt); ok {
				m.fullTitle = string(val)
				m.stage = 2 // Move to abbreviation input form
			}
			return m, nil
		}
	case []list.Item:
		m.list.SetItems(msg)
		m.fetched = true
	}
	if !m.fetched {
		return m, getTitles(m.seriesID)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View function for abbreviated title input (stage 2)
func UpdateAbbrevInput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []list.Item:
		m.list.SetItems(msg)
		m.fetched = true
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlLeft:
			m.stage = 1 // Return to list to choose a new title
			return m, nil

		case tea.KeyEnter:
			m.abbrevTitle = m.textInput.Value()
			backend.NewManga(meta, m.fullTitle, m.abbrevTitle, m.store)
			return m, tea.Quit
		}
	}

	if !m.textInputUpdated {
		m.textInput.Reset()
		m.textInput.Placeholder = "abbrev_title"
		m.textInput.Focus()
		m.textInputUpdated = true
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// Main view function, which calls the correct sub-function
func (m model) View() string {
	if m.stage == 0 {
		return ViewIDInput(m)
	} else if m.stage == 1 {
		return ViewChooser(m)
	} else if m.stage == 2 {
		return ViewAbbrevInput(m)
	}

	return "\n\nHow did we get here?\n\n"
}

// View function for ID input (stage 0)
func ViewIDInput(m model) string {
	return fmt.Sprintf("Input the ID: %s", m.textInput.View())
}

// View function for title chooser (stage 1)
func ViewChooser(m model) string {
	if !m.fetched {
		return "Querying Mangadex..."
	}
	return m.list.View()
}

// View function for abbreviated title input (stage 2)
func ViewAbbrevInput(m model) string {
	var view strings.Builder
	view.WriteString(fmt.Sprintf("\nYour chosen title is:\n'%s'\n", m.fullTitle))
	if m.textInputUpdated {
		view.WriteString(fmt.Sprintf("Input the abbreviated title: %s\n\n", m.textInput.View()))
	}
	view.WriteString("To go back and choose a different title, press ctrl-leftarrow")
	return view.String()
}

// Get the titles and put them into a slice of []list.Item
// The title chooser update function is "listening" for this
// type of tea.Msg message and adds them to the view.
func getTitles(MangaID string) tea.Cmd {
	return func() tea.Msg {
		meta = backend.PullMangaMeta(MangaID)
		titleOptions := []list.Item{tOpt(meta.Data.Attributes.Title.En)}
		for _, v := range meta.Data.Attributes.AltTitles {
			if len(v.En) > 0 {
				titleOptions = append(titleOptions, tOpt(v.En))
			} else if len(v.Ja) > 0 {
				titleOptions = append(titleOptions, tOpt(v.Ja))
			} else if len(v.JaRo) > 0 {
				titleOptions = append(titleOptions, tOpt(v.JaRo))
			}
		}
		return titleOptions
	}
}
