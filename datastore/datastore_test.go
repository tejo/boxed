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
}
