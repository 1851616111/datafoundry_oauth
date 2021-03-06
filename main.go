package main

import (
	"fmt"
	rsautil "github.com/asiainfoLDP/datafoundry_oauth2/util"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/cache"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/cache/redis"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/service"
	router "github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

var (
	tokenConfig                                           Config
	backingService_Redis                                  string
	GithubRedirectUrl, GithubClientID, GithubClientSecret string
	dbConf                                                storeConfig
	db                                                    Store
	DFHost_API                                            string
	DFHost_Key                                            string
	DF_API_Auth                                           string
	Redis_Addr, Redis_Port                                string
	Redis_Password, Redis_Cluster_Name                    string
	Cache                                                 cache.Cache
	CacheMan                                              cache.CacheMan
	KeyPool                                               *rsautil.Pool
)

func init() {

	initEnvs()
	backingService_Redis = RedisEnv.Get("Redis_BackingService_Name", nil)

	if RedisConfig, ok := <-service.NewBackingService(service.Redis, service.ValidateHP, checkRedis, service.ErrorBackingService).GetBackingServices(backingService_Redis); !ok {
		log.Fatal("init redis err")
	} else {
		Redis_Password = RedisConfig.Credential.Password
		log.Printf("redis url [%s@%s:%s/%s]", Redis_Password, Redis_Addr, Redis_Port, Redis_Cluster_Name)
		//Redis_Addr = "117.121.97.20"
		//Redis_Port = "9999"
	}

	initOauthConfig()

	initCache()
	initStorage()
	initOauth2Plugin()
	initDFHost()
	initAPI()
	initSSHKey()

}

func main() {

	runGitLabCacheController()
	log.Println("start gitlab cache contoller success")
	runGitHubCacheController()
	log.Println("start github cache contoller success")
	router := router.New()

	router.GET("/v1/repos/github-redirect", githubHandler)
	router.GET("/v1/repos/github/owner", githubOwnerReposHandler)
	router.GET("/v1/repos/github/orgs", githubOrgReposHandler)
	router.GET("/v1/repos/github/users/:user/repos/:repo", getGithubBranchHandler)

	router.POST("/v1/repos/gitlab", gitlabHandler)
	router.GET("/v1/repos/gitlab/:repo", gitLabOwnerReposHandler)
	router.GET("/v1/repos/gitlab/:repo/branches", gitLabBranchHandler)
	router.POST("/v1/repos/gitlab/authorize/deploy", gitLabSecretHandler)

	router.GET("/v1/repos/source/:source/webhooks", getWebHookHandler)
	router.POST("/v1/repos/source/:source/webhooks", createWebHookHandler)
	router.DELETE("/v1/repos/source/:source/webhooks", deleteWebHookHandler)

	router.POST("/v1/repos/gitlab/login", gitLabLoginHandler)

	log.Fatal(http.ListenAndServe(":9443", router))

	log.Println("service listen on :9443")
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
	dbConf = storeConfig{
		Addr:   httpAddrMaker(EtcdStorageEnv.Get("ETCD_HTTP_ADDR", nil)),
		Port:   EtcdStorageEnv.Get("ETCD_HTTP_PORT", nil),
		User:   EtcdStorageEnv.Get("ETCD_USER", nil),
		Passwd: EtcdStorageEnv.Get("ETCD_PASSWORD", nil),
	}

	refreshDB()
	fmt.Println("oauth init storage config success")
}

func initCache() {
	url := fmt.Sprintf("%s:%s", Redis_Addr, Redis_Port)
	Cache = redis.CreateCache(url, Redis_Password)
	CacheMan = cache.NewCacheMan(Cache)
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
	DFHost_API = DatafoundryEnv.Get("DATAFOUNDRY_HOST_ADDR", nil)
	DFHost_Key = etcdFormatUrl(DFHost_API)
}

func initAPI() {
	DF_API_Auth = DFHost_API + "/oapi/v1/users/~"
}

func initEnvs() {
	envNotNil := func(k string) {
		log.Fatalf("[Env] %s must not be nil.", k)
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

	RedisEnv.Init()
	RedisEnv.Print()
	RedisEnv.Validate(envNotNil)

}

func initSSHKey() {
	rsautil.Init("ssh-keygen")
	KeyPool = rsautil.NewKeyPool(10)
	go KeyPool.Run()
}

//https://github.com/login/oauth/authorize?client_id=2369ed831a59847924b4&scope=repo,user:email&state=ccc&redirect_uri=http://oauth2-oauth.app.asiainfodata.com/v1/github-redirect
//curl -v https://github.com/login/oauth/access_token -d "client_id=2369ed831a59847924b4&client_secret=510bb29970fcd684d0e7136a5947f92710332c98&code=4fda33093c9fc12711f1&state=ccc"
//access_token=f45feb6ff99f7b1be93d7dbcb8a4323431bc3321&scope=repo%2Cuser%3Aemail&token_type=bearer
//curl https://api.github.com/user -H "Authorization: token 620a4404e076f6cf1a10f9e00519924e43497091”

func checkRedis(svc service.Service) bool {
	const retryTimes = 3
	url := fmt.Sprintf("%s:%s", svc.Credential.Host, svc.Credential.Port)
	fmt.Printf("Redis Addr [%s]", url)
	for i := 1; i <= retryTimes; i++ {
		addr, port := getRedisMasterAddr(url, svc.Credential.Name)

		if len(addr) > 0 && len(port) > 0 {
			Redis_Addr, Redis_Port = addr, port
			log.Printf("dial redis[%s:%s] success", addr, port)
			return true
		}
		continue
	}

	return false
}

func fakeCheck(svc service.Service) bool {
	fmt.Printf("run fake check %v\n", svc)
	return true
}
