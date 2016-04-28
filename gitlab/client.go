package gitlab

import "fmt"

const (
	Gitlab_Credential_Key       = "PRIVATE-TOKEN"
	GitLab_Api_Url_Path_User    = "/api/v3/user"
	GitLab_Api_Url_Path_Project = "/api/v3/projects/owned"
	GitLab_Api_Url_Path_Keys    = "/api/v3/projects/%s/keys"
)

type Client interface {
	UserInterface
	//Groups
	ProjectInterface
	//Branches
	DeployKeyInterface
}

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

type Groups interface {
	ListGroups() ([]Group, error)
}

func (r *RestClient) ListGroups() ([]Group, error) {
	groups := []Group{}
	if err := r.Client.GetJson(groups, r.Url, r.Credential.Key, r.Credential.Value); err != nil {
		return nil, err
	}

	return groups, nil
}

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
	projects := new([]Project)
	if err := r.Client.GetJson(projects, r.Url, r.Credential.Key, r.Credential.Value); err != nil {
		return nil, err
	}

	return *projects, nil
}

type Branches interface {
	ListBranches(projectId int) ([]Branch, error)
}

func (r *RestClient) ListBranches(projectId int) ([]Branch, error) {
	r.Url = fmt.Sprintf(r.Url, projectId)

	branches := []Branch{}
	if err := r.Client.GetJson(branches, r.Url, r.Credential.Key, r.Credential.Value); err != nil {
		return nil, err
	}

	return branches, nil
}

type DeployKeyInterface interface {
	DeployKey(host, privateToken string) DeployKeys
}

func (c *HttpFactory) DeployKey(host, privateToken string) DeployKeys {
	return c.newClient(host, GitLab_Api_Url_Path_Keys, privateToken)
}

type DeployKeys interface {
	ListKeys() ([]DeployKey, error)
	CreateKey(option *NewDeployKeyOption) error
	DeleteKey(id int) error
}

func (r *RestClient) ListKeys() ([]DeployKey, error) {
	keys := []DeployKey{}
	if err := r.Client.GetJson(keys, r.Url, r.Credential.Key, r.Credential.Value); err != nil {
		return nil, err
	}

	return keys, nil
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
