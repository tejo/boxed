package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/shurcooL/go/github_flavored_markdown"
	"github.com/tejo/boxed/dropbox"
)

var DB *bolt.DB

type Article struct {
	ID        string
	Content   string
	Title     string `json:"title"`
	CreatedAt string `json:"created-at"`
	TimeStamp string `json:"timestamp"`
	Permalink string `json:"permalink"`
	Slug      string `json:"slug"`
	dropbox.FileMetadata
}

func (a *Article) GenerateID(email string) {
	a.generateSlug()
	a.ID = email + ":article:" + a.Slug
}

func (a *Article) generateSlug() {
	a.Slug = "/" + a.CreatedAt + "/" + a.Permalink
}

func (a *Article) Save() error {
	var err error
	err = DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserArticles"))
		article, err := json.Marshal(a)
		err = b.Put([]byte(a.ID), []byte(article))
		return err
	})
	return err
}

func (a *Article) sanitizeArticleMetadata() {
	if a.Permalink == "" {
		a.Permalink = a.Path
		for _, v := range [][]string{
			{" ", "-"},
			{"_", "-"},
			{"/published/", ""},
			{".md", ""}} {
			a.Permalink = strings.Replace(a.Permalink, v[0], v[1], -1)
		}
	}

	if a.Title == "" {
		a.Title = strings.Replace(a.Permalink, "-", " ", -1)
	}

	if a.CreatedAt == "" {
		t, _ := time.Parse(dropbox.TimeFormat, a.Modified)
		a.CreatedAt = fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
	}

}

func (a *Article) ParseTimeStamp() {
	t, err := time.Parse("2006-01-02", a.CreatedAt)
	if err != nil {
		log.Printf("%s for post %s\n", err, a.Path)
		return
	}
	a.TimeStamp = fmt.Sprintf("%d", t.Unix())
}

func LoadArticle(ID string) (Article, error) {
	var article Article
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserArticles"))
		a := b.Get([]byte(ID))
		json.Unmarshal(a, &article)
		return nil
	})
	return article, err
}

func Connect(dbname string) error {
	var err error
	DB, err = bolt.Open(dbname, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	//create buckets
	err = DB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("UserData")); err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		if _, err = tx.CreateBucketIfNotExists([]byte("UserArticles")); err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return err
	})
	return err
}

func Close() {
	DB.Close()
}

func SaveUserData(info *dropbox.AccountInfo, token dropbox.AccessToken) error {
	var err error
	err = DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserData"))
		t, _ := json.Marshal(token)
		err = b.Put([]byte(info.Email+":token"), []byte(t))
		i, _ := json.Marshal(info)
		err = b.Put([]byte(info.Email), []byte(i))
		return err
	})
	return err
}

func LoadUserToken(email string) (dropbox.AccessToken, error) {
	var AccessToken dropbox.AccessToken
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserData"))
		token := b.Get([]byte(email + ":token"))
		json.Unmarshal(token, &AccessToken)
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return AccessToken, err
}

func LoadUserData(email string) (*dropbox.AccountInfo, error) {
	var AccountInfo *dropbox.AccountInfo
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserData"))
		info := b.Get([]byte(email))
		json.Unmarshal(info, &AccountInfo)
		return nil
	})
	return AccountInfo, err
}

func ParseEntry(e dropbox.FileMetadata, c []byte) *Article {
	article := extractEntryData(c)
	article.Content = string(github_flavored_markdown.Markdown(c))
	article.FileMetadata = e
	article.sanitizeArticleMetadata()
	article.ParseTimeStamp()
	return article
}

// load user article keys, sorted by created at
func LoadArticleKeys(email string) []string {
	keys, sortedKeys := []string{}, []string{}
	DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("UserArticles")).Cursor()
		prefix := []byte(email + ":article:")
		for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			keys = append(keys, string(k))
		}
		return nil
	})

	sort.Strings(keys)

	for i := len(keys) - 1; i >= 0; i-- {
		sortedKeys = append(sortedKeys, keys[i])
	}
	return sortedKeys
}

func DeleteArtilcles(email string) {
	DB.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("UserArticles")).Cursor()
		prefix := []byte(email + ":article:")
		for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			c.Delete()
		}
		return nil
	})
}

func extractEntryData(c []byte) *Article {
	var article Article
	start := bytes.Index(c, []byte("<!--"))
	end := bytes.Index(c, []byte("-->"))
	// article metadata found
	if start > -1 && end > -1 {
		err := json.Unmarshal(c[(start+4):end], &article)
		if err != nil {
			fmt.Println("error:", err)
		}
	}
	return &article
}
