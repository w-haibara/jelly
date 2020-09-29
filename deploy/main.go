package main

import (
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"
)

func main() {
	client := &Client{
		api:    slack.New(os.Getenv("SLACK_BOT_TOKEN_DEPLOY")),
		secret: os.Getenv("SLACK_SIGNING_SECRET_DEPLOY"),
	}
	http.HandleFunc("/slack/events", client.getHandler())

	log.Println("[INFO] Server listening")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
