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
	test_pub = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQClvGPYscP7sNf+ReBismT5lT1It6/s+c8jMUZD4cUqKDwRAKIXcKG93tuBf52CPlk+xt77KfaE0s7lFmaVvjVFCaaBG6314G4D1KZ0fwjacNQhCr6VR8QtJXXm4jucKCHnqNPxScnlB2mO9pcjT9p5yMbFJfMrb9zT7db766oNg9mBSWsnWN/1fmvII7BQxLG9L7LKzU/nqVsgc5wPPzonm307k4oMvuBtPWIi6gkzlPh/fboSmm0vFOVAwxUdUt3wClWlPnJdhlZDFqvzYXuIRX9gD93MD7X6XTA6CnzkN2J9p7g5EfT9bpM6hGghMrKd/Uufu5fXsQyeaJH8uv7v michael@localhost"
	test_pri = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEApbxj2LHD+7DX/kXgYrJk+ZU9SLev7PnPIzFGQ+HFKig8EQCi
F3Chvd7bgX+dgj5ZPsbe+yn2hNLO5RZmlb41RQmmgRut9eBuA9SmdH8I2nDUIQq+
lUfELSV15uI7nCgh56jT8UnJ5QdpjvaXI0/aecjGxSXzK2/c0+3W++uqDYPZgUlr
J1jf9X5ryCOwUMSxvS+yys1P56lbIHOcDz86J5t9O5OKDL7gbT1iIuoJM5T4f326
EpptLxTlQMMVHVLd8ApVpT5yXYZWQxar82F7iEV/YA/dzA+1+l0wOgp85Ddifae4
ORH0/W6TOoRoITKynf1Ln7uX17EMnmiR/Lr+7wIDAQABAoIBAQCjXcet2cwtVGwU
IBzGLMKLoif+fdHT7YnYTsHMN8d5fp92wwEDyeokloAYbgp8T6j40F8LhJmS45k9
B4+nGw63NoQBz57yNn87F2ncezvm1kDDMSwbSdp+Bebp5yaLDqQdDbWcqfdw4pWS
bk8cZ6IbWWVU/8tqjaFG5bJ8MBg3qKE0a6JnRuFpJDcxxvZQR/BeOl5xlktK/ei1
1clqfEt+4b8PSadpqdff38bflZa9KIaEZz2SKzwjyYPopeeA7xbcK41KkcSqTuh8
WJyBsCGPd0YmxuJe2/UL8UypzxglexmN1kAZeyFfGMATNRr4kvKjREAtmQwONhtZ
Vi6yCgHBAoGBANhR5bYYZaBjmyKnzxMuus0J5geVk6k9gw1LO+m8UQkKvohGmZV4
KKL6zRYbhi7JUesPWLxYxuk6tlcoHcjy8nAYy246djOjnlRk4PNPx4GuiT6ZVWht
39YuGY9NjsQR2c3j8nX20Sl23K0sP7SGxW1S7eF92dxC9+HtrNLyjsQ1AoGBAMQj
Ig4x8eoBA/hKybEb1PJ3dCnohHNE8xbgipoycbP4IWlNVEgixNZHg7Q9C5FMLcwM
IY80UpZAPlCFWOzGxagpODV6ERHBZ9qDD1uE1NXOWcsqqtMPAATsGa70re5x2XJz
60neWEaGDeQ7pW8aStxE2dK+hTerKQoqXvO6mJMTAoGAKFdtnX4DRdwNjHL7HTqz
v5U+/t8YQJGmJQ6Ix9hEzIjia4uvDL7x5SMcqCjN51/IFSwxgj6UKd63Lp3eoCEe
sWUOWyov7QVwe5CsmvOf40FnevMhiG4lNk42mhD+tPYXRlxiVTmIXFE8alc8MjCI
FRFIJ6tOu9MJY2rtthFiKpkCgYBfVzipr8uJVT8JxcjvB7lmt3xHFtizc5O6ziFx
vQ2aTwZmuok6m3QVOSQjS/1AfshQRKFXjDaNBOOFnpxQVHsmOAszq4d6mwoRpN2l
Phd7atgpMy9gcw0uV1pQum2F19+8i+6WtLcyaN190SSksiIrmmhL0gLNwaysXVZU
oaKi8wKBgGT4lduxopzdUnxkFI7hfT0Ejtmhqb3kOhSNOE0ataB5FD1E9CYlnzc2
ZzPJMtsBYPND4n60SSyiSfMu51CdcvSbpjLXvAMAAe89bOWjI2m0ZnXf+oXiX6qv
Qc3vEl/Rcrg4Ak0hs+xjUbUZ9Py2Q4ueU/g7TZgbFIHujCMU09Ee
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
