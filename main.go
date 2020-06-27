package main

import (
	"flag"
	"vincit.fi/image-sorter/image"
	"vincit.fi/image-sorter/ui"
)

func main() {
	flag.Parse()
	root := flag.Arg(0)
	manager := image.ManagerForDir(root)

	ui.Init(&manager).Run([]string{})
}

