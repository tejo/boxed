package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"text/template"

	"github.com/GeertJohan/go.rice"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/tejo/boxed/datastore"
	"github.com/tejo/boxed/dropbox"
)

var templates = map[string]*template.Template{
	"index":   template.Must(template.New("layout").ParseFiles("templates/layout.html", "templates/index.html")),
	"article": template.Must(template.New("layout").ParseFiles("templates/layout.html", "templates/article.html")),
	"archive": template.Must(template.New("layout").ParseFiles("templates/layout.html", "templates/archive.html")),
	"feed":    template.Must(template.ParseFiles("templates/feed.atom")),
	"sitemap": template.Must(template.ParseFiles("templates/sitemap.xml")),
}

func main() {
	datastore.Connect("blog.db")
	defer datastore.Close()

	handleCommands()

	p := pat.New()
	p.Get("/sitemap.xml", Sitemap)
	p.Get("/feed.atom", Feed)
	p.Get("/login", Login)
	p.Get("/archive", Archive)
	p.Get(config.WebHookURL, WebHook)
	p.Post(config.WebHookURL, WebHook)
	p.Get("/account", Account)
	p.Get("/{id}", ArticleHandler)
	p.Get(config.CallbackURL, Callback)
	p.Get("/", Index)

	n := negroni.Classic()
	n.Use(negroni.NewStatic(rice.MustFindBox("static").HTTPBox()))
	n.UseHandler(p)

	log.Fatal(http.ListenAndServe(config.Port, n))
}

func WebHook(w http.ResponseWriter, r *http.Request) {
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

func ArticleHandler(w http.ResponseWriter, r *http.Request) {
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

func Index(w http.ResponseWriter, r *http.Request) {
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

func Archive(w http.ResponseWriter, r *http.Request) {
	index := datastore.LoadArticleIndex(config.DefaultUserEmail)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["archive"].ExecuteTemplate(w, "layout", struct {
		Index []datastore.Article
	}{
		Index: index,
	})
}

func Account(w http.ResponseWriter, r *http.Request) {
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

func Login(w http.ResponseWriter, r *http.Request) {
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
func Callback(w http.ResponseWriter, r *http.Request) {
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

func Sitemap(w http.ResponseWriter, r *http.Request) {
	index := datastore.LoadArticleIndex(config.DefaultUserEmail)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["sitemap"].ExecuteTemplate(w, "sitemap.xml", struct {
		Host  string
		Index []datastore.Article
	}{
		Host:  config.HostWithProtocol,
		Index: index,
	})
}

func Feed(w http.ResponseWriter, r *http.Request) {
	index := datastore.LoadArticleIndex(config.DefaultUserEmail)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	templates["feed"].ExecuteTemplate(w, "feed.atom", struct {
		Host  string
		Index []datastore.Article
	}{
		Host:  config.HostWithProtocol,
		Index: index,
	})
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
