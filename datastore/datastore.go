package datastore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
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
		b.Put([]byte(a.ID), []byte(article))
		b.Put([]byte(a.Path), []byte(a.ID))
		return err
	})
	return err
}

func (a *Article) Delete() error {
	var err error
	err = DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserArticles"))
		b.Delete([]byte(a.ID))
		b.Delete([]byte(a.Path))
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

		// saves user token
		t, _ := json.Marshal(token)
		err = b.Put([]byte(info.Email+":token"), []byte(t))

		// saves user data, use email as key
		i, _ := json.Marshal(info)
		err = b.Put([]byte(info.Email), []byte(i))

		// saves mapping for uid : email
		err = b.Put([]byte(strconv.Itoa(int(info.Uid))), []byte(info.Email))
		return err
	})
	return err
}

func SaveCurrentCursor(email string, delta *dropbox.Delta) error {
	var err error
	err = DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserData"))

		// saves user current cursor
		err = b.Put([]byte(email+":current_cursor"), []byte(delta.Cursor))
		return err
	})
	return err
}

func GetCurrenCursorByEmail(email string) (string, error) {
	var cursor string
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserData"))
		cursor = string(b.Get([]byte(email + ":current_cursor")))
		return nil
	})
	if cursor == "" {
		return cursor, errors.New("cursor not found with the provided email")
	}
	return cursor, nil
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

func LoadUserTokenByUID(uid int) (dropbox.AccessToken, error) {
	var email string
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserData"))
		email = string(b.Get([]byte(strconv.Itoa(uid))))
		return nil
	})

	return LoadUserToken(email)
}

func GetUserEmailByUID(uid int) (string, error) {
	var email string
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserData"))
		email = string(b.Get([]byte(strconv.Itoa(uid))))
		return nil
	})
	if email == "" {
		return email, errors.New("email not found with the provided uid")
	}
	return email, nil
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
	unsafe := blackfriday.MarkdownCommon(c)
	article.Content = string(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
	article.FileMetadata = e
	article.sanitizeArticleMetadata()
	article.ParseTimeStamp()
	return article
}

// load user article ids, sorted by created at
func LoadArticleIDs(email string) []string {
	ids, sortedIDs := []string{}, []string{}
	DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("UserArticles")).Cursor()
		prefix := []byte(email + ":article:")
		for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			ids = append(ids, string(k))
		}
		return nil
	})

	sort.Strings(ids)

	for i := len(ids) - 1; i >= 0; i-- {
		sortedIDs = append(sortedIDs, ids[i])
	}
	return sortedIDs
}

func LoadArticle(ID string) (*Article, error) {
	var a Article
	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("UserArticles"))
		v := b.Get([]byte(ID))
		json.Unmarshal(v, &a)
		return nil
	})
	if a.ID == "" {
		return &a, errors.New("article not found")
	}
	return &a, nil
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

func Delete(bucket []byte, key string) error {
	var err error
	err = DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		b.Delete([]byte(key))
		return err
	})
	return err
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
