package main

import (
	"encoding/json"
	api "github.com/openshift/origin/pkg/user/api/v1"
	"net/http"
)

func authDFRequest(r *http.Request) (*api.User, error) {
	token := r.Header.Get("Authorization")
	b, err := get(DF_API_Auth, "Authorization", token)
	if err != nil {
		return nil, err
	}

	user := new(api.User)
	if err := json.Unmarshal(b, user); err != nil {
		return nil, err
	}

	return user, nil
}
