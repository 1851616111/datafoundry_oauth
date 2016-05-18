package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"k8s.io/kubernetes/pkg/api"
	"strings"
)

const (
	PasswordSecret         = "password"
	GithubSecretLabelKey   = "openshift.io.oauth/github"
	GithubSecretLabelValue = "github"
	GitLabSecretLabel      = "openshift.io.oauth/gitlab"
	SecretsURL             = "/api/v1/namespaces/%s/secrets"
	SecretURL              = "/api/v1/namespaces/%s/secrets/%s"
	DFParamLabel           = "labelSelector"
)

var (
	post Post = httpPost
	get  Get  = httpGet
	put  Put  = httpPUT

	GithubSecretLabel = &Label{
		key:      GithubSecretLabelKey,
		operator: "=",
		value:    GithubSecretLabelValue,
	}
)

type secret interface {
	create(s *api.Secret, token string) error
	get(namespace, name string, token string) (*api.Secret, error)
	list(namespace, option GetOption, token string) (*api.SecretList, error)
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

	if _, err := p(apiURL, body, "Authorization", fmt.Sprintf("Bearer %s", token)); err != nil {
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

	if _, err := p(apiURL, body, "Authorization", fmt.Sprintf("Bearer %s", token)); err != nil {
		return err

	}

	return nil
}

type Label struct {
	key      string
	value    string
	operator string
}

type GetOption struct {
	Label *Label
	Name  string
}

func (l *Label) getLabel() string {
	return DFParamLabel + "=" + l.key + l.operator + l.value
}

func getUrl(namespace string, option GetOption) (string, error) {
	if option.Name == "" && option.Label == nil {
		return "", errors.New("param option must not all nil.")
	}

	apiURL := ""
	if option.Label != nil {
		apiURL = setSecretURL(namespace) + "?" + option.Label.getLabel()
	} else {
		apiURL = setSecretURLWithName(namespace, option.Name)
	}

	return apiURL, nil
}

func (g Get) get(namespace, name string, token string) (*api.Secret, error) {
	apiURL := setSecretURLWithName(namespace, name)
	b, err := g(apiURL, "Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return nil, err
	}

	secret := &api.Secret{}
	if err = json.Unmarshal(b, secret); err != nil {
		return nil, err
	}

	return secret, nil
}

//labelSelector
func (g Get) list(namespace string, option GetOption, token string) (*api.SecretList, error) {

	var url string
	var err error
	if url, err = getUrl(namespace, option); err != nil {
		return nil, err
	}
	b, err := g(url, "Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		return nil, err
	}
	secrets := &api.SecretList{}
	if err = json.Unmarshal(b, secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}

func (o *SecretTokenOptions) NewSecret() *api.Secret {
	secret := &api.Secret{}
	secret.Kind = "Secret"
	secret.APIVersion = "v1"
	secret.Name = o.SecretName
	secret.Namespace = o.NameSpace
	secret.Type = api.SecretTypeOpaque

	secret.Labels = map[string]string{
		GithubSecretLabelKey: GithubSecretLabelValue,
	}

	secret.Data = map[string][]byte{
		PasswordSecret: []byte(o.GitHubToken),
	}

	return secret
}

func (o *SecretSSHOptions) NewSecret() *api.Secret {
	secret := &api.Secret{}
	secret.Kind = "Secret"
	secret.APIVersion = "v1"
	secret.Name = o.SecretName
	secret.Namespace = o.NameSpace
	secret.Type = api.SecretTypeOpaque

	secret.Labels = map[string]string{
		GitLabSecretLabel: o.SecretName,
	}

	secret.Data = map[string][]byte{
		"ssh-privatekey": []byte(o.PrivateKey),
	}

	return secret
}

type SecretOption interface {
	NewSecret() *api.Secret
	GetDFToken() string
	GetDFNamespace() string
	GetSecretName() string

	IsSSHSecret() bool
	GetPrivateKey() string

	IsTokenSecret() bool
	GetToken() string

	Validate() error
}

func (o *SecretSSHOptions) GetDFToken() string {
	return o.DataFoundryToken
}

func (o *SecretTokenOptions) GetDFToken() string {
	return o.DataFoundryToken
}

func (o *SecretSSHOptions) GetDFNamespace() string {
	return o.NameSpace
}

func (o *SecretTokenOptions) GetDFNamespace() string {
	return o.NameSpace
}

func (o *SecretSSHOptions) GetSecretName() string {
	return o.SecretName
}

func (o *SecretTokenOptions) GetSecretName() string {
	return o.SecretName
}

func (o *SecretSSHOptions) IsSSHSecret() bool {
	return true
}

func (o *SecretTokenOptions) IsSSHSecret() bool {
	return false
}

func (o *SecretSSHOptions) IsTokenSecret() bool {
	return false
}

func (o *SecretTokenOptions) IsTokenSecret() bool {
	return true
}

func (o *SecretSSHOptions) GetPrivateKey() string {
	return o.PrivateKey
}

func (o *SecretTokenOptions) GetPrivateKey() string {
	return ""
}

func (o *SecretSSHOptions) GetToken() string {
	return ""
}

func (o *SecretTokenOptions) GetToken() string {
	return o.GitHubToken
}

func createSecret(o SecretOption) error {
	secret := o.NewSecret()
	return post.create(secret, o.GetDFToken())
}

func listSecrets(o SecretOption) (*api.SecretList, error) {
	option := GetOption{
		Label: GithubSecretLabel,
	}

	return get.list(o.GetDFNamespace(), option, o.GetDFToken())
}

func getSecret(o SecretOption) (*api.Secret, error) {
	return get.get(o.GetDFNamespace(), o.GetSecretName(), o.GetDFToken())
}

func upsertSecret(option SecretOption) error {
	secret, err := getSecret(option)
	if err != nil {
		if NotFound(err) {
			if err := createSecret(option); err != nil {
				return err
			}
			return nil
		}

		return err
	}

	if option.IsSSHSecret() {
		a, b := strings.TrimSpace(option.GetPrivateKey()), strings.TrimSpace(string(secret.Data["ssh-privatekey"]))
		if a == b {
			return nil
		}
	}

	if err := updateSecret(secret, option); err != nil {
		return err
	}

	return nil
}

func updateSecret(s *api.Secret, o SecretOption) error {

	if o.IsSSHSecret() {
		s.Data["ssh-privatekey"] = []byte(o.GetPrivateKey())
	}

	if o.IsTokenSecret() {
		s.Data[PasswordSecret] = []byte(o.GetToken())
	}

	return put.update(s, o.GetDFToken())
}

type SecretTokenOptions struct {
	NameSpace  string
	UserName   string
	SecretName string

	DataFoundryToken string
	GitHubToken      string
}

type SecretSSHOptions struct {
	NameSpace  string `json:"-"`
	UserName   string `json:"-"`
	SecretName string `json:"secret"`

	DataFoundryToken string `json:"-"`
	PrivateKey       string `json:"-"`
}

func (o *SecretTokenOptions) Validate() error {
	if len(o.NameSpace) == 0 {
		return errors.New("secret option namespace is null")
	}

	if len(o.UserName) == 0 {
		return errors.New("secret option user is null")
	}

	if len(o.SecretName) == 0 {
		return errors.New("secret option secret name is null")
	}

	if len(o.DataFoundryToken) == 0 {
		return errors.New("secret option df token is null")
	}

	if len(o.GitHubToken) == 0 {
		return errors.New("secret option github token is null")
	}

	return nil
}

func (o *SecretSSHOptions) Validate() error {
	if len(o.NameSpace) == 0 {
		return errors.New("secret option namespace is null")
	}

	if len(o.UserName) == 0 {
		return errors.New("secret option user is null")
	}

	if len(o.SecretName) == 0 {
		return errors.New("secret option secret name is null")
	}

	if len(o.DataFoundryToken) == 0 {
		return errors.New("secret option df token is null")
	}

	if len(o.PrivateKey) == 0 {
		return errors.New("secret option gitlab PrivateKey is null")
	}

	return nil
}

func setSecretURL(namespace string) string {
	return DFHost_API + fmt.Sprintf(SecretsURL, namespace)
}

func setSecretURLWithName(namespace string, name string) string {
	return DFHost_API + fmt.Sprintf(SecretURL, namespace, name)
}

func getUserKey(user, source string) string {
	return fmt.Sprintf("%s/%s/%s", EtcdUserRegistry, user, source)
}
