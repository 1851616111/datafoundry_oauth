package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"

	"encoding/json"
	"fmt"
	gitlabapi "github.com/asiainfoLDP/datafactory_oauth2/gitlab"
	api "github.com/openshift/origin/pkg/user/api/v1"
	"strings"
)

var (
	gitlab  = gitlabapi.ClientFactory()
)

//curl http://127.0.0.1:9443/v1/gitlab  -d '{"host":"https://code.dataos.io", "user":"mengjing","private_token":"fXYznpUCTQQe5sjM4FWm"}' -H "Authorization:bearer HrA7qZo1MA7TKxYgbx7htR_9ez-FSXGuA8aM2fZzRC4"
func gitlabHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authorization := r.Header.Get("Authorization")

	option := new(gitLabInfo)
	if err := parseRequestBody(r, option); err != nil {
		retHttpCodef(400, w, "read request body err %v", err)
		return
	}

	// todo validate option
	ret_gb, ret_df := make(chan *gitlabDumpling, 1), make(chan *datafoundryDumpling, 1)

	go func(gitlab *gitlabapi.HttpFactory, option *gitLabInfo, ret *chan *gitlabDumpling) {
		f, err := gitlab.User(option.Host, option.PrivateToken).GetUser()

		*ret <- &gitlabDumpling{
			filling: f,
			err:     err,
		}
	}(gitlab, option, &ret_gb)

	go func(authorization string, ret *chan *datafoundryDumpling) {
		f, err := authDF(authorization)

		*ret <- &datafoundryDumpling{
			filling: f,
			err:     err,
		}
	}(authorization, &ret_df)

	count := 0
	var oUser *api.User
res:
	for {
		select {
		case dump := <-ret_df:
			count++
			if dump.err != nil {
				retHttpCodef(401, w, "unauthorized from datafoundry,  err %v\n", dump.err)
				return
			}
			oUser = dump.filling
		case dump := <-ret_gb:
			count++
			if dump.err != nil {
				retHttpCodef(401, w, "unauthorized from gitlab %s,  err %v\n", option.Host, dump.err)
				return
			}
		default:
			if count == 2 {
				break res
			}
		}
	}

	key := fmt.Sprintf("/df_service/%s/df_user/%s/oauth/gitlabs/info", DFHost_Key, oUser.Name)
	if err := db.set(key, option); err != nil {
		retHttpCodef(400, w, "store gitlab err %v", err.Error())
		return
	}

	retHttpCodef(200, w, "ok")
}

//curl http://127.0.0.1:9443/v1/gitlab/repos/owner -H "Authorization:bearer HrA7qZo1MA7TKxYgbx7htR_9ez-FSXGuA8aM2fZzRC4"
//curl http://127.0.0.1:9443/v1/gitlab/repos/org -H "Authorization:bearer HrA7qZo1MA7TKxYgbx7htR_9ez-FSXGuA8aM2fZzRC4"
func gitLabOwnerReposHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userType := ps.ByName("user")

	token := r.Header.Get("Authorization")
	var user *api.User
	var err error
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, w, "auth err %s\n", err.Error())
		return
	}

	option, err := getGitLabOptionByDFUser(user.Name)
	if err != nil {
		retHttpCodef(400, w, "get gitlab info err %v", err.Error())
		return
	}

	projects, err := gitlab.Project(option.Host, option.PrivateToken).ListProjects()
	if err != nil {
		retHttpCodef(400, w, "get projects err %v", err.Error())
		return
	}

	var l interface{}
	switch userType {
	case "org":
		l = gitlabapi.ConverOrgProjects(projects)
	case "owner":
		l = gitlabapi.ConverOwnerProjects(projects)

	}

	b, err := json.Marshal(l)
	if err != nil {
		retHttpCodef(400, w, "convert projects err %v", err)
		return
	}

	retHttpCodef(200, w, "%s", string(b))
}

//curl http://etcdsystem.servicebroker.dataos.io:2379/v2/keys/df_service/https:/lab.asiainfodata.com:8443/df_user/mengjing/oauth/gitlab_service/https:/code.dataos.io -u asiainfoLDP:6ED9BA74-75FD-4D1B-8916-842CB936AC1A
//curl http://127.0.0.1:9443/v1/gitlab/authorize/deploy -H "Authorization:bearer HrA7qZo1MA7TKxYgbx7htR_9ez-FSXGuA8aM2fZzRC4" -H "namespace:oauth" -d '{"host":"https://code.dataos.io","project_id":43}'
func gitLabSecretHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	namespace := strings.TrimSpace(r.Header.Get("namespace"))
	if namespace == "" {
		retHttpCode(400, w, "param namespace must not be nil.\n")
		return
	}

	authorization := r.Header.Get("Authorization")
	token := stripBearToken(authorization)

	var err error
	dfUser, err := authDF(authorization)
	if err != nil {
		retHttpCodef(401, w, "unauthorized from datafoundry, err %v\n", err)
		return
	}

	bind := new(gitLabBindInfo)
	if err := parseRequestBody(r, bind); err != nil {
		retHttpCodef(400, w, "read request body err %v", err)
		return
	}

	//df_service/https_lab.asiainfodata.com:8443/df_user/mengjing/oauth/gitlabs/info
	registry := fmt.Sprintf("/df_service/%s/df_user/%s/oauth/gitlabs/info", DFHost_Key, dfUser.Name)
	str, err := db.get(registry, true, true)
	if err != nil {
		retHttpCodef(400, w, "find gilab host err %v", err.Error())
		return
	}

	gitLab := new(gitLabInfo)
	if err := json.Unmarshal([]byte(str), gitLab); err != nil {
		retHttpCodef(400, w, "find gilab host err %v", err.Error())
	}

	if _, err := gitlab.User(gitLab.Host, gitLab.PrivateToken).GetUser(); err != nil {
		retHttpCodef(400, w, "validate gitlab private_token err %v", err.Error())
		return
	}

	ks, err := gitlab.DeployKey(gitLab.Host, gitLab.PrivateToken).ListKeys(bind.Id)
	if err != nil {
		retHttpCodef(400, w, "get gitlab deploy keys err %v", err.Error())
		return
	}

	//gitlab project deploykey title format
	//df_host---https_lab.asiainfo.com:8443---df_user---panxy
	commonKey := generateGitLabTitle(DFHost_Key, dfUser.Name)
	if len(ks) > 0 {
		ks = gitlabapi.FilterDeployKeysByTitle(ks, commonKey, strings.HasPrefix)
	}

	//df_service 区分不同环境可能使用统一DB造成数据错乱
	//df_user 区分不通fd用户不能使用相同的密钥对
	//gitlab_service 一个环境(ex. project, release, develop)的某个用户可以接入不同的私有gitlab

	keyPair := KeyPool.Pop()

	fmt.Printf("public:  %s---------->", keyPair.Public)
	fmt.Printf("private: %s------------>", keyPair.Private)

	if len(ks) == 0 {
		keyOption := new(gitlabapi.NewDeployKeyOption)
		keyOption.ProjectId = bind.Id
		keyOption.Param = gitlabapi.NewDeployKeyParam{
			Title: commonKey,
			Key:   string(keyPair.Public),
		}
		if err := gitlab.DeployKey(gitLab.Host, gitLab.PrivateToken).CreateKey(keyOption); err != nil {
			retHttpCodef(400, w, "create deploy key err %v", err.Error())
			return
		}

		//逻辑有问题
		key := fmt.Sprintf("/df_service/%s/df_user/%s/oauth/gitlab_service/%s/deploykey", DFHost_Key, dfUser.Name, etcdFormatUrl(bind.Host))
		if err := db.set(key, keyPair); err != nil {
			retHttpCodef(400, w, "save private key err %v", err.Error())
			return
		}
	}

	option := &SecretSSHOptions{
		NameSpace:        namespace,
		UserName:         dfUser.Name,
		SecretName:       generateGitlabName(dfUser.Name, Schemastripper(bind.Host)),
		DatafactoryToken: token,
		PrivateKey:       string(keyPair.Private),
	}

	if err := option.Validate(); err != nil {
		retHttpCodef(400, w, "validate datafoundry ssh secret option err %s\n", err.Error())
		return
	}

	if err := upsertSecret(option); err != nil {
		retHttpCodef(400, w, "create datafoundry ssh secret err %s\n", err.Error())
		return
	}



	retHttpCode(200, w, "ok")
}

func getGitLabOptionByDFUser(name string) (*gitLabInfo, error) {
	key := fmt.Sprintf("/df_service/%s/df_user/%s/oauth/gitlabs/info", DFHost_Key, name)
	s, err := db.get(key, true, false)
	if err != nil {
		return nil, err
	}

	option := new(gitLabInfo)
	if err := json.Unmarshal([]byte(s), option); err != nil {
		return nil, err
	}

	return option, nil
}

func generateGitLabTitle(dfHost, dfUser string) string {
	return fmt.Sprintf("df_host---%s---df_user---%s", dfHost, dfUser)
}
