package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"github.com/markbates/going/wait"
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

func refreshArticles(email string) {
	currentCursor, _ := datastore.GetCurrenCursorByEmail(email)
	processUserDelta(email, currentCursor)
}

func processChanges(users []int) {
	for _, v := range users {
		email, err := datastore.GetUserEmailByUID(v)
		if err == nil {
			currentCursor, _ := datastore.GetCurrenCursorByEmail(email)
			processUserDelta(email, currentCursor)
		}
	}

}

func processUserDelta(email, cursor string) {
	at, err := datastore.LoadUserToken(email)
	if err != nil {
		log.Fatal(err)
		return
	}
	dbc := dropbox.NewClient(at, config.AppToken)
	d, _ := dbc.GetDelta("/published", cursor)

	for _, v := range d.Deleted {
		a, err := datastore.LoadArticle(email + ":article:" + v)
		if err == nil {
			a.Delete()
			log.Printf("deleted: %s", v)
		}
	}

	wait.Wait(len(d.Updated), func(index int) {
		entry, _ := dbc.GetMetadata(d.Updated[index], true)
		file, _ := dbc.GetFile(d.Updated[index])
		content, _ := ioutil.ReadAll(file)
		article := datastore.ParseEntry(*entry, content)
		article.GenerateID(email)
		article.Save()
		log.Printf("updated: %s", article.Path)
	})

	datastore.ArticlesReindex(email)
	datastore.SaveCurrentCursor(email, d.Cursor)
}
