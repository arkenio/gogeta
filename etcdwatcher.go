package main

import (
  "github.com/coreos/go-etcd/etcd"
  "log"
  "strings"
)

type watcher struct {
  client *etcd.Client
  config *Config
}

func NewEtcdWatcher(config *Config) *watcher {
  w := &watcher{}
  w.config = config
  w.client = etcd.NewClient([]string{config.etcdAdress})
  return w
}

/**
 * Init domains and environments.
 */
func (w *watcher) init() {
  w.loadAndWatch(w.config.domainPrefix, registerDomain)
  w.loadAndWatch(w.config.envPrefix, registerEnvironment)

}

/**
 * Loads and watch an etcd directory to register objects like domains, environments
 * etc... The register function is passed the etcd Node that has been loaded.
 */
func (w *watcher) loadAndWatch(etcdDir string, registerFunc func(*etcd.Node)) {
  w.loadPrefix(etcdDir, registerFunc)

  go func() {
    updateChannel := make(chan *etcd.Response, 10)
    w.watch(updateChannel, registerFunc)
    w.client.Watch(etcdDir, 0, true, updateChannel, nil)
  }()
}

func (w *watcher) loadPrefix(etcDir string, registerFunc func(*etcd.Node)) {
  response, err := w.client.Get(etcDir, true, false)

  if err == nil {
    for _, node := range response.Node.Nodes {
      registerFunc(&node)
    }
  }
}

func (w *watcher) watch(updateChannel chan *etcd.Response, registerFunc func(*etcd.Node)) {
  for {
    response := <-updateChannel
    registerFunc(response.Node)
  }
}

func registerDomain(node *etcd.Node) {
  domain := strings.Split(node.Key, "/")[2]
  log.Printf("Registering domain : %s with service %s", domain, node.Value)
  domains[domain] = node.Value
}

func registerEnvironment(node *etcd.Node) {
  environment := strings.Split(node.Key, "/")[2]
  log.Printf("Registering environement : %s", environment, node.Value)
  environments[environment] = node.Value
}
