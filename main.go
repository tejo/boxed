package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/tejo/dropbox"
)

var AppToken = dropbox.AppToken{
	Key:    "2vhv4i5dqyl92l1",
	Secret: "0k1q9zpbt1x3czk",
}

var callbackUrl = "http://localhost:8080/oauth/callback"
var db *bolt.DB

func withSession(w http.ResponseWriter, r *http.Request, fn func(*sessions.Session)) {
	store := sessions.NewCookieStore([]byte("182hetsgeih8765$aasdhj"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30 * 12,
		HttpOnly: true,
	}
	session, _ := store.Get(r, "godropblog")
	fn(session)
}

func init() {
	var err error
	db, err = bolt.Open("blog.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("UserData"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	gob.Register(dropbox.RequestToken{})
}

func main() {
	defer db.Close()

	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/login", Login)
	router.GET("/oauth/callback", Callback)

	log.Fatal(http.ListenAndServe(":8080", router))
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

func Callback(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	withSession(w, r, func(session *sessions.Session) {
		RequestToken := session.Values["RequestToken"].(dropbox.RequestToken)
		AccessToken, _ := dropbox.FinishAuth(AppToken, RequestToken)
		info, err := dbClient(AccessToken).GetAccountInfo()
		if err != nil {
			log.Println(err)
		}
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("UserData"))
			t, _ := json.Marshal(AccessToken)
			uid := strconv.Itoa(int(info.Uid))
			err := b.Put([]byte(uid+":token"), []byte(t))
			i, _ := json.Marshal(info)
			err = b.Put([]byte(uid), []byte(i))
			return err
		})
		session.Values["uid"] = info.Uid
		session.Save(r, w)
		fmt.Printf("AccessToken = %+v\n", AccessToken)
		http.Redirect(w, r, "/", 302)
	})
}

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	withSession(w, r, func(session *sessions.Session) {
		var AccessToken dropbox.AccessToken

		if uid := session.Values["uid"]; uid == nil {
			log.Println("no uid found")
			return
		} else {
			uid := strconv.Itoa(int(session.Values["uid"].(uint64)))
			db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("UserData"))
				token := b.Get([]byte(uid + ":token"))
				json.Unmarshal(token, &AccessToken)
				fmt.Printf("The answer is: %s\n", AccessToken)
				return nil
			})
		}
		db := dbClient(AccessToken)
		info, err := db.GetAccountInfo()
		fmt.Printf("err = %+v\n", err)

		if err != nil {
			//access token is not valid anymore
			fmt.Fprintf(w, " %+v\n", err)
			// reset all session
			session.Values["key"], session.Values["secret"] = "", ""
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
