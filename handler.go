package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	Option_Github_Code  = "code"
	Option_Github_State = "state"
)

//curl http://127.0.0.1:9443/v1/github-redirect?code=4fda33093c9fc12711f1&\state=ccc
//curl http://etcdsystem.servicebroker.dataos.io:2379/v2/keys/oauth/namespace/  -u asiainfoLDP:6ED9BA74-75FD-4D1B-8916-842CB936AC1A
//curl -H "namespace:namespace123" -H "user:panxy3" -H "beartoken:xxxxxxxxxxxxxxxx" http://127.0.0.1:9443/v1/github-redirect?code=4fda33093c9fc12711f1\&state=ccc
func githubHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := headers(r, "namespace", "user", "beartoken")

	if len(userInfo) != 3 {
		fmt.Fprintf(w, "request header not contains [namespace user beartoken]\n")
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
	userInfo["github_token"] = retRaw.Get("access_token")
	if len(userInfo["github_token"]) == 0 {
		fmt.Fprintf(w, "get github token null reaseon %s", string(b))
		return
	}

	if err := db.namespaceSet(userInfo["namespace"], userInfo["user"], userInfo); err != nil {
		fmt.Fprintf(w, "store namespace %s err %s", userInfo["namespace"], err.Error())
		return
	}

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
		DatafactoryToken: info["beartoken"],
		GitHubToken:      info["github_token"],
	}
}
