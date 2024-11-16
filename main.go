package main

import (
	"fmt"

	"github.com/twells46/gomangatool/internal/backend"
)

func main() {
	//p := tea.NewProgram(frontend.InitModel(), tea.WithAltScreen())
	//if _, err := p.Run(); err != nil {
	//	log.Fatalln(err)
	//}

	store := backend.Opendb("manga.sqlite3")
	tester := store.GetAll()
	new := backend.RefreshFeed(tester[0], store)
	fmt.Println(new)
	//fmt.Println(store.GetChapters(tester[0].MangaID))

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
