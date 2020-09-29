package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Client struct {
	api    *slack.Client
	secret string
}

func (client *Client) HandleDeployCmd(event *slackevents.AppMentionEvent, command string, w http.ResponseWriter) {
	if _, _, err := client.api.PostMessage(event.Channel, slack.MsgOptionText("--- deploy ---", false)); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (client *Client) WriteContent(eventsAPIEvent *slackevents.EventsAPIEvent, w http.ResponseWriter) {
	innerEvent := eventsAPIEvent.InnerEvent
	switch event := innerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		message := strings.Split(event.Text, " ")
		if len(message) < 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		command := message[1]
		switch command {
		case "deploy":
			client.HandleDeployCmd(event, command, w)
		}
	}
}

func (client *Client) HandleEvent(body []byte, w http.ResponseWriter) {
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(res.Challenge)); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case slackevents.CallbackEvent:
		client.WriteContent(&eventsAPIEvent, w)
	}
}

func (client *Client) handler(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, client.secret)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := verifier.Ensure(); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	client.HandleEvent(body, w)
}

func (client *Client) getHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		client.handler(w, r)
	}
}
