package main

import (
	"fmt"
	"net/http"
)

var (
	tokenConfig Config
	//RedirectUrl  = "http://oauth2-oauth.app.asiainfodata.com/v1/github-redirect"
	//ClientID     = "2369ed831a59847924b4"
	//ClientSecret = "510bb29970fcd684d0e7136a5947f92710332c98"
	GithubRedirectUrl, GithubClientID, GithubClientSecret string
	etcdClient                                            client

	store Store
)

func init() {
	initStorage()
	initOauth2Plugin()
	initOauthConfig()
}

func main() {

	http.HandleFunc("/v1/github-redirect", githubHandler)
	http.ListenAndServe(":9443", nil)

}

func initOauthConfig() {
	var err error
	tokenConfig, err = NewGitHub(GithubClientID, GithubClientSecret, GithubRedirectUrl, []string{"repo", "user:email"})
	if err != nil {
		fmt.Printf("oauth init fail %s", err.Error())
	}

	fmt.Println("oauth config init success")
}

func initStorage() {
	c := storeConfig{
		Addr:   httpAddrMaker(getEnv("ETCD_HTTP_ADDR", true)),
		Port:   getEnv("ETCD_HTTP_PORT", true),
		User:   getEnv("ETCD_USER", true),
		Passwd: getEnv("ETCD_PASSWORD", true),
	}

	printConfig(&c)

	etcdClient = c.newClient()
	etcdClient.SetCredentials(c.User, c.Passwd)

	fmt.Println("oauth init storage config success")
}

func initOauth2Plugin() {
	initGithubPlugin()
}

func initGithubPlugin() {
	GithubRedirectUrl = getEnv("GITHUB_REDIRECT_URL", true)
	GithubClientID = getEnv("GITHUB_CLIENT_ID", true)
	GithubClientSecret = getEnv("GITHUB_CLIENT_SECRET", true)
}
