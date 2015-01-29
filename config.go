package main

import (
	"os"

	"github.com/tejo/boxed/dropbox"
)

type Config struct {
	DefaultUserEmail string
	CallbackUrl      string
	Port             string
	AppToken         dropbox.AppToken
}

var config *Config

func init() {
	// initialize some config var
	config = &Config{}

	config.DefaultUserEmail = os.Getenv("DEFAULT_USER_EMAIL")
	config.AppToken = dropbox.AppToken{
		Key:    os.Getenv("KEY"),
		Secret: os.Getenv("SECRET"),
	}

	if os.Getenv("CALLBACK_URL") != "" {
		config.CallbackUrl = os.Getenv("CALLBACK_URL")
	} else {
		config.CallbackUrl = "http://localhost:8080/oauth/callback"
	}

	if os.Getenv("PORT") != "" {
		config.Port = ":" + os.Getenv("PORT")
	} else {
		config.Port = ":8080"
	}

}
