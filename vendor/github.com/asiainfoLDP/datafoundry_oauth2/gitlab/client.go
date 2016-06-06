package gitlab

import (
	"encoding/json"
	"fmt"
	httpuitl "github.com/asiainfoLDP/datafoundry_oauth2/util/http"
	"strings"
)

const (
	Gitlab_Credential_Key        = "PRIVATE-TOKEN"
	GitLab_Api_Url_Path_User     = "/api/v3/user"
	GitLab_Api_Url_Path_Project  = "/api/v3/projects"
	GitLab_Api_Url_Path_Keys     = "/api/v3/projects/%d/keys"
	GitLab_Api_Url_Path_Branch   = "/api/v3/projects/%d/repository/branches"
	GitLab_Api_Url_Path_CallBack = "/users/auth/ldapmain/callback"
	GitLab_Api_Url_Path_Session  = "/api/v3/session"
	GitLab_Api_Url_Path_WebHooks = "/api/v3/projects/%s/hooks"
)

type Client interface {
	UserInterface
	//Groups
	ProjectInterface
	BranchInterface
	DeployKeyInterface
	CallBackInterface
	SessionInterface
	WebHookInterface
}

//--------------------- User ---------------------
type UserInterface interface {
	User(host, privateToken string) Users
}

func (f *HttpFactory) User(host, privateToken string) Users {
	return f.newClient(host, GitLab_Api_Url_Path_User, privateToken)
}

type Users interface {
	GetUser() (*User, error)
}

func (c *RestClient) GetUser() (*User, error) {
	user := new(User)
	if err := c.Client.GetJson(user, c.Url, c.Credential.Key, c.Credential.Value); err != nil {
		return nil, err
	}

	return user, nil
}

//--------------------- Groups ---------------------
//type Groups interface {
//	ListGroups() ([]Group, error)
//}
//
//func (r *RestClient) ListGroups() ([]Group, error) {
//	groups := []Group{}
//	if err := r.Client.GetJson(groups, r.Url, r.Credential.Key, r.Credential.Value); err != nil {
//		return nil, err
//	}
//
//	return groups, nil
//}

//--------------------- Projects ---------------------
type ProjectInterface interface {
	Project(host, privateToken string) Projects
}

func (f *HttpFactory) Project(host, privateToken string) Projects {
	return f.newClient(host, GitLab_Api_Url_Path_Project, privateToken)
}

type Projects interface {
	ListProjects() ([]Project, error)
}

func (c *RestClient) ListProjects() ([]Project, error) {
	projects := []Project{}

	page := 1
	for {
		ps := new([]Project)
		url := fmt.Sprintf("%s?page=%d", c.Url, page)
		if err := c.Client.GetJson(ps, url, c.Credential.Key, c.Credential.Value); err != nil {
			return projects, err
		}
		if len(*ps) > 0 {
			projects = append(projects, *ps...)
			page = page + 1
			continue
		}

		return projects, nil
	}

}

//--------------------- Branches ---------------------
type BranchInterface interface {
	Branch(host, privateToken string) Branches
}

func (f *HttpFactory) Branch(host, privateToken string) Branches {
	return f.newClient(host, GitLab_Api_Url_Path_Branch, privateToken)
}

type Branches interface {
	ListBranches(projectId int) ([]Branch, error)
}

func (c *RestClient) ListBranches(projectId int) ([]Branch, error) {
	c.Url = fmt.Sprintf(c.Url, projectId)

	branches := new([]Branch)
	if err := c.Client.GetJson(branches, c.Url, c.Credential.Key, c.Credential.Value); err != nil {
		return nil, err
	}

	return *branches, nil
}

//--------------------- DeployKeys ---------------------
type DeployKeyInterface interface {
	DeployKey(host, privateToken string) DeployKeys
}

func (f *HttpFactory) DeployKey(host, privateToken string) DeployKeys {
	return f.newClient(host, GitLab_Api_Url_Path_Keys, privateToken)
}

type DeployKeys interface {
	ListKeys(projectId int) ([]DeployKey, error)
	CreateKey(option *NewDeployKeyOption) error
	DeleteKey(projectId, id int) error
}

func (c *RestClient) ListKeys(projectId int) ([]DeployKey, error) {
	keys := new([]DeployKey)
	url := fmt.Sprintf(c.Url, projectId)
	if err := c.Client.GetJson(keys, url, c.Credential.Key, c.Credential.Value); err != nil {
		return nil, err
	}

	return *keys, nil
}

func (c *RestClient) CreateKey(option *NewDeployKeyOption) error {
	url := fmt.Sprintf(c.Url, option.ProjectId)
	b, err := c.Client.Encode(option.Param)
	if err != nil {
		return err
	}

	_, err = c.Client.Post(url, "application/json", b, c.Credential.Key, c.Credential.Value)
	return err
}

func (c *RestClient) DeleteKey(projectId, id int) error {
	url := fmt.Sprintf(c.Url, projectId) + fmt.Sprintf("/%d", id)
	_, err := c.Client.Delete(url, c.Credential.Key, c.Credential.Value)

	return err
}

func (f *HttpFactory) newCallBackClient(host, api, username, password string) *RestClient {
	return &RestClient{
		Url:    fmt.Sprintf("%s?username=%s&password=%s", UrlMaker(host, api), username, password),
		Client: f,
	}
}

//callback is used for ldap to init login info
//--------------------- CallBack ---------------------
type CallBackInterface interface {
	CallBack(host, username, password string) CallBacks
}

type CallBacks interface {
	PostCallBack() (*CallBack, error)
}

func (f *HttpFactory) CallBack(host, user, password string) CallBacks {
	return f.newCallBackClient(host, GitLab_Api_Url_Path_CallBack, user, password)
}

func (c *RestClient) PostCallBack() (*CallBack, error) {
	b, err := c.Client.Post(c.Url, "application/json", []byte{})
	if err != nil {
		return nil, err
	}

	s := new(CallBack)
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}

	return s, nil
}

func (f *HttpFactory) newSessionClient(host, api, login, password string) *RestClient {
	return &RestClient{
		Url:    fmt.Sprintf("%s?login=%s&password=%s", UrlMaker(host, api), login, password),
		Client: f,
	}
}

//--------------------- Session ---------------------
type SessionInterface interface {
	Session(host, login, password string) Sessions
}

type Sessions interface {
	PostSession() (*Session, error)
}

func (f *HttpFactory) Session(host, login, password string) Sessions {
	return f.newSessionClient(host, GitLab_Api_Url_Path_Session, login, password)
}

func (c *RestClient) PostSession() (*Session, error) {
	b, err := c.Client.Post(c.Url, "application/json", []byte{})
	if err != nil {
		return nil, err
	}

	s := new(Session)
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}

	return s, nil
}

//--------------------- WebHook ---------------------
type WebHookInterface interface {
	WebHook(host, privateToken string) WebHooks
}

type WebHooks interface {
	CreateWebHook(projectId string, p httpuitl.Param) (*WebHookParam, error)
	UpdateWebHook(projectId string, id int, p httpuitl.Param) (*WebHookParam, error)
	DeleteWebHook(projectId string, id int) error
}

func (f *HttpFactory) WebHook(host, privateToken string) WebHooks {
	return f.newClient(host, GitLab_Api_Url_Path_WebHooks, privateToken)
}

func (c *RestClient) CreateWebHook(projectId string, params httpuitl.Param) (*WebHookParam, error) {
	return create(c, projectId, params)
}

func (c *RestClient) UpdateWebHook(projectId string, id int, params httpuitl.Param) (*WebHookParam, error) {
	return update(c, projectId, id, params)
}

func (c *RestClient) DeleteWebHook(projectId string, id int) error {
	return delete(c, projectId, id)
}

func create(c *RestClient, projectId string, p httpuitl.Param) (*WebHookParam, error) {
	url := fmt.Sprintf(c.Url, projectId) + fmt.Sprintf("/")

	var body interface{}
	var err error

	var bodyType string
	switch bodyType = p.GetBodyType(); bodyType {
	case "application/x-www-form-urlencoded":
		body = strings.NewReader(fmt.Sprint(p.GetParam()))
	case "application/json":
		fallthrough
	default:
		body, err = json.Marshal(p.GetParam())
		if err != nil {
			return nil, err
		}
	}

	b, err := c.Client.Post(url, bodyType, body, c.Credential.Key, c.Credential.Value)
	if err != nil {
		fmt.Println("create gitlab webhook err %v, %s", err, string(b))
	}

	ret := new(WebHookParam)
	if err := json.Unmarshal(b, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func update(c *RestClient, projectId string, id int, p httpuitl.Param) (*WebHookParam, error) {
	url := fmt.Sprintf(c.Url, projectId) + fmt.Sprintf("/%d", id)

	var body interface{}
	var err error

	var bodyType string
	switch bodyType = p.GetBodyType(); bodyType {
	case "application/x-www-form-urlencoded":
		body = strings.NewReader(fmt.Sprint(p.GetParam()))
	case "application/json":
		fallthrough
	default:
		body, err = json.Marshal(p.GetParam())
		if err != nil {
			return nil, err
		}
	}

	b, err := c.Client.Put(url, bodyType, body, c.Credential.Key, c.Credential.Value)
	if err != nil {
		fmt.Println("create gitlab webhook err %v, %s", err, string(b))
	}

	ret := new(WebHookParam)
	if err := json.Unmarshal(b, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func delete(c *RestClient, projectId string, id int) error {
	url := fmt.Sprintf(c.Url, projectId) + fmt.Sprintf("/%d", id)
	_, err := c.Client.Delete(url, c.Credential.Key, c.Credential.Value)
	return err
}
