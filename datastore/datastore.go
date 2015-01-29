package datastore

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/tejo/boxed/dropbox"
)

var DB *bolt.DB

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
