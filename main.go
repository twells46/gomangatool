package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/twells46/gomangatool/internal/backend"
	"github.com/twells46/gomangatool/internal/frontend"
)

func main() {
	store := backend.Opendb("manga.sqlite3")
	all := store.GetAll()
	for i, m := range all {
		all[i] = backend.RefreshFeed(m, store)
	}

	p := tea.NewProgram(frontend.InitModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalln(err)
	}

	/*
		tester := all[0]
			tester.Chapters = backend.DownloadAll(tester.Chapters, store)
			fmt.Println(tester)
	*/
	//new := backend.RefreshFeed(tester, store)
	//fmt.Println(new)

	//fmt.Println(manga)
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
