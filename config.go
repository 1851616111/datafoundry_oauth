package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	EtcdStorageEnv Env = &EnvOnce{
		envs: map[string]string{"ETCD_HTTP_ADDR": "", "ETCD_HTTP_PORT": "", "ETCD_USER": "", "ETCD_PASSWORD": ""},
	}
	GithubApplicationEnv Env = &EnvOnce{
		envs: map[string]string{"GITHUB_REDIRECT_URL": "", "GITHUB_CLIENT_ID": "", "GITHUB_CLIENT_SECRET": ""},
	}
	DatafoundryEnv Env = &EnvOnce{
		envs: map[string]string{"DATAFOUNDRY_HOST_ADDR": ""},
	}
	RedisEnv Env = &EnvOnce{
		envs: map[string]string{"Redis_BackingService_Name": ""},
	}
)

type Env interface {
	Init()
	Validate(func(k string))
	Get(name string, decoder interface{}) string
	Print()
}

type EnvOnce struct {
	envs map[string]string
	once sync.Once
}

func (e *EnvOnce) Init() {
	fn := func() {
		for k := range e.envs {
			e.envs[k] = os.Getenv(k)
		}
	}

	e.once.Do(fn)
}

func (e *EnvOnce) Validate(fn func(k string)) {
	for k, v := range e.envs {
		if strings.TrimSpace(v) == "" {
			fn(k)
		}
	}
}

func (e *EnvOnce) Get(name string, decoder interface{}) string {
	return e.envs[name]
}

func (e *EnvOnce) Print() {
	for k, v := range e.envs {
		fmt.Printf("[Env] %s=%s\n", k, v)
	}
}
