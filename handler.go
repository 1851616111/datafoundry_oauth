package main

import (
	"net/http"
	"fmt"
)
const (
	ReDirectUrl = "https://api.daocloud.io/v1/github-redirect?redirect-url=https://dashboard.daocloud.io/settings/profile?third_party=github"
	CodeRedirect = 302
)
func githubHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("------------------> %#v\n%", r)
	c, err := NewGitHub("2369ed831a59847924b4", "510bb29970fcd684d0e7136a5947f92710332c98", ReDirectUrl, []string{"repo", "user:email"})
	if err != nil {
		fmt.Fprint(w, err)
	}
	t := Auth(c, "/")

	http.Redirect(w, r, t, CodeRedirect)
}