package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/kubernetes/pkg/api"
)

const (
	PasswordSecret = "password"
	SecretLabel    = "openshift.io.oauth/github"
	SecretsURL     = "/api/v1/namespaces/%s/secrets"
	SecretURL      = "/api/v1/namespaces/%s/secrets/%s"
)

var (
	post Post = httpPost
	get  Get  = httpGet
	put  Put  = httpPUT
)

type secret interface {
	create(s *api.Secret, token string) error
	get(namespace, name string, token string) (*api.Secret, error)
	update(s *api.Secret, token string) error
}

type Get func(url string, credential ...string) ([]byte, error)

type Post func(url string, body []byte, credential ...string) ([]byte, error)

type Put func(url string, body []byte, credential ...string) ([]byte, error)

func (p Put) update(s *api.Secret, token string) error {
	body, err := json.Marshal(s)
	if err != nil {
		return err
	}

	apiURL := setSecretURLWithName(s.Namespace, s.Name)
	if _, err := p(apiURL, body, getTokenCredential(token)...); err != nil {
		return err
	}

	return nil
}

func (p Post) create(s *api.Secret, token string) error {
	body, err := json.Marshal(s)
	if err != nil {
		return err
	}

	apiURL := setSecretURL(s.Namespace)

	if _, err := p(apiURL, body, getTokenCredential(token)...); err != nil {
		return err

	}

	return nil
}

func (g Get) get(namespace, name string, token string) (*api.Secret, error) {
	apiURL := setSecretURLWithName(namespace, name)

	b, err := g(apiURL, getTokenCredential(token)...)
	if err != nil {
		return nil, err
	}

	secret := &api.Secret{}
	if err = json.Unmarshal(b, secret); err != nil {
		return nil, err
	}

	return secret, nil
}

func (o *secretOptions) NewBasicAuthSecret() *api.Secret {
	secret := &api.Secret{}
	secret.Kind = "Secret"
	secret.APIVersion = "v1"
	secret.Name = o.SecretName
	secret.Namespace = o.NameSpace
	secret.Type = api.SecretTypeOpaque

	secret.Labels = map[string]string{
		SecretLabel: o.SecretName,
	}

	secret.Data = map[string][]byte{
		PasswordSecret: []byte(o.GitHubToken),
	}

	return secret
}

func createSecret(o *secretOptions) error {
	secret := o.NewBasicAuthSecret()

	return post.create(secret, o.DatafactoryToken)
}

func getSecret(o *secretOptions) (*api.Secret, error) {
	return get.get(o.NameSpace, o.SecretName, o.DatafactoryToken)
}

func upsertSecret(option *secretOptions) error {
	secret, err := getSecret(option)
	if err != nil {
		if NotFount(err) {
			if err := createSecret(option); err != nil {
				return err
			}
			return nil
		}

		return err
	}

	if err := updateSecret(secret, option); err != nil {
		return err
	}

	return nil
}

func updateSecret(s *api.Secret, o *secretOptions) error {
	s.Data[PasswordSecret] = []byte(o.GitHubToken)
	return put.update(s, o.DatafactoryToken)
}

type secretOptions struct {
	NameSpace  string
	UserName   string
	SecretName string

	DatafactoryToken string
	GitHubToken      string
}

func (o *secretOptions) validate() error {
	if len(o.NameSpace) == 0 {
		return errors.New("secret option namespace is null")
	}

	if len(o.UserName) == 0 {
		return errors.New("secret option user is null")
	}

	if len(o.SecretName) == 0 {
		return errors.New("secret option secret name is null")
	}

	if len(o.DatafactoryToken) == 0 {
		return errors.New("secret option df token is null")
	}

	if len(o.GitHubToken) == 0 {
		return errors.New("secret option github token is null")
	}

	return nil
}

func setSecretURL(namespace string) string {
	return DFHost + fmt.Sprintf(SecretsURL, namespace)
}

func setSecretURLWithName(namespace string, name string) string {
	return DFHost + fmt.Sprintf(SecretURL, namespace, name)
}

func getUserKey(user, source string) string {
	return fmt.Sprintf("%s/%s/%s", EtcdUserRegistry, user, source)
}
