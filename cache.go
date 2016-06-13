package main

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"time"

	"github.com/asiainfoLDP/datafoundry_oauth2/gitlab"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/cache"
	chanutil "github.com/asiainfoLDP/datafoundry_oauth2/util/channel"
	etcdutil "github.com/asiainfoLDP/datafoundry_oauth2/util/etcd"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/pprof"
)

const (
	IdlerTimerSec     = 20
	CacheFrequencySec = 60
)

var scale uint8

//func init() {
//	go func() {
//		scale = uint8(initAvgIdle())
//		fmt.Printf("init cpu idle: %d\n", scale)
//	}()
//}

type pair struct {
	key   string
	value string
}

func runController() {
	const loopPeriod = 30 * time.Second

	ret := make(chan interface{}, 1)
	reduceChan := make(chan string, 2000)

	producer := func() {
		if rsp, err := db.getDir("/df_service"); err != nil {
			log.Printf("controller looper %ds for gitlab orgs err %v\n", loopPeriod, err)
			return
		} else {
			if rsp != nil {
				ret <- rsp
			}

		}
	}

	middler := func() {
		select {
		case rsp := <-ret:
			etcdutil.RangeNodeFunc(rsp.(*etcd.Response).Node, func(n *etcd.Node) {
				reduceChan <- n.Value
			})
		}
	}

	consumer := func() {
		select {
		case p := <-reduceChan:
			go func() {
				m := new(map[string]string)
				json.Unmarshal([]byte(p), m)
				host, user, privateToken := (*m)["host"], (*m)["user"], (*m)["private_token"]
				if host != "" && user != "" && privateToken != "" {

					projects, err := glApi.Project(host, privateToken).ListProjects()
					if err != nil {
						fmt.Printf("controller loop get project err %v\n", err)
						return
					}

					if len(projects) == 0 {
						return
					}

					if b, err := json.Marshal(projects); err != nil {
						fmt.Printf("controller looper %ds for gitlab projects err %v\n", err)
					} else {
						if err := Cache.HCache("gitlab_"+host+"_repo", "user_"+user+"_repos", b); err != nil {
							fmt.Printf("controller cache gitlab projects err %v\n", err)
						}
						b, err := Cache.HFetch("gitlab_"+host+"_repo", "user_"+user+"_repos")
						if err != nil {
							fmt.Println(err)
						}
						fmt.Println(string(b))
					}

					gitlab.RangeProjectsFunc(projects, func(pid int) {
						go func() {
							branches, err := glApi.Branch(host, privateToken).ListBranches(pid)
							if err != nil {
								fmt.Printf("controller loop get branches err %v\n", err)
								return
							}

							if len(branches) == 0 {
								return
							}

							if b, err := json.Marshal(branches); err != nil {
								fmt.Printf("looper %ds for gitlab project %d err %v\n", pid, err)
							} else {
								if err := Cache.HCache("gitlab_"+host+"_branch", fmt.Sprintf("project_%d", pid), b); err != nil {
									fmt.Printf("looper %ds for gitlab project %d err %v\n", pid, err)
								}

								b, err := Cache.HFetch("gitlab_"+host+"_branch", fmt.Sprintf("project_%d", pid))
								if err != nil {
									fmt.Println(err)
								}
								fmt.Println(string(b))
							}
						}()
					})
				}
			}()
		default:
		}
	}

	cacher := cache.NewCacher(cache.ProducerFn(producer), cache.MiddlerFn(middler), cache.ConsumerFn(consumer), loopPeriod, 0, 0)
	cacher.Run()
}

func initAvgIdle() float32 {
	done := make(chan struct{}, 1)
	defer close(done)
	newout := chanutil.TimeReader(time.Second*IdlerTimerSec, pprof.GetStat(pprof.Line_CPU, done))

	return averageIdle(newout)
}

func averageIdle(c <-chan interface{}) float32 {
	length := len(c)
	sum := float32(0.0)

	for i := range c {
		if cpu, ok := i.(*pprof.CPU); ok {
			sum += cpu.Idle
		}
	}

	return sum / float32(length)
}
