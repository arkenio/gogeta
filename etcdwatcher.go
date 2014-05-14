package main

import (
	"encoding/json"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"regexp"
	"strings"
)

// A watcher loads and watch the etcd hierarchy for domains and services.
type watcher struct {
	client       *etcd.Client
	config       *Config
	domains      map[string]*Domain
	environments map[string]*EnvironmentCluster
}


// Constructor for a new watcher
func NewEtcdWatcher(config *Config, domains map[string]*Domain, envs map[string]*EnvironmentCluster) *watcher {
	client := etcd.NewClient([]string{config.etcdAddress})
	return &watcher{client, config, domains, envs}
}


//Init domains and environments.
func (w *watcher) init() {
	go w.loadAndWatch(w.config.domainPrefix, w.registerDomain)
	go w.loadAndWatch(w.config.envPrefix, w.registerEnvironment)

}


// Loads and watch an etcd directory to register objects like domains, environments
// etc... The register function is passed the etcd Node that has been loaded.
func (w *watcher) loadAndWatch(etcdDir string, registerFunc func(*etcd.Node, string)) {
	w.loadPrefix(etcdDir, registerFunc)

	updateChannel := make(chan *etcd.Response, 10)
	go w.watch(updateChannel, registerFunc)
	w.client.Watch(etcdDir, (uint64)(0), true, updateChannel, nil)

}

func (w *watcher) loadPrefix(etcDir string, registerFunc func(*etcd.Node, string)) {
	response, err := w.client.Get(etcDir, true, true)
	if err == nil {
		for _, serviceNode := range response.Node.Nodes {
			registerFunc(serviceNode, response.Action)


		}
	}


}

func (w *watcher) watch(updateChannel chan *etcd.Response, registerFunc func(*etcd.Node, string)) {
	for {
		response := <-updateChannel
		if(response != nil) {
			registerFunc(response.Node, response.Action)
		}

	}
}

func (w *watcher) registerDomain(node *etcd.Node, action string) {

	domainName := w.getDomainForNode(node)
	domainKey := w.config.domainPrefix + "/" + domainName
	response, err := w.client.Get(domainKey, true, false)

	if err == nil {
		domain := &Domain{}
		for _, node := range response.Node.Nodes {
			switch node.Key {
			case domainKey + "/type":
				domain.typ = node.Value
			case domainKey + "/value":
				domain.value = node.Value
			}
		}
		if domain.typ != "" && domain.value != "" {
			w.domains[domainName] = domain
			log.Printf("Registering domain %s with service (%s):%s", domainName, domain.typ, domain.value)
		}
	}

}

func (w *watcher) getDomainForNode(node *etcd.Node) string {
	r := regexp.MustCompile(w.config.domainPrefix + "/(.*)")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[0]
}

func (w *watcher) getEnvForNode(node *etcd.Node) string {
	r := regexp.MustCompile(w.config.envPrefix + "/(.*)(/.*)*")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[0]
}

func (w *watcher) getEnvIndexForNode(node *etcd.Node) string {
	r := regexp.MustCompile(w.config.envPrefix + "/(.*)(/.*)*")
	return strings.Split(r.FindStringSubmatch(node.Key)[1], "/")[1]
}

func (w *watcher) Remove(key string) {

}

func (w *watcher) registerEnvironment(node *etcd.Node, action string) {
	envName := w.getEnvForNode(node)
	// Get service's root node instead of changed node.
	envNode, _ := w.client.Get(w.config.envPrefix + "/" + envName, true, true)

	for _, indexNode := range envNode.Node.Nodes {

		envIndex := w.getEnvIndexForNode(indexNode)
		envKey := w.config.envPrefix + "/" + envName + "/" + envIndex
		statusKey := envKey + "/status"

		response, err := w.client.Get(envKey, true, true)

		if err == nil {


			if w.environments[envName] == nil {
				w.environments[envName] = &EnvironmentCluster{}
			}

			env := &Environment{}
			env.key = envIndex


			if action == "delete" || action == "expire" {
				w.Remove(env.key)
				return
			}


			for _, node := range response.Node.Nodes {
				switch node.Key {
				case envKey + "/location":
					location := &service{}
					err := json.Unmarshal([]byte(node.Value), location)
					if err != nil {
						panic(err)
					}

					env.location.Host = location.Host
					env.location.Port = location.Port
				case envKey + "/domain":
					env.domain = node.Value

				case statusKey:
					env.status = &Status{}

					for _, subNode := range node.Nodes {
						switch subNode.Key {
						case statusKey + "/alive":
							env.status.alive = subNode.Value
						case statusKey + "/current":
							env.status.current = subNode.Value
						case statusKey + "/expected":
							env.status.expected = subNode.Value
						}
					}
				}
			}

			if env.location.Host != "" && env.location.Port != 0 {
				w.environments[envName].Add(env)
				log.Printf("Registering environment %s with address : http://%s:%d/", envName, env.location.Host, env.location.Port)
				if env.domain != "" && w.domains[env.domain] != nil {
					w.domains[env.domain].server = nil
					log.Printf("Reset domain %s", env.domain)
				}
			}
			if env.status != nil && env.status.current != "" {
				w.environments[envName].Add(env)
				log.Printf("Watching environment %s status : Alive: %s - Current: %s - Expected: %s", envName, env.status.alive, env.status.current, env.status.expected)
			}
			w.environments[envName].Dump("onRegister")
		}
	}
}
