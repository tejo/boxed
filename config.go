package main

import (
	"os"

	"github.com/tejo/boxed/dropbox"
)

//config struct used for configure boxed
type Config struct {
	DefaultUserEmail string
	HostWithProtocol string
	CallbackURL      string
	WebHookURL       string
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

	if os.Getenv("HOST_WITH_PROTOCOL") != "" {
		config.HostWithProtocol = os.Getenv("HOST_WITH_PROTOCOL")
	} else {
		config.HostWithProtocol = "http://localhost:8080"
	}

	if os.Getenv("WEBHOOK_URL") != "" {
		config.WebHookURL = os.Getenv("WEBHOOK_URL")
	} else {
		config.WebHookURL = "/webhook"
	}

	if os.Getenv("CALLBACK_URL") != "" {
		config.CallbackURL = os.Getenv("CALLBACK_URL")
	} else {
		config.CallbackURL = "/oauth/callback"
	}

	if os.Getenv("PORT") != "" {
		config.Port = ":" + os.Getenv("PORT")
	} else {
		config.Port = ":8080"
	}

}
