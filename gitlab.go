package main

import (
	"encoding/json"
	"fmt"
	gitlabapi "github.com/asiainfoLDP/datafoundry_oauth2/gitlab"
	oapi "github.com/openshift/origin/pkg/user/api/v1"
	"io/ioutil"
	"net/http"
)

type gitLabInfo struct {
	Host         string `json:"host"`
	User         string `json:"user"`
	PrivateToken string `json:"private_token"`
}

type gitLabBindInfo struct {
	Host string `json:"host"`
	Id   int    `json:"project_id"`
}

type gitlabDumpling struct {
	filling *gitlabapi.User
	err     error
}

type datafoundryDumpling struct {
	filling *oapi.User
	err     error
}

func parseRequestBody(r *http.Request, i interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, i); err != nil {
		return err
	}

	return nil
}

func setDeployKey(source, key string, value interface{}) error {
	deployKey := fmt.Sprintf("/oauth/deploykeys/%s/%s", source, key)
	return db.set(deployKey, value)
}

func getDeployKey(source, key string) (string, error) {
	deployKey := fmt.Sprintf("/oauth/deploykeys/%s/%s", source, key)
	return db.get(deployKey, true, false)
}

func setWebHook(source, host, namespace, build string, value interface{}) error {
	webHook := fmt.Sprintf("/oauth/webhooks/%s/host/%s/namespaces/%s/builds/%s", source, host, namespace, build)
	return db.set(webHook, value)
}

func getWebHook(source, host, namespace, build string) (string, error) {
	webHook := fmt.Sprintf("/oauth/webhooks/%s/host/%s/namespaces/%s/builds/%s", source, host, namespace, build)
	return db.get(webHook, true, false)
}

func deleteWebHook(source, host, namespace, build string) error {
	webHook := fmt.Sprintf("/oauth/webhooks/%s/host/%s/namespaces/%s/builds/%s", source, host, namespace, build)
	return db.delete(webHook, true)
}
