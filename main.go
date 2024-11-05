package main

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/twells46/gomangatool/internal/backend"
)

func main() {
	/*
		md.DlChapter(`362936f9-2456-4120-9bea-b247df21d0bc`)
		feed := md.GetFeed(`6941f16b-b56e-404a-b4ba-2fc7e009d38f`, 0)
		fmt.Println(feed)
	*/

	backend.SqlTester()
}
