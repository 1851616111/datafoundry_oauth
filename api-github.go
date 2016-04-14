package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	Github_API_Owner_Repos = "https://api.github.com/user/repos"
)

func GetUserRepos(userInfo map[string]string) (*Repos, error) {
	credentials := strings.SplitN(userInfo["credential_key"], ":", 2)
	if len(credentials) != 2 {
		return nil, errors.New(fmt.Sprintf("[GET]/user/repos, user info credential_key %v\n not right\n", credentials))
	}

	credKey, credValue := credentials[0], fmt.Sprintf("%s %s", credentials[1], userInfo["credential_value"])

	b, err := get(Github_API_Owner_Repos, credKey, credValue)
	if err != nil {
		return nil, err
	}

	repos := &Repos{}
	if err := json.Unmarshal(b, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}
