package main

import (
	"errors"
	"log"
	"sync"
)

type EnvironmentCluster struct {
	instances []*Environment
	lastIndex int
	lock      sync.RWMutex
}

func (cl *EnvironmentCluster) Next() (*Environment, error) {
	if cl == nil {
		return nil, StatusError{}
	}
	cl.lock.RLock()
	defer cl.lock.RUnlock()
	if len(cl.instances) == 0 {
		return nil, errors.New("no alive instance found")
	}
	var instance *Environment
	for tries := 0; tries < len(cl.instances); tries++ {
		index := (cl.lastIndex + 1) % len(cl.instances)
		cl.lastIndex = index

		instance = cl.instances[index]
		if (instance.status == nil || instance.status.compute() == STARTED_STATUS) {
			return instance, nil
		}
	}
	log.Printf("No instance started for %s", instance.domain)
	lastStatus := instance.status
	return nil, StatusError{instance.status.compute(), lastStatus }
}

func (cl *EnvironmentCluster) Remove(key string) {

	match := -1
	for k, v := range cl.instances {
		if v.key == key {
			match = k
		}
	}

	cl.instances = append(cl.instances[:match], cl.instances[match+1:]...)
	cl.Dump("remove")
}



func (cl *EnvironmentCluster) Add(env *Environment) {
	for i, v := range cl.instances {
		if v.key == env.key {
			cl.instances[i] = env
			return
		}
	}

	cl.instances = append(cl.instances, env)
}

func (cl *EnvironmentCluster) Dump(action string) {
	for _, v := range cl.instances {
		log.Printf("Dump after %s %s -> %s:%d", action, v.key, v.location.Host, v.location.Port)
	}
}
