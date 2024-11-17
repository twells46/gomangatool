package frontend

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/backend"
)

const (
	idInput int = iota
	chooser
	abbrevInput
)

// The components of the Adder form, for adding new Manga
type Adder struct {
	textInput        textinput.Model
	list             list.Model
	mangaID          string
	fullTitle        string
	abbrevTitle      string
	meta             backend.MangaMeta
	stage            int
	fetched          bool // For stage 1: have the title options been fetched?
	textInputUpdated bool // For stage 2: has textInput been cleared and updated?
}

// Return an adder with initialized textinput and list
func newAdder() Adder {
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

// Clear the adder and return to the library
func adderExit(m model) model {
	m.view = library
	m.adder = newAdder()
	return m
}

type (
	errMsg error
	tOpt   string
)

// Implement list.Item and list.DefaultItem for tOpt
func (t tOpt) FilterValue() string { return string(t) }
func (t tOpt) Title() string       { return string(t) }
func (t tOpt) Description() string { return "" }

// Overall update function for the Adder
func AdderUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return adderExit(m), nil
		}
	}

	switch m.adder.stage {
	case idInput:
		return AdderUpdateIDInput(msg, m)
	case chooser:
		return AdderUpdateChooser(msg, m)
	case abbrevInput:
		return AdderUpdateAbbrevInput(msg, m)
	default:
		return m, nil
	}
}

// Update function for the inputting the manga ID (stage 0)
func AdderUpdateIDInput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.adder.mangaID = m.adder.textInput.Value()
			m.adder.stage = chooser // Move to title choices list
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
				m.adder.stage = abbrevInput // Move to abbreviation input form
			}
			return m, nil
		}
	}
	if !m.adder.fetched {
		return getTitles(m), nil
	}

	var cmd tea.Cmd
	m.adder.list, cmd = m.adder.list.Update(msg)
	return m, cmd
}

// View function for abbreviated title input (stage 2)
func AdderUpdateAbbrevInput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlLeft:
			m.adder.stage = chooser // Return to list to choose a new title
			return m, nil

		case tea.KeyEnter:
			m.adder.abbrevTitle = m.adder.textInput.Value()
			backend.NewManga(m.adder.meta, m.adder.fullTitle, m.adder.abbrevTitle, m.store)
			m.library.toAddID = m.adder.mangaID // Tell the library view it has a new series to display
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

// Overall view function for the Adder
func AdderView(m model) string {
	switch m.adder.stage {
	case idInput:
		return AdderViewIDInput(m)
	case chooser:
		return AdderViewChooser(m)
	case abbrevInput:
		return AdderViewAbbrevInput(m)
	default:
		return "\n\nAdder got confused 🤮😭😨👿💔🔥💯💯💯\n\n"
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
// Returns the model with the items set and stores the metadata for later use
func getTitles(m model) model {
	meta := backend.PullMangaMeta(m.adder.mangaID)
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
	m.adder.list.SetItems(titleOptions)
	m.adder.fetched = true
	m.adder.meta = meta

	return m
}
