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
  w.client = etcd.NewClient([]string{})
  return w
}

func (w *watcher) init() {
  w.loadApplications()

  go func() {
    appsChannel := make(chan *etcd.Response, 10)
    w.watchApplication(appsChannel)
    w.client.Watch(w.config.etcdPrefix, 0, true, appsChannel, nil)
  }()
}

func (w *watcher) loadApplications() {
  response, err := w.client.Get(w.config.etcdPrefix, true, false)

  if err == nil {
    for _, node := range response.Node.Nodes {
      w.registerServer(&node)
    }
  }
}

func (w *watcher) watchApplication(appsChannel chan *etcd.Response) {
  for {
    response := <- appsChannel
    w.registerServer(response.Node)
  }
}

func (w *watcher) registerServer(node *etcd.Node) {
      domain := strings.Split(node.Key, "/")[2]
      log.Printf("Detected domain : %s", domain);
      //TODO add domain to global domain config
}
