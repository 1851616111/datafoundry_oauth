package main

import (
	"encoding/json"
	"fmt"
	Go "github.com/asiainfoLDP/datafoundry_oauth2/util/goroutine"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	Github_API_User          = "https://api.github.com/user"
	Github_API_User_Repos    = "https://api.github.com/users/%s/repos"
	Github_API_Owner_Orgs    = "https://api.github.com/user/orgs"
	Github_API_Org_Repos     = "https://api.github.com/orgs/%s/repos?page=%d"
	Github_API_Repo_Branches = "https://api.github.com/repos/%s/%s/branches"
)

func GetUserInfo(userInfo map[string]string) (*Owner, error) {
	credKey, credValue := getCredentials(userInfo)

	b, err := get(Github_API_User, credKey, credValue)
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

	url := fmt.Sprintf(Github_API_User_Repos, user.Login)
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

	var firstPage int = 1
	var lastPage int
	var err error

	repos := &Repos{}
	resp_header_value_c := make(chan []string, 1)

	url := fmt.Sprintf(Github_API_Org_Repos, org, firstPage)

	b, err := httpGetFunc(url, func(resp *http.Response) {
		const GitHubApiLastPageKey = "Link"
		asyncHeaderKey(resp.Header, GitHubApiLastPageKey, resp_header_value_c)
		return
	}, credKey, credValue)

	if err := json.Unmarshal(b, &repos); err != nil {
		log.Printf("get github orgs reps err %v", err)
		return repos, err
	}

	if s, ok := <-resp_header_value_c; ok {
		//s = <https://api.github.com/organizations/14065116/repos?page=2>; rel="next", <https://api.github.com/organizations/14065116/repos?page=4>; rel="last"
		//need 4
		str := s[0][strings.Index(s[0], ","):]
		lastPageStr := strings.TrimSpace(middleStr(str, "page=", ">"))
		if lastPage, err = strconv.Atoi(lastPageStr); err != nil {
			log.Println("get github orgs last page err", err)
		}
	}

	if lastPage >= 2 {
		goNum := lastPage - 1
		var goFunc = func(goTime int) interface{} {
			page := goTime + 1

			url := fmt.Sprintf(Github_API_Org_Repos, org, page)
			b, err := get(url, credKey, credValue)
			if err != nil {
				log.Printf("go get github orgs reps times=%d err %v", goTime, err)
				return nil
			}

			repos := &Repos{}

			if err := json.Unmarshal(b, &repos); err != nil {
				log.Printf("go get github orgs reps times=%d  err %v", goTime, err)
				return nil
			}

			return *repos
		}
		resCh := make(chan interface{}, goNum)

		Go.Go(goNum, goFunc, resCh)

		for res := range resCh {
			if reps, ok := res.(Repos); ok {
				*repos = append(*repos, reps...)
			}
		}
	}

	return repos, nil
}

func asyncHeaderKey(header http.Header, key string, value_c chan []string) {
	if header == nil {
		return
	}

	if vs, ok := header[key]; ok {
		value_c <- vs
	}

	//no matter success or fail, close value_c to info main goroutine
	close(value_c)
	return
}

//"Link":[]string{"<https://api.github.com/organizations/14065116/repos?page=2>; rel=\"next\", <https://api.github.com/organizations/14065116/repos?page=4>; rel=\"last\""}
func middleStr(s, start, end string) string {
	if len(s) == 0 {
		return ""
	}

	startIndex, endIndex := strings.Index(s, start), strings.LastIndex(s, end)

	if startIndex == -1 || endIndex == -1 {
		return ""
	}

	if startIndex+len(start) >= endIndex {
		return ""
	}

	return s[startIndex+len(start) : endIndex]
}

func GetRepoBranck(userInfo map[string]string, user, repo string) ([]map[string]interface{}, error) {
	credKey, credValue := getCredentials(userInfo)

	url := fmt.Sprintf(Github_API_Repo_Branches, user, repo)

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
