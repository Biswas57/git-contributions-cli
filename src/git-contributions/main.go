package main

import (
	"flag"
)

func main() {
	var folder string
	var emailPath string
	flag.StringVar(&folder, "add", "", "add a new folder to scan for Git repositories")
	flag.StringVar(&emailPath, "path", "", "path to your emails")
	flag.Parse()

	if folder != "" {
		scan(folder)
		return
	}

	stats(emailPath)
}
