package main

import (
	"fmt"
	api "github.com/openshift/origin/pkg/user/api/v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encoding/json"
	"github.com/julienschmidt/httprouter"
)

const (
	Option_Github_Code  = "code"
	Option_Github_State = "state"
)

//curl http://127.0.0.1:9443/v1/github-redirect?code=4fda33093c9fc12711f1&\state=ccc
//curl http://etcdsystem.servicebroker.dataos.io:2379/v2/keys/oauth/namespace/  -u asiainfoLDP:6ED9BA74-75FD-4D1B-8916-842CB936AC1A
//curl -H "namespace:namespace123" -H "user:panxy3" -H "bearer:xxxxxxxxxxxxxxxx" http://127.0.0.1:9443/v1/github-redirect?code=4fda33093c9fc12711f1\&state=ccc
func githubHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userInfo := headers(r, "namespace", "user", "bearer")

	if len(userInfo) != 3 {
		fmt.Fprintf(w, "request header not contains [namespace user bearer]\n")
		return
	}

	raw, err := queryRequestURI(r)
	if err != nil {
		fmt.Fprintf(w, "parse request url %s err %s", r.RequestURI, err.Error())
		return
	}

	code, state := raw.Get(Option_Github_Code), raw.Get(Option_Github_State)

	retriveTokenUrl := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s&state=%s", GithubClientID, GithubClientSecret, code, state)
	retriveTokenURL, err := url.ParseQuery(retriveTokenUrl)
	if err != nil {
		fmt.Fprintf(w, "generate token request url %s err %s", retriveTokenUrl, err.Error())
		return
	}

	res, err := http.PostForm("https://github.com/login/oauth/access_token", retriveTokenURL)
	if err != nil {
		fmt.Fprintf(w, "retrive token err %s", err.Error())
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Fprintf(w, "retrive token err %s", err.Error())
		return
	}

	retRaw, _ := url.ParseQuery(string(b))
	userInfo["credential_value"] = retRaw.Get("access_token")
	if len(userInfo["credential_value"]) == 0 {
		fmt.Fprintf(w, "get github token null reaseon %s", string(b))
		return
	}
	completeGithubInfo(userInfo)

	key := fmt.Sprintf("/oauth/namespaces/%s/%s/%s", userInfo["namespace"], userInfo["user"], userInfo["source"])
	if err := db.set(key, userInfo); err != nil {
		fmt.Fprintf(w, "store namespace %s err %s", userInfo["namespace"], err.Error())
		return
	}
	go syncOauthUser(db, userInfo)

	option := setSecretOption(userInfo)
	if err := option.validate(); err != nil {
		fmt.Fprintf(w, "validate datafoundry secret option err %s\n", err.Error())
		return
	}

	secret, err := getSecret(option)
	if NotFount(err) {
		if err := createSecret(option); err != nil {
			fmt.Fprintf(w, "create datafoundry secret err %s\n", err.Error())
			return
		}

		fmt.Fprint(w, "create datafaoundry secret success\n")
		return
	}

	if err != nil {
		fmt.Fprintf(w, "get datafoundry secret err %s\n", err.Error())
		return
	}

	if err := updateSecret(secret, option); err != nil {
		fmt.Fprintf(w, "update datafoundry secret err %s\n", err.Error())
		return
	}

	fmt.Fprintf(w, "ok")

}

func githubUserOwnerReposHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user *api.User
	var err error
	if user, err = authDFRequest(r); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/user/repos, auth err %s\n", err.Error())
		return
	}

	var userInfo map[string]string
	if userInfo, err = getGithubInfoByDFUser(user); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/user/repos, get user info err %s\n", err.Error())
		return
	}

	var repos *Repos
	if repos, err = GetOwnerRepos(userInfo); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/user/repos, request github err %s\n", err.Error())
		return
	}

	newRepos := repos.Convert()
	b, err := json.Marshal(newRepos)
	if err != nil {
		fmt.Fprintf(w, "[GET]/github.com/user/repos, convert return err %s\n", string(b))
		return
	}

	fmt.Fprintf(w, "%s", string(b))
}

func githubOrgOwnerReposHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user *api.User
	var err error
	if user, err = authDFRequest(r); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/user/orgs, auth err %s\n", err.Error())
		return
	}

	var userInfo map[string]string
	if userInfo, err = getGithubInfoByDFUser(user); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/user/orgs, get user info err %s\n", err.Error())
		return
	}

	var orgs []Org
	if orgs, err = GetOwnerOrgs(userInfo); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/user/orgs, get orgs info err %s\n", err.Error())
		return
	}

	repos := Repos{}
	for _, v := range orgs {
		var l *Repos
		if l, err = GetOrgReps(userInfo, v.Login); err != nil {
			fmt.Printf("[GET]/github.com/user/orgs, get org %s info err %s\n", v.Login, err.Error())
			continue
		}
		repos = append(repos, *l...)
	}

	newRepos := repos.Convert()
	b, err := json.Marshal(newRepos)
	if err != nil {
		fmt.Fprintf(w, "[GET]/github.com/user/orgs, convert return err %s\n", string(b))
		return
	}

	fmt.Fprintf(w, "%s", string(b))
}

func getGithubBranchHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userName, repoName := ps.ByName("user"), ps.ByName("repo")
	var user *api.User
	var err error
	if user, err = authDFRequest(r); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/users/%s/repos/%s, auth err %s\n", userName, repoName, err.Error())
		return
	}

	var userInfo map[string]string
	if userInfo, err = getGithubInfoByDFUser(user); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/users/%s/repos/%s, get user info err %s\n", userName, repoName, err.Error())
		return
	}

	var tmp []map[string]interface{}
	if tmp, err = GetRepoBranck(userInfo, userName, repoName); err != nil {
		fmt.Fprintf(w, "[GET]/github.com/users/%s/repos/%s, request err %s\n", userName, repoName, err.Error())
		return
	}

	b, err := json.Marshal(tmp)
	if err != nil {
		fmt.Fprintf(w, "[GET]/github.com/users/%s/repos/%s, convert return err %s\n", userName, repoName, err.Error())
		return
	}

	fmt.Fprintf(w, "%s", string(b))
}

//ex. /v1/github-redirect?code=8fdf6827d52a1aca5052&state=ppp
func queryRequestURI(r *http.Request) (url.Values, error) {
	uri, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		return nil, err
	}
	return url.ParseQuery(strings.TrimPrefix(uri.RawQuery, uri.Path+"?"))
}

func setSecretOption(info map[string]string) *secretOptions {
	return &secretOptions{
		NameSpace:        info["namespace"],
		UserName:         info["user"],
		SecretName:       generateName(info["namespace"], info["user"]),
		DatafactoryToken: info["bearer"],
		GitHubToken:      info["credential_value"],
	}
}

func syncOauthUser(db Store, userInfo map[string]string) {
	k := getUserKey(userInfo["user"], userInfo["source"])
	if err := db.set(k, userInfo); err != nil {
		//todo err handler
		fmt.Printf("add user info err %s", err.Error())
	}
}

func completeGithubInfo(info map[string]string) {
	info["time"] = time.Now().String()
	info["source"] = "github.com"
	info["credential_key"] = "Authorization:token"
}
