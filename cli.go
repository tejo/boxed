package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/tejo/boxed/datastore"
	"github.com/tejo/boxed/dropbox"
)

func handleCommands() {
	refresh := flag.Bool("refresh", false, "refresh articles for the provided email if present, otherwise refresh the default email")
	oauth := flag.Bool("oauth", false, "authorize oauth app from command line")
	flag.Parse()

	if *refresh {
		if len(flag.Args()) > 0 {
			refreshArticles(flag.Args()[0])
		} else {
			refreshArticles(config.DefaultUserEmail)
		}
		os.Exit(1)
	}

	if *oauth {
		requestToken, _ := dropbox.StartAuth(config.AppToken)
		url, _ := url.Parse("")
		fmt.Println("open in a web browser the following url and authorize boxed app:")
		fmt.Println(dropbox.GetAuthorizeURL(requestToken, url))
		fmt.Println("\n\npress enter when ready")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		accessToken, _ := dropbox.FinishAuth(config.AppToken, requestToken)
		dbc := dropbox.NewClient(accessToken, config.AppToken)
		info, err := dbc.GetAccountInfo()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("authorized: %s %s\n", info.DisplayName, info.Email)
			datastore.SaveUserData(info, accessToken)
			dbc.CreateDir("drafts")
			dbc.CreateDir("published")
		}
		os.Exit(1)
	}

}
