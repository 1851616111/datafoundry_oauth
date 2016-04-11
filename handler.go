package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
)
const (
	ReDirectUrl = "https://api.daocloud.io/v1/github-redirect?redirect-url=https://dashboard.daocloud.io/settings/profile?third_party=github"
	CodeRedirect = 302
)
func githubHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("------------------> %#v\n%", r)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("-----------> read body err %s\n", err.Error())
	}

	fmt.Println("----------------> get request from  github ", string(b))

	c, err := NewGitHub("2369ed831a59847924b4", "510bb29970fcd684d0e7136a5947f92710332c98", ReDirectUrl, []string{"repo", "user:email"})
	if err != nil {
		fmt.Fprint(w, err)
	}
	t := Auth(c, "/")

	http.Redirect(w, r, t, CodeRedirect)
}