package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"jelly/configure"
	"jelly/deployer"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	confirmDeploymentAction = "confirm-deployment"
	releaseURL              = "https://github.com/w-haibara/portfolio/releases"
	gitBranch               = "master"
)

type secrets struct {
	signingSecret string
}

// Client is infomation of connection to API
type Client struct {
	API     *slack.Client
	secrets secrets
}

// InitClient is ...
func InitClient(path string, client *Client) error {
	var conf configure.Conf
	if bytes, err := ioutil.ReadFile(path); err == nil {
		if err = configure.NewConf(bytes, &conf); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("NewConf failed")
			return err
		}
	} else {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("ReadFile failed")
		return err
	}
	*client = Client{
		API: slack.New(conf.Secrets.OauthAccessToken),
		secrets: secrets{
			signingSecret: conf.Secrets.SigningSecret,
		},
	}
	return nil
}

// verify Request Verifier
func (client *Client) verify(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	verifier, err := slack.NewSecretsVerifier(r.Header, client.secrets.signingSecret)
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

// handleDeployCmd handle deploy command
func (client *Client) handleDeployCmd(event *slackevents.AppMentionEvent, arg string, w http.ResponseWriter) {
	msg := ""
	if arg == "" {
		arg = "latest"
		msg = fmt.Sprintf("The latest version will be deploy (%v)", releaseURL+"/latest")
	} else {
		msg = fmt.Sprintf("Commit: `%v` will be deploy (%v)", arg, releaseURL+"/tag/"+arg)
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

	if _, err := client.API.PostEphemeral(event.Channel, event.User, fallbackText, blocks); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("API PostEphemeral failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

//eventsWriteContent write content to w
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

// eventsHandler is http handler of events
func (client *Client) eventsHandler(body []byte, w http.ResponseWriter) {
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("ParseEvent failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Json Unmarshal failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(res.Challenge)); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Write bytes failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case slackevents.CallbackEvent:
		client.eventsWriteContent(&eventsAPIEvent, w)
	}
}

// handleConfirmDeploymentAction handle ConfirmDeployment action
func (client *Client) handleConfirmDeploymentAction(action *slack.BlockAction, payload *slack.InteractionCallback, w http.ResponseWriter) {
	arg := action.Value

	if arg == "" {
		cancelMsg := slack.MsgOptionText(
			fmt.Sprintf("<@%s> Canceled", payload.User.ID),
			false)
		if _, _, err := client.API.PostMessage(payload.Channel.ID, cancelMsg); err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("API PostMessage failed")
		}
	} else {
		go func() {
			startMsg := slack.MsgOptionText(
				fmt.Sprintf("<@%s> deploying commit: `%s`", payload.User.ID, arg),
				false)
			if _, _, err := client.API.PostMessage(payload.Channel.ID, startMsg); err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("API PostMessage failed")
			}

			resultMsg := ""
			if result, err := deployer.Deploy(arg); err == nil {
				resultMsg = fmt.Sprintf("deployment completed\n%v", result)
			} else {
				resultMsg = fmt.Sprintf("deployment failed!\n%v", err)
			}

			endMsg := slack.MsgOptionText(resultMsg, false)
			if _, _, err := client.API.PostMessage(payload.Channel.ID, endMsg); err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Error("API PostMessage failed")
			}
		}()
	}

	deleteOriginal := slack.MsgOptionDeleteOriginal(payload.ResponseURL)
	if _, _, _, err := client.API.SendMessage("", deleteOriginal); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("API SendMessage failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// actionsHandler is a http handler of action
func (client *Client) actionsHandler(w http.ResponseWriter, r *http.Request) {
	var payload *slack.InteractionCallback
	if err := json.Unmarshal([]byte(r.FormValue("payload")), &payload); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Json Unmarshal failed")
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

// GetEventsHandler returns a http handler of events
func (client *Client) GetEventsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL)
		body, err := client.verify(w, r)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Verify failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		client.eventsHandler(body, w)
	}
}

// GetActionsHandler returns a http handler of action
func (client *Client) GetActionsHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := client.verify(w, r)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("Verify failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		client.actionsHandler(w, r)
	}
}
