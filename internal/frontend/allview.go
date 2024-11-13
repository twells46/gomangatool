package frontend

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/backend"
)

type AllView struct {
	list list.Model
}

func initAllView(store *backend.SQLite) AllView {
	items := make([]list.Item, 0)
	for _, v := range store.GetAll() {
		items = append(items, list.Item(v))
	}
	d := list.NewDefaultDelegate()

	list := list.New(items, d, 80, 25)
	list.Title = "Choose a title:"

	return AllView{list}
}

func AllViewUpdate(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			m.view = adder
		}
	}

	var cmd tea.Cmd
	m.allView.list, cmd = m.allView.list.Update(msg)
	return m, cmd
}

func AllViewView(m model) string {
	return m.allView.list.View()
}
