package service

import "fmt"

const (
	MongoDB = "MongoDB"
	Kafka   = "Kafka"
	Redis   = "Redis"
)

type ServiceList []Service

type Service struct {
	Name       string     `json:"name"`
	Label      string     `json:"label"`
	Plan       string     `json:"plan"`
	Credential Credential `json:"credentials"`
}

type Credential struct {
	Host     string `json:"Host"`
	Name     string `json:"Name"`
	Password string `json:"Password"`
	Port     string `json:"Port"`
	Uri      string `json:"Uri"`
	Username string `json:"Username"`
	VHost    string `json:"Vhost"`
}

func (c Credential) String() string {

	if len(c.Username)+len(c.Name) == 0 {

		return fmt.Sprintf("%s:%s", c.Host, c.Port)

	} else if len(c.Name) != 0 && len(c.Username) == 0 {

		return fmt.Sprintf("%s:%s/%s", c.Host, c.Port, c.Name)

	} else if len(c.Username) != 0 && len(c.Name) == 0 {

		return fmt.Sprintf(`%s:%s@%s:%s`, c.Username, c.Password, c.Host, c.Port)

	} else {

		return fmt.Sprintf(`%s:%s@%s:%s/%s`, c.Username, c.Password, c.Host, c.Port, c.Name)
	}
}

type Params map[string]interface{}

func (p Params) String() string {
	if p == nil || len(p) == 0 {
		return ""
	}

	uri := "?"
	for k, v := range p {
		uri += fmt.Sprint(k, "=", v, "&")
	}
	return uri[:len(uri)-1]
}
