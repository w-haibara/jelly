package main

import (
	"io/ioutil"
	"jelly/configure"
	"jelly/handler"
	"log"
	"net/http"

	"github.com/slack-go/slack"
)

func main() {
	path := "./conf.json"

	var conf configure.Conf
	if bytes, err := ioutil.ReadFile(path); err == nil {
		if err = configure.NewConf(bytes, conf); err != nil {
			log.Fatal(err)
			return
		}
	} else {
		log.Fatal(err)
		return
	}

	client := handler.NewClient(slack.New(conf.Secrets.OauthAccessToken))

	http.HandleFunc("/slack/events",
		client.GetEventsHandler(conf.Secrets.SigningSecret))
	http.HandleFunc("/slack/actions",
		client.GetActionsHandler(conf.Secrets.SigningSecret))

	log.Println("[INFO] Server listening")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
