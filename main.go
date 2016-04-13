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
	db                                                    Store
	DFHost                                                string
)

func init() {
	initStorage()
	initOauth2Plugin()
	initOauthConfig()
	initDFHost()
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
	db = c.newClient()

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

func initDFHost() {
	DFHost = getEnv("DATAFACTORY_HOST_ADDR", true)
}

//https://github.com/login/oauth/authorize?client_id=2369ed831a59847924b4&scope=repo,user:email&state=ccc&redirect_uri=http://oauth2-oauth.app.asiainfodata.com/v1/github-redirect
//curl -v https://github.com/login/oauth/access_token -d "client_id=2369ed831a59847924b4&client_secret=510bb29970fcd684d0e7136a5947f92710332c98&code=4fda33093c9fc12711f1&state=ccc"
//access_token=f45feb6ff99f7b1be93d7dbcb8a4323431bc3321&scope=repo%2Cuser%3Aemail&token_type=bearer
//curl https://api.github.com/user -H "Authorization: token 620a4404e076f6cf1a10f9e00519924e43497091”
