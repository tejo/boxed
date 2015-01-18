package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/scottferg/Dropbox-Go/dropbox"
)

var s dropbox.Session
var store = sessions.NewCookieStore([]byte("182hetsgeih8765$aasdhj"))

type User struct {
	DropboxID string
	Name      string
	Token     dropbox.AccessToken
}

func NewDropboxSession() dropbox.Session {
	return dropbox.Session{
		AppKey:     "72ton4woqnari86",
		AppSecret:  "12a0p8gtsg7fp3i",
		AccessType: "app_folder",
	}
}

func main() {
	s = NewDropboxSession()

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/login", Login)
	router.GET("/oauth/callback", Callback)

	log.Fatal(http.ListenAndServe(":8080", router))
}
func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	_, err := s.ObtainRequestToken()
	if err != nil {
		log.Fatal(err)
		s = NewDropboxSession()
		s.ObtainRequestToken()
	}
	http.Redirect(w, r, dropbox.GenerateAuthorizeUrl(s.Token.Key, &dropbox.Parameters{
		OAuthCallback: "http://localhost:8080/oauth/callback",
	}), 301)
}

func Callback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "godropblog")
	token, _ := s.ObtainAccessToken()
	session.Values["key"] = token.Key
	session.Values["secret"] = token.Secret
	session.Save(r, w)
	http.Redirect(w, r, "/", 301)
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	session, _ := store.Get(r, "godropblog")

	if key, secret := session.Values["key"], session.Values["secret"]; key == nil && secret == nil {
		http.Redirect(w, r, "/login", 301)
		return
	} else {
		s.Token.Key = key.(string)
		s.Token.Secret = secret.(string)
	}

	dropbox.CreateFolder(
		s,
		dropbox.Uri{
			Root: "sandbox",
			Path: "/published",
		},
		&dropbox.Parameters{},
	)

	account, err := dropbox.GetAccount(s, &dropbox.Parameters{})

	if err != nil {
		//access toke is not valid anymore
		fmt.Fprintf(w, " %+v\n", err)
		// reset all session
		s = NewDropboxSession()
		session.Values["key"], session.Values["secret"] = "", ""
		session.Save(r, w)
		return
	}

	fmt.Printf("account = %+v\n", account)

	dropbox.CreateFolder(
		s,
		dropbox.Uri{
			Root: "sandbox",
			Path: "/draft",
		},
		&dropbox.Parameters{},
	)

	p := &dropbox.Parameters{
		FileLimit: "10000",
		List:      "true",
	}

	u := dropbox.Uri{
		Root: "sandbox",
		Path: "/",
	}

	data, err := dropbox.GetMetadata(s, u, p)
	fmt.Fprintf(w, " %+v\n %+v\n", data, err)
}
