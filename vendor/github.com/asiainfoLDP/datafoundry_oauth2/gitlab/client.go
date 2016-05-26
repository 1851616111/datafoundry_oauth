package gitlab

import (
	"fmt"
	"encoding/json"
)

const (
	Gitlab_Credential_Key       = "PRIVATE-TOKEN"
	GitLab_Api_Url_Path_User    = "/api/v3/user"
	GitLab_Api_Url_Path_Project = "/api/v3/projects"
	GitLab_Api_Url_Path_Keys    = "/api/v3/projects/%d/keys"
	GitLab_Api_Url_Path_Branch  = "/api/v3/projects/%d/repository/branches"
	GitLab_Api_Url_Path_Session = "/api/v3/session"
)

type Client interface {
	UserInterface
	//Groups
	ProjectInterface
	BranchInterface
	DeployKeyInterface
	SessionInterface
}

//--------------------- User ---------------------
type UserInterface interface {
	User(host, privateToken string) Users
}

func (c *HttpFactory) User(host, privateToken string) Users {
	return c.newClient(host, GitLab_Api_Url_Path_User, privateToken)
}

type Users interface {
	GetUser() (*User, error)
}

func (r *RestClient) GetUser() (*User, error) {
	user := new(User)
	if err := r.Client.GetJson(user, r.Url, r.Credential.Key, r.Credential.Value); err != nil {
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

func (c *HttpFactory) Project(host, privateToken string) Projects {
	return c.newClient(host, GitLab_Api_Url_Path_Project, privateToken)
}

type Projects interface {
	ListProjects() ([]Project, error)
}

func (r *RestClient) ListProjects() ([]Project, error) {
	projects := []Project{}

	page := 1
	for {
		ps := new([]Project)
		url := fmt.Sprintf("%s?page=%d", r.Url, page)
		if err := r.Client.GetJson(ps, url, r.Credential.Key, r.Credential.Value); err != nil {
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

func (c *HttpFactory) Branch(host, privateToken string) Branches {
	return c.newClient(host, GitLab_Api_Url_Path_Branch, privateToken)
}

type Branches interface {
	ListBranches(projectId int) ([]Branch, error)
}

func (r *RestClient) ListBranches(projectId int) ([]Branch, error) {
	r.Url = fmt.Sprintf(r.Url, projectId)

	branches := new([]Branch)
	if err := r.Client.GetJson(branches, r.Url, r.Credential.Key, r.Credential.Value); err != nil {
		return nil, err
	}

	return *branches, nil
}

//--------------------- DeployKeys ---------------------
type DeployKeyInterface interface {
	DeployKey(host, privateToken string) DeployKeys
}

func (c *HttpFactory) DeployKey(host, privateToken string) DeployKeys {
	return c.newClient(host, GitLab_Api_Url_Path_Keys, privateToken)
}

type DeployKeys interface {
	ListKeys(projectId int) ([]DeployKey, error)
	CreateKey(option *NewDeployKeyOption) error
	DeleteKey(id int) error
}

func (r *RestClient) ListKeys(projectId int) ([]DeployKey, error) {
	keys := new([]DeployKey)
	url := fmt.Sprintf(r.Url, projectId)
	if err := r.Client.GetJson(keys, url, r.Credential.Key, r.Credential.Value); err != nil {
		return nil, err
	}

	return *keys, nil
}

func (r *RestClient) CreateKey(option *NewDeployKeyOption) error {
	url := fmt.Sprintf(r.Url, option.ProjectId)
	b, err := r.Client.Encode(option.Param)
	if err != nil {
		return err
	}

	_, err = r.Client.Post(url, b, r.Credential.Key, r.Credential.Value)
	return err
}

func (r *RestClient) DeleteKey(id int) error {
	url := fmt.Sprintf(r.Url, id)
	_, err := r.Client.Delete(url, r.Credential.Key, r.Credential.Value)

	return err
}

func (h *HttpFactory) newClient(host, api, privateToken string) *RestClient {
	return &RestClient{
		Url: UrlMaker(host, api),
		Credential: Credential{
			Key:   Gitlab_Credential_Key,
			Value: privateToken,
		},
		Client: h,
	}
}

func (h *HttpFactory) newSessionClient(host, api, login, password string) *RestClient {
	return &RestClient{
		Url:    fmt.Sprintf("%s?login=%s&password=%s", UrlMaker(host, api), login, password),
		Client: h,
	}
}

//--------------------- Session ---------------------
type SessionInterface interface {
	Session(host, login, password string) Sessions
}

type Sessions interface {
	PostSession() (*Session, error)
}

func (c *HttpFactory) Session(host, login, password string) Sessions {
	return c.newSessionClient(host, GitLab_Api_Url_Path_Session, login, password)
}

func (r *RestClient) PostSession() (*Session, error) {
	b, err := r.Client.Post(r.Url, []byte{})
	if err != nil {
		return nil, err
	}

	s := new(Session)
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}

	return s, nil
}
