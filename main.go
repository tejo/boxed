package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/boltdb/bolt"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/markbates/going/wait"
	"github.com/shurcooL/go/github_flavored_markdown"
	"github.com/tejo/g-blog/datastore"
	"github.com/tejo/g-blog/dropbox"
)

var AppToken dropbox.AppToken
var defaultUserEmail string

var callbackUrl = "http://localhost:8080/oauth/callback"
var db *bolt.DB

func main() {
	defaultUserEmail = os.Getenv("DEFAULT_USER_EMAIL")
	AppToken = dropbox.AppToken{
		Key:    os.Getenv("KEY"),
		Secret: os.Getenv("SECRET"),
	}

	datastore.Connect("blog.db")
	defer datastore.Close()

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/login", Login)
	router.GET("/account", Account)
	router.GET("/r", Refresh)
	router.GET("/oauth/callback", Callback)
	router.GET("/a/:year/:month/day/:slug", ArticleHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}

type Article struct {
	Key     string
	Content string
	dropbox.FileMetadata
}

func Refresh(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var at dropbox.AccessToken
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserData"))
		token := b.Get([]byte(defaultUserEmail + ":token"))
		json.Unmarshal(token, &at)
		fmt.Printf("The answer is: %s\n", at)
		return nil
	})
	meta, _ := dbClient(at).GetMetadata("/published", true)
	fmt.Printf("meta.Contents = %+v\n", meta.Contents)
	wait.Wait(len(meta.Contents), func(index int) {
		entry := meta.Contents[index]
		if entry.IsDir {
			return
		}
		file, _ := dbClient(at).GetFile(entry.Path)
		content, _ := ioutil.ReadAll(file)
		article := &Article{
			Key:          defaultUserEmail + ":article:" + entry.Path,
			Content:      string(github_flavored_markdown.Markdown(content)),
			FileMetadata: entry,
		}
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("UserArticles"))
			a, err := json.Marshal(article)
			err = b.Put([]byte(article.Key), []byte(a))
			return err
		})
	})
}

func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	withSession(w, r, func(session *sessions.Session) {
		RequestToken, _ := dropbox.StartAuth(AppToken)
		session.Values["RequestToken"] = RequestToken
		url, _ := url.Parse(callbackUrl)
		authUrl := dropbox.GetAuthorizeURL(RequestToken, url)
		session.Save(r, w)
		http.Redirect(w, r, authUrl.String(), 302)
	})
}

// saves the user id in session, save used data and access token in
// db, creates the default folders
func Callback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	withSession(w, r, func(session *sessions.Session) {
		RequestToken := session.Values["RequestToken"].(dropbox.RequestToken)
		AccessToken, _ := dropbox.FinishAuth(AppToken, RequestToken)
		dbc := dbClient(AccessToken)
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

func ArticleHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Printf("ps = %+v\n", ps.ByName("articleslug"))
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("UserArticles")).Cursor()

		prefix := []byte(defaultUserEmail + ":article:")
		for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var a Article
			json.Unmarshal(v, &a)
			fmt.Fprint(w, a.Path)
		}

		return nil
	})
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("UserArticles")).Cursor()

		prefix := []byte(defaultUserEmail + ":article:")
		for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var a Article
			json.Unmarshal(v, &a)
			fmt.Fprint(w, a.Path)
		}

		return nil
	})
}

func Account(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	withSession(w, r, func(session *sessions.Session) {
		var AccessToken dropbox.AccessToken

		if email := session.Values["email"]; email == nil {
			fmt.Fprint(w, "no email found")
			return
		} else {
			email := session.Values["email"].(string)
			AccessToken, _ = datastore.LoadUserToken(email)
		}
		db := dbClient(AccessToken)
		info, err := db.GetAccountInfo()
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

func dbClient(t dropbox.AccessToken) *dropbox.Client {
	return &dropbox.Client{
		AppToken:    AppToken,
		AccessToken: t,
		Config: dropbox.Config{
			Access: dropbox.AppFolder,
			Locale: "us",
		}}
}

func withSession(w http.ResponseWriter, r *http.Request, fn func(*sessions.Session)) {
	gob.Register(dropbox.RequestToken{})
	store := sessions.NewCookieStore([]byte("182hetsgeih8765$aasdhj"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30 * 12,
		HttpOnly: true,
	}
	session, _ := store.Get(r, "godropblog")
	fn(session)
}
