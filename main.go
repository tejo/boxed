package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/tejo/boxed/datastore"
	"github.com/tejo/boxed/dropbox"
)

func main() {
	datastore.Connect("blog.db")
	defer datastore.Close()

	handleCommands()

	p := pat.New()
	p.Get("/sitemap.xml", sitemap)
	p.Get("/feed.atom", feed)
	p.Get("/login", login)
	p.Get("/archive", archive)
	p.Get(config.WebHookURL, webHook)
	p.Post(config.WebHookURL, webHook)
	p.Get("/account", account)
	p.Get("/{id}", articleHandler)
	p.Get(config.CallbackURL, callback)
	p.Get("/", index)

	n := negroni.Classic()
	n.Use(negroni.NewStatic(rice.MustFindBox("static").HTTPBox()))
	n.UseHandler(p)

	log.Fatal(http.ListenAndServe(config.Port, n))
}

//dropbox endpoint, must be configured on db site
func webHook(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintf(w, "%s", r.URL.Query().Get("challenge"))
		return
	}

	if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var d dropbox.DeltaPayLoad
		err := decoder.Decode(&d)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("processing %+v\n", d.Delta.Users)
		go processChanges(d.Delta.Users)
	}
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	index := datastore.LoadArticleIndex(config.DefaultUserEmail)
	var article *datastore.Article
	for _, v := range index {
		if v.Permalink == r.URL.Query().Get(":id") {
			article, _ = datastore.LoadArticle(v.ID)
			continue
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["article"].ExecuteTemplate(w, "layout", struct {
		Article *datastore.Article
		Index   []datastore.Article
	}{
		Article: article,
		Index:   index,
	})
}

func index(w http.ResponseWriter, r *http.Request) {
	index := datastore.LoadArticleIndex(config.DefaultUserEmail)
	var articles []*datastore.Article
	var i []datastore.Article
	if len(index) > 3 {
		i = index[:3]
	} else {
		i = index
	}
	for _, v := range i {
		a, _ := datastore.LoadArticle(v.ID)
		articles = append(articles, a)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["index"].ExecuteTemplate(w, "layout", struct {
		Articles []*datastore.Article
		Index    []datastore.Article
	}{
		Articles: articles,
		Index:    index,
	})
}

func archive(w http.ResponseWriter, r *http.Request) {
	index := datastore.LoadArticleIndex(config.DefaultUserEmail)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["archive"].ExecuteTemplate(w, "layout", struct {
		Index []datastore.Article
	}{
		Index: index,
	})
}

func sitemap(w http.ResponseWriter, r *http.Request) {
	index := datastore.LoadArticleIndex(config.DefaultUserEmail)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["sitemap.xml"].ExecuteTemplate(w, "T", struct {
		Host  string
		Index []datastore.Article
	}{
		Host:  config.HostWithProtocol,
		Index: index,
	})
}

func feed(w http.ResponseWriter, r *http.Request) {
	index := datastore.LoadArticleIndex(config.DefaultUserEmail)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := templates["feed.atom"].ExecuteTemplate(w, "T", struct {
		Host  string
		Index []datastore.Article
	}{
		Host:  config.HostWithProtocol,
		Index: index,
	})

	log.Println(err)
}

func withSession(w http.ResponseWriter, r *http.Request, fn func(*sessions.Session)) {
	gob.Register(dropbox.RequestToken{})
	store := sessions.NewCookieStore([]byte("182hetsgeih8765$aasdhj"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30 * 12,
		HttpOnly: true,
	}
	session, _ := store.Get(r, "boxedsession")
	fn(session)
}
