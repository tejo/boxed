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

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/login", Login)
	router.GET("/oauth/callback", Callback)

	log.Fatal(http.ListenAndServe(":8080", router))
}
func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	RequestToken, _ = dropbox.StartAuth(AppToken)
	fmt.Printf("dropbox.StartAuth() = %+v\n", RequestToken)
	u, _ := url.Parse("http://localhost:8080/oauth/callback")
	authUrl := dropbox.GetAuthorizeURL(RequestToken, u)
	http.Redirect(w, r, authUrl.String(), 302)
}

func Callback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	AccessToken, _ = dropbox.FinishAuth(AppToken, RequestToken)
	fmt.Printf("AccessToken = %+v\n", AccessToken)
	http.Redirect(w, r, "/", 302)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	db := &dropbox.Client{
		AppToken:    AppToken,
		AccessToken: AccessToken,
		Config: dropbox.Config{
			Access: dropbox.AppFolder,
			Locale: "us",
		}}
	info, err := db.GetAccountInfo()
	fmt.Printf("err = %+v\n", err)
	fmt.Printf("err = %+v\n", info)
	db.CreateDir("drafts")
	db.CreateDir("published")
	db.GetDelta()
}
