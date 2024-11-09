package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"
	"github.com/twells46/gomangatool/internal/backend"
)

type (
	errMsg error
	tOpt   string
)

func (t tOpt) FilterValue() string { return string(t) }
func (t tOpt) Title() string       { return string(t) }
func (t tOpt) Description() string { return "" }

type model struct {
	textinput textinput.Model
	seriesID  string
	fullTitle string
	list      list.Model
	err       error
	stage     int
	fetched   bool
	quitting  bool // NOTE: Currently unused
}

// Initialize a new model
func initModel() model {
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
		textinput: ti,
		list:      l,
		err:       nil,
		stage:     0,
		fetched:   false,
		quitting:  false, // NOTE: Currently unused
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
			m.seriesID = m.textinput.Value()
			m.stage = 1 // Move to title choices list
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil
	}
	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

// Update function for choosing the title (stage 1)
func UpdateChooser(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []list.Item:
		m.list.SetItems(msg)
		m.fetched = true
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if val, ok := m.list.SelectedItem().(tOpt); ok {
				m.fullTitle = string(val)
				m.stage = 2 // Move to abbreviation input form
			}
			return m, nil
		}
	}
	if !m.fetched {
		return m, getMeta(m.seriesID)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View function for abbreviated title input (stage 2)
func UpdateAbbrevInput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	panic("unimplemented")
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

	return "\n\nSomething went wrong...\n\n"
}

// View function for ID input (stage 0)
func ViewIDInput(m model) string {
	return fmt.Sprintf("Input the ID: %s", m.textinput.View())
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
	return fmt.Sprintf("\n\nUNDER CONSTRUCTION\n\n")
}

func getMeta(MangaID string) tea.Cmd {
	return func() tea.Msg {
		meta := backend.PullMangaMeta(MangaID)
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

func main() {
	p := tea.NewProgram(initModel())
	if _, err := p.Run(); err != nil {
		log.Fatalln(err)
	}

	//store := backend.Opendb("manga.sqlite3")
	//backend.NewManga(seriesID, store)
	/*
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		md.DlChapter(`362936f9-2456-4120-9bea-b247df21d0bc`)
		feed := md.GetFeed(`6941f16b-b56e-404a-b4ba-2fc7e009d38f`, 0)
		fmt.Println(feed)

		store := backend.Opendb("manga.sqlite3")
		//backend.SqlTester()
		backend.NewManga("ee51d8fb-ba27-46a5-b204-d565ea1b11aa", store)
	*/
}
