package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

func NewBackingService(name string, validate ValidateService, check CheckService, err func(err error)) BackingService {
	return &service{
		name:     name,
		validate: validate,
		check:    check,
		err:      err,
	}
}

type BackingService interface {
	GetBackingServices(name string) <-chan Service
}

type service struct {
	name     string
	validate ValidateService
	check    CheckService
	err      func(err error)
}

func (s *service) GetBackingServices(name string) <-chan Service {
	return initBackingServicesFunc(name, s.validate, s.check, s.err)
}

func initBackingServicesFunc(name string, validate ValidateService, check CheckService, errFunc func(err error)) <-chan Service {

	svcs := getCredentials(name)

	if len(svcs) == 0 {
		if errFunc != nil {
			errFunc(errors.New(fmt.Sprintf("backingservice %s config nil.", name)))
		}
		return nil
	}

	c := make(chan Service, len(svcs))

	go func() {
		for _, svc := range svcs {
			c <- svc
		}
		close(c)
	}()

	return checkBackingServices(validateBackingServices(c, validate), check)
}

func validateBackingServices(sc <-chan Service, validate ValidateService) <-chan Service {
	c := make(chan Service)

	go func() {
		for svc := range sc {
			if validate(svc) {
				c <- svc
			}
		}
		close(c)
	}()

	return c
}

func checkBackingServices(sc <-chan Service, checkFunc CheckService) <-chan Service {
	c := make(chan Service, len(sc))

	var wg sync.WaitGroup

	for svc := range sc {
		wg.Add(1)
		go func() {
			if checkFunc(svc) {
				c <- svc
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	return c
}

type CheckService func(svc Service) bool

type ValidateService func(svc Service) bool

func ValidateHPN(svc Service) bool {
	if len(svc.Credential.Host) == 0 || len(svc.Credential.Port) == 0 || len(svc.Credential.Name) == 0 {
		return false
	}
	return true
}

func ValidateHP(svc Service) bool {
	if len(svc.Credential.Host) == 0 || len(svc.Credential.Port) == 0 {
		return false
	}
	return true
}

func GenerateBackingServiceUrl(svc Service, param Params) string {
	return fmt.Sprint(svc.Credential) + fmt.Sprint(param)
}

const EnvKey = "VCAP_SERVICES"

func getCredentials(name string) ServiceList {
	s := os.Getenv(EnvKey)
	if len(s) == 0 {
		return nil
	}

	m := new(map[string]ServiceList)
	if err := json.Unmarshal([]byte(s), m); err != nil {
		return nil
	}

	return (*m)[name]
}
