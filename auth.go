package main

import (
	"encoding/json"
	"fmt"
	api "github.com/openshift/origin/pkg/user/api/v1"
	"net/http"
	"strings"
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

func getCredentials(userInfo map[string]string) (string, string) {
	credentials := strings.SplitN(userInfo["credential_key"], ":", 2)
	if len(credentials) != 2 {
		fmt.Printf("auth credentials not rigth %v\n", credentials)
		return "", ""
	}

	return credentials[0], fmt.Sprintf("%s %s", credentials[1], userInfo["credential_value"])

}
