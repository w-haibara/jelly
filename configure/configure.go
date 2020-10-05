package configure

import (
	"encoding/json"
	"io/ioutil"
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
func NewConf(path string) (Conf, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return Conf{}, err
	}
	var conf Conf
	if err := json.Unmarshal(bytes, &conf); err != nil {
		log.Fatal(err)
		return Conf{}, err
	}
	return conf, nil
}

// GetSigningSecret is returns a Signing Secret
func (conf *Conf) GetSigningSecret() string {
	return conf.Secrets.SigningSecret
}

// GetOauthAccessToken is returns a Oauth Access Token
func (conf *Conf) GetOauthAccessToken() string {
	return conf.Secrets.OauthAccessToken
}
