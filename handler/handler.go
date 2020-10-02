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
	"jelly/deployer"
)

const (
	confirmDeploymentAction = "confirm-deployment"
	releaseUrl              = "https://github.com/w-haibara/portfolio/releases"
	gitBranch               = "master"
)

type Client struct {
	Api    *slack.Client
	secret string
}

/*
 * Create new Client
 */
func NewClient(api *slack.Client) *Client {
	return &Client{
		Api: api,
	}
}

/*
 * Request Verifier
 */
func (client *Client) verify(secret string, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	verifier, err := slack.NewSecretsVerifier(r.Header, secret)
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
func (client *Client) handleDeployCmd(event *slackevents.AppMentionEvent, arg string, w http.ResponseWriter) {
	msg := ""
	if arg == "" {
		arg = "latest"
		msg = fmt.Sprintf("The latest version will be deploy (%v)", releaseUrl+"/latest")
	} else {
		msg = fmt.Sprintf("Commit: `%v` will be deploy (%v)", arg, releaseUrl+"/tag/"+arg)
	}
	msg += "\nDo you want to continue?"
	text := slack.NewTextBlockObject(slack.MarkdownType, msg, false, false)
	textSection := slack.NewSectionBlock(text, nil, nil)

	confirmButtonText := slack.NewTextBlockObject(slack.PlainTextType, "OK", false, false)
	confirmButton := slack.NewButtonBlockElement("", arg, confirmButtonText)
	confirmButton.WithStyle(slack.StylePrimary)

	denyButtonText := slack.NewTextBlockObject(slack.PlainTextType, "Cancel", false, false)
	denyButton := slack.NewButtonBlockElement("", "", denyButtonText)
	denyButton.WithStyle(slack.StyleDanger)

	actionBlock := slack.NewActionBlock(confirmDeploymentAction, confirmButton, denyButton)

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
		args := strings.Split(event.Text, " ")
		if len(args) < 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		command := args[1]
		switch command {
		case "deploy":
			if len(args) >= 3 {
				client.handleDeployCmd(event, args[2], w)
			} else {
				client.handleDeployCmd(event, "", w)
			}
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
func (client *Client) handleConfirmDeploymentAction(action *slack.BlockAction, payload *slack.InteractionCallback, w http.ResponseWriter) {
	arg := action.Value

	if arg == "" {
		cancelMsg := slack.MsgOptionText(
			fmt.Sprintf("<@%s> Canceled", payload.User.ID),
			false)
		if _, _, err := client.Api.PostMessage(payload.Channel.ID, cancelMsg); err != nil {
			log.Println(err)
		}
	} else {
		go func() {
			startMsg := slack.MsgOptionText(
				fmt.Sprintf("<@%s> deploying commit: `%s`", payload.User.ID, arg),
				false)
			if _, _, err := client.Api.PostMessage(payload.Channel.ID, startMsg); err != nil {
				log.Println(err)
			}

			resultMsg := ""
			if result, err := deployer.Deploy(arg); err == nil {
				resultMsg = fmt.Sprintf("deployment completed\n%v", result)
			} else {
				resultMsg = fmt.Sprintf("deployment failed!\n%v", err)
			}

			endMsg := slack.MsgOptionText(resultMsg, false)
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
		case confirmDeploymentAction:
			client.handleConfirmDeploymentAction(action, payload, w)
		}
	}
}

/*
 * Handler Getters
 */
func (client *Client) GetEventsHandler(secret string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := client.verify(secret, w, r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		client.eventsHandler(body, w)
	}
}

func (client *Client) GetActionsHandler(secret string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := client.verify(secret, w, r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		client.actionsHandler(w, r)
	}
}
