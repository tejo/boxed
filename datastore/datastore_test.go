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
		_, err = tx.CreateBucket([]byte("UserArticles"))
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

	at, _ = datastore.LoadUserTokenByUID(1234)
	a.Equal(at, token)

	userData, _ := datastore.LoadUserData("foo@bar.it")
	a.Equal(userData, info)
}

func Test_SaveCurrentCursor(t *testing.T) {
	a := assert.New(t)
	datastore.SaveCurrentCursor("foo@bar.it", "foobar")
	c, err := datastore.GetCurrenCursorByEmail("foo@bar.it")
	a.Equal(c, "foobar")
	a.Equal(err, nil)
}

func Test_GetUserEmailByUID(t *testing.T) {
	a := assert.New(t)
	info := &dropbox.AccountInfo{
		Email:       "foo@bar.it",
		Uid:         1234,
		DisplayName: "pippo",
	}
	token := dropbox.AccessToken{Key: "a", Secret: "b"}
	datastore.SaveUserData(info, token)

	email, _ := datastore.GetUserEmailByUID(1234)
	a.Equal(email, "foo@bar.it")

	_, err := datastore.GetUserEmailByUID(1)
	a.NotEqual(err, nil)
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
	a.Equal(article.ID, "foo@bar.it:article:/published/this_is_my-first article.md")
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
	a.Equal(article.TimeStamp, "1311033600")
}

func Test_Article_Save(t *testing.T) {
	a := assert.New(t)
	func() { //I am a horrible person
		article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContent())
		article.GenerateID("foo@bar.it")
		article.Save()
	}()
	article, _ := datastore.LoadArticle("foo@bar.it:article:/published/this_is_my-first article.md")
	a.Equal(article.Title, "this is my first article")
	a.Equal(article.Permalink, "this-is-my-first-article")
	a.Equal(article.CreatedAt, "2015-10-10")
	a.Equal(article.FileMetadata, fakeFileMetaData())
}

func Test_Article_Delete(t *testing.T) {
	a := assert.New(t)
	article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContent())
	article.GenerateID("foo@bar.it")
	article.Save()

	article.Delete()

	_, err := datastore.LoadArticle(article.ID)
	a.NotEqual(err, nil)
}

func Test_Datastore_Bucket_Delete(t *testing.T) {
	a := assert.New(t)
	article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContent())
	article.GenerateID("foo@bar.it")
	article.Save()

	datastore.Delete([]byte("UserArticles"), article.ID)

	_, err := datastore.LoadArticle(article.ID)
	a.NotEqual(err, nil)
}

func Test_LoadArticle(t *testing.T) {
	a := assert.New(t)
	article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContent())
	article.GenerateID("foo@bar.it")
	article.Save()

	//load article by ID (key)
	loadedArticle, _ := datastore.LoadArticle(article.ID)
	a.Equal(article, loadedArticle)

	//test article not found
	_, err := datastore.LoadArticle("foo")
	a.NotEqual(err, nil)
}

func Test_DeleteArticles(t *testing.T) {
	a := assert.New(t)
	article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContentWithNoMetadata())
	article.GenerateID("foo@bar.it")
	article.Save()
	datastore.DeleteArtilcles("foo@bar.it")
	a.Equal(len(datastore.LoadArticleIDs("foo@bar.it")), 0)
}

func Test_LoadArticleIndex(t *testing.T) {
	a := assert.New(t)

	datastore.DeleteArtilcles("foo@bar.it")

	article := datastore.ParseEntry(fakeFileMetaData(), fakeFileContentWithNoMetadata())
	article.Path = "p1"
	article.Permalink = "a1"
	article.CreatedAt = "2014-12-01"
	article.GenerateID("foo@bar.it")
	article.Save()
	anotherArticle := datastore.ParseEntry(fakeFileMetaData(), fakeFileContentWithNoMetadata())
	article.Path = "p2"
	anotherArticle.Permalink = "a2"
	anotherArticle.Title = "t2"
	anotherArticle.CreatedAt = "2014-12-02"
	anotherArticle.GenerateID("foo@bar.it")
	anotherArticle.Save()

	datastore.ArticlesReindex("foo@bar.it")
	index := datastore.LoadArticleIndex("foo@bar.it")

	a.Contains(index[0][0], "a2")
	a.Contains(index[0][1], anotherArticle.ID)
	a.Contains(index[0][2], "t2")
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
