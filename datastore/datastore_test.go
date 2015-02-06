package datastore_test

import (
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/tejo/boxed/datastore"
	"github.com/tejo/boxed/dropbox"
)

func TestMain(m *testing.M) {
	datastore.Connect("test.db")
	exitVal := m.Run()
	os.Remove("test.db")
	os.Exit(exitVal)
}

func Test_Connect(t *testing.T) {
	a := assert.New(t)
	a.Equal(datastore.DB.Path(), "test.db")
}

func Test_CreateDefaultBuckets(t *testing.T) {
	a := assert.New(t)
	datastore.DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("UserData"))
		a.NotEqual(err, nil)
		_, err = tx.CreateBucket([]byte("UserData"))
		a.NotEqual(err, nil)
		return nil
	})
}

func Test_SaveUserData(t *testing.T) {
	a := assert.New(t)
	info := &dropbox.AccountInfo{
		Email:       "foo@bar.it",
		Uid:         1234,
		DisplayName: "pippo",
	}
	token := dropbox.AccessToken{Key: "a", Secret: "b"}
	datastore.SaveUserData(info, token)

	at, _ := datastore.LoadUserToken("foo@bar.it")
	a.Equal(at, token)

	userData, _ := datastore.LoadUserData("foo@bar.it")
	a.Equal(userData, info)
}

func Test_ParseArticle(t *testing.T) {
	a := assert.New(t)
	article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContent())
	a.Contains(article.Content, "my first article</h1>")
	a.Equal(article.Title, "this is my first article")
	a.Equal(article.Permalink, "this-is-my-first-article")
	a.Equal(article.CreatedAt, "2015-10-10")
	a.Equal(article.FileMetadata, fakeFileMetaData())
	article.GenerateID("foo@bar.it")
	a.Equal(article.ID, "foo@bar.it:article:/2015-10-10/this-is-my-first-article")
	a.Equal(article.TimeStamp, "1444435200")
}

func Test_ParseArticle_WithNoMetadata(t *testing.T) {
	a := assert.New(t)
	article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContentWithNoMetadata())
	a.Contains(article.Content, "my first article</h1>")
	a.Equal(article.Title, "this is my first article")
	a.Equal(article.Permalink, "this-is-my-first-article")
	a.Equal(article.CreatedAt, "2011-07-19")
	a.Equal(article.FileMetadata, fakeFileMetaData())
	article.GenerateID("foo@bar.it")
	a.Equal(article.ID, "foo@bar.it:article:/2011-07-19/this-is-my-first-article")
	a.Equal(article.TimeStamp, "1311033600")
}

func Test_SaveArticle(t *testing.T) {
	a := assert.New(t)
	func() { //I am a horrible peson
		article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContent())
		article.GenerateID("foo@bar.it")
		article.Save()
	}()
	article, _ := datastore.LoadArticle("foo@bar.it:article:/2015-10-10/this-is-my-first-article")
	a.Equal(article.Title, "this is my first article")
	a.Equal(article.Permalink, "this-is-my-first-article")
	a.Equal(article.CreatedAt, "2015-10-10")
	a.Equal(article.FileMetadata, fakeFileMetaData())
}

func Test_generateSlug(t *testing.T) {
	a := assert.New(t)
	article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContent())
	article.GenerateID("foo@bar.it")
	a.Equal(article.Slug, "/2015-10-10/this-is-my-first-article")
}

func fakeFileMetaData() dropbox.FileMetadata {
	return dropbox.FileMetadata{
		Path:     "/published/this_is_my-first article.md",
		IsDir:    false,
		Modified: "Tue, 19 Jul 2011 21:55:38 +0000",
	}
}

func fakeFileContent() []byte {
	b := `
<!--{
		"created-at": "2015-10-10",
		"permalink": "this-is-my-first-article",
		"title": "this is my first article"
}-->

# my first article
	`
	return []byte(b)
}

func fakeFileContentWithNoMetadata() []byte {
	b := `

# my first article
	`
	return []byte(b)
}
