package main

import (
	"log"
	"text/template"

	"github.com/GeertJohan/go.rice"
)

var templates = map[string]*template.Template{}

func loadTemplates() {
	templateBox := rice.MustFindBox("templates")
	layout, err := templateBox.String("layout.html")
	if err != nil {
		log.Fatal(err)
	}

	t := []string{
		"index",
		"article",
		"archive",
	}

	for _, tplName := range t {
		tplContent, err := templateBox.String(tplName + ".html")
		if err != nil {
			log.Fatal(err)
		}
		templates[tplName] = template.Must(template.New("layout").Parse(layout + tplContent))
	}

	feed, err := templateBox.String("feed.atom")
	if err != nil {
		log.Fatal(err)
	}

	templates["feed.atom"] = template.Must(template.New("feed.atom").Parse(feed))

	sitemap, err := templateBox.String("sitemap.xml")
	if err != nil {
		log.Fatal(err)
	}

	templates["sitemap.xml"] = template.Must(template.New("sitemap.xml").Parse(sitemap))
}
