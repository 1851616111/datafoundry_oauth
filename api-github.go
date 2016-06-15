package main

import (
	"encoding/json"
	"fmt"
	"log"
)

const (
	GitHub_API_User          = "https://api.github.com/user"
	GitHub_API_User_Repos    = "https://api.github.com/users/%s/repos"
	GitHub_API_Owner_Orgs    = "https://api.github.com/user/orgs"
	GitHub_API_Org_Repos     = "https://api.github.com/orgs/%s/repos"
	GitHub_API_Repo_Branches = "https://api.github.com/repos/%s/%s/branches"
	GitHub_API_Repo_WebHook  = "https://api.github.com/repos/%s/%s/hooks"
)

func GetUserInfo(userInfo map[string]string) (*Owner, error) {
	credKey, credValue := getCredentials(userInfo)

	b, err := get(GitHub_API_User, credKey, credValue)
	if err != nil {
		return nil, err
	}

	user := &Owner{}
	if err := json.Unmarshal(b, &user); err != nil {
		return nil, err
	}

	return user, nil
}

func GetOwnerRepos(userInfo map[string]string) (*Repos, error) {

	user, err := GetUserInfo(userInfo)
	if err != nil {
		return nil, err
	}

	credKey, credValue := getCredentials(userInfo)

	url := fmt.Sprintf(GitHub_API_User_Repos, user.Login)
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

func GetOwnerOrgs(userInfo map[string]string) ([]Org, error) {
	credKey, credValue := getCredentials(userInfo)

	b, err := get(GitHub_API_Owner_Orgs, credKey, credValue)
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

	const paramMaxPerPage = 100
	page := 1

	repos := &Repos{}
	for {
		url := fmt.Sprintf(GitHub_API_Org_Repos, org) + fmt.Sprintf("?per_page=%d&page=%d", paramMaxPerPage, page)
		b, err := get(url, credKey, credValue)
		if err != nil {
			log.Printf("get github orgs repos err %v", err)
			return nil, err
		}

		section := &Repos{}
		if err := json.Unmarshal(b, &section); err != nil {
			log.Printf("get github orgs reps err %v", err)
			return repos, err
		}

		if len(*section) > 0 {
			*repos = append(*repos, *section...)
		}

		//最后一页
		if len(*section) < 100 {
			break
		}
		page += 1
	}

	return repos, nil
}

func GetRepoBranck(userInfo map[string]string, user, repo string) ([]map[string]interface{}, error) {
	credKey, credValue := getCredentials(userInfo)

	url := fmt.Sprintf(GitHub_API_Repo_Branches, user, repo)

	b, err := get(url, credKey, credValue)
	if err != nil {
		return nil, err
	}

	tmp := []map[string]interface{}{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil, err
	}

	return tmp, nil
}

func CreateRepoWebHook(user, repo string, option *GitHubWebHookOption, credentialKey, credentialValue string) (*GitHubWebHook, error) {
	url := fmt.Sprintf(GitHub_API_Repo_WebHook, user, repo)

	byte, err := json.Marshal(option)
	if err != nil {
		return nil, err
	}

	b, err := post(url, byte, credentialKey, credentialValue)
	if err != nil {
		return nil, err
	}

	hook := new(GitHubWebHook)
	if err := json.Unmarshal(b, hook); err != nil {
		return nil, err
	}

	return hook, nil
}

func UpdateRepoWebHook(user, repo string, id int, option *GitHubWebHookOption, credentialKey, credentialValue string) (*GitHubWebHook, error) {
	url := fmt.Sprintf(GitHub_API_Repo_WebHook, user, repo) + fmt.Sprintf("/%d", id)

	byte, err := json.Marshal(option)
	if err != nil {
		return nil, err
	}

	b, err := patch(url, byte, credentialKey, credentialValue)
	if err != nil {
		return nil, err
	}

	hook := new(GitHubWebHook)
	if err := json.Unmarshal(b, hook); err != nil {
		return nil, err
	}

	return hook, nil
}

func DeleteRepoWebHook(user, repo string, id int, credentialKey, credentialValue string) error {
	url := fmt.Sprintf(GitHub_API_Repo_WebHook, user, repo) + fmt.Sprintf("/%d", id)
	_, err := delete(url, credentialKey, credentialValue)
	return err
}
