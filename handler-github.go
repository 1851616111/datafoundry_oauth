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
//curl http://127.0.0.1:9443/v1/github-redirect?code=4fda33093c9fc12711f1\&state=ccc?namespace=auth\&bearer=xxxxxxxxxxxxxxxxxxxxxxxxx
func githubHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var ns, bearer string
	if ns = r.FormValue("namespace"); strings.TrimSpace(ns) == "" {
		retHttpCode(400, w, "param namespace must not be nil.\n")
		return
	}
	if bearer = r.FormValue("bearer"); strings.TrimSpace(bearer) == "" {
		retHttpCode(400, w, "param bearer must not be nil.\n")
		return
	}

	var user *api.User
	var err error
	if user, err = authDF("bearer " + bearer); err != nil {
		retHttpCode(401, w, "unauthorized\n")
		return
	}

	userInfo := map[string]string{
		"namespace": ns,
		"bearer":    bearer,
		"user":      user.Name,
	}

	raw, err := queryRequestURI(r)
	if err != nil {
		retHttpCodef(400, w, "parse request url %s err %s", r.RequestURI, err.Error())
		return
	}

	code, state := raw.Get(Option_Github_Code), raw.Get(Option_Github_State)

	retriveTokenUrl := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s&state=%s", GithubClientID, GithubClientSecret, code, state)
	retriveTokenURL, err := url.ParseQuery(retriveTokenUrl)
	if err != nil {
		retHttpCodef(400, w, "generate token request url %s err %s", retriveTokenUrl, err.Error())
		return
	}

	res, err := http.PostForm("https://github.com/login/oauth/access_token", retriveTokenURL)
	if err != nil {
		retHttpCodef(400, w, "retrive token err %s", err.Error())
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		retHttpCodef(400, w, "retrive token err %s", err.Error())
		return
	}

	retRaw, _ := url.ParseQuery(string(b))
	userInfo["credential_value"] = retRaw.Get("access_token")
	if len(userInfo["credential_value"]) == 0 {
		retHttpCodef(400, w, "get github token null reaseon %s", string(b))
		return
	}

	completeGithubInfo(userInfo)

	key := fmt.Sprintf("/oauth/namespaces/%s/%s/%s", userInfo["namespace"], userInfo["user"], userInfo["source"])
	if err := db.set(key, userInfo); err != nil {
		retHttpCodef(400, w, "store namespace %s err %s", userInfo["namespace"], err.Error())
		return
	}
	go syncOauthUser(db, userInfo)

	option := setTokenSecretOption(userInfo)
	if err := option.validate(); err != nil {
		retHttpCodef(400, w, "validate datafoundry secret option err %s\n", err.Error())
		return
	}

	if err := upsertSecret(option); err != nil {
		retHttpCodef(400, w, "operate datafoundry secret err %s\n", err.Error())
		return
	}

	retHttpCode(200, w, "ok")
}

func githubOwnerReposHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user *api.User
	var err error
	token := r.Header.Get("Authorization")
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, w, "auth err %s\n", err.Error())
		return
	}

	var userInfo map[string]string
	if userInfo, err = getGithubInfo(user); err != nil {
		retHttpCodef(400, w, "get user info err %s\n", err.Error())
		return
	}

	var repos *Repos
	if repos, err = GetOwnerRepos(userInfo); err != nil {
		retHttpCodef(400, w, "request github err %s\n", err.Error())
		return
	}

	newRepos := repos.Convert()
	b, err := json.Marshal(newRepos)
	if err != nil {
		retHttpCodef(400, w, "convert return err %s\n", string(b))
		return
	}

	retHttpCodef(200, w, "%s", string(b))
}

func githubOrgReposHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user *api.User
	var err error
	token := r.Header.Get("Authorization")
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, w, "auth err %s\n", err.Error())
		return
	}

	var userInfo map[string]string
	if userInfo, err = getGithubInfo(user); err != nil {
		retHttpCodef(400, w, "get user info err %s\n", err.Error())
		return
	}

	var orgs []Org
	if orgs, err = GetOwnerOrgs(userInfo); err != nil {
		retHttpCodef(400, w, "get orgs info err %s\n", err.Error())
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
		retHttpCodef(400, w, "convert return err %s\n", string(b))
		return
	}

	retHttpCode(200, w, "%s", string(b))
}

func getGithubBranchHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userName, repoName := ps.ByName("user"), ps.ByName("repo")
	var user *api.User
	var err error
	token := r.Header.Get("Authorization")
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, w, "auth err %s\n", err.Error())
		return
	}

	var userInfo map[string]string
	if userInfo, err = getGithubInfo(user); err != nil {
		retHttpCodef(400, w, "get user info err %s\n", err.Error())
		return
	}

	var tmp []map[string]interface{}
	if tmp, err = GetRepoBranck(userInfo, userName, repoName); err != nil {
		retHttpCodef(400, w, "get repo branch err %s\n", err.Error())
		return
	}

	b, err := json.Marshal(tmp)
	if err != nil {
		retHttpCodef(400, w, "convert return err %s\n", err.Error())
		return
	}

	retHttpCode(200, w, "%s", string(b))
}

//ex. /v1/github-redirect?code=8fdf6827d52a1aca5052&state=ppp
func queryRequestURI(r *http.Request) (url.Values, error) {
	uri, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		return nil, err
	}
	return url.ParseQuery(strings.TrimPrefix(uri.RawQuery, uri.Path+"?"))
}

func setTokenSecretOption(info map[string]string) *SecretTokenOptions {
	return &SecretTokenOptions{
		NameSpace:        info["namespace"],
		UserName:         info["user"],
		SecretName:       generateGithubName(info["user"]),
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
