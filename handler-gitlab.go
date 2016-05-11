package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"

	"encoding/json"
	"fmt"
	gitlabapi "github.com/asiainfoLDP/datafactory_oauth2/gitlab"
	gitlabutil "github.com/asiainfoLDP/datafactory_oauth2/util"
	dfapi "github.com/openshift/origin/pkg/user/api/v1"
	"strconv"
	"strings"
)

var (
	glApi = gitlabapi.ClientFactory()
)

//curl http://127.0.0.1:9443/v1/gitlab  -d '{"host":"https://code.dataos.io", "user":"mengjing","private_token":"fXYznpUCTQQe5sjM4FWm"}' -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8"
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

//curl http://127.0.0.1:9443/v1/gitlab/repos/owner?page=1 -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8"
//curl http://127.0.0.1:9443/v1/gitlab/repos/org?page=2 -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8"
func gitLabOwnerReposHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userType := ps.ByName("user")

	token := r.Header.Get("Authorization")
	var user *dfapi.User
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

	p := r.FormValue("page")
	var page int

	if page, err = strconv.Atoi(p); err != nil {
		retHttpCodef(400, w, "invalide param page %d", p)
		return
	}
	if page < 1 {
		retHttpCodef(400, w, "invalide param page %d", p)
		return
	}

	projects, err := glApi.Project(option.Host, option.PrivateToken).ListProjects(uint32(page))
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

//curl http://127.0.0.1:9443/v1/gitlab/repo/43/branches -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8"
func gitLabBranchHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	repo := ps.ByName("repo")
	var projectId int
	var err error
	if projectId, err = strconv.Atoi(repo); err != nil {
		retHttpCode(400, w, "invalide param repo ")
		return
	}

	if projectId < 0 {
		retHttpCode(400, w, "invalide param repo ")
		return
	}

	token := r.Header.Get("Authorization")
	var user *dfapi.User
	if user, err = authDF(token); err != nil {
		retHttpCodef(401, w, "auth err %s\n", err.Error())
		return
	}

	option, err := getGitLabOptionByDFUser(user.Name)
	if err != nil {
		retHttpCodef(400, w, "get gitlab info err %v", err.Error())
		return
	}

	branches, err := glApi.Branch(option.Host, option.PrivateToken).ListBranches(projectId)
	if err != nil {
		retHttpCodef(400, w, "get project branches err %v", err.Error())
		return
	}

	b, err := json.Marshal(branches)
	if err != nil {
		retHttpCodef(400, w, "convert branches err %v", err)
		return
	}

	retHttpCodef(200, w, "%s", string(b))

}

//test case1: 同一用户针对不同gitlab
//test case2: 不同用户针对不同gitlab
//curl http://etcdsystem.servicebroker.dataos.io:2379/v2/keys/df_service/https:/lab.asiainfodata.com:8443/df_user/mengjing/oauth/gitlab_service/https:/code.dataos.io -u asiainfoLDP:6ED9BA74-75FD-4D1B-8916-842CB936AC1A
//curl http://127.0.0.1:9443/v1/gitlab/authorize/deploy -H "Authorization:bearer 7TlqnRS1S-x18MVqaKIhGRSvyTLhAd5t5Ca3JjH5Uu8" -H "namespace:oauth" -d '{"host":"https://code.dataos.io","project_id":43}'
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

	//根据当前oauth的配置的DataFoundry服务器,查询当前用户绑定的Gitlab服务HostA
	var gitLab *gitLabInfo
	if gitLab, err = getGitLabOptionByDFUser(dfUser.Name); err != nil {
		retHttpCodef(400, w, "find gilab host err %v", err.Error())
		return
	}

	//HostA与用户提供的Host不一致,应该返回并定失败
	if gitLab.Host != bind.Host {
		//todo 一个用户对应多主机情况,需要向党莎确认
		retHttpCodef(400, w, "unknow host %s %v", bind.Host, err.Error())
		return
	}

	//验证DF用户绑定gilab时的PrivateToken是否生效, 若{"message":"401 Unauthorized"},则PrivateToken失效
	ks, err := glApi.DeployKey(gitLab.Host, gitLab.PrivateToken).ListKeys(bind.Id)
	if err != nil {
		if gitlabapi.IsUnauthorized(err) {
			retHttpCodef(401, w, "gitlab %s private_token invalid\n", bind.Host)
			return
		}
		retHttpCodef(400, w, "get gitlab deploy keys err %v", err.Error())
		return
	}

	//gitlab project deploykey title format
	//df_host---https_lab.asiainfo.com:8443---df_user---panxy
	filter := generateGitLabTitle(DFHost_Key, dfUser.Name)
	if len(ks) > 0 {
		//过滤出DF_oauth2的对应的主机和用户名下面的
		ks = gitlabapi.FilterDeployKeysByTitle(ks, filter, strings.HasPrefix)
	}

	//df_service 区分不同环境可能使用统一DB造成数据错乱
	//df_user 区分不通DF用户不能使用相同的密钥对
	//gitlab_service 一个环境(ex. project, release, develop)的某个用户可以接入不同的私有gitlab

	//若gitlab的project上没有相应的deploykey,则在gitlab上的project创建deployke,并存储
	//若gitlab的project上面找到相应的deploykey,切存储中有记录,则不重新生成sshkey
	//若gitlab的project上面找到相应的deploykey,由于恶劣的情况导致存储丢失,暂时无能为力.

	errTryCount := 0
	var deployKey *gitlabutil.KeyPair
	var exist bool
	exist, deployKey = hasDeployKey(DFHost_Key, dfUser.Name, bind.Host)
retry:
	{
		if !exist {
			if deployKey == nil {
				//若deploy不存在,且过程无差错,则重新生成
				deployKey = KeyPool.Pop()
			} else {
				//查询过程出错(网络,数据库等一场),则直接返回,因为不确定是否存储中存在之前使用的deploykey,所以不能往下走
				errTryCount++
				if errTryCount == 2 {
					retHttpCodef(400, w, "get gitlab deploy key in store err %v", err.Error())
					return
				}
				goto retry
			}
		}
	}

	if len(ks) == 0 {
		//todo make it transaction
		keyOption := &gitlabapi.NewDeployKeyOption{
			ProjectId: bind.Id,
			Param: gitlabapi.NewDeployKeyParam{
				Title: filter,
				Key:   string(deployKey.Public),
			},
		}

		if err := glApi.DeployKey(gitLab.Host, gitLab.PrivateToken).CreateKey(keyOption); err != nil {
			retHttpCodef(400, w, "create deploy key err %v", err.Error())
			return
		}

		if err := setDeployKey(DFHost_Key, dfUser.Name, bind.Host, deployKey); err != nil {
			retHttpCodef(400, w, "save private key err %v", err.Error())
			return
		}
	}

	option := &SecretSSHOptions{
		NameSpace:        namespace,
		UserName:         dfUser.Name,
		SecretName:       generateGitlabName(dfUser.Name, Schemastripper(bind.Host)),
		DatafactoryToken: token,
		PrivateKey:       string(deployKey.Private),
	}

	if err := option.Validate(); err != nil {
		retHttpCodef(400, w, "validate datafoundry ssh secret option err %s\n", err.Error())
		return
	}

	if err := upsertSecret(option); err != nil {
		retHttpCodef(400, w, "create datafoundry ssh secret err %s\n", err.Error())
		return
	}

	retHttpCode(200, w, fmt.Sprintf("{\"secret\":\"%s\"}", option.SecretName))
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
