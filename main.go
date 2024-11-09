package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"
)

type (
	errMsg error
)

type model struct {
	textinput textinput.Model
	seriesID  string
	err       error
	stage     int
	quitting  bool // NOTE: Currently unused
}

// Initialize a new model
func initModel() model {
	ti := textinput.New()
	ti.Placeholder = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 64

	return model{
		textinput: ti,
		err:       nil,
		stage:     0,
		quitting:  false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// Main update function, which handles universal quit keys
// then passes off to the appropriate sub-function
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	if m.stage == 0 {
		return UpdateIDInput(msg, m)
	}

	if m.stage == 1 {
		return UpdateChooser(msg, m)
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
			m.stage++
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
	panic("unimplemented")
}

// Main view function, which calls the correct sub-function
func (m model) View() string {
	if m.stage == 0 {
		return ViewIDInput(m)
	} else if m.stage == 1 {
		return ViewChooser(m)
	}

	return "\n\nSomething went wrong...\n\n"
}

// View function for ID input (stage 0)
func ViewIDInput(m model) string {
	return fmt.Sprintf("Input the ID: %s", m.textinput.View())
}

func ViewChooser(m model) string {
	return "\n\nUNDER CONSTRUCTION\n\n"
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
