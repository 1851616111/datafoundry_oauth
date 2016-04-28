package main

import (
	"encoding/json"
	gitlabapi "github.com/asiainfoLDP/datafactory_oauth2/gitlab"
	oapi "github.com/openshift/origin/pkg/user/api/v1"
	"io/ioutil"
	"net/http"
)

type gitlabBindOption struct {
	Host         string `json:"host"`
	User         string `json:"user"`
	PrivateToken string `json:"private_token"`
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
