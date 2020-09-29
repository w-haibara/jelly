package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	selectVersionAction     = "select-version"
	confirmDeploymentAction = "confirm-deployment"
)

type Client struct {
	Api    *slack.Client
	Secret string
}

/*
 * Request Verifier
 */
func (client *Client) verify(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	verifier, err := slack.NewSecretsVerifier(r.Header, client.Secret)
	if err != nil {
		return nil, err
	}

	bodyReader := io.TeeReader(r.Body, &verifier)
	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}

	if err := verifier.Ensure(); err != nil {
		return nil, err
	}

	return body, nil
}

/*
 * Events Handler
 */
func (client *Client) handleDeployCmd(event *slackevents.AppMentionEvent, command string, w http.ResponseWriter) {
	text := slack.NewTextBlockObject(slack.MarkdownType, "Please select *version*.", false, false)
	textSection := slack.NewSectionBlock(text, nil, nil)

	versions := []string{"v1.0.0", "v1.1.0", "v1.1.1"}
	options := make([]*slack.OptionBlockObject, 0, len(versions))
	for _, v := range versions {
		optionText := slack.NewTextBlockObject(slack.PlainTextType, v, false, false)
		options = append(options, slack.NewOptionBlockObject(v, optionText))
	}

	placeholder := slack.NewTextBlockObject(slack.PlainTextType, "Select version", false, false)
	selectMenu := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, placeholder, "", options...)

	actionBlock := slack.NewActionBlock(selectVersionAction, selectMenu)
	fallbackText := slack.MsgOptionText("This client is not supported.", false)
	blocks := slack.MsgOptionBlocks(textSection, actionBlock)

	if _, err := client.Api.PostEphemeral(event.Channel, event.User, fallbackText, blocks); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (client *Client) eventsWriteContent(eventsAPIEvent *slackevents.EventsAPIEvent, w http.ResponseWriter) {
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
			client.handleDeployCmd(event, command, w)
		}
	}
}

func (client *Client) eventsHandler(body []byte, w http.ResponseWriter) {
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
		client.eventsWriteContent(&eventsAPIEvent, w)
	}
}

/*
 * Actions Handler
 */
func deploy(version string) {
	log.Printf("deploy %s", version)
}

func (client *Client) handleSelectVersionAction(action *slack.BlockAction, payload *slack.InteractionCallback, w http.ResponseWriter) {
	version := action.SelectedOption.Value

	text := slack.NewTextBlockObject(slack.MarkdownType,
		fmt.Sprintf("`%s` is about to deployed...", version), false, false)
	textSection := slack.NewSectionBlock(text, nil, nil)

	confirmButtonText := slack.NewTextBlockObject(slack.PlainTextType, "OK", false, false)
	confirmButton := slack.NewButtonBlockElement("", version, confirmButtonText)
	confirmButton.WithStyle(slack.StylePrimary)

	denyButtonText := slack.NewTextBlockObject(slack.PlainTextType, "Canccel", false, false)
	denyButton := slack.NewButtonBlockElement("", "deny", denyButtonText)
	denyButton.WithStyle(slack.StyleDanger)

	actionBlock := slack.NewActionBlock(confirmDeploymentAction, confirmButton, denyButton)

	fallbackText := slack.MsgOptionText("This client is not supported.", false)
	blocks := slack.MsgOptionBlocks(textSection, actionBlock)

	replaceOriginal := slack.MsgOptionReplaceOriginal(payload.ResponseURL)
	if _, _, _, err := client.Api.SendMessage("", replaceOriginal, fallbackText, blocks); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (client *Client) handleConfirmDeploymentAction(action *slack.BlockAction, payload *slack.InteractionCallback, w http.ResponseWriter) {
	if strings.HasPrefix(action.Value, "v") {
		version := action.Value
		go func() {
			startMsg := slack.MsgOptionText(
				fmt.Sprintf("<@%s> deploying `%s`.", payload.User.ID, version), false)
			if _, _, err := client.Api.PostMessage(payload.Channel.ID, startMsg); err != nil {
				log.Println(err)
			}

			deploy(version)

			endMsg := slack.MsgOptionText(
				fmt.Sprintf("`%s` deployment completed!", version), false)
			if _, _, err := client.Api.PostMessage(payload.Channel.ID, endMsg); err != nil {
				log.Println(err)
			}
		}()
	}

	deleteOriginal := slack.MsgOptionDeleteOriginal(payload.ResponseURL)
	if _, _, _, err := client.Api.SendMessage("", deleteOriginal); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (client *Client) actionsHandler(w http.ResponseWriter, r *http.Request) {
	var payload *slack.InteractionCallback
	if err := json.Unmarshal([]byte(r.FormValue("payload")), &payload); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch payload.Type {
	case slack.InteractionTypeBlockActions:
		if len(payload.ActionCallback.BlockActions) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		action := payload.ActionCallback.BlockActions[0]
		switch action.BlockID {
		case selectVersionAction:
			client.handleSelectVersionAction(action, payload, w)
		case confirmDeploymentAction:
			client.handleConfirmDeploymentAction(action, payload, w)
		}
	}
}

/*
 * Handler Getters
 */
func (client *Client) GetEventsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := client.verify(w, r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		client.eventsHandler(body, w)
	}
}

func (client *Client) GetActionsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := client.verify(w, r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		client.actionsHandler(w, r)
	}
}
