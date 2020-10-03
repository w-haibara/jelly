package configure

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type Conf struct {
	Secrets Secrets
}

type Secrets struct {
	Signing_secret     string `json:"signing_secret"`
	Oauth_access_token string `json:"oauth_access_token"`
}

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

func (conf *Conf) GetSigningSecret() string {
	return conf.Secrets.Signing_secret
}

func (conf *Conf) GetOauthAccessToken() string {
	return conf.Secrets.Oauth_access_token
}
