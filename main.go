package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/tejo/dropbox"
)

var store = sessions.NewCookieStore([]byte("182hetsgeih8765$aasdhj"))

type User struct {
	DropboxID string
	Name      string
}

var AppToken = dropbox.AppToken{
	Key:    "72ton4woqnari86",
	Secret: "12a0p8gtsg7fp3i",
}

var RequestToken dropbox.RequestToken
var AccessToken dropbox.AccessToken

var callbackUrl = "http://localhost:8080/oauth/callback"

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/login", Login)
	router.GET("/oauth/callback", Callback)

	log.Fatal(http.ListenAndServe(":8080", router))
}
func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	RequestToken, _ = dropbox.StartAuth(AppToken)
	url, _ := url.Parse(callbackUrl)
	authUrl := dropbox.GetAuthorizeURL(RequestToken, url)
	http.Redirect(w, r, authUrl.String(), 302)
}

func Callback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "godropblog")
	AccessToken, _ = dropbox.FinishAuth(AppToken, RequestToken)
	session.Values["key"] = AccessToken.Key
	session.Values["secret"] = AccessToken.Secret
	session.Save(r, w)
	fmt.Printf("AccessToken = %+v\n", AccessToken)
	http.Redirect(w, r, "/", 302)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "godropblog")

	if key, secret := session.Values["key"], session.Values["secret"]; key == nil && secret == nil {
		// http.Redirect(w, r, "/login", 302)
		return
	} else {
		AccessToken.Key = key.(string)
		AccessToken.Secret = secret.(string)
	}
	db := &dropbox.Client{
		AppToken:    AppToken,
		AccessToken: AccessToken,
		Config: dropbox.Config{
			Access: dropbox.AppFolder,
			Locale: "us",
		}}
	info, err := db.GetAccountInfo()

	if err != nil {
		//access token is not valid anymore
		fmt.Fprintf(w, " %+v\n", err)
		// reset all session
		session.Values["key"], session.Values["secret"] = "", ""
		session.Save(r, w)
		// http.Redirect(w, r, "/login", 302)
		return
	}

	fmt.Printf("err = %+v\n", err)
	fmt.Printf("err = %+v\n", info)
	db.CreateDir("drafts")
	db.CreateDir("published")
	delta, err := db.GetDelta()
	fmt.Printf("delta = %+v\n", delta)
	fmt.Printf("delta err = %+v\n", err)
}
