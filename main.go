package main

import (
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"jelly/handler"
)

func getSecret(v string) string {
	switch v {
	case "bot_token":
		return os.Getenv("SLACK_BOT_TOKEN_DEPLOY")
	case "signing_secret":
		return os.Getenv("SLACK_SIGNING_SECRET_DEPLOY")
	}
	return ""
}

func main() {
	client := handler.NewClient(
		slack.New(getSecret("bot_token")),
	)

	http.HandleFunc("/slack/events", client.GetEventsHandler(getSecret("signing_secret")))
	http.HandleFunc("/slack/actions", client.GetActionsHandler(getSecret("signing_secret")))

	log.Println("[INFO] Server listening")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
