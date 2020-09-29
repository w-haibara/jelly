package main

import (
	"log"
	"net/http"
	"os"

	"deploy_bot/handler"
	"github.com/slack-go/slack"
)

func main() {
	client := &handler.Client{
		Api:    slack.New(os.Getenv("SLACK_BOT_TOKEN_DEPLOY")),
		Secret: os.Getenv("SLACK_SIGNING_SECRET_DEPLOY"),
	}

	http.HandleFunc("/slack/events", client.GetEventsHandler())
	http.HandleFunc("/slack/actions", client.GetActionsHandler())

	log.Println("[INFO] Server listening")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
