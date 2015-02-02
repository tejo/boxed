package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
	a.ID = email + ":article:" + a.FileMetadata.Path
	a.generateSlug()
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
		fmt.Printf("err = %+v\n", err)
		return err
	})
	return err
}

func LoadArticle(ID string) (Article, error) {
	var article Article
	err := DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserArticles"))
		a := b.Get([]byte(ID))
		fmt.Printf("a = %+v\n", string(a))
		json.Unmarshal(a, &article)
		return nil
	})
	return article, err
}

func (a *Article) ParseTimeStamp() {
	test, err := time.Parse("2006-02-01", a.CreatedAt)
	if err == nil {
		a.TimeStamp = fmt.Sprintf("%d", test.Unix())
	} else {
		log.Printf("time parse error = %+v\n", err)
		a.TimeStamp = "0000000000"
	}
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
	article.ParseTimeStamp()
	return article
}

func extractEntryData(c []byte) *Article {
	var article Article
	start := bytes.Index(c, []byte("<!--"))
	end := bytes.Index(c, []byte("-->"))
	if start > -1 && end > -1 {
		err := json.Unmarshal(c[(start+4):end], &article)
		if err != nil {
			fmt.Println("error:", err)
		}
	}
	return &article
}
