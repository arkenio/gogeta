package main

import (
  "log"
)

const (
  progname = "etcd-reverse-proxy"
)

func main() {
  log.Printf("%s starting",progname)
  c := parseConfig()

  w := NewEtcdWatcher(c)
  w.init()

  p := NewProxy(c, &DummyResolver{})
  p.start()

}





