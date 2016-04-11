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

//curl http://127.0.0.1:9443/v1/github-redirect?code=8fdf6827d52a1aca5052\&state=9999
func githubHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.Header.Get("namespace")
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
	githubToken := retRaw.Get("access_token")
	if len(githubToken) > 0 {
		if err := etcd.set(namespace, githubToken); err != nil {
			fmt.Fprintf(w, "store namespace %s, token  %s\n err", namespace, githubToken)
			return
		}
	}

	fmt.Fprintf(w, "get request %#s\n", githubToken)
}

//ex. /v1/github-redirect?code=8fdf6827d52a1aca5052&state=ppp
func queryRequestURI(r *http.Request) (url.Values, error) {
	uri, err := url.ParseRequestURI(r.RequestURI)
	if err != nil {
		return nil, err
	}
	return url.ParseQuery(strings.TrimPrefix(uri.RawQuery, uri.Path+"?"))
}
