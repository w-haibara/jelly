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
	case "token":
		return os.Getenv("SLACK_BOT_TOKEN_DEPLOY")
	case "secret":
		return os.Getenv("SLACK_SIGNING_SECRET_DEPLOY")
	}
	return ""
}

func main() {
	client := handler.NewClient(
		slack.New(getSecret("token")),
	)

	http.HandleFunc("/slack/events", client.GetEventsHandler(getSecret("secret")))
	http.HandleFunc("/slack/actions", client.GetActionsHandler(getSecret("secret")))

	log.Println("[INFO] Server listening")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
