package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"

	"bytes"
	"encoding/json"
	"fmt"
	gitlabapi "github.com/asiainfoLDP/datafoundry_oauth2/gitlab"
	//gitlabutil "github.com/asiainfoLDP/datafoundry_oauth2/util"
	dfapi "github.com/openshift/origin/pkg/user/api/v1"
	"log"
	"strconv"
	"strings"
)

var (
	glApi = gitlabapi.ClientFactory()
)

//curl http://127.0.0.1:9443/v1/repos/gitlab  -d '{"host":"https://code.dataos.io", "user":"root","private_token":"hvXbXHKTPNxqzUDuSyLw"}' -H "Authorization:bearer i1TerZwHQSsveIrHs53wr6lKdzxbJL2mVNCu8fs5Ao0"
//curl http://127.0.0.1:9443/v1/repos/gitlab  -d '{"host":"https://code.dataos.io", "user":"mengjing","private_token":"fXYznpUCTQQe5sjM4FWm"}' -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8"
func gitlabHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	authorization := r.Header.Get("Authorization")

	option := new(gitLabInfo)
	if err := parseRequestBody(r, option); err != nil {
		retHttpCodef(400, 1400, w, "read request body err %v", err)
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
	}(glApi, option, &ret_gb)

	go func(authorization string, ret *chan *datafoundryDumpling) {
		f, err := authDF(authorization)

		*ret <- &datafoundryDumpling{
			filling: f,
			err:     err,
		}
	}(authorization, &ret_df)

	count := 0
	var oUser *dfapi.User
res:
	for {
		select {
		case dump := <-ret_df:
			count++
			if dump.err != nil {
				retHttpCodef(400, 401, w, "unauthorized from datafoundry,  err %v", dump.err)
				return
			}
			oUser = dump.filling
		case dump := <-ret_gb:
			count++
			if dump.err != nil {
				retHttpCodef(400, 1401, w, "unauthorized from gitlab %s,  err %v", option.Host, dump.err)
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
		retHttpCodef(400, 1400, w, "store gitlab err %v", err.Error())
		return
	}

	retHttpCodef(200, 1200, w, "ok")
}

//curl http://127.0.0.1:9443/v1/repos/gitlab/owner -H "Authorization:Bearer twizX0NaWxdbtoFhD7wvH5L3ioClX6iSBVaF83cuAes"
//curl http://127.0.0.1:9443/v1/repos/gitlab/orgs -H "Authorization:Bearer i1TerZwHQSsveIrHs53wr6lKdzxbJL2mVNCu8fs5Ao0"
//curl http://127.0.0.1:9443/v1/repos/gitlab/orgs -H "Authorization:Bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8"
func gitLabOwnerReposHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userType := ps.ByName("repo")

	token := r.Header.Get("Authorization")
	var user *dfapi.User
	var err error
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, 1401, w, "auth err %s", err.Error())
		return
	}

	option, err := getGitLabOptionByDFUser(user.Name)
	if err != nil {
		if EtcdKeyNotFound(err) {
			retHttpCode(400, 1401, w, "unauthorized")
		} else {
			retHttpCodef(400, 1400, w, "get user info err %s", err.Error())
		}

		return
	}

	projects := []gitlabapi.Project{}

	b, err := Cache.HFetch("host_"+option.Host, "user_"+option.User)
	if err != nil {
		retHttpCodef(400, 1400, w, "get projects(cached) err %v", err.Error())
		return
	}

	json.Unmarshal(b, &projects)

	//projects, err := glApi.Project(option.Host, option.PrivateToken).ListProjects()
	//if err != nil {
	//	retHttpCodef(400, 1400, w, "get projects err %v", err.Error())
	//	return
	//}

	var p interface{}
	switch userType {
	case "orgs":
		p = gitlabapi.ConverOrgProjects(projects)
	case "owner":
		p = gitlabapi.ConverOwnerProjects(projects)
	}

	type ret struct {
		Host string      `json:"host"`
		Info interface{} `json:"infos"`
	}

	rt := ret{
		Host: option.Host,
		Info: p,
	}

	b, err = json.Marshal(rt)
	if err != nil {
		retHttpCodef(400, 1400, w, "convert projects err %v", err)
		return
	}

	retHttpCodeJson(200, 1200, w, string(b))
}

//curl http://127.0.0.1:9443/v1/gitlab/repo/43/branches -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8"
func gitLabBranchHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	repo := ps.ByName("repo")
	var projectId int
	var err error
	if projectId, err = strconv.Atoi(repo); err != nil {
		retHttpCode(400, 1400, w, "invalide param repo ")
		return
	}

	if projectId < 0 {
		retHttpCode(400, 1400, w, "invalide param repo ")
		return
	}

	token := r.Header.Get("Authorization")
	var user *dfapi.User
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, 1401, w, "auth err %s", err.Error())
		return
	}

	option, err := getGitLabOptionByDFUser(user.Name)
	if err != nil {
		retHttpCodef(400, 1400, w, "get gitlab info err %v", err.Error())
		return
	}

	branches, err := glApi.Branch(option.Host, option.PrivateToken).ListBranches(projectId)
	if err != nil {
		retHttpCodef(400, 1400, w, "get project branches err %v", err.Error())
		return
	}

	b, err := json.Marshal(branches)
	if err != nil {
		retHttpCodef(400, 1400, w, "convert branches err %v", err)
		return
	}

	retHttpCodeJson(200, 1200, w, string(b))

}

//curl http://etcdsystem.servicebroker.dataos.io:2379/v2/keys/oauth/deploykeys/gitlab  -u asiainfoLDP:6ED9BA74-75FD-4D1B-8916-842CB936AC1A
//curl http://127.0.0.1:9443/v1/repos/gitlab/authorize/deploy -H "Authorization:Bearer DWqXQ0N0YhFqBfjIyT0oHpxcTtIwR9nmCpvMaqUKx70" -H "namespace:oauth" -d '{"host":"https://code.dataos.io","project_id":43}'
func gitLabSecretHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	namespace := strings.TrimSpace(r.Header.Get("namespace"))
	if namespace == "" {
		retHttpCode(400, 1400, w, "param namespace must not be nil.")
		return
	}

	authorization := r.Header.Get("Authorization")
	//token := stripBearToken(authorization)

	var err error
	dfUser, err := authDF(authorization)
	if err != nil {
		retHttpCodef(401, 1401, w, "unauthorized from datafoundry, err %v", err)
		return
	}

	bind := new(gitLabBindInfo)
	if err := parseRequestBody(r, bind); err != nil {
		retHttpCodef(400, 1400, w, "read request body err %v", err)
		return
	}

	//根据当前oauth的配置的DataFoundry服务器,查询当前用户绑定的Gitlab服务HostA
	var gitLab *gitLabInfo
	if gitLab, err = getGitLabOptionByDFUser(dfUser.Name); err != nil {
		retHttpCodef(400, 1400, w, "find gilab host err %v", err.Error())
		return
	}

	//HostA与用户提供的Host不一致,应该返回并定失败
	if gitLab.Host != bind.Host {
		retHttpCodef(400, 1400, w, "unknow host %s %v", bind.Host, err.Error())
		return
	}

	//验证DF用户绑定gilab时的PrivateToken是否生效, 若{"message":"401 Unauthorized"},则PrivateToken失效
	ks, err := glApi.DeployKey(gitLab.Host, gitLab.PrivateToken).ListKeys(bind.Id)
	if err != nil {
		if gitlabapi.IsUnauthorized(err) {
			retHttpCodef(401, 1401, w, "gitlab %s private_token invalid", bind.Host)
			return
		}
		retHttpCodef(400, 1400, w, "get gitlab deploy keys err %v", err.Error())
		return
	}

	const DeployKeyTitle = "DataFoundry@ci.dataos.io"
	create := func(w http.ResponseWriter) {
		deployKey := KeyPool.Pop()
		fmt.Printf("generage deploy keys:\n%s", deployKey)

		keyOption := &gitlabapi.NewDeployKeyOption{
			ProjectId: bind.Id,
			Param: gitlabapi.NewDeployKeyParam{
				Title: DeployKeyTitle,
				Key:   string(deployKey.Public),
			},
		}

		keyId := getMd5(deployKey.Public)
		secretName := generateReposDeployName("gitlab", Schemastripper(bind.Host))

		pair := Pair{
			DFSecret:   secretName,
			PrivateKey: base64Encode(deployKey.Private),
		}

		if err := glApi.DeployKey(gitLab.Host, gitLab.PrivateToken).CreateKey(keyOption); err != nil {
			retHttpCodef(400, 1400, w, "create deploy key err %v", err.Error())
			return
		}

		if err := setDeployKey("gitlab", keyId, pair); err != nil {
			retHttpCodef(400, 1400, w, "save private key err %v", err.Error())
			return
		}

		option := &SecretSSHOptions{
			NameSpace:        namespace,
			UserName:         dfUser.Name,
			SecretName:       secretName,
			DataFoundryToken: stripBearToken(authorization),
			PrivateKey:       string(deployKey.Private),
		}

		if err := upsertSecret(option); err != nil {
			retHttpCodef(400, 1400, w, "create datafoundry ssh secret err %s", err.Error())
			return
		}

		retHttpCodeJson(200, 1200, w, fmt.Sprintf("{\"secret\":\"%s\"}", option.SecretName))
	}

	if len(ks) > 0 {
		ks = gitlabapi.FilterDeployKeysByTitle(ks, gitlabapi.Equals, DeployKeyTitle)
	}

	if len(ks) == 1 {

		id := getTextId(ks[0].Key)

		key, err := getDeployKey("gitlab", id)
		if err != nil {
			if EtcdKeyNotFound(err) {
				if err := glApi.DeployKey(gitLab.Host, gitLab.PrivateToken).DeleteKey(bind.Id, ks[0].Id); err != nil {
					retHttpCodef(400, 1400, w, "delete deploy key err %v", err.Error())
				} else {
					//if err := deleteSecret(option); err != nil {
					//	log.Printf("delete old secret err %v", err)
					//}
					//若etcd内容查找不到, 清空deploykey, 清理环境走新建流程
					create(w)
				}
				return
			}
			retHttpCodef(400, 1400, w, "get deploy key err %v\n", err.Error())
			return
		}

		pair, err := parsePair(key)
		if err != nil {
			retHttpCodef(400, 1400, w, "parse deploy key err %v\n", err.Error())
			return
		}

		pk, err := base64Decode(pair.PrivateKey)
		if err != nil {
			retHttpCodef(400, 1400, w, "parse deploy key err %v\n", err.Error())
			return
		}

		option := &SecretSSHOptions{
			NameSpace:        namespace,
			UserName:         dfUser.Name,
			SecretName:       pair.DFSecret,
			DataFoundryToken: stripBearToken(authorization),
			PrivateKey:       string(pk),
		}

		if err := upsertSecret(option); err != nil {
			retHttpCodef(400, 1400, w, "create datafoundry ssh secret err %s", err.Error())
			return
		}
		retHttpCodeJson(200, 1200, w, fmt.Sprintf("{\"secret\":\"%s\"}", option.SecretName))
		return
	}

	if len(ks) == 0 {
		create(w)
		return
	}
}

//curl -XPOST  http://127.0.0.1:9443/v1/repos/gitlab/login?host=https://code.dataos.io\&username=panxy3\&password=eadsch6ju -H "Authorization:Bearer i1TerZwHQSsveIrHs53wr6lKdzxbJL2mVNCu8fs5Ao0"
func gitLabLoginHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	host, username, password := strings.TrimSpace(r.FormValue("host")), strings.TrimSpace(r.FormValue("username")), r.FormValue("password")
	if len(host) == 0 {
		retHttpCode(400, 1400, w, "param host must not nil.")
		return
	}
	if len(username) == 0 {
		retHttpCode(400, 1400, w, "param username must not nil.")
		return
	}

	authorization := r.Header.Get("Authorization")
	if len(authorization) == 0 {
		retHttpCode(401, 401, w, "header Authorization must not nil.")
		return
	}

	s, err := glApi.Session(host, username, password).PostSession()
	if err != nil {
		log.Printf("[POST]/v1/repos/gitlab/login. user %s err %v\n", username, err)
		retHttpCode(401, 1401, w, "unauthorized")
		return
	}

	option := &gitLabInfo{
		Host:         host,
		User:         username,
		PrivateToken: s.PrivateToken,
	}

	b, err := json.Marshal(option)
	if err != nil {
		retHttpCode(400, 1400, w, err)
		return
	}

	nr, err := http.NewRequest("POST", "/v1/repos/gitlab", bytes.NewReader(b))
	if err != nil {
		retHttpCode(400, 1400, w, err)
		return
	}
	nr.Header.Set("Authorization", authorization)

	gitlabHandler(w, nr, nil)
	return
}

func getGitLabOptionByDFUser(name string) (*gitLabInfo, error) {
	key := fmt.Sprintf("/df_service/%s/df_user/%s/oauth/gitlabs/info", DFHost_Key, name)
	s, err := db.getValue(key)
	if err != nil {
		return nil, err
	}

	option := new(gitLabInfo)
	if err := json.Unmarshal([]byte(s), option); err != nil {
		return nil, err
	}

	return option, nil
}

func getTextId(privateKey string) string {
	p := []byte(privateKey)
	//增加回车
	p = append(p, 0xa)
	return getMd5(p)
}

type Pair struct {
	DFSecret   string
	PrivateKey string
}

func parsePair(pairStr string) (*Pair, error) {
	pair := new(Pair)
	if err := json.Unmarshal([]byte(pairStr), pair); err != nil {
		return nil, err
	}

	return pair, nil
}
