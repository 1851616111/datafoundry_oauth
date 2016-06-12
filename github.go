package main

import (
	"encoding/json"
	"errors"
	api "github.com/openshift/origin/pkg/user/api/v1"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	"strings"
)

func NewGitHub(id, secret, redirectUrl string, scopes []string) (Config, error) {
	github := Github{new(oauth2.Config)}
	if err := github.setClientId(id); err != nil {
		return nil, err
	}

	if err := github.setClientSecret(secret); err != nil {
		return nil, err
	}

	if err := github.setEndpoint(githuboauth.Endpoint); err != nil {
		return nil, err
	}

	if err := github.setRedirectURL(redirectUrl); err != nil {
		return nil, err
	}

	if err := github.setScope(scopes); err != nil {
		return nil, err
	}
	return &github, nil
}

type Github struct {
	*oauth2.Config
}

func (c *Github) setClientId(id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("invalid oauth2 Config client id")
	}

	c.ClientID = id
	return nil
}

func (c *Github) setClientSecret(secret string) error {
	if strings.TrimSpace(secret) == "" {
		return errors.New("invalid oauth2 Config client secret")
	}

	c.ClientSecret = secret
	return nil
}

func (c *Github) setEndpoint(endPoint oauth2.Endpoint) error {
	if strings.TrimSpace(endPoint.AuthURL) == "" || strings.TrimSpace(endPoint.TokenURL) == "" {
		return errors.New("invalid oauth2 Config endpoint")
	}

	c.Endpoint = endPoint
	return nil
}

func (c *Github) setRedirectURL(redirectUrl string) error {
	if strings.TrimSpace(redirectUrl) == "" {
		return errors.New("invalid oauth2 Config redirectl url")
	}

	c.RedirectURL = redirectUrl
	return nil
}

func (c *Github) setScope(scopes []string) error {
	c.Scopes = scopes
	return nil
}

func Auth(c Config, status string, opts ...oauth2.AuthCodeOption) string {
	return c.(*Github).AuthCodeURL(status, opts...)
}

func Exchange(code string) (*oauth2.Token, error) {
	return tokenConfig.(*Github).Exchange(oauth2.NoContext, code)
}

func getGithubInfo(user *api.User) (map[string]string, error) {
	key := getUserKey(user.Name, "github.com")
	info, err := db.getValue(key)
	if err != nil {
		return nil, err
	}

	userInfo := map[string]string{}
	if err := json.Unmarshal([]byte(info), &userInfo); err != nil {
		return nil, err
	}

	return userInfo, err
}
