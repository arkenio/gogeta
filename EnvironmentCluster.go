package main

import (
	"errors"
	"sync"
	"github.com/golang/glog"
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
		glog.V(5).Infof("Checking instance %d status : %s", index, instance.status.compute())
		if ( instance.status.compute() == STARTED_STATUS) {
			return instance, nil
		}
	}
	glog.V(5).Infof("No instance started for %s", instance.name)

	lastStatus := instance.status
	glog.V(5).Infof("Last status :")
	glog.V(5).Infof("   current  : %s", lastStatus.current)
	glog.V(5).Infof("   expected : %s", lastStatus.expected)
	glog.V(5).Infof("   alive : %s", lastStatus.alive)
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

// Get an environment by its key (index). Returns nil if not found.
func (cl *EnvironmentCluster) Get(key string) *Environment {
	for i, v := range cl.instances {
		if v.key == key {
			return cl.instances[i]
		}
	}
	return nil
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
		glog.Infof("Dump after %s %s -> %s:%d", action, v.key, v.location.Host, v.location.Port)
	}
}
