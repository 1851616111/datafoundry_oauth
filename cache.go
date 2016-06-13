package main

import (
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"time"

	"github.com/asiainfoLDP/datafoundry_oauth2/gitlab"
	chanutil "github.com/asiainfoLDP/datafoundry_oauth2/util/channel"
	etcdutil "github.com/asiainfoLDP/datafoundry_oauth2/util/etcd"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/pprof"
	"github.com/asiainfoLDP/datafoundry_oauth2/util/wait"
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
	defer close(ret)
	timerFn := func() {
		if rsp, err := db.getDir("/df_service"); err != nil {
			log.Printf("controller looper %ds for gitlab orgs err %v\n", loopPeriod, err)
			return
		} else {
			if rsp != nil {
				ret <- rsp
			}

		}
	}

	go wait.Until(timerFn, loopPeriod, wait.NeverStop)

	reduceChan := make(chan string, 2000)

	go func() {
		for {
			select {
			case rsp := <-ret:
				etcdutil.RangeNodeFunc(rsp.(*etcd.Response).Node, func(n *etcd.Node) {
					reduceChan <- n.Value
				})
			}
		}
	}()

	go func () {
		for {
			select {
			case p := <-reduceChan:
			fmt.Printf("read gitlab inf %v\n", p)
				m := new(map[string]string)
				json.Unmarshal([]byte(p), m)
				host, user, privateToken := (*m)["host"], (*m)["user"], (*m)["private_token"]
				if host != "" && user != "" && privateToken != "" {
					projects, _ := glApi.Project(host, privateToken).ListProjects()
					if b, err := json.Marshal(projects); err != nil {
						fmt.Printf("controller looper %ds for gitlab projects err %v\n", err)
					} else {
						if err := Cache.HCache("gitlab_"+host+"_repo", "user_"+user+"_repos", b); err != nil {
							fmt.Printf("controller cache gitlab projects err %v\n", err)
						}
					}

					if len(projects) > 0 {
						gitlab.RangeProjectsFunc(projects, func(pid int) {
							go func() {
								branches, _ := glApi.Branch(host, privateToken).ListBranches(pid)
								if b, err := json.Marshal(branches); err != nil {
									fmt.Printf("looper %ds for gitlab project %d err %v\n", pid, err)
								} else {
									if err := Cache.HCache("gitlab_"+host+"_branch", fmt.Sprintf("project_%d", pid), b); err != nil {
										fmt.Printf("looper %ds for gitlab project %d err %v\n", pid, err)
									}
								}
							}()

						})
					}
				}
			default:
			}
		}
	}()

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
