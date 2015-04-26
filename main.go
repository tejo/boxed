package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/pat"
	"github.com/tejo/boxed/datastore"
	"github.com/tejo/boxed/dropbox"
)

func main() {
	datastore.Connect("blog.db")
	defer datastore.Close()

	handleCommands()
	loadTemplates()

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
	p.Get("/", home)

	n := negroni.Classic()
	n.Use(negroni.NewStatic(rice.MustFindBox("static").HTTPBox()))
	n.UseHandler(p)

	log.Fatal(http.ListenAndServe(config.Port, n))
}

// dropbox endpoint, must be configured on db site
// when a file is updated in the db folder, this
// endpoint will be hit by db
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
	var article *datastore.Article
	for _, v := range datastore.LoadArticleIndex(config.DefaultUserEmail) {
		if v.Permalink == r.URL.Query().Get(":id") {
			article, _ = datastore.LoadArticle(v.ID)
			continue
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["article"].ExecuteTemplate(w, "layout",
		map[string]interface{}{
			"SiteName": config.SiteName,
			"Article":  article,
			"CssClass": "article",
		})
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["index"].ExecuteTemplate(w, "layout",
		map[string]interface{}{
			"SiteName": config.SiteName,
			"Articles": getLatestArticles(),
			"CssClass": "home",
		})
}

func archive(w http.ResponseWriter, r *http.Request) {
	articles := datastore.LoadArticleIndex(config.DefaultUserEmail)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["archive"].ExecuteTemplate(w, "layout",
		map[string]interface{}{
			"SiteName": config.SiteName,
			"Index":    articles,
			"CssClass": "archive",
		})
}

func sitemap(w http.ResponseWriter, r *http.Request) {
	articles := datastore.LoadArticleIndex(config.DefaultUserEmail)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["sitemap.xml"].ExecuteTemplate(w, "T",
		map[string]interface{}{
			"Host":  config.HostWithProtocol,
			"Index": articles,
		})
}

func feed(w http.ResponseWriter, r *http.Request) {
	articles := datastore.LoadArticleIndex(config.DefaultUserEmail)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["feed.atom"].ExecuteTemplate(w, "T",
		map[string]interface{}{
			"SiteName": config.SiteName,
			"Host":     config.HostWithProtocol,
			"Index":    articles,
		})
}

func getLatestArticles() []*datastore.Article {
	articleIndex := datastore.LoadArticleIndex(config.DefaultUserEmail)
	var articles []*datastore.Article
	var firstThreeArticles []datastore.Article
	if len(articleIndex) > 3 {
		firstThreeArticles = articleIndex[:3]
	} else {
		firstThreeArticles = articleIndex
	}
	for _, v := range firstThreeArticles {
		a, _ := datastore.LoadArticle(v.ID)
		articles = append(articles, a)
	}
	return articles
}
