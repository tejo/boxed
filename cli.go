package main

import (
	"flag"
	"os"
)

func handleCommands() {
	refresh := flag.Bool("refresh", false, "refresh posts for the provided email if present, otherwise refresh the default email")
	flag.Parse()

	if *refresh {
		if len(flag.Args()) > 0 {
			refreshPosts(flag.Args()[0])
		} else {
			refreshPosts(config.DefaultUserEmail)
		}
		os.Exit(1)
	}

}
