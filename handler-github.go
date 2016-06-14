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

//curl http://etcdsystem.servicebroker.dataos.io:2379/v2/keys/oauth/namespace/  -u asiainfoLDP:6ED9BA74-75FD-4D1B-8916-842CB936AC1A
//curl http://127.0.0.1:9443/v1/github-redirect?code=d13f63cc79c2907f9e55\&state=xcv&namespace=oauthtest\&bearer=Uzl65t8jzNc46ZoZEqS4Rg8R9JVbQ5plOH7Nf0gsJV4&redirect_url=https://baidu.com
func githubHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	r.ParseForm()
	var ns, bearer, redirect_url string
	if ns = r.FormValue("namespace"); strings.TrimSpace(ns) == "" {
		retHttpCode(400, 1400, w, "param namespace must not be nil.")
		return
	}
	if bearer = r.FormValue("bearer"); strings.TrimSpace(bearer) == "" {
		retHttpCode(400, 1400, w, "param bearer must not be nil.")
		return
	}
	if redirect_url = r.FormValue("redirect_url"); strings.TrimSpace(redirect_url) == "" {
		retHttpCode(400, 1400, w, "param redirect_url must not be nil.")
		return
	}

	var user *api.User
	var err error
	if user, err = authDF("bearer " + bearer); err != nil {
		retHttpCode(401, 1400, w, "unauthorized")
		return
	}

	userInfo := map[string]string{
		"namespace": ns,
		"bearer":    bearer,
		"user":      user.Name,
	}

	raw, err := queryRequestURI(r)
	if err != nil {
		retHttpCodef(400, 1400, w, "parse request url %s err %s", r.RequestURI, err.Error())
		return
	}

	code, state := raw.Get(Option_Github_Code), raw.Get(Option_Github_State)

	retriveTokenUrl := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s&state=%s", GithubClientID, GithubClientSecret, code, state)
	retriveTokenURL, err := url.ParseQuery(retriveTokenUrl)
	if err != nil {
		retHttpCodef(400, 1400, w, "generate token request url %s err %s", retriveTokenUrl, err.Error())
		return
	}

	res, err := http.PostForm("https://github.com/login/oauth/access_token", retriveTokenURL)
	if err != nil {
		retHttpCodef(400, 1400, w, "retrive token err %s", err.Error())
		return
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		retHttpCodef(400, 1400, w, "retrive token err %s", err.Error())
		return
	}

	retRaw, _ := url.ParseQuery(string(b))
	userInfo["credential_value"] = retRaw.Get("access_token")
	if len(userInfo["credential_value"]) == 0 {
		retHttpCodef(400, 1400, w, "get github token null reaseon %s", string(b))
		return
	}

	completeGithubInfo(userInfo)

	key := fmt.Sprintf("/oauth/namespaces/%s/%s/%s", userInfo["namespace"], userInfo["user"], userInfo["source"])
	if err := db.set(key, userInfo); err != nil {
		retHttpCodef(400, 1400, w, "store namespace %s err %s", userInfo["namespace"], err.Error())
		return
	}
	go syncOauthUser(db, userInfo)

	http.Redirect(w, r, redirect_url, 302)
}

//curl http://127.0.0.1:9443/v1/repos/github/owner?namespace=oauth -H  "Authorization: Bearer V3nszMTMHl_IJalZMuZVADjAxzDJBhuhzrcb01U6AKg"
func githubOwnerReposHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user *api.User
	var err error
	var ns string
	token := r.Header.Get("Authorization")
	if len(token) == 0 {
		retHttpCode(400, 1400, w, "no header Authorization")
		return
	}
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, 1401, w, "auth err %s", err.Error())
		return
	}
	if ns = r.FormValue("namespace"); strings.TrimSpace(ns) == "" {
		retHttpCode(400, 1400, w, "param namespace must not be nil.")
		return
	}

	var userInfo map[string]string
	if userInfo, err = getGithubInfo(user); err != nil {
		if EtcdKeyNotFound(err) {
			url := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&scope=repo,user:email&state=${{}}&redirect_uri=%s", GithubClientID, GithubRedirectUrl)
			retHttpCode(400, 1401, w, url)
		} else {
			retHttpCodef(400, 1400, w, "get user info err %s", err.Error())
		}
		return
	}

	//go concurrency
	const TotalConcurrency = 2
	result := make(chan Result, TotalConcurrency)
	done := make(chan struct{}, 1)
	defer close(result)

	go func(result chan Result) {
		option := &SecretTokenOptions{
			NameSpace:        ns,
			UserName:         user.Name,
			SecretName:       generateGithubName(user.Name),
			DataFoundryToken: stripBearToken(token),
			GitHubToken:      userInfo["credential_value"],
		}

		if err := option.Validate(); err != nil {
			Done(done, result, 400, 1400, fmt.Sprintf("validate datafoundry secret option err %v", err))
			return
		}

		secrets, err := listSecrets(option)
		if err != nil {
			Done(done, result, 400, 1400, fmt.Sprintf("list secret err %v", err))
			return
		}

		if len(secrets.Items) == 0 {
			if err := createSecret(option); err != nil {
				Done(done, result, 400, 1400, fmt.Sprintf("create secret  err %v", err))
				return
			}
			Done(done, result, 200, 1200, fmt.Sprintf(`"secret":"%s"`, option.SecretName))
			return
		}

		Done(done, result, 200, 1200, fmt.Sprintf(`"secret":"%s"`, secrets.Items[0].Name))
	}(result)

	go func(result chan Result) {
		var repos *Repos
		var data []byte
		cached := r.FormValue("cache")
		switch cached {
		case "false":
			if repos, err = GetOwnerRepos(userInfo); err != nil {
				Done(done, result, 400, 1400, fmt.Sprintf("request github err %v", err))
				return
			}

			newRepos := repos.Convert()
			data, err = json.Marshal(newRepos)
			if err != nil {
				Done(done, result, 400, 1400, fmt.Sprintf("convert return err %v", err))
				return
			}
		default:

			data, err = Cache.HFetch("www.github.com", "user_"+user.Name+"@owner_repos")
			if err != nil {
				retHttpCodef(400, 1400, w, "get github owner repos(cached) err %v", err.Error())
				return
			}
		}

		Done(done, result, 200, 1200, fmt.Sprintf(`"infos":%s`, string(data)))
	}(result)

	msg := []string{}
	for i := 1; i <= TotalConcurrency; i++ {
		select {
		case res := <-result:
			if res.code != 200 {
				done <- struct{}{}
				retHttpCode(res.code, res.bodyCode, w, res.msg)
				return
			}
			msg = append(msg, res.msg)
		}
	}

	retHttpCodeJson(200, 1200, w, fmt.Sprintf("{%s,%s}", msg[0], msg[1]))
}

//curl http://127.0.0.1:9443/v1/repos/github/orgs -H  "Authorization: Bearer V3nszMTMHl_IJalZMuZVADjAxzDJBhuhzrcb01U6AKg"
//curl http://127.0.0.1:9443/v1/repos/github/orgs?cache=false -H  "Authorization: Bearer V3nszMTMHl_IJalZMuZVADjAxzDJBhuhzrcb01U6AKg"
func githubOrgReposHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user *api.User
	var err error
	token := r.Header.Get("Authorization")
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, 1401, w, "auth err %s", err.Error())
		return
	}

	var data []byte

	cached := r.FormValue("cache")
	switch cached {
	case "false":
		var userInfo map[string]string
		if userInfo, err = getGithubInfo(user); err != nil {
			retHttpCodef(400, 1400, w, "get user info err %s", err.Error())
			return
		}

		var orgs []Org
		if orgs, err = GetOwnerOrgs(userInfo); err != nil {
			retHttpCodef(400, 1400, w, "get orgs info err %s", err.Error())
			return
		}

		repos := Repos{}
		for _, v := range orgs {
			var l *Repos
			if l, err = GetOrgReps(userInfo, v.Login); err != nil {
				fmt.Printf("[GET]/github.com/user/orgs, get org %s info err %s", v.Login, err.Error())
				continue
			}
			repos = append(repos, *l...)
		}

		newRepos := repos.Convert()
		data, err = json.Marshal(newRepos)
		if err != nil {
			retHttpCodef(400, 1400, w, "convert return err %v", err)
			return
		}
	default:
		if data, err = Cache.HFetch("www.github.com", "user_"+user.Name+"@orgs_repos"); err != nil {
			retHttpCodef(400, 1400, w, "get github orgs repos(cached) err %v", err.Error())
			return
		}

	}

	retHttpCodeJson(200, 1200, w, string(data))
}

func getGithubBranchHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userName, repoName := ps.ByName("user"), ps.ByName("repo")
	var user *api.User
	var err error
	token := r.Header.Get("Authorization")
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, 1401, w, "auth err %s", err.Error())
		return
	}

	var userInfo map[string]string
	if userInfo, err = getGithubInfo(user); err != nil {
		retHttpCodef(400, 1400, w, "get user info err %s", err.Error())
		return
	}

	var tmp []map[string]interface{}
	if tmp, err = GetRepoBranck(userInfo, userName, repoName); err != nil {
		retHttpCodef(400, 1400, w, "get repo branch err %s", err.Error())
		return
	}

	b, err := json.Marshal(tmp)
	if err != nil {
		retHttpCodef(400, 1400, w, "convert return err %s", err.Error())
		return
	}

	retHttpCodeJson(200, 1200, w, string(b))
}

//ex. /v1/github-redirect?code=8fdf6827d52a1aca5052&state=ppp
func queryRequestURI(r *http.Request) (url.Values, error) {
	uri, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		return nil, err
	}
	return url.ParseQuery(strings.TrimPrefix(uri.RawQuery, uri.Path+"?"))
}

func setTokenSecretOption(info map[string]string) SecretOption {
	return &SecretTokenOptions{
		NameSpace:        info["namespace"],
		UserName:         info["user"],
		SecretName:       generateGithubName(info["user"]),
		DataFoundryToken: info["bearer"],
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
