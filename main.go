package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/mattn/go-sqlite3"
	"github.com/twells46/gomangatool/internal/backend"
)

type (
	errMsg error
)

var seriesID string

type inputIDModel struct {
	textinput textinput.Model
	err       error
}

func InitInputIDModel() inputIDModel {
	ti := textinput.New()
	ti.Placeholder = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 64

	return inputIDModel{
		textinput: ti,
		err:       nil,
	}
}
func (m inputIDModel) Init() tea.Cmd {
	return nil
}

func (m inputIDModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC:
			seriesID = m.textinput.Value()
			return m, tea.Quit
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m inputIDModel) View() string {
	return fmt.Sprintf("Input the ID: %s", m.textinput.View())
}

func main() {
	p := tea.NewProgram(InitInputIDModel())
	if _, err := p.Run(); err != nil {
		log.Fatalln(err)
	}

	store := backend.Opendb("manga.sqlite3")
	backend.NewManga(seriesID, store)
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
