package main

import (
	"fmt"

	md "github.com/twells46/gomangatool/internal/mdapi"
)

func main() {
	//md.DlChapter(`362936f9-2456-4120-9bea-b247df21d0bc`)
	md.GetFeed(`6941f16b-b56e-404a-b4ba-2fc7e009d38f`, 0)
	fmt.Println("asdf")
}
