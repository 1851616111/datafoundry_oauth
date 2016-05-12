package util

import (
	"fmt"
	"code.dataos.io/asiainfoLDP/datafoundry_oauth2/util/rand"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

const TmpPath = "./tmp"

var Cmd_Path string

func Init(cmd string) {
	//ssh-keygen -t rsa -f .ssh/id_rsa -N ""
	var err error
	Cmd_Path, err = exec.LookPath(cmd)
	if err != nil {
		if err.Error() == cmdErrNotFound(cmd) {
			log.Fatalf("cmd %s not found", cmd)
		}
	}
}

func NewKeyPool(size int) *Pool {

	if _, err := os.Stat(TmpPath); os.IsNotExist(err) {
		if err := os.Mkdir(TmpPath, os.FileMode(0700)); err != nil {
			log.Fatalf("mkdir %s err", TmpPath)
		}
	}

	return &Pool{
		file:        TmpPath,
		max:         size,
		intervalSec: 1,
		pool:        make(chan *KeyPair, size),
	}
}

func (p *Pool) Run() {
	select {
	case <-time.After(time.Duration(p.intervalSec)):
		if len(p.pool) < p.max {
			for i := 1; i <= p.max-len(p.pool); i++ {
				pair, err := generateKeyPair(p.file+"/"+rand.String(6), "rsa")
				if err != nil {
					fmt.Printf("err %#v\n", err)
					return
				}
				p.pool <- pair
			}
		}
	}
}

func (p *Pool) Pop() *KeyPair {
	return <-p.pool
}

type Pool struct {
	file        string
	max         int
	intervalSec int
	pool        chan *KeyPair
}

type KeyPair struct {
	Public  []byte
	Private []byte
}

func generateKeyPair(key string, keyType string) (*KeyPair, error) {
	defer os.Remove(key)
	defer os.Remove(key + ".pub")
	if err := createKey(key, keyType); err != nil {
		return nil, err
	}

	return getKeyPair(key)

}

func createKey(file string, keyType string) error {
	cmd := exec.Command(Cmd_Path, "-t", keyType, "-f", file, "-N", "")
	return cmd.Run()
}

func getKeyPair(Key string) (*KeyPair, error) {
	pubKey := Key + ".pub"
	pub, err := ioutil.ReadFile(pubKey)
	if err != nil {
		return nil, err
	}

	pri, err := ioutil.ReadFile(Key)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		Public:  pub,
		Private: pri,
	}, nil
}

func cmdErrNotFound(file string) string {
	errStr := &exec.Error{file, exec.ErrNotFound}
	return errStr.Error()
}
