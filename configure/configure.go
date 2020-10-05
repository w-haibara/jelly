package configure

import (
	"encoding/json"
	"log"
)

// Conf is configuration of app
type Conf struct {
	Secrets Secrets
}

// Secrets is a configuration of secrets
type Secrets struct {
	SigningSecret    string `json:"signing_secret"`
	OauthAccessToken string `json:"oauth_access_token"`
}

// NewConf returns a new Conf
func NewConf(bytes []byte, conf Conf) error {
	if err := json.Unmarshal(bytes, &conf); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
