package main

import (
	"encoding/json"
	"fmt"
)

const (
	Github_API_Owner_Repos = "https://api.github.com/user/repos"
	Github_API_Owner_Orgs  = "https://api.github.com/user/orgs"
	Github_API_Org_Repos   = "https://api.github.com/orgs/%s/repos"
)

func GetOwnerRepos(userInfo map[string]string) (*Repos, error) {
	credKey, credValue := getCredentials(userInfo)

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

func GetOwnerOrgs(userInfo map[string]string) ([]Org, error) {
	credKey, credValue := getCredentials(userInfo)

	b, err := get(Github_API_Owner_Orgs, credKey, credValue)
	if err != nil {
		return nil, err
	}

	orgs := []Org{}
	if err := json.Unmarshal(b, &orgs); err != nil {
		return nil, err
	}

	return orgs, nil
}

func GetOrgReps(userInfo map[string]string, org string) (*Repos, error) {
	credKey, credValue := getCredentials(userInfo)
	url := fmt.Sprintf(Github_API_Org_Repos, org)

	b, err := get(url, credKey, credValue)
	if err != nil {
		return nil, err
	}

	repos := &Repos{}
	if err := json.Unmarshal(b, &repos); err != nil {
		return nil, err
	}

	return repos, nil
}
