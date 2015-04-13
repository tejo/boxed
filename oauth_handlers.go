package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/tejo/boxed/datastore"
	"github.com/tejo/boxed/dropbox"
)

func account(w http.ResponseWriter, r *http.Request) {
	withSession(w, r, func(session *sessions.Session) {
		var AccessToken dropbox.AccessToken

		if email := session.Values["email"]; email == nil {
			fmt.Fprint(w, "no email found")
			return
		}
		email := session.Values["email"].(string)
		AccessToken, _ = datastore.LoadUserToken(email)

		dbc := dropbox.NewClient(AccessToken, config.AppToken)
		info, err := dbc.GetAccountInfo()
		if err != nil {
			// access token is not valid anymore
			// reset session
			session.Values["email"] = ""
			session.Save(r, w)
			fmt.Fprint(w, "access token not valid")
			return
		}
		fmt.Fprintf(w, "info = %+v\n", info)
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	withSession(w, r, func(session *sessions.Session) {
		RequestToken, _ := dropbox.StartAuth(config.AppToken)
		session.Values["RequestToken"] = RequestToken
		url, _ := url.Parse(config.HostWithProtocol + config.CallbackURL)
		authURL := dropbox.GetAuthorizeURL(RequestToken, url)
		session.Save(r, w)
		http.Redirect(w, r, authURL.String(), 302)
	})
}

// saves the user id in session, save used data and access token in
// db, creates the default folders
func callback(w http.ResponseWriter, r *http.Request) {
	withSession(w, r, func(session *sessions.Session) {
		RequestToken := session.Values["RequestToken"].(dropbox.RequestToken)
		AccessToken, _ := dropbox.FinishAuth(config.AppToken, RequestToken)
		dbc := dropbox.NewClient(AccessToken, config.AppToken)
		info, err := dbc.GetAccountInfo()
		if err != nil {
			log.Println(err)
		}
		datastore.SaveUserData(info, AccessToken)
		session.Values["email"] = info.Email
		session.Save(r, w)
		dbc.CreateDir("drafts")
		dbc.CreateDir("published")
		http.Redirect(w, r, "/", 302)
	})
}
