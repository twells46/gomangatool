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
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case Adder:
		m.adder = msg
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
		cmds = append(cmds, getTitles(m.adder))
	}

	var cmd tea.Cmd
	m.adder.list, cmd = m.adder.list.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// Update function for abbreviated title input
func AdderUpdateAbbrevInput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)

	switch msg := msg.(type) {
	case backend.Manga:
		cmds = append(cmds, m.library.list.InsertItem(2147483647, msg))
		m = adderExit(m)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlLeft:
			m.adder.stage = chooser // Return to list to choose a new title
			return m, nil

		case tea.KeyEnter:
			m.adder.abbrevTitle = m.adder.textInput.Value()
			cmds = append(cmds, adderNewManga(&m.adder, m.store))
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
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
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

// View function for ID input
func AdderViewIDInput(m model) string {
	return fmt.Sprintf("Input the ID: %s", m.adder.textInput.View())
}

// View function for title chooser
func AdderViewChooser(m model) string {
	if !m.adder.fetched {
		return "Querying Mangadex..."
	}
	return m.adder.list.View()
}

// View function for abbreviated title input
// TODO: Styling
func AdderViewAbbrevInput(m model) string {
	var view strings.Builder
	view.WriteString(fmt.Sprintf("\nYour chosen title is:\n'%s'\n", m.adder.fullTitle))
	if m.adder.textInputUpdated {
		view.WriteString(fmt.Sprintf("Input the abbreviated title: %s\n\n", m.adder.textInput.View()))
	}
	view.WriteString("To go back and choose a different title, press ctrl-leftarrow")
	return view.String()
}

// Get the title options and store the rest of the metadata so that we
// don't have to query the API multiple times.
// For now it needs to take and return the whole Adder
// since it does so many things.
func getTitles(m Adder) tea.Cmd {
	return func() tea.Msg {
		meta := backend.PullMangaMeta(m.mangaID)
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
		m.list.SetItems(titleOptions)
		m.fetched = true
		m.meta = meta

		return m
	}
}

func adderNewManga(adder *Adder, store *backend.SQLite) tea.Cmd {
	return func() tea.Msg {
		return backend.NewManga(adder.meta, adder.fullTitle, adder.abbrevTitle, store)
	}
}
