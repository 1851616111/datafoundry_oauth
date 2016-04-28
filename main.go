package main

import (
	"fmt"
	router "github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

var (
	tokenConfig                                           Config
	GithubRedirectUrl, GithubClientID, GithubClientSecret string
	db                                                    Store
	DFHost                                                string
	DF_API_Auth                                           string
)

func init() {
	initEnvs()
	initOauthConfig()

	initStorage()
	initOauth2Plugin()
	initDFHost()
	initAPI()
}

func main() {
	router := router.New()

	router.GET("/v1/github-redirect", githubHandler)
	router.GET("/v1/repos/github/owner", githubOwnerReposHandler)
	router.GET("/v1/repos/github/orgs", githubOrgReposHandler)
	router.GET("/v1/repos/github/users/:user/repos/:repo", getGithubBranchHandler)

	router.POST("/v1/gitlab", gitlabHandler)
	router.GET("/v1/gitlab/repos/:user", gitLabOwnerReposHandler)
	router.POST("/v1/gitlab/authorize/deploy", gitLabSecretHandler)

	log.Fatal(http.ListenAndServe(":9443", router))

}

func initOauthConfig() {
	var err error
	tokenConfig, err = NewGitHub(GithubClientID, GithubClientSecret, GithubRedirectUrl, []string{"repo", "user:email"})
	if err != nil {
		fmt.Errorf("oauth init fail %s\n", err.Error())

	}

	fmt.Println("oauth config init success")
}

func initStorage() {
	c := storeConfig{
		Addr:   httpAddrMaker(EtcdStorageEnv.Get("ETCD_HTTP_ADDR", nil)),
		Port:   EtcdStorageEnv.Get("ETCD_HTTP_PORT", nil),
		User:   EtcdStorageEnv.Get("ETCD_USER", nil),
		Passwd: EtcdStorageEnv.Get("ETCD_PASSWORD", nil),
	}

	db = c.newClient()
	fmt.Println("oauth init storage config success")
}

func initOauth2Plugin() {
	initGithubPlugin()
}

func initGithubPlugin() {
	GithubRedirectUrl = GithubApplicationEnv.Get("GITHUB_REDIRECT_URL", nil)
	GithubClientID = GithubApplicationEnv.Get("GITHUB_CLIENT_ID", nil)
	GithubClientSecret = GithubApplicationEnv.Get("GITHUB_CLIENT_SECRET", nil)
}

func initDFHost() {
	DFHost = DatafoundryEnv.Get("DATAFACTORY_HOST_ADDR", nil)
}

func initAPI() {
	DF_API_Auth = DFHost + "/oapi/v1/users/~"
}

func initEnvs() {
	envNotNil := func(k string) {
		fmt.Errorf("[Env] %s must not be nil.", k)
	}

	EtcdStorageEnv.Init()
	EtcdStorageEnv.Print()
	EtcdStorageEnv.Validate(envNotNil)

	GithubApplicationEnv.Init()
	GithubApplicationEnv.Print()
	GithubApplicationEnv.Validate(envNotNil)

	DatafoundryEnv.Init()
	DatafoundryEnv.Print()
	DatafoundryEnv.Validate(envNotNil)
}

//https://github.com/login/oauth/authorize?client_id=2369ed831a59847924b4&scope=repo,user:email&state=ccc&redirect_uri=http://oauth2-oauth.app.asiainfodata.com/v1/github-redirect
//curl -v https://github.com/login/oauth/access_token -d "client_id=2369ed831a59847924b4&client_secret=510bb29970fcd684d0e7136a5947f92710332c98&code=4fda33093c9fc12711f1&state=ccc"
//access_token=f45feb6ff99f7b1be93d7dbcb8a4323431bc3321&scope=repo%2Cuser%3Aemail&token_type=bearer
//curl https://api.github.com/user -H "Authorization: token 620a4404e076f6cf1a10f9e00519924e43497091”
