package main

import (
	"github.com/slack-go/slack"
	"jelly/configure"
	"jelly/handler"
	"log"
	"net/http"
)

func main() {
	conf, err := configure.NewConf("./conf.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	client := handler.NewClient(
		slack.New(conf.GetOauthAccessToken()),
	)

	http.HandleFunc("/slack/events",
		client.GetEventsHandler(conf.GetSigningSecret()))
	http.HandleFunc("/slack/actions",
		client.GetActionsHandler(conf.GetSigningSecret()))

	log.Println("[INFO] Server listening")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
