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

const (
	test_pub = "ssh-rsa MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAoP/iNBFmX9xf7tXeXjD3Sjg6o23dYVQZFsqAutXfKrR8jlU4Y9Na+NjbG9EDmzjvdeJnOht7rA/LLem4SUVnmimpji4YyTb6G0qUAIUzLOes328tpAOXOIYVglv+LEhHbNSUc0yFEvsJK3x+Fkl4TKye/a2Z2J5eRhUe2LKJvX7l7dawMw7vBRWXgQg2O9yr2vS7ylZVGJvyKUsj2rjvHvPLMn/SL+IPQLaGTRgNM1Eu7fcDkdniv9hqp1dz9l27akSod6Jr5ozNPdcZhScQxseVAMHa+Uk4fTr1Y+t7pi/4SsgS0q7qorzxGQaPGdLMgsl1FZ4l5NF/25AVPpgNUQIDAQAB"
	test_pri = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAoP/iNBFmX9xf7tXeXjD3Sjg6o23dYVQZFsqAutXfKrR8jlU4
Y9Na+NjbG9EDmzjvdeJnOht7rA/LLem4SUVnmimpji4YyTb6G0qUAIUzLOes328t
pAOXOIYVglv+LEhHbNSUc0yFEvsJK3x+Fkl4TKye/a2Z2J5eRhUe2LKJvX7l7daw
Mw7vBRWXgQg2O9yr2vS7ylZVGJvyKUsj2rjvHvPLMn/SL+IPQLaGTRgNM1Eu7fcD
kdniv9hqp1dz9l27akSod6Jr5ozNPdcZhScQxseVAMHa+Uk4fTr1Y+t7pi/4SsgS
0q7qorzxGQaPGdLMgsl1FZ4l5NF/25AVPpgNUQIDAQABAoIBAHPaDbT3/Fnoo2Oi
pCPVSm0u7wshCJd7w54B1AYd1jvNqn9lVXGH6kN9EJYAnn64xp9mbm+CzUhwCP2d
3A1lkvp9FlSIWS+ZjvnKfZuuPbgHcf9J6mbGaq+y1JF8jvbgf/0RL02Ud4HEAMI5
ECYLcjSCVtombLlSpHX7xrmaJ8GiBZNSZp/3znsu8IbXOHEH85pgeny59Rl34O7u
mghcf8PEGhisuLCUmBL7hr1XHyPy360q1KuDul9FFbILCMG8aNBlCwX/1yG+UwJD
hJId2GpeW2iQQMubAFsQ1D+FKbSQjvn2/NssJfggxugjXOHuPLomCc9PClre5cHM
1jsthhECgYEAxNrxT3R2M7+OBE3ArUD0M/TgGoao7UeEKwxTM8aHSMM0XKBq2Gd8
+uVfLbFMoUXeOU39F0S2zpdwOLSITuXvBnxm0gpQLW6uT3mp/4VDq36Jr9VFzQ4K
rUl1i6B16E23WiFoCsxIrbsUyKtdfLvDTiORd1Z1DwzMW6x16Z3L6r0CgYEA0V8c
isOcMqdsgqZx++Ye1K4QZBygzpuDpk9ijt3TP55DYyUcwhRzhAAfsJaNLo7/3h01
LOdCXnDQhkodxX4D+wW1hK4Lm79ZL7JYp6VcfCw7dB95KLhMqdF7lEQuhuDANev6
x+40a5lQn8Mmy14Bbi6m1bcqS9IMJABF814doCUCgYAa+4OuB2GYUD5QGrQ5Szjt
0jfRivmmpHHaULMq2qB6eb84nwhmJzE7VqtIIRBG3sPKCQWS5elEwf8w1pYEcoHj
2rNhQOaig5RC8oM5sfOHky2eO1Z4997Ax9vjype+wsBKC2AucrfXkFgV9V84FKh9
kmSC/gfHi1KLkkULQ4TK5QKBgQDCjfI00+46d69yfH6gx8bQdOsQTDX1pzcffNcl
0OVzUXpnD954To7FE2RfMJcCs6j52gRGtKLMpWJv10Fw+ldylGyHXT+2O4oBs2WE
aznUvTmF/5UTjKbYiqueK/lcJk8WDDFeRXB6p93uh2ZuRe1oWHt5TppEGGxlq8dU
jZlT7QKBgQCwx9j8LqU8nNIxjhJG6KJZjIBoUnV4n75DKcwsjeRIqp6B7poddIhw
l/mcd0p9pYw1XeCr84ihdPLdqZFb/S1qTPVrW6qAVjvE06skgmBgIW4F4vwlVT6M
rq3db1JMLffc8zOi4zTVzcazcX9MkGSKNXaGNFJcFiZpo5skX1AOHg==
-----END RSA PRIVATE KEY-----
`
)

var (
	gitlab  = gitlabapi.ClientFactory()
	testKey = deployKey{
		Private: test_pri,
		Public:  test_pub,
	}
)

//curl http://127.0.0.1:9443/v1/gitlab  -d '{"host":"https://code.dataos.io", "user":"mengjing","private_token":"fXYznpUCTQQe5sjM4FWm"}' -H "Authorization:bearer uEgOIepT95OkbbNFY9zQcTO_8Ae445fjrBD9uLPGEKc"
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

//curl http://127.0.0.1:9443/v1/gitlab/repos/owner -H "Authorization:bearer uEgOIepT95OkbbNFY9zQcTO_8Ae445fjrBD9uLPGEKc"
//curl http://127.0.0.1:9443/v1/gitlab/repos/org -H "Authorization:bearer uEgOIepT95OkbbNFY9zQcTO_8Ae445fjrBD9uLPGEKc"
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
//curl http://127.0.0.1:9443/v1/gitlab/authorize/deploy -H "Authorization:bearer uEgOIepT95OkbbNFY9zQcTO_8Ae445fjrBD9uLPGEKc" -H "namespace:oauth" -d '{"host":"https://code.dataos.io","project_id":43}'
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

	if len(ks) == 0 {
		keyOption := new(gitlabapi.NewDeployKeyOption)
		keyOption.ProjectId = bind.Id
		keyOption.Param = gitlabapi.NewDeployKeyParam{
			Title: commonKey,
			Key: test_pub,
		}
		if err := gitlab.DeployKey(gitLab.Host, gitLab.PrivateToken).CreateKey(keyOption); err != nil {
			retHttpCodef(400, w, "create deploy key err %v", err.Error())
			return
		}

		key := fmt.Sprintf("/df_service/%s/df_user/%s/oauth/gitlab_service/%s/deploykey", DFHost_Key, dfUser.Name, etcdFormatUrl(bind.Host))
		if err := db.set(key, testKey); err != nil {
			retHttpCodef(400, w, "save private key err %v", err.Error())
			return
		}
	}

	option := &SecretSSHOptions{
		NameSpace:        namespace,
		UserName:         dfUser.Name,
		SecretName:       generateGitlabName(dfUser.Name, Schemastripper(bind.Host)),
		DatafactoryToken: token,
		PrivateKey:       test_pri,
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
