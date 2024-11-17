package frontend

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/backend"
)

type Adder struct {
	textInput        textinput.Model
	list             list.Model
	seriesID         string
	fullTitle        string
	abbrevTitle      string
	stage            int  // Since theres only 3 stages I don't define a const for this
	fetched          bool // For stage 1: have the titles been fetched?
	textInputUpdated bool // For stage 2: has textInput been cleared and updated?
}

func initAdder() Adder {
	ti := textinput.New()
	ti.Placeholder = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 64

	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	l := list.New([]list.Item{}, d, 80, 20)
	l.Title = "Choose a title:"
	return Adder{
		textInput: ti,
		list:      l,
	}
}

type (
	errMsg error
	tOpt   string
)

// TODO: This should be integrated into the Adder struct
var meta backend.MangaMeta

func (t tOpt) FilterValue() string { return string(t) }
func (t tOpt) Title() string       { return string(t) }
func (t tOpt) Description() string { return "" }

func AdderUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.adder.stage = 0
			m.view = library
		}
	}

	switch m.adder.stage {
	case 0:
		return AdderUpdateIDInput(msg, m)
	case 1:
		return AdderUpdateChooser(msg, m)
	case 2:
		return AdderUpdateAbbrevInput(msg, m)
	default:
		return m, nil
	}
}

func adderExit(m model) model {
	m.adder.list.SetItems([]list.Item{})
	m.adder.textInput.Reset()
	m.adder.stage = 0
	m.view = library
	return m
}

// Update function for the inputting the manga ID (stage 0)
func AdderUpdateIDInput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return adderExit(m), nil

		case tea.KeyEnter:
			m.adder.seriesID = m.adder.textInput.Value()
			m.adder.stage = 1 // Move to title choices list
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil
	}
	var cmd tea.Cmd
	m.adder.textInput, cmd = m.adder.textInput.Update(msg)
	return m, cmd
}

// Update function for choosing the title (stage 1)
func AdderUpdateChooser(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if val, ok := m.adder.list.SelectedItem().(tOpt); ok {
				m.adder.fullTitle = string(val)
				m.adder.stage = 2 // Move to abbreviation input form
			}
			return m, nil
		}
	case []list.Item:
		m.adder.list.SetItems(msg)
		m.adder.fetched = true
	}
	if !m.adder.fetched {
		return m, getTitles(m.adder.seriesID)
	}

	var cmd tea.Cmd
	m.adder.list, cmd = m.adder.list.Update(msg)
	return m, cmd
}

// View function for abbreviated title input (stage 2)
func AdderUpdateAbbrevInput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []list.Item:
		m.adder.list.SetItems(msg)
		m.adder.fetched = true
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlLeft:
			m.adder.stage = 1 // Return to list to choose a new title
			return m, nil

		case tea.KeyEnter:
			m.adder.abbrevTitle = m.adder.textInput.Value()
			backend.NewManga(meta, m.adder.fullTitle, m.adder.abbrevTitle, m.store)
			m.library.toAddID = m.adder.seriesID
			return adderExit(m), nil
		}
	}

	if !m.adder.textInputUpdated {
		m.adder.textInput.Reset()
		m.adder.textInput.Placeholder = "abbrev_title"
		m.adder.textInput.Focus()
		m.adder.textInputUpdated = true
	}

	var cmd tea.Cmd
	m.adder.textInput, cmd = m.adder.textInput.Update(msg)
	return m, cmd
}

func AdderView(m model) string {
	switch m.adder.stage {
	case 0:
		return AdderViewIDInput(m)
	case 1:
		return AdderViewChooser(m)
	case 2:
		return AdderViewAbbrevInput(m)
	default:
		return "Adder got confused 🤮😭😨👿💔🔥💯💯💯"
	}
}

// View function for ID input (stage 0)
func AdderViewIDInput(m model) string {
	return fmt.Sprintf("Input the ID: %s", m.adder.textInput.View())
}

// View function for title chooser (stage 1)
func AdderViewChooser(m model) string {
	if !m.adder.fetched {
		return "Querying Mangadex..."
	}
	return m.adder.list.View()
}

// View function for abbreviated title input (stage 2)
func AdderViewAbbrevInput(m model) string {
	var view strings.Builder
	view.WriteString(fmt.Sprintf("\nYour chosen title is:\n'%s'\n", m.adder.fullTitle))
	if m.adder.textInputUpdated {
		view.WriteString(fmt.Sprintf("Input the abbreviated title: %s\n\n", m.adder.textInput.View()))
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
